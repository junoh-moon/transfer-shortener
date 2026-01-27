package http

import (
	"io"
	"net/http"
	"strings"
	"time"
)

type TransferProxy struct {
	backendURL string
	client     *http.Client
}

func NewTransferProxy(backendURL string) *TransferProxy {
	return &TransferProxy{
		backendURL: backendURL,
		client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (p *TransferProxy) ProxyUpload(w http.ResponseWriter, r *http.Request) (string, error) {
	targetURL := p.backendURL + r.URL.Path

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		return "", err
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fullURL := strings.TrimSpace(string(body))
	return fullURL, nil
}
