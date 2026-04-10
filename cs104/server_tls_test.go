package cs104

import (
	"context"
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
	caPool          *x509.CertPool
	serverTLSConfig *tls.Config
	clientTLSConfig *tls.Config
}

func generateTestPKI(t *testing.T) testPKI {
	t.Helper()

	// CA key and self-signed cert
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	caPool := x509.NewCertPool()
	caPool.AddCert(caCert)

	// Server cert signed by CA
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Test Server"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	serverTLSCert := tls.Certificate{
		Certificate: [][]byte{serverDER},
		PrivateKey:  serverKey,
	}

	// Client cert signed by CA
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "Test Client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	clientTLSCert := tls.Certificate{
		Certificate: [][]byte{clientDER},
		PrivateKey:  clientKey,
	}

	return testPKI{
		caPool: caPool,
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

// waitForListener polls the server until its listener is ready and returns the address.
func waitForListener(t *testing.T, srv *Server) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		srv.mux.Lock()
		l := srv.listen
		srv.mux.Unlock()
		if l != nil {
			return l.Addr().String()
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not start listening in time")
	return ""
}

func TestTLSHappyPath(t *testing.T) {
	pki := generateTestPKI(t)

	activeCh := make(chan *SrvSession, 1)
	handler := &captureHandler{}
	srv := NewServer(handler)
	srv.TLSConfig = pki.serverTLSConfig
	srv.ConnState = func(c asdu.Connect, state ConnState) {
		if state == ConnStateActive {
			if sess, ok := c.(*SrvSession); ok {
				activeCh <- sess
			}
		}
	}

	go srv.ListenAndServe("127.0.0.1:0")
	defer srv.Close()

	addr := waitForListener(t, srv)

	opt := NewOption()
	if err := opt.SetRemoteServer("tls://" + addr); err != nil {
		t.Fatal(err)
	}
	opt.SetTLSConfig(pki.clientTLSConfig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for client ConnStateNew before sending STARTDT
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
	defer client.Close()

	// Wait for client to connect, then send STARTDT
	select {
	case <-clientNewCh:
		client.SendStartDt()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for client to connect")
	}

	select {
	case sess := <-activeCh:
		certs := sess.PeerCertificates()
		if len(certs) == 0 {
			t.Fatal("expected peer certificates, got none")
		}
		if certs[0].Subject.CommonName != "Test Client" {
			t.Fatalf("unexpected CN: %s", certs[0].Subject.CommonName)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for TLS connection to become active")
	}
}

func TestPlainTCPRegression(t *testing.T) {
	activeCh := make(chan *SrvSession, 1)
	handler := &captureHandler{}
	srv := NewServer(handler)
	// TLSConfig is nil — plain TCP
	srv.ConnState = func(c asdu.Connect, state ConnState) {
		if state == ConnStateActive {
			if sess, ok := c.(*SrvSession); ok {
				activeCh <- sess
			}
		}
	}

	go srv.ListenAndServe("127.0.0.1:0")
	defer srv.Close()

	addr := waitForListener(t, srv)

	opt := NewOption()
	if err := opt.SetRemoteServer("tcp://" + addr); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	defer client.Close()

	select {
	case <-clientNewCh:
		client.SendStartDt()
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for client to connect")
	}

	select {
	case sess := <-activeCh:
		certs := sess.PeerCertificates()
		if certs != nil {
			t.Fatalf("expected nil peer certificates on plain TCP, got %d", len(certs))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for plain TCP connection to become active")
	}
}

func TestPlainClientToTLSServerFails(t *testing.T) {
	pki := generateTestPKI(t)

	closedCh := make(chan struct{}, 1)
	handler := &captureHandler{}
	srv := NewServer(handler)
	srv.TLSConfig = pki.serverTLSConfig
	srv.ConnState = func(_ asdu.Connect, state ConnState) {
		if state == ConnStateClosed {
			select {
			case closedCh <- struct{}{}:
			default:
			}
		}
	}

	go srv.ListenAndServe("127.0.0.1:0")
	defer srv.Close()

	addr := waitForListener(t, srv)

	// Plain TCP client — no TLS. Use raw TCP dial to avoid the client's scheme-based TLS logic.
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	// Send garbage (IEC 104 start frame) which is not a TLS ClientHello
	conn.Write([]byte{0x68, 0x04, 0x07, 0x00, 0x00, 0x00}) // STARTDT-Active U-frame
	defer conn.Close()

	// The server session should fail the TLS handshake and close
	select {
	case <-closedCh:
		// Expected — the connection failed
	case <-time.After(5 * time.Second):
		t.Fatal("timed out — expected plain client connection to TLS server to fail")
	}
}

func TestUntrustedCertRejected(t *testing.T) {
	pki := generateTestPKI(t)

	// Generate a separate, untrusted client cert (self-signed, not by the server's CA)
	rogueKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rogueTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(99),
		Subject:      pkix.Name{CommonName: "Rogue Client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	rogueDER, err := x509.CreateCertificate(rand.Reader, rogueTemplate, rogueTemplate, &rogueKey.PublicKey, rogueKey)
	if err != nil {
		t.Fatal(err)
	}
	rogueTLSCert := tls.Certificate{
		Certificate: [][]byte{rogueDER},
		PrivateKey:  rogueKey,
	}

	closedCh := make(chan struct{}, 1)
	handler := &captureHandler{}
	srv := NewServer(handler)
	srv.TLSConfig = pki.serverTLSConfig
	srv.ConnState = func(_ asdu.Connect, state ConnState) {
		if state == ConnStateClosed {
			select {
			case closedCh <- struct{}{}:
			default:
			}
		}
	}

	go srv.ListenAndServe("127.0.0.1:0")
	defer srv.Close()

	addr := waitForListener(t, srv)

	opt := NewOption()
	if err := opt.SetRemoteServer("tls://" + addr); err != nil {
		t.Fatal(err)
	}
	opt.SetTLSConfig(&tls.Config{
		RootCAs:      pki.caPool,
		Certificates: []tls.Certificate{rogueTLSCert},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := NewClient(handler, opt)
	go client.Start(ctx)
	defer client.Close()

	select {
	case <-closedCh:
		// Expected — untrusted cert rejected
	case <-time.After(5 * time.Second):
		t.Fatal("timed out — expected untrusted cert to be rejected")
	}
}
