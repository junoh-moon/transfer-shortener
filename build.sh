#!/bin/bash
set -e

IMAGE="hub.sixtyfive.me/transfer-shortener"
TAG="${1:-latest}"

echo "Building ${IMAGE}:${TAG}..."
docker build -t "${IMAGE}:${TAG}" .

echo "Pushing ${IMAGE}:${TAG}..."
docker push "${IMAGE}:${TAG}"

echo "Done! watch-cluster will pick up the new image automatically."
