#!/bin/bash
set -e

DEPLOY_DIR="deploy"
VPS="${1:?Usage: ./deploy.sh user@host}"

echo "==> Building Go binary (linux/amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$DEPLOY_DIR/api-server" ./cmd/api

echo "==> Building Expo web..."
cd mobile
EXPO_PUBLIC_API_URL=/api/v1 npx expo export --platform web
rm -rf "../$DEPLOY_DIR/web"
cp -r dist "../$DEPLOY_DIR/web"
cd ..

echo "==> Uploading deploy/ to $VPS..."
rsync -avz --exclude='node_modules' "$DEPLOY_DIR/" "$VPS:~/ai-calories/"

echo "==> Starting on server..."
ssh "$VPS" 'cd ~/ai-calories && docker compose up -d --build'

echo "==> Done!"
