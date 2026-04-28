package client

import "net/http"

type Client struct {
	endpoint   string
	userAgent  string
	httpClient *http.Client
}

// NewClient создает новый HTTP-клиент с заданными параметрами.
func NewClient(config *Config, roundTripper http.RoundTripper) *Client {
	return &Client{
		endpoint:  config.EndpointUrl,
		userAgent: config.UserAgent,
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: roundTripper,
		},
	}
}

func (c *Client) HttpClient() *http.Client {
	return c.httpClient
}
