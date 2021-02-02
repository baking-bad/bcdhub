package noderpc

import "net/http"

type testClient struct {
	handler func(*http.Request) (*http.Response, error)
}

// Do -
func (c *testClient) Do(req *http.Request) (*http.Response, error) {
	return c.handler(req)
}
