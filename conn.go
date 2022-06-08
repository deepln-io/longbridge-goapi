// Copyright 2022 Deep Learning Limited. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package longbridge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deepln-io/longbridge-goapi/internal/pb/control"
	"github.com/deepln-io/longbridge-goapi/internal/pb/trade"
	"github.com/deepln-io/longbridge-goapi/internal/protocol"
	"github.com/golang/glog"
	"golang.org/x/net/websocket"
	"google.golang.org/protobuf/proto"
)

type otpProvider interface {
	getOTP() (string, error)
}

type respPackage struct {
	Header protocol.PkgHeader
	Body   []byte
	Err    error
}

type longConn struct {
	UseGzip           bool
	reconnectInterval time.Duration
	endPoint          string
	cancel            context.CancelFunc
	// Used to block API call if the client is not in authenticated state
	reqChan     chan struct{}
	conn        net.Conn                       // The net connection for long-polling connection (tcp or web socket)
	mu          sync.RWMutex                   // protecting responses
	responses   map[uint32]chan<- *respPackage // Map from series number to a receiving channel
	onPush      func(header *protocol.PushPkgHeader, body []byte, pkgErr error)
	otpProvider otpProvider
	sessionID   string // The session ID from last auth result, used for reconnection if not expired
	expires     int64
	seriesNum   uint32
	recover     func() // The hook to run when connection is restored
}

type pushContentType uint8

const (
	pushContentTypeUndefined pushContentType = pushContentType(trade.ContentType_CONTENT_UNDEFINED)
	pushContentTypeJSON      pushContentType = pushContentType(trade.ContentType_CONTENT_JSON)
	pushContentTypeProto     pushContentType = pushContentType(trade.ContentType_CONTENT_PROTO)
)

const defaultTimeout uint16 = 10000 // In milleseconds, default 10s, max 60s

// connState represents the client connection state for long polling notification service.
type connState int8

const (
	stateInit      connState = iota
	stateConnected           // Network connected, but the session is not authenticated
	stateAuthenticated
	stateDisconnected
)

type report struct {
	newState connState
	reason   string
}

func newLongConn(endPoint string, otpProvider otpProvider) *longConn {
	c := &longConn{
		endPoint:    endPoint,
		otpProvider: otpProvider,
		reqChan:     make(chan struct{}),
	}
	c.onPush = c.handlePushPkg
	return c
}

func (c *longConn) send(header protocol.PkgHeader, message proto.Message) error {
	glog.V(3).Infof("Primitive send: cmd=%d", header.GetCommonHeader().CmdCode)
	body, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message body: %v", err)
	}
	return protocol.Send(c.conn, header, body)
}

func (c *longConn) recv(message proto.Message) (protocol.PkgHeader, error) {
	h, body, err := protocol.Recv(c.conn)
	if err != nil {
		return nil, fmt.Errorf("error getting the auth response: %w", err)
	}
	glog.V(3).Infof("Primitive recv: pkg: %d, cmd: %d, message len: %d",
		h.GetCommonHeader().PkgType, h.GetCommonHeader().CmdCode, h.GetCommonHeader().BodyLen)
	return h, proto.Unmarshal(body, message)
}

func (c *longConn) getReqHeader(cmdCode protocol.Command) *protocol.ReqPkgHeader {
	header := &protocol.ReqPkgHeader{}
	header.CmdCode = byte(cmdCode)
	header.RequestID = atomic.AddUint32(&c.seriesNum, 1)
	header.Timeout = defaultTimeout
	header.UseGzip = c.UseGzip
	return header
}

// auth sents the One Time Password (OTP) to authenticate the session.
// It relies on the primitive send/recv functions to communicate with long bridge server.
// Once a session is authenticated, the further API can use high level API Call() for API request.
func (c *longConn) auth(token string) error {
	glog.V(3).Infof("Auth with token %s", token)
	header := c.getReqHeader(protocol.CmdAuth)
	if err := c.send(header, &control.AuthRequest{Token: token}); err != nil {
		return err
	}
	var resp control.AuthResponse
	if _, err := c.recv(&resp); err != nil {
		return fmt.Errorf("invalid auth response: %v", err)
	}
	c.sessionID = resp.GetSessionId()
	c.expires = resp.GetExpires() / 1000 // Converted to seconds
	glog.V(3).Infof("Got auth session %v, expires on %v", c.sessionID, time.Unix(c.expires, 0))
	return nil
}

func (c *longConn) reAuth() error {
	glog.V(3).Infof("ReAuth with session %s", c.sessionID)
	header := c.getReqHeader(protocol.CmdReconnect)
	if err := c.send(header, &control.ReconnectRequest{SessionId: c.sessionID}); err != nil {
		return err
	}
	var resp control.ReconnectResponse
	if _, err := c.recv(&resp); err != nil {
		return fmt.Errorf("invalid reconnect resp: %v", err)
	}
	c.sessionID = resp.GetSessionId()
	c.expires = resp.GetExpires() / 1000
	return nil
}

func (c *longConn) ping() error {
	defer trace("ping")()
	header := c.getReqHeader(protocol.CmdHeartbeat)
	return c.send(header, &control.Heartbeat{Timestamp: time.Now().Unix()})
}

// Call sends the request and waits for the response with protobuf protocol.
// It is a high level function for various application API to use.
// Setting 0 to timeout value means no timeout for the API call.
func (c *longConn) Call(apiName string, header *protocol.ReqPkgHeader, message proto.Message,
	resp proto.Message, timeout time.Duration) error {
	glog.V(3).Infof("Call %s, waiting for connection", apiName)
	c.reqChan <- struct{}{}
	glog.V(3).Infof("Call %s, connection ready", apiName)
	defer glog.V(3).Infof("Call return resp with %#v", resp)

	ch := make(chan *respPackage, 1)
	c.mu.Lock()
	if c.responses == nil {
		c.mu.Unlock()
		return fmt.Errorf("client is not ready for API %s", apiName)
	}
	c.responses[header.RequestID] = ch
	c.mu.Unlock()
	if err := c.send(header, message); err != nil {
		return fmt.Errorf("error sending %s request: %v", apiName, err)
	}
	var pkg *respPackage
	if timeout == 0 {
		pkg = <-ch
	} else {
		select {
		case <-time.After(timeout):
			pkg = &respPackage{Err: fmt.Errorf("time out for %s", apiName)}
		case pkg = <-ch:
		}
	}
	if pkg.Err != nil {
		return fmt.Errorf("error receiving %s response: %v", apiName, pkg.Err)
	}
	if err := proto.Unmarshal(pkg.Body, resp); err != nil {
		return fmt.Errorf("error receiving response: %v", err)
	}
	return nil
}

// The method connect establishes the network connection to remote end points.
// It chooses TCP or web socket according to the end point protocol scheme.
// It is secured to use web socket for trade related APIs.
func (c *longConn) connect(ctx context.Context) error {
	defer trace("connect")()
	if strings.HasPrefix(c.endPoint, "tcp://") {
		glog.V(2).Infof("Connected using TCP protocol to addr %v", c.endPoint)
		tc, err := net.Dial("tcp", c.endPoint[len("tcp://"):])
		if err != nil {
			return err
		}
		if _, err := tc.Write([]byte{0b00010001, 0b00001001}); err != nil {
			return err
		}
		c.conn = tc
	} else {
		glog.V(2).Infof("Connected using Web socket protocol to %v", c.endPoint)
		ws, err := websocket.Dial(c.endPoint, "", "http://localhost")
		if err != nil {
			return err
		}
		c.conn = ws
	}
	return nil
}

func (c *longConn) establishSession(ctx context.Context) error {
	if c.sessionID == "" || time.Now().After(time.Unix(c.expires, 0)) {
		otp, err := c.otpProvider.getOTP()
		if err != nil {
			return fmt.Errorf("error getting access token: %v", err)
		}
		if err := c.auth(otp); err != nil {
			return fmt.Errorf("error auth to long bridge gateway: %v", err)
		}
	} else {
		if err := c.reAuth(); err != nil {
			c.sessionID = ""
			c.expires = 0
			return fmt.Errorf("error reconnecting: %v", err)
		}
	}
	return nil
}

// keepAlive uses a simple method to keep connection alive by sending a hearbeat message every 10 seconds.
func (c *longConn) keepAlive(ctx context.Context, reporter chan<- *report) {
	defer trace("hearbeat")()
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.ping(); err != nil {
				glog.V(2).Infof("Error sending hearbeat: %v", err)
				go func(err error) { reporter <- &report{newState: stateDisconnected, reason: err.Error()} }(err)
			}
		}
	}
}

func (c *longConn) handleRespPkg(header *protocol.RespPkgHeader, body []byte, pkgErr error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.responses == nil {
		return
	}
	ch, ok := c.responses[header.RequestID]
	if !ok {
		level := glog.Level(4)
		if header.CmdCode == byte(protocol.CmdHeartbeat) {
			level = glog.Level(3)
		}
		glog.V(level).Infof("Discard package with request ID %v, pkgType %v, cmd %v", header.RequestID, header.PkgType, header.CmdCode)
		return
	}
	delete(c.responses, header.RequestID)
	switch header.StatusCode {
	case 0:
	default:
		if pkgErr == nil {
			pkgErr = fmt.Errorf("error response with status code %v", header.StatusCode)
		}
	}
	go func() { ch <- &respPackage{Header: header, Body: body, Err: pkgErr} }()
}

func (c *longConn) handlePushPkg(header *protocol.PushPkgHeader, body []byte, pkgErr error) {
	glog.V(2).Infof("Discard push package: cmd=%d len=%d err=%v", header.CmdCode, len(body), pkgErr)
}

// Function readLoop waits for the incoming packages, delivers notifications to subscriber, sends package to requester,
// or reports errors to reporter, which will cause the FSM in runLoop to change state.
func (c *longConn) readLoop(ctx context.Context, reporter chan<- *report) {
	defer trace("readLoop")()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if c.conn == nil {
				return
			}
			h, body, err := protocol.Recv(c.conn)
			if errors.Is(err, io.EOF) {
				// Connection lost
				go func(err error) { reporter <- &report{newState: stateDisconnected, reason: err.Error()} }(err)
				return
			}
			switch header := h.(type) {
			case *protocol.ReqPkgHeader:
				if header.CmdCode == byte(protocol.CmdHeartbeat) {
					// heartbeat request, reply with same body
					protocol.Send(c.conn, &protocol.RespPkgHeader{Common: header.Common, RequestID: header.RequestID, StatusCode: 0}, body)
					return
				}
				if header.CmdCode == byte(protocol.CmdClose) {
					// Server initializes the closing event
					var resp control.Close
					if err := proto.Unmarshal(body, &resp); err != nil {
						glog.V(2).Infof("Discard ill format server close request: %v", err)
						return
					}
					switch resp.Code {
					case control.Close_HeartbeatTimeout, control.Close_ServerError, control.Close_ServerShutdown:
						glog.Infof("Server initializes disconnection with code %v. Will reconnect", resp.Code)
						go func(err error) { reporter <- &report{newState: stateDisconnected, reason: resp.Reason} }(err)
					default:
						glog.Infof("Server initializes disconnection with code %v. Will re-establish the session", resp.Code)
						go func(err error) { reporter <- &report{newState: stateConnected, reason: resp.Reason} }(err)
					}
					return
				}
				glog.V(2).Infof("Unknown request from server with command code=%d", header.CmdCode)

			case *protocol.RespPkgHeader:
				c.handleRespPkg(header, body, err)

			case *protocol.PushPkgHeader:
				c.onPush(header, body, err)
			}
		}
	}
}

func (c *longConn) drainResponses() {
	c.mu.Lock()
	for k, ch := range c.responses {
		delete(c.responses, k)
		close(ch)
	}
	// Note that notification handlers are kept
	c.responses = nil
	c.mu.Unlock()
}

// Function runLoop is the main loop that manages the long connection when it is enabled.
// It supports auto connection, authentication, and reconnection upon connection failure.
// It uses a finite state machine (FSM) to organize the control flow.
// The update to state is through channel instead of sharing variable.
func (c *longConn) runLoop(ctx context.Context) {
	defer trace("runLoop")()
	stateChan := make(chan connState, 1)
	setState := func(state connState) {
		stateChan <- state
		glog.V(2).Infof("Client transit to %v", state)
	}
	// The reporter is used by the receiving go routine to report network error and instructs this go routine to change state.
	reporter := make(chan *report)
	reset := make(chan struct{}) // Instruct the API approving go routine to leave when client is leaving authenticated state.
	defer func() {
		c.drainResponses()
	}()
	stateChan <- stateInit
	for {
		select {
		case <-ctx.Done():
			return

		case report := <-reporter:
			setState(report.newState)
			if report.newState != stateAuthenticated {
				reset <- struct{}{}
			}
			glog.Infof("Something wrong reason: %s", report.reason)

		case state := <-stateChan:
			switch state {
			case stateDisconnected:
				c.drainResponses()
				if c.recover != nil {
					c.recover()
				}
				fallthrough

			case stateInit:
				if err := c.connect(ctx); err != nil {
					glog.V(3).Infof("Error connecting to long bridge push service: %v. Retry in %d seconds",
						err, c.reconnectInterval/time.Second)
					time.Sleep(c.reconnectInterval)
					setState(stateInit)
					continue
				}
				setState(stateConnected)

			case stateConnected:
				if err := c.establishSession(ctx); err != nil {
					glog.V(3).Infof("Error establishing session to long bridge push service: %v. Retry in %d seconds",
						err, c.reconnectInterval/time.Second)
					time.Sleep(c.reconnectInterval)
					// To avoid being stuck in this state, randomly switch back to reconnection.
					// Another approach is to set a max retrial limit.
					if rand.Float64() < 0.5 {
						setState(stateConnected)
					} else {
						setState(stateInit)
					}
					continue
				}
				setState(stateAuthenticated)

			case stateAuthenticated:
				c.mu.Lock()
				c.responses = make(map[uint32]chan<- *respPackage)
				c.mu.Unlock()
				// The go routine is used to open a service window to approve API requests
				// until reporter advices a door closed signal upon connection failure.
				// It also maintains two child go routines so that when it exits, the children exit too.
				go func(ctx context.Context) {
					defer trace("approving")()
					cctx, cancel := context.WithCancel(ctx)
					defer cancel()
					go c.keepAlive(cctx, reporter)
					go c.readLoop(cctx, reporter)
					for {
						select {
						case <-c.reqChan:
							// grant API calls
						case <-reset:
							return
						case <-ctx.Done():
							if c.conn != nil {
								c.conn.Close() // Force the readLoop to exit from recv()
								c.conn = nil
							}
							return
						}
					}
				}(ctx)
			}
		}
	}
}

// Enable enables/disables the long connection for receiving notifications.
// Since Long Bridge limits one long polling connection per account, we need to close the connection for an account with different
// markets to support running different trading processes for different markets.
// So when a market is opened, the connection is established, and is closed when market closed.
func (c *longConn) Enable(enable bool) {
	if enable {
		if c.cancel != nil {
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		c.cancel = cancel
		go c.runLoop(ctx)
	} else {
		if c.cancel == nil {
			return
		}
		c.cancel()
		c.cancel = nil
		c.seriesNum = 0
		c.sessionID = ""
		c.expires = 0
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
	}
}
