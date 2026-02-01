package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TransferProxy struct {
	backendURL string
	publicURL  string
	client     *http.Client
}

func NewTransferProxy(backendURL, publicURL string) *TransferProxy {
	return &TransferProxy{
		backendURL: backendURL,
		publicURL:  publicURL,
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

	req.ContentLength = r.ContentLength
	for key, values := range r.Header {
		if key == "Host" {
			continue
		}
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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(body))
	}

	// Transform internal backend URL to public URL
	publicParsed, _ := url.Parse(p.publicURL)
	returnedURL, err := url.Parse(strings.TrimSpace(string(body)))
	if err != nil {
		return "", err
	}
	returnedURL.Scheme = publicParsed.Scheme
	returnedURL.Host = publicParsed.Host
	return returnedURL.String(), nil
}

func (p *TransferProxy) ProxyGet(w http.ResponseWriter, r *http.Request) {
	targetURL := p.backendURL + r.URL.Path

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, targetURL, nil)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Set Host header to public URL so backend generates correct URLs
	publicParsed, _ := url.Parse(p.publicURL)
	req.Host = publicParsed.Host

	for key, values := range r.Header {
		if key == "Host" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		http.Error(w, "Backend error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
