package netcheck

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const TorCheckURL = "https://check.torproject.org/"

var strongTagPattern = regexp.MustCompile(`(?is)<strong>\s*([^<]+)\s*</strong>`)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Result struct {
	IP    string
	IsTor bool
}

type Checker struct {
	Client HTTPClient
	URL    string
}

func NewChecker() Checker {
	return Checker{
		Client: &http.Client{Timeout: 20 * time.Second},
		URL:    TorCheckURL,
	}
}

func (c Checker) Check(ctx context.Context) (Result, error) {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	url := c.URL
	if url == "" {
		url = TorCheckURL
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}

	response, err := client.Do(request)
	if err != nil {
		return Result{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Result{}, fmt.Errorf("Tor check returned HTTP %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return Result{}, err
	}
	return ParseTorCheckResponse(string(body))
}

func ParseTorCheckResponse(body string) (Result, error) {
	ip := parseStrongText(body)
	if ip == "" {
		return Result{}, fmt.Errorf("Tor check response does not include a public IP")
	}

	switch {
	case strings.Contains(body, "Congratulations. This browser is configured to use Tor."):
		return Result{IP: ip, IsTor: true}, nil
	case strings.Contains(body, "Sorry. You are not using Tor."):
		return Result{IP: ip, IsTor: false}, nil
	default:
		return Result{}, fmt.Errorf("Tor check response does not include a known Tor status")
	}
}

func parseStrongText(body string) string {
	match := strongTagPattern.FindStringSubmatch(body)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(html.UnescapeString(match[1]))
}
