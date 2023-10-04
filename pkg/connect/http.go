package connect

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type HttpMethod int

const (
	Undefined HttpMethod = iota
	Get
	Post
	Head
	Put
	Delete
	Connect
	Options
	Trace
	Patch
)

const defaultTimeout = 5 * time.Second

type BuilderFunc = func(httpConfig *HttpConfig) (*http.Request, error)

type Authenticator interface {
	Authenticate(httpConfig *HttpConfig, req *http.Request) error
}

type HttpConfig struct {
	Method        HttpMethod
	Url           string
	Body          any
	Headers       map[string]string
	Builder       BuilderFunc
	Authenticator Authenticator
	Timeout       time.Duration
}

type HttpConnector interface {
	Connector

	SetReqBuilder(builder BuilderFunc)
	SetAuthenticator(authenticator Authenticator)
	SetMethod(method HttpMethod)
	SetUrl(url string)
	SetBody(body any)
	SetHeaders(headers map[string]string)
	SetTimeout(timeout time.Duration)
}

type httpConnector struct {
	Config HttpConfig
}

func (c *httpConnector) SetConfig(config any) {
	conf, ok := config.(HttpConfig)
	if !ok {
		panic("Expected an HttpConfig struct.")
	}
	c.Config = conf
}

func (c *httpConnector) Request() ([]byte, error) {
	if c.Config.Builder != nil {
		req, err := c.Config.Builder(&c.Config)
		if err != nil {
			return nil, MakeConnectErr(err, 0)
		}

		res, reqErr := c.httpRequest(req, c.Config.Timeout)
		return []byte(res), reqErr
	}

	switch c.Config.Method {
	case Get:
		res, err := c.httpGet()
		return []byte(res), err

	case Post:
		switch body := c.Config.Body.(type) {
		case io.Reader:
			res, err := c.httpPost(body)
			return []byte(res), err
		case string:
			res, err := c.httpPost(strings.NewReader(body))
			return []byte(res), err
		case []byte:
			res, err := c.httpPost(bytes.NewReader(body))
			return []byte(res), err
		default:
			return nil, MakeConnectErr(errors.New("unsupported type for body"), 0)
		}

	default:
		return nil, MakeConnectErr(errors.New("unsupported HTTP method"), 0)
	}
}

func (c *httpConnector) ConnectorID() string {
	return "HTTP"
}

func (c *httpConnector) SetReqBuilder(builder BuilderFunc) {
	c.Config.Builder = builder
}

func (c *httpConnector) SetAuthenticator(authenticator Authenticator) {
	c.Config.Authenticator = authenticator
}

func (c *httpConnector) SetMethod(method HttpMethod) {
	c.Config.Method = method
}

func (c *httpConnector) SetUrl(url string) {
	c.Config.Url = url
}

func (c *httpConnector) SetBody(body any) {
	c.Config.Body = body
}

func (c *httpConnector) SetHeaders(headers map[string]string) {
	c.Config.Headers = headers
}

func (c *httpConnector) SetTimeout(timeout time.Duration) {
	c.Config.Timeout = timeout
}

// Build and Connector for HTTP GET requests.
func MakeHttpGetConnector(url string, headers map[string]string) HttpConnector {
	return &httpConnector{
		Config: HttpConfig{
			Method:  Get,
			Url:     url,
			Headers: headers,
			Timeout: defaultTimeout,
		},
	}
}

// Build and Connector for HTTP POST requests.
func MakeHttpPostConnector(url string, body any, headers map[string]string) HttpConnector {
	return &httpConnector{
		Config: HttpConfig{
			Method:  Post,
			Url:     url,
			Body:    body,
			Headers: headers,
			Timeout: defaultTimeout,
		},
	}
}

// Build and Connector for HTTP GET requests.
func MakeHttpConnectorWithBuilder(builder BuilderFunc) HttpConnector {
	return &httpConnector{
		Config: HttpConfig{
			Builder: builder,
			Timeout: defaultTimeout,
		},
	}
}

func (c *httpConnector) httpGet() (string, error) {
	req, err := http.NewRequest("GET", c.Config.Url, nil)
	if err != nil {
		return "", MakeConnectErr(err, 0)
	}

	for key, val := range c.Config.Headers {
		req.Header.Add(key, val)
	}

	if c.Config.Authenticator != nil {
		err = c.Config.Authenticator.Authenticate(&c.Config, req)
		if err != nil {
			return "", MakeConnectErr(err, 0)
		}
	}

	client := http.DefaultClient
	client.Timeout = c.Config.Timeout
	resp, err := client.Do(req)

	if err != nil {
		if resp != nil {
			return "", MakeConnectErr(err, resp.StatusCode)
		} else {
			return "", MakeConnectErr(err, 0)
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", MakeConnectErr(err, resp.StatusCode)
	}

	return string(body), nil
}

func (c *httpConnector) httpPost(reqBody io.Reader) (string, error) {
	req, err := http.NewRequest("POST", c.Config.Url, reqBody)
	if err != nil {
		return "", MakeConnectErr(err, 0)
	}

	for key, val := range c.Config.Headers {
		req.Header.Add(key, val)
	}

	if c.Config.Authenticator != nil {
		err = c.Config.Authenticator.Authenticate(&c.Config, req)
		if err != nil {
			return "", MakeConnectErr(err, 0)
		}
	}

	client := http.DefaultClient
	client.Timeout = c.Config.Timeout
	resp, err := client.Do(req)

	if err != nil {
		return "", MakeConnectErr(err, 0)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", MakeConnectErr(err, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", MakeConnectErr(err, resp.StatusCode)
	}

	return string(body), nil
}

func (c *httpConnector) httpRequest(req *http.Request, timeout time.Duration) (string, error) {
	for key, val := range c.Config.Headers {
		req.Header.Add(key, val)
	}

	if c.Config.Authenticator != nil {
		err := c.Config.Authenticator.Authenticate(&c.Config, req)
		if err != nil {
			return "", MakeConnectErr(err, 0)
		}
	}

	client := http.DefaultClient
	client.Timeout = timeout
	resp, err := client.Do(req)

	if err != nil {
		if resp != nil {
			return "", MakeConnectErr(err, resp.StatusCode)
		} else {
			return "", MakeConnectErr(err, 0)
		}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", MakeConnectErr(err, resp.StatusCode)
	}

	return string(body), nil
}
