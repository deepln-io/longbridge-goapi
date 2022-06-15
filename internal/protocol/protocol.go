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

// Package protocol provides the binary encoding/decoding for long bridge communication protocol.
// The design principle is to follow minimalism.
// It provides to main API Send and Recv to send and receive binary data from io.Reader and io.Writer following
// longbridge communication protocol.
// No specific network communication protocl (e.g., TCP or Websocket) is assumed.
package protocol

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"

	"github.com/deepln-io/longbridge-goapi/internal/pb/control"
	"github.com/deepln-io/longbridge-goapi/internal/pb/quote"
	"github.com/deepln-io/longbridge-goapi/internal/pb/trade"
)

const BodyLenLimit = 16 * 1024 * 1024 // 16MB

var defaultSignature = [16]byte{'d', 'e', 'e', 'p', 'l', 'n', '.', 'i', 'o'}

type Command byte

// The command code
const (
	CmdClose     = Command(control.Command_CMD_CLOSE)
	CmdHeartbeat = Command(control.Command_CMD_HEARTBEAT)
	CmdAuth      = Command(control.Command_CMD_AUTH)
	CmdReconnect = Command(control.Command_CMD_RECONNECT)

	CmdSub    = Command(trade.Command_CMD_SUB)
	CmdUnSub  = Command(trade.Command_CMD_UNSUB)
	CmdNotify = Command(trade.Command_CMD_NOTIFY)

	CmdInfo                        = Command(quote.Command_QuerySecurityStaticInfo)
	CmdRealtimeQuote               = Command(quote.Command_QuerySecurityQuote)
	CmdRealtimeOptionQuote         = Command(quote.Command_QueryOptionQuote)
	CmdRealtimeWarrantQuote        = Command(quote.Command_QueryWarrantQuote)
	CmdOrderBook                   = Command(quote.Command_QueryDepth)
	CmdBrokerQueue                 = Command(quote.Command_QueryBrokers)
	CmdBrokerInfo                  = Command(quote.Command_QueryParticipantBrokerIds)
	CmdTicker                      = Command(quote.Command_QueryTrade)
	CmdIntraday                    = Command(quote.Command_QueryIntraday)
	CmdKLine                       = Command(quote.Command_QueryCandlestick)
	CmdOptionExpiryDate            = Command(quote.Command_QueryOptionChainDate)
	CmdOptionDateStrikeInfo        = Command(quote.Command_QueryOptionChainDateStrikeInfo)
	CmdWarrantIssuers              = Command(quote.Command_QueryWarrantIssuerInfo)
	CmdSearchWarrant               = Command(quote.Command_QueryWarrantFilterList)
	CmdMarketTradePeriod           = Command(quote.Command_QueryMarketTradePeriod)
	CmdTradeDate                   = Command(quote.Command_QueryMarketTradeDay)
	CmdIntradayCapFlow             = Command(quote.Command_QueryCapitalFlowIntraday)
	CmdIntradayCapFlowDistribution = Command(quote.Command_QueryCapitalFlowDistribution)
	CmdSecurityFinanceIndex        = Command(quote.Command_QuerySecurityCalcIndex)
	CmdSubscription                = Command(quote.Command_Subscription)
	CmdSubscribe                   = Command(quote.Command_Subscribe)
	CmdUnsubscribe                 = Command(quote.Command_Unsubscribe)
	CmdPushQuoteData               = Command(quote.Command_PushQuoteData)
	CmdPushOrderBookData           = Command(quote.Command_PushDepthData)
	CmdPushBrokersData             = Command(quote.Command_PushBrokersData)
	CmdPushTickersData             = Command(quote.Command_PushTradeData)
)

type PkgType int8

const (
	UnknownPackage PkgType = 0
	ReqPackage     PkgType = 1
	RespPackage    PkgType = 2
	PushPackage    PkgType = 3
)

type Flag struct {
	PkgType    PkgType
	NeedVerify bool
	UseGzip    bool
}

type Common struct {
	Flag
	CmdCode byte
	BodyLen uint32
}

type PkgHeader interface {
	GetCommonHeader() *Common
}

type ReqPkgHeader struct {
	Common
	RequestID uint32
	Timeout   uint16
}

func (r *ReqPkgHeader) GetCommonHeader() *Common {
	return &r.Common
}

type RespPkgHeader struct {
	Common
	RequestID  uint32
	StatusCode uint8
}

func (r *RespPkgHeader) GetCommonHeader() *Common {
	return &r.Common
}

type PushPkgHeader struct {
	Common
}

func (p *PushPkgHeader) GetCommonHeader() *Common {
	return &p.Common
}

type OptionalSignature struct {
	Nonce     uint64
	Signature [16]byte
}

func NewFlag(b byte) Flag {
	return Flag{
		PkgType:    PkgType(b & 0xf),
		NeedVerify: b&0x10 == 0x10,
		UseGzip:    b&0x20 == 0x20,
	}
}

func (f Flag) ToByte() byte {
	var b byte
	if f.UseGzip {
		b |= 0x20
	}
	if f.NeedVerify {
		b |= 0x10
	}
	b |= byte(f.PkgType & 0xf)
	return b
}

type encoder struct {
	writer io.Writer
	err    error
}

func (e *encoder) pack(data interface{}) {
	if e.err != nil {
		return
	}
	e.err = binary.Write(e.writer, binary.BigEndian, data)
}

type decoder struct {
	reader io.Reader
	err    error
}

func (d *decoder) unpack(data interface{}) {
	if d.err != nil {
		return
	}
	d.err = binary.Read(d.reader, binary.BigEndian, data)
}

// Send marshals the header, body and optional signature into byte stream according to long bridge binary communication protocol and
// writes it to writer. The common fields PkgType and BodyLen are filled automatically by Send().
func Send(w io.Writer, header PkgHeader, body []byte) error {
	if header == nil || reflect.ValueOf(header).IsNil() {
		return fmt.Errorf("nil header %#v", header)
	}
	cm := header.GetCommonHeader()
	switch header.(type) {
	case *ReqPkgHeader:
		cm.PkgType = ReqPackage
	case *RespPkgHeader:
		cm.PkgType = RespPackage
	case *PushPkgHeader:
		cm.PkgType = PushPackage
	default:
		return fmt.Errorf("unsupported type for header: %v", cm)
	}

	if cm.UseGzip {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write(body); err != nil {
			return fmt.Errorf("error zipping body data: %v", err)
		}
		if err := zw.Close(); err != nil {
			return fmt.Errorf("error flushing zip buffer: %v", err)
		}
		body = buf.Bytes()
	}
	if len(body) >= BodyLenLimit {
		return fmt.Errorf("body len %d is beyond limit %d", len(body), BodyLenLimit)
	}

	var buf bytes.Buffer
	if err := buf.WriteByte(cm.Flag.ToByte()); err != nil {
		return err
	}
	enc := &encoder{writer: &buf}
	enc.pack(cm.CmdCode)
	switch r := header.(type) {
	case *ReqPkgHeader:
		enc.pack(r.RequestID)
		enc.pack(r.Timeout)

	case *RespPkgHeader:
		enc.pack(r.RequestID)
		enc.pack(r.StatusCode)

	case *PushPkgHeader:
	}
	cm.BodyLen = uint32(len(body))
	enc.pack([]uint8{uint8(cm.BodyLen & 0xff0000 >> 16), uint8(cm.BodyLen & 0xff00 >> 8), uint8(cm.BodyLen & 0x0ff)})

	if enc.err != nil {
		return fmt.Errorf("error packing request package: %v", enc.err)
	}
	if _, err := buf.Write(body); err != nil {
		return fmt.Errorf("error writing body (size=%d): %v", len(body), err)
	}
	if cm.NeedVerify {
		// Currently the signature is unused, just send emtpy data if the flag instructs to verify
		sig := &OptionalSignature{Signature: defaultSignature}
		if err := binary.Write(&buf, binary.BigEndian, sig.Nonce); err != nil {
			return fmt.Errorf("error writing nonce %v: %v", sig.Nonce, err)
		}
		if _, err := buf.Write(sig.Signature[:]); err != nil {
			return fmt.Errorf("error writing signature %v: %v", sig.Signature, err)
		}
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func unpackCommonHeader(r io.Reader) (*Common, error) {
	b := make([]byte, 2)
	n, err := r.Read(b)
	if err != nil {
		// Use %w to wrap the underlying io error, using errors.Is(err, io.EOF) to check if the connection is lost.
		return nil, fmt.Errorf("error unpacking common header: %w", err)
	}
	if n < 2 {
		return nil, fmt.Errorf("not enough data (%d) for header", n)
	}
	return &Common{Flag: NewFlag(b[0]), CmdCode: b[1]}, nil
}

func unpackReqPackageHeader(r io.Reader, rh *ReqPkgHeader) error {
	dec := &decoder{reader: r}
	dec.unpack(&rh.RequestID)
	dec.unpack(&rh.Timeout)
	var b [3]uint8
	dec.unpack(&b[0])
	dec.unpack(&b[1])
	dec.unpack(&b[2])
	rh.BodyLen = uint32(b[0])<<16 + uint32(b[1])<<8 + uint32(b[2])
	return dec.err
}

func unpackRespPackageHeader(r io.Reader, rh *RespPkgHeader) error {
	dec := &decoder{reader: r}
	dec.unpack(&rh.RequestID)
	dec.unpack(&rh.StatusCode)
	var b [3]uint8
	dec.unpack(&b[0])
	dec.unpack(&b[1])
	dec.unpack(&b[2])
	rh.BodyLen = uint32(b[0])<<16 + uint32(b[1])<<8 + uint32(b[2])
	return dec.err
}

func unpackPushPackageHeader(r io.Reader, ph *PushPkgHeader) error {
	dec := &decoder{reader: r}
	var b [3]uint8
	dec.unpack(&b[0])
	dec.unpack(&b[1])
	dec.unpack(&b[2])
	ph.BodyLen = uint32(b[0])<<16 + uint32(b[1])<<8 + uint32(b[2])
	return dec.err
}

func unpackSignature(r io.Reader, sig *OptionalSignature) error {
	dec := &decoder{reader: r}
	dec.unpack(&sig.Nonce)
	dec.unpack(&sig.Signature)
	return dec.err
}

// Recv gets the package header and body from the reader.
func Recv(r io.Reader) (header PkgHeader, body []byte, err error) {
	cm, err := unpackCommonHeader(r)
	if err != nil {
		return
	}
	switch cm.PkgType {
	case ReqPackage:
		rh := &ReqPkgHeader{Common: *cm}
		cm = &rh.Common
		err = unpackReqPackageHeader(r, rh)
		header = rh
	case RespPackage:
		rh := &RespPkgHeader{Common: *cm}
		cm = &rh.Common
		err = unpackRespPackageHeader(r, rh)
		header = rh
	case PushPackage:
		ph := &PushPkgHeader{Common: *cm}
		cm = &ph.Common
		err = unpackPushPackageHeader(r, ph)
		header = ph
	}
	if err != nil {
		return
	}
	body = make([]byte, cm.BodyLen)
	_, err = io.ReadFull(r, body)
	if err != nil {
		return
	}
	if cm.UseGzip {
		var zr *gzip.Reader
		zr, err = gzip.NewReader(bytes.NewBuffer(body))
		if err != nil {
			return
		}
		body, err = io.ReadAll(zr)
		cm.BodyLen = uint32(len(body)) // Set it to uncompressed size
		zr.Close()
		if err != nil {
			return
		}
	}
	if header.GetCommonHeader().NeedVerify {
		var sig OptionalSignature
		err = unpackSignature(r, &sig)
		if err != nil {
			return
		}
		// Note that currently the signature is not really verified, according to Long bridge developer group, we simply discard it here.
	}
	return
}
