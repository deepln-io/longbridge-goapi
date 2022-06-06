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

package protocol_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/deepln-io/longbridge-goapi/internal/pb/control"
	"github.com/deepln-io/longbridge-goapi/internal/pb/trade"
	"github.com/deepln-io/longbridge-goapi/internal/protocol"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSendRecvReqPkg(t *testing.T) {
	assert := assert.New(t)
	var buf bytes.Buffer
	header := &protocol.ReqPkgHeader{RequestID: 1, Timeout: 1000}
	header.CmdCode = byte(protocol.CmdHeartbeat)

	err := protocol.Send(&buf, header, nil)
	assert.Nil(err, "Checking empty body request")
	assert.Equal("01 01 00 00 00 01 03 e8 00 00 00", fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes")

	rh, rb, err := protocol.Recv(&buf)
	assert.Nil(err, "Checking recv empty body")
	assert.Equal(0, len(rb), "Checking recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")

	buf.Reset()
	header.CmdCode = byte(protocol.CmdAuth)
	body, _ := proto.Marshal(&control.AuthRequest{Token: "one time password"})
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking Send()")
	assert.Equal(uint32(len(body)), header.BodyLen, "Verify body length")
	assert.Equal("01 02 00 00 00 01 03 e8 00 00 13 0a 11 6f 6e 65 20 74 69 6d 65 20 70 61 73 73 77 6f 72 64",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv non-empty body")
	assert.Equal(len(body), len(rb), "Checking non-empty recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")
	assert.Equal(body, rb, "Checking recv body content")

	buf.Reset()
	header.UseGzip = true
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking Send() with gzip")
	assert.Equal("21 02 00 00 00 01 03 e8 00 00 2b 1f 8b 08 00 00 00 00 00 00 ff e2 12 cc cf 4b 55 28 c9 cc 4d 55 28 48 2c 2e 2e cf 2f 4a 01 04 00 00 ff ff 3a eb bc c7 13 00 00 00",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes with zipped body")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv gzipped non-empty body")
	assert.Equal(len(body), len(rb), "Checking gzipped non-empty recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")
	assert.Equal(body, rb, "Checking recv body content")

	buf.Reset()
	header.UseGzip = false
	header.NeedVerify = true
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking empty body request")
	assert.Equal("11 02 00 00 00 01 03 e8 00 00 13 0a 11 6f 6e 65 20 74 69 6d 65 20 70 61 73 73 77 6f 72 64 00 00 00 00 00 00 00 00 64 65 65 70 6c 6e 2e 69 6f 00 00 00 00 00 00 00",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes with signature and gzipped")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv non-empty body with verification")
	assert.Equal(len(body), len(rb), "Checking non-empty recv body length with verification")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag with verification")
	assert.Equal(body, rb, "Checking recv body content with verification")
}

func TestSendRecvRespPkg(t *testing.T) {
	assert := assert.New(t)
	var buf bytes.Buffer
	header := &protocol.RespPkgHeader{RequestID: 1, StatusCode: 3}
	header.CmdCode = byte(protocol.CmdAuth)

	err := protocol.Send(&buf, header, nil)
	assert.Nil(err, "Checking empty resp body")
	assert.Equal("02 02 00 00 00 01 03 00 00 00", fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes")

	rh, rb, err := protocol.Recv(&buf)
	assert.Nil(err, "Checking recv empty body")
	assert.Equal(0, len(rb), "Checking recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")

	buf.Reset()
	header.CmdCode = byte(protocol.CmdAuth)
	body, _ := proto.Marshal(&control.AuthResponse{SessionId: "session-id", Expires: 1653632133001})
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking Send()")
	assert.Equal(uint32(len(body)), header.BodyLen, "Verify body length")
	assert.Equal("02 02 00 00 00 01 03 00 00 13 0a 0a 73 65 73 73 69 6f 6e 2d 69 64 10 89 cf 9f a1 90 30",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv non-empty body")
	assert.Equal(len(body), len(rb), "Checking non-empty recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")
	assert.Equal(body, rb, "Checking recv body content")

	buf.Reset()
	header.UseGzip = true
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking Send() with gzip")
	assert.Equal("22 02 00 00 00 01 03 00 00 2c 1f 8b 08 00 00 00 00 00 00 ff e2 e2 2a 4e 2d 2e ce cc cf d3 cd 4c 11 e8 3c 3f 7f e1 04 03 40 00 00 00 ff ff 51 a1 64 d0 13 00 00 00",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes with zipped body")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv gzipped non-empty body")
	assert.Equal(len(body), len(rb), "Checking gzipped non-empty recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")
	assert.Equal(body, rb, "Checking recv body content")

	header.UseGzip = false
	header.NeedVerify = true
	buf.Reset()
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking empty body request")
	assert.Equal("12 02 00 00 00 01 03 00 00 13 0a 0a 73 65 73 73 69 6f 6e 2d 69 64 10 89 cf 9f a1 90 30 00 00 00 00 00 00 00 00 64 65 65 70 6c 6e 2e 69 6f 00 00 00 00 00 00 00",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes with signature and gzipped")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv non-empty body with verification")
	assert.Equal(len(body), len(rb), "Checking non-empty recv body length with verification")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag with verification")
	assert.Equal(body, rb, "Checking recv body content with verification")
}

func TestSendRecvPushPkg(t *testing.T) {
	assert := assert.New(t)
	var buf bytes.Buffer
	header := &protocol.PushPkgHeader{}
	header.CmdCode = byte(protocol.CmdNotify)

	err := protocol.Send(&buf, header, nil)
	assert.Nil(err, "Checking empty resp body")
	assert.Equal("03 12 00 00 00", fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes")

	rh, rb, err := protocol.Recv(&buf)
	assert.Nil(err, "Checking recv empty body")
	assert.Equal(0, len(rb), "Checking recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")

	buf.Reset()
	header.CmdCode = byte(protocol.CmdAuth)
	body, _ := proto.Marshal(&trade.Notification{
		Topic:       "stock market",
		ContentType: trade.ContentType_CONTENT_JSON,
		Data:        []byte("[1,2,3]")})
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking Send()")
	assert.Equal(uint32(len(body)), header.BodyLen, "Verify body length")
	assert.Equal("03 02 00 00 19 0a 0c 73 74 6f 63 6b 20 6d 61 72 6b 65 74 10 01 22 07 5b 31 2c 32 2c 33 5d",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv non-empty body")
	assert.Equal(len(body), len(rb), "Checking non-empty recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")
	assert.Equal(body, rb, "Checking recv body content")

	buf.Reset()
	header.UseGzip = true
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking Send() with gzip")
	assert.Equal("23 02 00 00 31 1f 8b 08 00 00 00 00 00 00 ff e2 e2 29 2e c9 4f ce 56 c8 4d 2c ca 4e 2d 11 60 54 62 8f 36 d4 31 d2 31 8e 05 04 00 00 ff ff 83 bc 10 32 19 00 00 00",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes with zipped body")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv gzipped non-empty body")
	assert.Equal(len(body), len(rb), "Checking gzipped non-empty recv body length")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag")
	assert.Equal(body, rb, "Checking recv body content")

	header.UseGzip = false
	header.NeedVerify = true
	buf.Reset()
	err = protocol.Send(&buf, header, body)
	assert.Nil(err, "Checking empty body request")
	assert.Equal("13 02 00 00 19 0a 0c 73 74 6f 63 6b 20 6d 61 72 6b 65 74 10 01 22 07 5b 31 2c 32 2c 33 5d 00 00 00 00 00 00 00 00 64 65 65 70 6c 6e 2e 69 6f 00 00 00 00 00 00 00",
		fmt.Sprintf("% x", buf.Bytes()), "Verify the encoded bytes with signature and gzipped")

	rh, rb, err = protocol.Recv(&buf)
	assert.Nil(err, "Checking recv non-empty body with verification")
	assert.Equal(len(body), len(rb), "Checking non-empty recv body length with verification")
	assert.Equal(header.Flag, rh.GetCommonHeader().Flag, "Verify recv header flag with verification")
	assert.Equal(body, rb, "Checking recv body content with verification")
}

func TestInvalidPackage(t *testing.T) {
	assert := assert.New(t)
	var buf bytes.Buffer
	header := &protocol.ReqPkgHeader{}
	bigBody := make([]byte, protocol.BodyLenLimit)
	err := protocol.Send(&buf, header, bigBody)
	assert.NotNil(err, "Check invalid body size")

	buf.Reset()
	var r *protocol.RespPkgHeader
	err = protocol.Send(&buf, r, bigBody)
	assert.NotNil(err, "Check nil pkg")

	buf.Reset()
	_, _, err = protocol.Recv(&buf)
	assert.NotNil(err, "empty body for recv")

	buf.WriteByte(1)
	_, _, err = protocol.Recv(&buf)
	assert.NotNil(err, "not enough headerfor recv")

	buf.Reset()
	buf.WriteByte(1)
	buf.WriteByte(1)
	_, _, err = protocol.Recv(&buf)
	assert.NotNil(err, "missing request ID recv")

	buf.Reset()
	for _, b := range []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x01, 0x03, 0xe8, 0x00, 0x00, 0x13, 0x0a, 0x11, 0x6f, 0x6e, 0x65, 0x20, 0x74, 0x69, 0x6d, 0x65, 0x20, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72} {
		buf.WriteByte(b)
	}
	_, _, err = protocol.Recv(&buf)
	assert.NotNil(err, "missing last byte causing body length unmatched.")

	buf.Reset()
	for _, b := range []byte{0x11, 0x02, 0x00, 0x00, 0x00, 0x01, 0x03, 0xe8, 0x00, 0x00, 0x13, 0x0a, 0x11, 0x6f, 0x6e, 0x65, 0x20, 0x74, 0x69, 0x6d, 0x65, 0x20, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x00} {
		buf.WriteByte(b)
	}
	_, _, err = protocol.Recv(&buf)
	assert.NotNil(err, "checking no signature")
}
