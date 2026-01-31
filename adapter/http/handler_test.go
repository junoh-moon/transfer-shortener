package http_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"transfer-shortener/domain/entity"
	handler "transfer-shortener/adapter/http"
)

type mockCreateShortURL struct {
	executeFunc func(ctx context.Context, fullURL string) (*entity.ShortURL, error)
}

func (m *mockCreateShortURL) Execute(ctx context.Context, fullURL string) (*entity.ShortURL, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, fullURL)
	}
	return nil, errors.New("not implemented")
}

type mockResolveShortURL struct {
	executeFunc func(ctx context.Context, token string) (string, error)
}

func (m *mockResolveShortURL) Execute(ctx context.Context, token string) (string, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, token)
	}
	return "", errors.New("not implemented")
}

type mockBackendProxy struct {
	proxyUploadFunc func(w http.ResponseWriter, r *http.Request) (string, error)
	proxyGetFunc    func(w http.ResponseWriter, r *http.Request)
}

func (m *mockBackendProxy) ProxyUpload(w http.ResponseWriter, r *http.Request) (string, error) {
	if m.proxyUploadFunc != nil {
		return m.proxyUploadFunc(w, r)
	}
	return "", errors.New("not implemented")
}

func (m *mockBackendProxy) ProxyGet(w http.ResponseWriter, r *http.Request) {
	if m.proxyGetFunc != nil {
		m.proxyGetFunc(w, r)
	}
}

func TestHandler_Upload_PUT_Success(t *testing.T) {
	backendURL := "https://transfer.sixtyfive.me/abc12/file.txt"

	createUC := &mockCreateShortURL{
		executeFunc: func(ctx context.Context, fullURL string) (*entity.ShortURL, error) {
			return &entity.ShortURL{
				Token:     "xyz1",
				FullURL:   fullURL,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{
		proxyUploadFunc: func(w http.ResponseWriter, r *http.Request) (string, error) {
			return backendURL, nil
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodPut, "/file.txt", strings.NewReader("file content"))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	expected := "https://transfer.sixtyfive.me/xyz1\n"
	if string(body) != expected {
		t.Errorf("expected body %q, got %q", expected, string(body))
	}
}

func TestHandler_Upload_POST_Success(t *testing.T) {
	backendURL := "https://transfer.sixtyfive.me/abc12/file.txt"

	createUC := &mockCreateShortURL{
		executeFunc: func(ctx context.Context, fullURL string) (*entity.ShortURL, error) {
			return &entity.ShortURL{
				Token:     "xyz1",
				FullURL:   fullURL,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{
		proxyUploadFunc: func(w http.ResponseWriter, r *http.Request) (string, error) {
			return backendURL, nil
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("file content"))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHandler_Redirect_Success(t *testing.T) {
	fullURL := "https://transfer.sixtyfive.me/abc12/file.txt"

	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{
		executeFunc: func(ctx context.Context, token string) (string, error) {
			if token == "xyz1" {
				return fullURL, nil
			}
			return "", errors.New("not found")
		},
	}
	proxy := &mockBackendProxy{}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/xyz1", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != fullURL {
		t.Errorf("expected Location %s, got %s", fullURL, location)
	}
}

func TestHandler_UnknownToken_ProxiesToBackend(t *testing.T) {
	// If short token not found, proxy to backend (might be a full URL token)
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{
		executeFunc: func(ctx context.Context, token string) (string, error) {
			return "", errors.New("not found")
		},
	}

	var proxyCalled bool
	proxy := &mockBackendProxy{
		proxyGetFunc: func(w http.ResponseWriter, r *http.Request) {
			proxyCalled = true
			w.WriteHeader(http.StatusOK)
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/abc12", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if !proxyCalled {
		t.Error("expected proxy to be called for unknown token")
	}
}

func TestHandler_Index(t *testing.T) {
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHandler_Health(t *testing.T) {
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHandler_FullPath_ProxiesToBackend(t *testing.T) {
	// GET /{token}/{filename} should proxy to backend (not redirect)
	// Start a mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("file content from backend"))
	}))
	defer backend.Close()

	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{
		executeFunc: func(ctx context.Context, token string) (string, error) {
			return "", errors.New("not found") // short token not found
		},
	}
	proxy := handler.NewTransferProxy(backend.URL, "https://transfer.sixtyfive.me")

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/abc12/file.txt", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "file content from backend" {
		t.Errorf("expected body from backend, got %q", string(body))
	}
}

func TestHandler_ShortToken_Redirects(t *testing.T) {
	// GET /{short} (4 chars) should redirect, not proxy
	fullURL := "https://transfer.sixtyfive.me/abc12/file.txt"

	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{
		executeFunc: func(ctx context.Context, token string) (string, error) {
			if token == "xyz1" {
				return fullURL, nil
			}
			return "", errors.New("not found")
		},
	}
	proxy := &mockBackendProxy{}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/xyz1", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != fullURL {
		t.Errorf("expected Location %s, got %s", fullURL, location)
	}
}

func TestHandler_Index_BrowserRequest_ProxiesToBackend(t *testing.T) {
	// Browser requests (Accept: text/html) should proxy to backend for web frontend
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}

	var proxyCalled bool
	proxy := &mockBackendProxy{
		proxyGetFunc: func(w http.ResponseWriter, r *http.Request) {
			proxyCalled = true
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html>Web Frontend</html>"))
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if !proxyCalled {
		t.Error("expected proxy to be called for browser request with Accept: text/html")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHandler_Index_CLIRequest_ReturnsUsageText(t *testing.T) {
	// CLI requests (no Accept header or Accept: */*) should return usage text
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}

	var proxyCalled bool
	proxy := &mockBackendProxy{
		proxyGetFunc: func(w http.ResponseWriter, r *http.Request) {
			proxyCalled = true
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	// Test with no Accept header (like curl default)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if proxyCalled {
		t.Error("proxy should NOT be called for CLI request without Accept: text/html")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "curl") {
		t.Errorf("expected usage text with curl example, got %q", string(body))
	}
}

func TestHandler_Index_CLIRequestWithWildcard_ReturnsUsageText(t *testing.T) {
	// CLI with Accept: */* should return usage text, not proxy
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}

	var proxyCalled bool
	proxy := &mockBackendProxy{
		proxyGetFunc: func(w http.ResponseWriter, r *http.Request) {
			proxyCalled = true
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "*/*")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if proxyCalled {
		t.Error("proxy should NOT be called for CLI request with Accept: */*")
	}
}

func TestHandler_Index_VaryHeader(t *testing.T) {
	// All index responses should include Vary: Accept header for CDN caching
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{
		proxyGetFunc: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	tests := []struct {
		name   string
		accept string
	}{
		{"browser request", "text/html"},
		{"CLI request", ""},
		{"wildcard request", "*/*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			vary := rec.Header().Get("Vary")
			if !strings.Contains(vary, "Accept") {
				t.Errorf("expected Vary header to contain 'Accept', got %q", vary)
			}
		})
	}
}

func TestHandler_Upload_ProxyError_ReturnsBadGateway(t *testing.T) {
	// When proxy fails, should return 502 Bad Gateway
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{
		proxyUploadFunc: func(w http.ResponseWriter, r *http.Request) (string, error) {
			return "", errors.New("backend connection failed")
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodPut, "/file.txt", strings.NewReader("content"))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandler_Upload_CreateShortURLError_ReturnsInternalError(t *testing.T) {
	// When creating short URL fails, should return 500 Internal Server Error
	createUC := &mockCreateShortURL{
		executeFunc: func(ctx context.Context, fullURL string) (*entity.ShortURL, error) {
			return nil, errors.New("database error")
		},
	}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{
		proxyUploadFunc: func(w http.ResponseWriter, r *http.Request) (string, error) {
			return "https://transfer.sixtyfive.me/abc12/file.txt", nil
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodPut, "/file.txt", strings.NewReader("content"))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	// Unsupported methods should return 405 Method Not Allowed
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://transfer.sixtyfive.me")

	methods := []string{http.MethodDelete, http.MethodPatch, http.MethodOptions}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/file.txt", nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405 for %s, got %d", method, rec.Code)
			}
		})
	}
}

func TestTransferProxy_ProxyUpload_Success(t *testing.T) {
	// Test real proxy upload with mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/file.txt" {
			t.Errorf("expected path /file.txt, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("http://backend:5327/abc12/file.txt\n"))
	}))
	defer backend.Close()

	proxy := handler.NewTransferProxy(backend.URL, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodPut, "/file.txt", strings.NewReader("file content"))
	rec := httptest.NewRecorder()

	fullURL, err := proxy.ProxyUpload(rec, req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should transform backend URL to public URL
	expected := "https://transfer.sixtyfive.me/abc12/file.txt"
	if fullURL != expected {
		t.Errorf("expected %s, got %s", expected, fullURL)
	}
}

func TestTransferProxy_ProxyUpload_BackendError(t *testing.T) {
	// Test proxy when backend returns error
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer backend.Close()

	proxy := handler.NewTransferProxy(backend.URL, "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodPut, "/file.txt", strings.NewReader("content"))
	rec := httptest.NewRecorder()

	_, err := proxy.ProxyUpload(rec, req)

	if err == nil {
		t.Error("expected error when backend returns 500")
	}
}

func TestTransferProxy_ProxyGet_BackendError(t *testing.T) {
	// Test proxy GET when backend connection fails
	proxy := handler.NewTransferProxy("http://localhost:1", "https://transfer.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/abc12/file.txt", nil)
	rec := httptest.NewRecorder()

	proxy.ProxyGet(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}
