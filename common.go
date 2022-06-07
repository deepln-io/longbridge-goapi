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
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

type httpMethod string

const (
	httpGET    httpMethod = "GET"
	httpPOST   httpMethod = "POST"
	httpPUT    httpMethod = "PUT"
	httpDELETE httpMethod = "DELETE"
)

// Symbol is a stock symbol (of format <StockCode>.<Market>) or fund symbol (ISIN format)
type Symbol string

// ToStock convert symbol into stock code and market.
func (s Symbol) ToStock() (string, Market, error) {
	parts := strings.Split(string(s), ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%v is not of format <StockCode>.<Market>", s)
	}
	return parts[0], Market(strings.ToUpper(parts[1])), nil
}

type urlPath string

const (
	version               = "/v1"
	getOTPURLPath urlPath = version + "/socket/token"
	refreshToken  urlPath = version + "/token/refresh"

	HK Market = "HK"
	US Market = "US"
)

func signature(doc []byte, secret string) string {
	sum := sha1.Sum(doc)
	payload := "HMAC-SHA256|" + hex.EncodeToString(sum[:])
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *client) request(method httpMethod, path urlPath, params url.Values, body []byte, reply interface{}) error {
	c.limiter.Wait(context.Background())
	req, err := c.createSignedRequest(method, path, params, body, time.Now())
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error calling long bridge API: %v", err)
	}
	defer resp.Body.Close()
	return decodeResp(resp.Body, reply)
}

func decodeResp(reader io.Reader, reply interface{}) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return fmt.Errorf("cannot read response: %v", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	if err := decoder.Decode(reply); err != nil {
		return fmt.Errorf("error decoding the JSON data in response body '%v': %v", buf.String(), err)
	}
	return nil
}

// createSignedRequest creates the signed request for longbridge API. It is a low level API.
// The signing method is described in https://open.longbridgeapp.com/docs/how-to-access-api.
// The data can be map from string to string or []string for long bridge API.
// For GET/DELETE the data is encoded in query. For POST/PUT the data is marshalled into JSON and keep in the body.
func (c *client) createSignedRequest(method httpMethod, urlPath urlPath, params url.Values, body []byte, tm time.Time) (*http.Request, error) {
	switch method {
	case httpGET, httpDELETE:
		if len(body) > 0 {
			glog.Warningf("http method %q should have empty body, got (%d), body discarded", method, len(body))
			body = nil
		}
	case httpPOST, httpPUT:
	default:
		return nil, fmt.Errorf("unsupported method %q for longbridge API", method)
	}

	var br io.Reader
	if len(body) > 0 {
		br = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(string(method), c.config.BaseURL+string(urlPath), br)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	ts := fmt.Sprintf("%.3f", float64(tm.UnixMilli())/1e3)
	req.Header.Add("X-Api-Key", c.config.AppKey)
	req.Header.Add("Authorization", c.config.AccessToken)
	req.Header.Add("X-Timestamp", ts)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	q := req.URL.Query()
	for k, vals := range params {
		for _, v := range vals {
			q.Add(k, v)
		}
	}
	req.URL.RawQuery = q.Encode()

	var buf bytes.Buffer
	buf.WriteString(string(method))
	buf.WriteString("|")
	buf.WriteString(string(urlPath))
	buf.WriteString("|")
	buf.WriteString(req.URL.RawQuery)
	buf.WriteString("|authorization:")
	buf.WriteString(c.config.AccessToken)
	buf.WriteString("\nx-api-key:")
	buf.WriteString(c.config.AppKey)
	buf.WriteString("\nx-timestamp:")
	buf.WriteString(ts)
	buf.WriteString("\n|authorization;x-api-key;x-timestamp|")
	if len(body) > 0 {
		sum := sha1.Sum(body)
		buf.WriteString(hex.EncodeToString(sum[:]))
	}
	req.Header.Add("X-Api-Signature", "HMAC-SHA256 SignedHeaders=authorization;x-api-key;x-timestamp, Signature="+
		signature(buf.Bytes(), c.config.AppSecret))
	return req, nil
}

type statusResp struct {
	Code    int
	Message string
}

func (r *statusResp) CheckSuccess() error {
	if r.Code != 0 {
		return fmt.Errorf("response %+v is not successful", r)
	}
	return nil
}

type joinErrors struct {
	errs []string
}

// Add adds error to errors. Nil error will be skipped.
func (errs *joinErrors) Add(name string, err interface{}) {
	if err != nil {
		errs.errs = append(errs.errs, fmt.Sprintf("%s: %v", name, err))
	}
}

// ToError returns error from joined error messages by new line.
func (errs *joinErrors) ToError() error {
	if errs.errs == nil {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(errs.errs, "\n"))
}

type parser struct {
	errs joinErrors
}

func (p *parser) parseInt(fieldName, fieldText string) int64 {
	val, err := strconv.ParseInt(fieldText, 10, 64)
	if err != nil {
		p.errs.Add(fieldName, err)
		return 0
	}
	return val
}

func (p *parser) parseFloat(fieldName, fieldText string) float64 {
	if fieldText == "" {
		return 0
	}
	val, err := strconv.ParseFloat(fieldText, 64)
	if err != nil {
		p.errs.Add(fieldName, err)
		return 0
	}
	return val
}

func (p *parser) parse(fieldName, fieldText string, target interface{}) {
	switch v := target.(type) {
	case *float64:
		*v = p.parseFloat(fieldName, fieldText)
	case *uint64:
		val, err := strconv.ParseUint(fieldText, 10, 64)
		if err != nil {
			p.errs.Add(fieldName, err)
			return
		}
		*v = val
	case *int64:
		*v = p.parseInt(fieldName, fieldText)
	default:
		p.errs.Add(fieldName, fmt.Errorf("unsupported type %T, should be one of: *int64, *uint64, and *float64", target))
	}
}

func (p *parser) Error() error {
	return p.errs.ToError()
}

type params map[string]string

func (p params) Add(key string, val string) {
	if len(val) > 0 {
		p[key] = val
	}
}

func (p params) AddInt(key string, val int64) {
	p[key] = strconv.FormatInt(val, 10)
}

// float value = 0 means optional
func (p params) AddOptFloat(key string, val float64) {
	if math.Abs(val) >= 1e-6 {
		p[key] = strconv.FormatFloat(val, 'f', -1, 64)
	}
}

func (p params) AddDate(key string, val time.Time) {
	if !val.IsZero() {
		p[key] = val.Format("2006-01-02")
	}
}

// integer value = 0 means optional
func (p params) AddOptInt(key string, val int64) {
	if val != 0 {
		p.AddInt(key, val)
	}
}

func (p params) AddOptUint(key string, val uint64) {
	if val != 0 {
		p[key] = strconv.FormatUint(val, 10)
	}
}

func (p params) Values() url.Values {
	vals := url.Values{}
	for k, v := range p {
		vals.Add(k, v)
	}
	return vals
}

func trace(msg string) func() {
	start := time.Now()
	glog.V(4).Infof("Enter %s", msg)
	return func() { glog.V(4).Infof("Exit %s (%s)", msg, time.Since(start)) }
}
