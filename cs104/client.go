// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package cs104

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/clog"
)

const (
	inactive = iota
	active
)

// Client is an IEC104 master
type Client struct {
	option  ClientOption
	conn    net.Conn
	handler Handler

	// channel
	rcvASDU  chan []byte // for received asdu
	sendASDU chan []byte // for send asdu
	rcvRaw   chan []byte // for recvLoop raw cs104 frame
	sendRaw  chan []byte // for sendLoop raw cs104 frame

	// Send and receive sequence numbers for I-frames
	seqNoSend uint16 // sequence number of next outbound I-frame
	ackNoSend uint16 // outbound sequence number yet to be confirmed
	seqNoRcv  uint16 // sequence number of next inbound I-frame
	ackNoRcv  uint16 // inbound sequence number yet to be confirmed

	// maps sendTime I-frames to their respective sequence number
	pending []seqPending

	startDtActiveSendSince atomic.Value // Timeout interval while waiting for confirmation after sending StartDT-Active
	stopDtActiveSendSince  atomic.Value // Timeout while waiting for confirmation after initiating StopDT-Active

	// Connection status
	status   uint32
	rwMux    sync.RWMutex
	isActive uint32

	// Miscellaneous
	clog.Clog

	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	closeCancel context.CancelFunc

	onConnect        func(c *Client)
	onConnectionLost func(c *Client)
	onActivated      func(c *Client)
	onDeactivated    func(c *Client)
}

// NewClient returns an IEC104 master,default config and default asdu.ParamsWide params
func NewClient(handler Handler, o *ClientOption) *Client {
	return &Client{
		option:           *o,
		handler:          handler,
		rcvASDU:          make(chan []byte, o.config.RecvUnAckLimitW<<4),
		sendASDU:         make(chan []byte, o.config.SendUnAckLimitK<<4),
		rcvRaw:           make(chan []byte, o.config.RecvUnAckLimitW<<5),
		sendRaw:          make(chan []byte, o.config.SendUnAckLimitK<<5), // may not block!
		Clog:             clog.NewLogger("cs104 client => "),
		onConnect:        func(*Client) {},
		onConnectionLost: func(*Client) {},
		onActivated:      func(*Client) {},
		onDeactivated:    func(*Client) {},
	}
}

// SetOnConnectHandler set on connect handler
func (sf *Client) SetOnConnectHandler(f func(c *Client)) *Client {
	if f != nil {
		sf.onConnect = f
	}
	return sf
}

// SetConnectionLostHandler set connection lost handler
func (sf *Client) SetConnectionLostHandler(f func(c *Client)) *Client {
	if f != nil {
		sf.onConnectionLost = f
	}
	return sf
}

// SetOnActivatedHandler sets a handler invoked when StartDT is confirmed by the server
func (sf *Client) SetOnActivatedHandler(f func(c *Client)) *Client {
	if f != nil {
		sf.onActivated = f
	}
	return sf
}

// SetOnDeactivatedHandler sets a handler invoked when StopDT is confirmed by the server
func (sf *Client) SetOnDeactivatedHandler(f func(c *Client)) *Client {
	if f != nil {
		sf.onDeactivated = f
	}
	return sf
}

// Start manages the connection lifecycle to the server, handling connection attempts, failures, and disconnections.
func (sf *Client) Start(ctx context.Context) error {
	sf.rwMux.Lock()
	if !atomic.CompareAndSwapUint32(&sf.status, initial, disconnected) {
		sf.rwMux.Unlock()
		return errors.New("client already started")
	}
	ctx, sf.closeCancel = context.WithCancel(ctx)
	sf.rwMux.Unlock()
	defer sf.setConnectStatus(initial)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	sf.Debug("connecting server %+v", sf.option.server)
	conn, err := openConnection(ctx, sf.option.server, sf.option.TLSConfig, sf.option.config.ConnectTimeout0, sf.option.DialContext)
	if err != nil {
		sf.Error("connect failed, %v", err)
		return err
	}
	sf.Debug("connect success")
	sf.conn = conn
	err = sf.run(ctx)

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		sf.Debug("disconnected, %v", err)
	} else {
		sf.Error("run failed, %v", err)
	}
	return err
}

func (sf *Client) recvLoop() {
	sf.Debug("recvLoop started")
	defer func() {
		sf.cancel()
		sf.wg.Done()
		sf.Debug("recvLoop stopped")
	}()

	for {
		rawData := make([]byte, APDUSizeMax)
		for rdCnt, length := 0, 2; rdCnt < length; {
			byteCount, err := io.ReadFull(sf.conn, rawData[rdCnt:length])
			if err != nil {
				// See: https://github.com/golang/go/issues/4373
				if err != io.EOF && err != io.ErrClosedPipe ||
					strings.Contains(err.Error(), "use of closed network connection") {
					sf.Error("receive failed, %v", err)
					return
				}
				if e, ok := err.(net.Error); ok && !e.Temporary() {
					sf.Error("receive failed, %v", err)
					return
				}
				if rdCnt == 0 && err == io.EOF {
					sf.Error("remote connect closed, %v", err)
					return
				}
			}

			rdCnt += byteCount
			if rdCnt == 0 {
				continue
			} else if rdCnt == 1 {
				if rawData[0] != startFrame {
					rdCnt = 0
					continue
				}
			} else {
				if rawData[0] != startFrame {
					rdCnt, length = 0, 2
					continue
				}
				length = int(rawData[1]) + 2
				if length < APCICtlFiledSize+2 || length > APDUSizeMax {
					rdCnt, length = 0, 2
					continue
				}
				if rdCnt == length {
					apdu := rawData[:length]
					sf.Debug("RX Raw[% x]", apdu)
					sf.rcvRaw <- apdu
				}
			}
		}
	}
}

func (sf *Client) sendLoop() {
	sf.Debug("sendLoop started")
	defer func() {
		sf.cancel()
		sf.wg.Done()
		sf.Debug("sendLoop stopped")
	}()
	for {
		select {
		case <-sf.ctx.Done():
			return
		case apdu := <-sf.sendRaw:
			sf.Debug("TX Raw[% x]", apdu)
			for wrCnt := 0; len(apdu) > wrCnt; {
				byteCount, err := sf.conn.Write(apdu[wrCnt:])
				if err != nil {
					// See: https://github.com/golang/go/issues/4373
					if err != io.EOF && err != io.ErrClosedPipe ||
						strings.Contains(err.Error(), "use of closed network connection") {
						sf.Error("sendRaw failed, %v", err)
						return
					}
					if e, ok := err.(net.Error); !ok || !e.Temporary() {
						sf.Error("sendRaw failed, %v", err)
						return
					}
					// temporary error may be recoverable
				}
				wrCnt += byteCount
			}
		}
	}
}

// run is the big fat state machine.
func (sf *Client) run(ctx context.Context) error {
	sf.Debug("run started!")
	// before anything make sure init
	sf.cleanUp()

	sf.ctx, sf.cancel = context.WithCancel(ctx)
	sf.setConnectStatus(connected)
	sf.wg.Add(3)
	go sf.recvLoop()
	go sf.sendLoop()
	go sf.handlerLoop()

	var checkTicker = time.NewTicker(timeoutResolution)

	// transmission timestamps for timeout calculation
	var willNotTimeout = time.Now().Add(time.Hour * 24 * 365 * 100)

	var unAckRcvSince = willNotTimeout
	var idleTimeout3Sine = time.Now()         // Idle interval checkpoint for initiating TestFrAct
	var testFrAliveSendSince = willNotTimeout // Timeout interval while waiting for confirmation after initiating TestFrAct

	sf.startDtActiveSendSince.Store(willNotTimeout)
	sf.stopDtActiveSendSince.Store(willNotTimeout)

	sendSFrame := func(rcvSN uint16) {
		sf.Debug("TX sFrame %v", sAPCI{rcvSN})
		sf.sendRaw <- newSFrame(rcvSN)
	}

	sendIFrame := func(asdu1 []byte) {
		seqNo := sf.seqNoSend

		iframe, err := newIFrame(seqNo, sf.seqNoRcv, asdu1)
		if err != nil {
			return
		}
		sf.ackNoRcv = sf.seqNoRcv
		sf.seqNoSend = (seqNo + 1) & 32767
		sf.pending = append(sf.pending, seqPending{seqNo & 32767, time.Now()})

		sf.Debug("TX iFrame %v", iAPCI{seqNo, sf.seqNoRcv})
		sf.sendRaw <- iframe
	}

	defer func() {
		// default: STOPDT, when connection established and not enabled "data transfer" yet
		atomic.StoreUint32(&sf.isActive, inactive)
		sf.setConnectStatus(disconnected)
		checkTicker.Stop()
		_ = sf.conn.Close() // Trigger cancel indirectly; closing the connection causes loops to abort
		sf.wg.Wait()
		sf.onConnectionLost(sf)
		sf.Debug("run stopped!")
	}()

	sf.onConnect(sf)
	for {
		if atomic.LoadUint32(&sf.isActive) == active && seqNoCount(sf.ackNoSend, sf.seqNoSend) <= sf.option.config.SendUnAckLimitK {
			select {
			case o := <-sf.sendASDU:
				sendIFrame(o)
				idleTimeout3Sine = time.Now()
				continue
			case <-sf.ctx.Done():
				return sf.ctx.Err()
			default: // make no block
			}
		}
		select {
		case <-sf.ctx.Done():
			return sf.ctx.Err()
		case now := <-checkTicker.C:
			// check all timeouts
			if now.Sub(testFrAliveSendSince) >= sf.option.config.SendUnAckTimeout1 ||
				now.Sub(sf.startDtActiveSendSince.Load().(time.Time)) >= sf.option.config.SendUnAckTimeout1 ||
				now.Sub(sf.stopDtActiveSendSince.Load().(time.Time)) >= sf.option.config.SendUnAckTimeout1 {
				sf.Error("test frame alive confirm timeout t₁")
				return errors.New("test frame alive confirm timeout t₁")
			}
			// check oldest unacknowledged outbound
			if sf.ackNoSend != sf.seqNoSend &&
				//now.Sub(sf.peek()) >= sf.SendUnAckTimeout1 {
				now.Sub(sf.pending[0].sendTime) >= sf.option.config.SendUnAckTimeout1 {
				sf.ackNoSend++
				sf.Error("fatal transmission timeout t₁")
				return errors.New("fatal transmission timeout t₁")
			}

			// If the earliest sent I-frame has timed out, send an S-frame in response
			if sf.ackNoRcv != sf.seqNoRcv &&
				(now.Sub(unAckRcvSince) >= sf.option.config.RecvUnAckTimeout2 ||
					now.Sub(idleTimeout3Sine) >= timeoutResolution) {
				sendSFrame(sf.seqNoRcv)
				sf.ackNoRcv = sf.seqNoRcv
			}

			// When idle timeout elapses, send TestFrActive frame to keep the connection alive
			if now.Sub(idleTimeout3Sine) >= sf.option.config.IdleTimeout3 {
				sf.sendUFrame(uTestFrActive)
				testFrAliveSendSince = time.Now()
				idleTimeout3Sine = testFrAliveSendSince
			}

		case apdu := <-sf.rcvRaw:
			idleTimeout3Sine = time.Now() // Upon receiving any I, S, or U frame, reset the idle timer (t3)
			apci, asduVal := parse(apdu)
			switch head := apci.(type) {
			case sAPCI:
				sf.Debug("RX sFrame %v", head)
				if !sf.updateAckNoOut(head.rcvSN) {
					sf.Error("fatal incoming acknowledge either earlier than previous or later than sendTime")
					return errors.New("fatal incoming acknowledge either earlier than previous or later than sendTime")
				}

			case iAPCI:
				sf.Debug("RX iFrame %v", head)
				if atomic.LoadUint32(&sf.isActive) == inactive {
					sf.Warn("station not active")
					break // not active, discard apdu
				}
				if !sf.updateAckNoOut(head.rcvSN) || head.sendSN != sf.seqNoRcv {
					sf.Error("fatal incoming acknowledge either earlier than previous or later than sendTime")
					return errors.New("fatal incoming acknowledge either earlier than previous or later than sendTime")
				}

				sf.rcvASDU <- asduVal
				if sf.ackNoRcv == sf.seqNoRcv { // first unacked
					unAckRcvSince = time.Now()
				}

				sf.seqNoRcv = (sf.seqNoRcv + 1) & 32767
				if seqNoCount(sf.ackNoRcv, sf.seqNoRcv) >= sf.option.config.RecvUnAckLimitW {
					sendSFrame(sf.seqNoRcv)
					sf.ackNoRcv = sf.seqNoRcv
				}

			case uAPCI:
				sf.Debug("RX uFrame %v", head)
				switch head.function {
				//case uStartDtActive:
				//	sf.sendUFrame(uStartDtConfirm)
				//	atomic.StoreUint32(&sf.isActive, active)
				case uStartDtConfirm:
					atomic.StoreUint32(&sf.isActive, active)
					sf.startDtActiveSendSince.Store(willNotTimeout)
					// notify activation
					sf.onActivated(sf)
				//case uStopDtActive:
				//	sf.sendUFrame(uStopDtConfirm)
				//	atomic.StoreUint32(&sf.isActive, inactive)
				case uStopDtConfirm:
					atomic.StoreUint32(&sf.isActive, inactive)
					sf.stopDtActiveSendSince.Store(willNotTimeout)
					// notify deactivation
					sf.onDeactivated(sf)
				case uTestFrActive:
					sf.sendUFrame(uTestFrConfirm)
				case uTestFrConfirm:
					testFrAliveSendSince = willNotTimeout
				default:
					sf.Error("illegal U-Frame functions[0x%02x] ignored", head.function)
				}
			}
		}
	}
}

func (sf *Client) handlerLoop() {
	sf.Debug("handlerLoop started")
	defer func() {
		sf.wg.Done()
		sf.Debug("handlerLoop stopped")
	}()

	for {
		select {
		case <-sf.ctx.Done():
			return
		case rawAsdu := <-sf.rcvASDU:
			asduPack := asdu.NewEmptyASDU(&sf.option.params)
			if err := asduPack.UnmarshalBinary(rawAsdu); err != nil {
				sf.Warn("asdu UnmarshalBinary failed,%+v", err)
				continue
			}
			if err := sf.clientHandler(asduPack); err != nil {
				sf.Warn("Falied handling I frame, error: %v", err)
			}
		}
	}
}

func (sf *Client) setConnectStatus(status uint32) {
	sf.rwMux.Lock()
	atomic.StoreUint32(&sf.status, status)
	sf.rwMux.Unlock()
}

func (sf *Client) connectStatus() uint32 {
	sf.rwMux.RLock()
	status := atomic.LoadUint32(&sf.status)
	sf.rwMux.RUnlock()
	return status
}

func (sf *Client) cleanUp() {
	sf.ackNoRcv = 0
	sf.ackNoSend = 0
	sf.seqNoRcv = 0
	sf.seqNoSend = 0
	sf.pending = nil
	// clear sending chan buffer
loop:
	for {
		select {
		case <-sf.sendRaw:
		case <-sf.rcvRaw:
		case <-sf.rcvASDU:
		case <-sf.sendASDU:
		default:
			break loop
		}
	}
}

func (sf *Client) sendUFrame(which byte) {
	sf.Debug("TX uFrame %v", uAPCI{which})
	sf.sendRaw <- newUFrame(which)
}

func (sf *Client) updateAckNoOut(ackNo uint16) (ok bool) {
	if ackNo == sf.ackNoSend {
		return true
	}
	// Validate new acknowledgements: ACK cannot precede the request sequence number; treat as error
	if seqNoCount(sf.ackNoSend, sf.seqNoSend) < seqNoCount(ackNo, sf.seqNoSend) {
		return false
	}

	// confirm reception
	for i, v := range sf.pending {
		if v.seq == (ackNo - 1) {
			sf.pending = sf.pending[i+1:]
			break
		}
	}

	sf.ackNoSend = ackNo
	return true
}

// IsConnected get server session connected state
func (sf *Client) IsConnected() bool {
	return sf.connectStatus() == connected
}

// IsActive returns whether the data transfer is active (StartDT confirmed)
func (sf *Client) IsActive() bool {
	return atomic.LoadUint32(&sf.isActive) == active
}

// clientHandler hand response handler
func (sf *Client) clientHandler(asduPack *asdu.ASDU) error {
	sf.Debug("ASDU %+v", asduPack)
	msg, err := asdu.ParseASDU(asduPack)
	if err != nil {
		return err
	}
	return sf.handler.Handle(sf, msg)
}

// Params returns params of client
func (sf *Client) Params() *asdu.Params {
	return &sf.option.params
}

// Send send asdu
func (sf *Client) Send(a *asdu.ASDU) error {
	if !sf.IsConnected() {
		return ErrUseClosedConnection
	}
	if atomic.LoadUint32(&sf.isActive) == inactive {
		return ErrNotActive
	}
	data, err := a.MarshalBinary()
	if err != nil {
		return err
	}
	select {
	case sf.sendASDU <- data:
	default:
		return ErrBufferFulled
	}
	return nil
}

// UnderlyingConn returns underlying conn of client
func (sf *Client) UnderlyingConn() net.Conn {
	return sf.conn
}

// Close close all
func (sf *Client) Close() error {
	sf.rwMux.Lock()
	if sf.closeCancel != nil {
		sf.closeCancel()
	}
	sf.rwMux.Unlock()
	return nil
}

// SendStartDt start data transmission on this connection
func (sf *Client) SendStartDt() {
	sf.startDtActiveSendSince.Store(time.Now())
	sf.sendUFrame(uStartDtActive)
}

// SendStopDt stop data transmission on this connection
func (sf *Client) SendStopDt() {
	sf.stopDtActiveSendSince.Store(time.Now())
	sf.sendUFrame(uStopDtActive)
}

// InterrogationCmd wrap asdu.InterrogationCmd
func (sf *Client) InterrogationCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, qoi asdu.QualifierOfInterrogation) error {
	return asdu.InterrogationCmd(sf, coa, ca, qoi)
}

// CounterInterrogationCmd wrap asdu.CounterInterrogationCmd
func (sf *Client) CounterInterrogationCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, qcc asdu.QualifierCountCall) error {
	return asdu.CounterInterrogationCmd(sf, coa, ca, qcc)
}

// ReadCmd wrap asdu.ReadCmd
func (sf *Client) ReadCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, ioa asdu.InfoObjAddr) error {
	return asdu.ReadCmd(sf, coa, ca, ioa)
}

// ClockSynchronizationCmd wrap asdu.ClockSynchronizationCmd
func (sf *Client) ClockSynchronizationCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, t time.Time) error {
	return asdu.ClockSynchronizationCmd(sf, coa, ca, t)
}

// ResetProcessCmd wrap asdu.ResetProcessCmd
func (sf *Client) ResetProcessCmd(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, qrp asdu.QualifierOfResetProcessCmd) error {
	return asdu.ResetProcessCmd(sf, coa, ca, qrp)
}

// DelayAcquireCommand wrap asdu.DelayAcquireCommand
func (sf *Client) DelayAcquireCommand(coa asdu.CauseOfTransmission, ca asdu.CommonAddr, msec uint16) error {
	return asdu.DelayAcquireCommand(sf, coa, ca, msec)
}

// TestCommand  wrap asdu.TestCommand
func (sf *Client) TestCommand(coa asdu.CauseOfTransmission, ca asdu.CommonAddr) error {
	return asdu.TestCommand(sf, coa, ca)
}
