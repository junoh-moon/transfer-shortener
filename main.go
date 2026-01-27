package main

import (
	"log"
	"net/http"
	"os"

	httpAdapter "transfer-shortener/adapter/http"
	"transfer-shortener/adapter/sqlite"
	"transfer-shortener/usecase"
)

func main() {
	config := loadConfig()

	repo, err := sqlite.NewRepository(config.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repo.Close()

	createUC := usecase.NewCreateShortURL(repo)
	resolveUC := usecase.NewResolveShortURL(repo)
	proxy := httpAdapter.NewTransferProxy(config.BackendURL)

	handler := httpAdapter.NewHandler(createUC, resolveUC, proxy, config.PublicURL)

	log.Printf("Starting server on %s", config.ListenAddr)
	log.Printf("Backend: %s", config.BackendURL)
	log.Printf("Public URL: %s", config.PublicURL)

	if err := http.ListenAndServe(config.ListenAddr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type Config struct {
	ListenAddr string
	BackendURL string
	PublicURL  string
	DBPath     string
}

func loadConfig() Config {
	return Config{
		ListenAddr: getEnv("LISTEN_ADDR", ":8080"),
		BackendURL: getEnv("BACKEND_URL", "http://transfer:5327"),
		PublicURL:  getEnv("PUBLIC_URL", "https://t.sixtyfive.me"),
		DBPath:     getEnv("DB_PATH", "/data/shortener.db"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
