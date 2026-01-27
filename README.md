# transfer-shortener

URL shortener proxy for [transfer.sh](https://github.com/dutchcoders/transfer.sh) service.

## Features

- Shortens transfer.sh URLs from `https://host/token/filename` to `https://host/short`
- Supports PUT and POST (multipart) uploads
- SQLite storage for URL mappings
- 4-character random tokens (16M+ combinations)

## Usage

```bash
# PUT upload
curl --upload-file ./file.txt https://transfer.sixtyfive.me/file.txt
# Returns: https://transfer.sixtyfive.me/x0pe

# POST multipart upload
curl -F "file=@./file.txt" https://transfer.sixtyfive.me/
# Returns: https://transfer.sixtyfive.me/p7WQ

# Access shortened URL (redirects to full URL)
curl -L https://transfer.sixtyfive.me/x0pe
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_ADDR` | `:8080` | Server listen address |
| `BACKEND_URL` | `http://transfer:5327` | Backend transfer.sh URL |
| `PUBLIC_URL` | `https://transfer.sixtyfive.me` | Public-facing URL |
| `DB_PATH` | `/data/shortener.db` | SQLite database path |

## Build

```bash
# Local build
go build -o shortener .

# Docker build (linux/amd64)
./build.sh
```

## Deployment

Kubernetes manifests in `k8s/` directory:

```bash
kubectl apply -f k8s/
```

## Architecture

```
Client → Ingress → transfer-shortener → transfer.sh backend
                         ↓
                    SQLite (token → full URL)
```

## License

MIT
