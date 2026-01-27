package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"transfer-shortener/domain/entity"
)

type CreateShortURLUseCase interface {
	Execute(ctx context.Context, fullURL string) (*entity.ShortURL, error)
}

type ResolveShortURLUseCase interface {
	Execute(ctx context.Context, token string) (string, error)
}

type BackendProxy interface {
	ProxyUpload(w http.ResponseWriter, r *http.Request) (string, error)
	ProxyGet(w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	createUC  CreateShortURLUseCase
	resolveUC ResolveShortURLUseCase
	proxy     BackendProxy
	publicURL string
}

func NewHandler(
	createUC CreateShortURLUseCase,
	resolveUC ResolveShortURLUseCase,
	proxy BackendProxy,
	publicURL string,
) *Handler {
	return &Handler{
		createUC:  createUC,
		resolveUC: resolveUC,
		proxy:     proxy,
		publicURL: publicURL,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/health":
		h.handleHealth(w, r)
	case r.Method == http.MethodPut || r.Method == http.MethodPost:
		h.handleUpload(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/":
		h.handleIndex(w, r)
	case r.Method == http.MethodGet:
		h.handleGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	fullURL, err := h.proxy.ProxyUpload(w, r)
	if err != nil {
		log.Printf("proxy error: %v", err)
		http.Error(w, "Backend error", http.StatusBadGateway)
		return
	}

	shortURL, err := h.createUC.Execute(r.Context(), fullURL)
	if err != nil {
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	result := fmt.Sprintf("%s/%s\n", h.publicURL, shortURL.Token)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	// If path contains slash (e.g., "abc12/file.txt"), proxy to backend
	if strings.Contains(path, "/") {
		h.proxy.ProxyGet(w, r)
		return
	}

	// Try to resolve as short token
	fullURL, err := h.resolveUC.Execute(r.Context(), path)
	if err != nil {
		// Not a short token, proxy to backend
		h.proxy.ProxyGet(w, r)
		return
	}

	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Transfer Shortener\n\n")
	fmt.Fprintf(w, "Upload: curl --upload-file ./file.txt %s/file.txt\n", h.publicURL)
	fmt.Fprintf(w, "Or:     curl -F filedata=@./file.txt %s/\n", h.publicURL)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
