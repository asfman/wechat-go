package wxweb

import (
	"net/http/cookiejar"
	"net/http"
	"io/ioutil"
	"net"
  "time"
	"bytes"
)

type Client struct {
	httpClient *http.Client
	userAgent string
}

type Header map[string]string

func NewClient(userAgent string) *Client {
	var netTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial:  (&net.Dialer{Timeout: 100 * time.Second}).Dial,
	}

	cookieJar, _ := cookiejar.New(nil)
  return &Client{
    httpClient: &http.Client{
      Jar:     cookieJar,
      Transport: netTransport,
    },
    userAgent: userAgent,
  }
}

func (c *Client) SetJar(jar http.CookieJar) {
	c.httpClient.Jar = jar
}

func (c *Client) fetchResponse(method string, uri string, body []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
  if c.userAgent != "" {
    req.Header.Set("User-Agent", c.userAgent)
  }
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.httpClient.Do(req)
}

func (c *Client) Get(url string) ([]byte, error) {
  res, err := c.fetchResponse("GET", url, nil, Header{})
	if err != nil {
		return nil, err
	}
	body, _:= ioutil.ReadAll(res.Body)
	defer res.Body.Close()
  return body, nil
}

func (c *Client) PostJsonBytes(url string, json []byte) ([]byte, error) {
  res, err := c.fetchResponse("POST", url, json, Header{ "Content-Type": "application/json; charset=UTF-8" })
	if err != nil {
		return nil, err
	}
	body, _:= ioutil.ReadAll(res.Body)
	defer res.Body.Close()
  return body, nil
}

