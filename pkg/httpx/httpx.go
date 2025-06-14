package httpx

import (
	"crypto/tls"
	"fmt"
	"time"

	reqv3 "github.com/imroc/req/v3"
)

type HttpxClient interface {
	SetToken(token string)
	Get(url string, query map[string]string, header map[string]string, data any) error
	Put(url string, query map[string]string, header map[string]string, body any, data any) error
	Post(url string, query map[string]string, header map[string]string, body any, data any) error
	Delete(url string, query map[string]string, header map[string]string, data any) error
}

type httpxClient struct {
	client *reqv3.Client
	token  string
}

func NewHttpXClient(tlsFunc func() *tls.Config, token string) HttpxClient {
	httpC := reqv3.C().SetTimeout(20 * time.Second)
	if tlsFunc != nil {
		httpC.TLSClientConfig = &tls.Config{
			GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
				return tlsFunc(), nil
			},
		}
	} else {
		httpC.TLSClientConfig.InsecureSkipVerify = true
	}
	return &httpxClient{
		client: httpC,
		token:  token,
	}
}

func (hx *httpxClient) SetToken(token string) {
	if token != "" {
		hx.token = token
	}
}

func (hx *httpxClient) request() *reqv3.Request {
	if hx.token == "" {
		return hx.client.R().SetBearerAuthToken(hx.token)
	}
	return hx.client.R()
}

func (hx *httpxClient) Get(url string, query map[string]string, header map[string]string, data any) error {
	resp, err := hx.request().SetQueryParams(query).SetHeaders(header).SetSuccessResult(data).Get(url)
	if err != nil {
		return err
	}
	if resp.IsSuccessState() {
		return nil
	}
	respBody, err := resp.ToBytes()
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", respBody)

}

func (hx *httpxClient) Put(url string, query map[string]string, header map[string]string, body any, data any) error {
	resp, err := hx.request().SetQueryParams(query).SetHeaders(header).SetSuccessResult(data).SetBody(body).Put(url)
	if err != nil {
		return err
	}
	if resp.IsSuccessState() {
		return nil
	}
	respBody, err := resp.ToBytes()
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", respBody)
}

func (hx *httpxClient) Post(url string, query map[string]string, header map[string]string, body any, data any) error {
	resp, err := hx.request().SetQueryParams(query).SetHeaders(header).SetSuccessResult(data).SetBody(body).Post(url)
	if err != nil {
		return err
	}
	if resp.IsSuccessState() {
		return nil
	}
	respBody, err := resp.ToBytes()
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", respBody)
}

func (hx *httpxClient) Delete(url string, query map[string]string, header map[string]string, data any) error {
	resp, err := hx.request().SetQueryParams(query).SetHeaders(header).SetSuccessResult(data).Delete(url)
	if err != nil {
		return err
	}
	if resp.IsSuccessState() || resp.GetStatusCode() == 404 {
		return nil
	}
	respBody, err := resp.ToBytes()
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", respBody)
}
