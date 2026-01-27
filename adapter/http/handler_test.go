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
	proxyFunc func(w http.ResponseWriter, r *http.Request) (string, error)
}

func (m *mockBackendProxy) ProxyUpload(w http.ResponseWriter, r *http.Request) (string, error) {
	if m.proxyFunc != nil {
		return m.proxyFunc(w, r)
	}
	return "", errors.New("not implemented")
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
		proxyFunc: func(w http.ResponseWriter, r *http.Request) (string, error) {
			return backendURL, nil
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://t.sixtyfive.me")

	req := httptest.NewRequest(http.MethodPut, "/file.txt", strings.NewReader("file content"))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	expected := "https://t.sixtyfive.me/xyz1\n"
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
		proxyFunc: func(w http.ResponseWriter, r *http.Request) (string, error) {
			return backendURL, nil
		},
	}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://t.sixtyfive.me")

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

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://t.sixtyfive.me")

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

func TestHandler_Redirect_NotFound(t *testing.T) {
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{
		executeFunc: func(ctx context.Context, token string) (string, error) {
			return "", errors.New("not found")
		},
	}
	proxy := &mockBackendProxy{}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://t.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandler_Index(t *testing.T) {
	createUC := &mockCreateShortURL{}
	resolveUC := &mockResolveShortURL{}
	proxy := &mockBackendProxy{}

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://t.sixtyfive.me")

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

	h := handler.NewHandler(createUC, resolveUC, proxy, "https://t.sixtyfive.me")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}
