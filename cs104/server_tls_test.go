package cs104

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/marrasen/go-iecp5/asdu"
)

// testPKI holds certificates generated for a test run.
type testPKI struct {
	serverTLSConfig *tls.Config
	clientTLSConfig *tls.Config
}

// issueCert generates a key and certificate from tmpl. When parent is nil the
// certificate is self-signed; otherwise it is signed by parent/parentKey.
func issueCert(t *testing.T, tmpl, parent *x509.Certificate, parentKey crypto.Signer) (tls.Certificate, *x509.Certificate) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signKey := crypto.Signer(key)
	if parent == nil {
		parent = tmpl
	} else {
		signKey = parentKey
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, parent, &key.PublicKey, signKey)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, cert
}

func generateTestPKI(t *testing.T) testPKI {
	t.Helper()

	caTLSCert, caCert := issueCert(t, &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}, nil, nil)
	caKey := caTLSCert.PrivateKey.(crypto.Signer)

	serverTLSCert, _ := issueCert(t, &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Test Server"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}, caCert, caKey)

	clientTLSCert, _ := issueCert(t, &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "Test Client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}, caCert, caKey)

	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	return testPKI{
		serverTLSConfig: &tls.Config{
			Certificates: []tls.Certificate{serverTLSCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    caPool,
		},
		clientTLSConfig: &tls.Config{
			ServerName:   "127.0.0.1",
			RootCAs:      caPool,
			Certificates: []tls.Certificate{clientTLSCert},
		},
	}
}

// startServer serves srv on an ephemeral port and returns the bound address.
func startServer(t *testing.T, srv *Server) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	return ln.Addr().String()
}

func TestSessionActivation(t *testing.T) {
	pki := generateTestPKI(t)

	cases := []struct {
		name      string
		serverTLS *tls.Config
		scheme    string
		clientTLS *tls.Config
		wantCN    string // "" means expect no peer certificates
	}{
		{"TLS", pki.serverTLSConfig, "tls://", pki.clientTLSConfig, "Test Client"},
		{"PlainTCP", nil, "tcp://", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := &captureHandler{}
			srv := NewServer(handler).SetTLSConfig(tc.serverTLS)
			activeCh := make(chan *SrvSession, 1)
			srv.ConnState = func(c asdu.Connect, state ConnState) {
				if state == ConnStateActive {
					if sess, ok := c.(*SrvSession); ok {
						select {
						case activeCh <- sess:
						default:
						}
					}
				}
			}
			addr := startServer(t, srv)

			opt := NewOption()
			if err := opt.SetRemoteServer(tc.scheme + addr); err != nil {
				t.Fatal(err)
			}
			if tc.clientTLS != nil {
				opt.SetTLSConfig(tc.clientTLS)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			clientNewCh := make(chan struct{}, 1)
			client := NewClient(handler, opt)
			client.SetConnStateHandler(func(_ asdu.Connect, state ConnState) {
				if state == ConnStateNew {
					select {
					case clientNewCh <- struct{}{}:
					default:
					}
				}
			})
			go client.Start(ctx)
			t.Cleanup(func() { _ = client.Close() })

			select {
			case <-clientNewCh:
				client.SendStartDt()
			case <-ctx.Done():
				t.Fatal("timed out waiting for client to connect")
			}

			select {
			case sess := <-activeCh:
				certs := sess.PeerCertificates()
				if tc.wantCN == "" {
					if certs != nil {
						t.Fatalf("expected nil peer certificates on plain TCP, got %d", len(certs))
					}
				} else {
					if len(certs) == 0 {
						t.Fatal("expected peer certificates, got none")
					}
					if certs[0].Subject.CommonName != tc.wantCN {
						t.Fatalf("unexpected CN: %s", certs[0].Subject.CommonName)
					}
				}
			case <-ctx.Done():
				t.Fatal("timed out waiting for connection to become active")
			}
		})
	}
}

// startRejectingServer starts a TLS server and returns its address plus a
// channel that reports any session creation, which these tests treat as a
// failure: a connection that fails the handshake must never become a session.
func startRejectingServer(t *testing.T, pki testPKI) (string, <-chan struct{}) {
	t.Helper()
	newCh := make(chan struct{}, 1)
	srv := NewServer(&captureHandler{}).SetTLSConfig(pki.serverTLSConfig)
	srv.ConnState = func(_ asdu.Connect, state ConnState) {
		if state == ConnStateNew {
			select {
			case newCh <- struct{}{}:
			default:
			}
		}
	}
	return startServer(t, srv), newCh
}

func TestPlainClientToTLSServerFails(t *testing.T) {
	pki := generateTestPKI(t)
	addr, newCh := startRejectingServer(t, pki)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// An IEC 104 start frame is not a TLS ClientHello; the handshake must fail.
	if _, err := conn.Write([]byte{0x68, 0x04, 0x07, 0x00, 0x00, 0x00}); err != nil {
		t.Fatal(err)
	}

	// The server must close the connection, possibly after sending a TLS alert.
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 256)
	var readErr error
	for readErr == nil {
		_, readErr = conn.Read(buf)
	}
	if ne, ok := readErr.(net.Error); ok && ne.Timeout() {
		t.Fatal("expected server to close the connection after a failed handshake")
	}

	select {
	case <-newCh:
		t.Fatal("server created a session for a connection that failed the handshake")
	default:
	}
}

func TestUntrustedCertRejected(t *testing.T) {
	pki := generateTestPKI(t)

	// A self-signed client cert, not issued by the CA the server trusts.
	rogueTLSCert, _ := issueCert(t, &x509.Certificate{
		SerialNumber: big.NewInt(99),
		Subject:      pkix.Name{CommonName: "Rogue Client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}, nil, nil)

	addr, newCh := startRejectingServer(t, pki)

	cfg := pki.clientTLSConfig.Clone()
	cfg.Certificates = []tls.Certificate{rogueTLSCert}

	raw, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	_ = raw.SetDeadline(time.Now().Add(5 * time.Second))

	conn := tls.Client(raw, cfg)
	err = conn.Handshake()
	if err == nil {
		// Under TLS 1.3 the server's bad-certificate alert arrives after the
		// client considers the handshake done; it surfaces on the first read.
		_, err = conn.Read(make([]byte, 1))
	}
	if err == nil {
		t.Fatal("expected handshake with untrusted client cert to be rejected")
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		t.Fatal("expected certificate rejection, got timeout")
	}

	select {
	case <-newCh:
		t.Fatal("server created a session for an untrusted client cert")
	default:
	}
}
