#!/bin/bash
set -e

IMAGE="hub.sixtyfive.me/transfer-shortener"
TAG="${1:-latest}"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "Building ${IMAGE}:${TAG} for linux/amd64..."
echo "Commit: ${COMMIT}, Build time: ${BUILD_TIME}"

docker buildx build --platform linux/amd64 \
  --build-arg COMMIT="${COMMIT}" \
  --build-arg BUILD_TIME="${BUILD_TIME}" \
  -t "${IMAGE}:${TAG}" --push .

echo "Done! watch-cluster will pick up the new image automatically."
