package m3u8

import "net/http"

// HeaderMapTransport implements custom header injection
type HeaderMapTransport struct {
	Headers map[string]string
	Base    http.RoundTripper
}

func (t *HeaderMapTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}
	return t.Base.RoundTrip(req)
}
