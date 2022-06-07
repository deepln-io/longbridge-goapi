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
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateSignedRequest(t *testing.T) {
	assert := assert.New(t)
	c, err := newClient(&Config{
		BaseURL:     "https://openapi.longbridgeapp.com",
		AccessToken: "fake access token",
		AppKey:      "fake App key",
		AppSecret:   "fake secret",
	})
	if err != nil {
		t.Fatalf("Error creating longbridge client: %v", err)
	}
	hkt, _ := time.LoadLocation("Asia/Hong_Kong")
	req, err := c.createSignedRequest(httpGET, "/v1/trade/order/history",
		map[string][]string{"order id": {"0001"}, "type": {"limited"}}, nil, time.Date(2022, time.January, 2, 0, 0, 1, 2, hkt))
	assert.Nil(err, "Cheating signed GET request")
	assert.Equal("https://openapi.longbridgeapp.com/v1/trade/order/history?order+id=0001&type=limited", req.URL.String(), "Checking url")
	assert.Equal("HMAC-SHA256 SignedHeaders=authorization;x-api-key;x-timestamp, Signature=7971ae0992cc6509fbb9f9dff14686c07ed316e3366479ed84ec00e47d1c8515",
		req.Header.Get("X-Api-Signature"), "Checking GET signature")

	body := []byte(`{"symbol":"AAPL.US", "order_type":"LO", "trigger_price":"123.0","side":"Buy","submitted_quantity":"20"}`)
	req, err = c.createSignedRequest(httpPOST, "/v1/trade/order", nil, body,
		time.Date(2022, time.January, 2, 0, 0, 0, 0, hkt))
	assert.Nil(err, "Cheating signed POST request")
	assert.NotNil(req.Body, "Checking POST body")
	assert.Equal("https://openapi.longbridgeapp.com/v1/trade/order", req.URL.String(), "Checking url")
	assert.Equal("HMAC-SHA256 SignedHeaders=authorization;x-api-key;x-timestamp, Signature=6fb79bfc9384428ae8136882917001032c03094b0048a858616318004c3e6b4d",
		req.Header.Get("X-Api-Signature"), "Checking POST signature")
}

func ExampleTradeClient_Enable() {
	// Before running the example, set the access key, app key and app secret in the env. i.e.,
	// export LB_ACCESS_TOKEN="<your access token>"
	// export LB_APP_KEY="<your app key>"
	// export LB_APP_SECRET="<your app secret>"
	// Then run with command to see the detailed log.
	// go test -v -run ExampleTradeClient_Enable -args -logtostderr -v=4
	c, err := NewTradeClient(&Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	c.OnOrderChange = func(order *Order) {
		log.Printf("Got order: %#v\n", order)
	}
	c.Enable(true)
	time.Sleep(11 * time.Second)
	log.Println("Start to close subscription")
	c.Enable(false)
	time.Sleep(10 * time.Second)
	log.Println("Start to re-enable subscription")
	c.Enable(true)
	time.Sleep(15 * time.Second)
	log.Println("Closing")
	c.Close()
	// Output:
}

func Exampleclient_RefreshAccessToken() {
	// Before running the example, set the access key, app key and app secret in the env. i.e.,
	// export LB_ACCESS_TOKEN="<your access token>"
	// export LB_APP_KEY="<your app key>"
	// export LB_APP_SECRET="<your app secret>"
	// Then run with command to see the detailed log.
	// go test -v -run Exampleclient_RefreshAccessToken -args -logtostderr -v=4
	c, err := newClient(&Config{
		AccessToken: os.Getenv("LB_ACCESS_TOKEN"),
		AppKey:      os.Getenv("LB_APP_KEY"),
		AppSecret:   os.Getenv("LB_APP_SECRET"),
	})
	if err != nil {
		log.Fatalf("Error creating longbridge client: %v", err)
	}
	defer c.Close()
	token, err := c.RefreshAccessToken(time.Now().AddDate(0, 1, 0))
	if err != nil {
		log.Fatalf("Error refreshing access token: %v", err)
	}
	log.Printf("Access token: %#v", token)
	// Output:
}
