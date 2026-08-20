package netcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	TorCheckURL     = "https://check.torproject.org/api/ip"
	maxResponseSize = 64 << 10
)

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

type apiResponse struct {
	IP    string `json:"IP"`
	IsTor *bool  `json:"IsTor"`
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
		client = &http.Client{Timeout: 20 * time.Second}
	}
	url := c.URL
	if url == "" {
		url = TorCheckURL
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	request.Header.Set("Accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return Result{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Result{}, fmt.Errorf("Tor check returned HTTP %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return Result{}, err
	}
	if len(body) > maxResponseSize {
		return Result{}, fmt.Errorf("Tor check response exceeds %d bytes", maxResponseSize)
	}
	return ParseTorCheckResponse(string(body))
}

func ParseTorCheckResponse(body string) (Result, error) {
	var response apiResponse
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return Result{}, fmt.Errorf("decode Tor check response: %w", err)
	}
	ip := strings.TrimSpace(response.IP)
	if net.ParseIP(ip) == nil {
		return Result{}, fmt.Errorf("Tor check response has invalid public IP %q", response.IP)
	}
	if response.IsTor == nil {
		return Result{}, fmt.Errorf("Tor check response does not include Tor status")
	}
	return Result{IP: ip, IsTor: *response.IsTor}, nil
}
