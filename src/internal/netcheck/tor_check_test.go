package netcheck

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type clientFunc func(request *http.Request) (*http.Response, error)

func (f clientFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestParseTorCheckResponse(t *testing.T) {
	for _, test := range []struct {
		name  string
		body  string
		isTor bool
	}{
		{name: "Tor", body: `{"IsTor":true,"IP":"203.0.113.1"}`, isTor: true},
		{name: "clearnet", body: `{"IsTor":false,"IP":"2001:db8::1"}`, isTor: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseTorCheckResponse(test.body)
			if err != nil {
				t.Fatal(err)
			}
			if result.IsTor != test.isTor {
				t.Fatalf("IsTor = %v, want %v", result.IsTor, test.isTor)
			}
		})
	}
}

func TestParseTorCheckResponseRejectsInvalidPayload(t *testing.T) {
	for _, body := range []string{
		`{"IP":"203.0.113.1"}`,
		`{"IsTor":true,"IP":"invalid"}`,
		`not-json`,
	} {
		if _, err := ParseTorCheckResponse(body); err == nil {
			t.Fatalf("ParseTorCheckResponse(%q) succeeded", body)
		}
	}
}

func TestCheckerUsesJSONEndpoint(t *testing.T) {
	checker := Checker{
		URL: "https://example.test/api/ip",
		Client: clientFunc(func(request *http.Request) (*http.Response, error) {
			if request.URL.String() != "https://example.test/api/ip" {
				t.Fatalf("URL = %q", request.URL)
			}
			if request.Header.Get("Accept") != "application/json" {
				t.Fatalf("Accept = %q", request.Header.Get("Accept"))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"IsTor":true,"IP":"203.0.113.1"}`)),
			}, nil
		}),
	}

	result, err := checker.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsTor || result.IP != "203.0.113.1" {
		t.Fatalf("result = %#v", result)
	}
}
