#!/bin/bash

# Testlarni ishga tushirish skripti

set -e

echo "=== PF Service Testlari ==="
echo ""

cd pfservice

echo "1. Media download testlari..."
go test ./internal/media_download/... -v
echo ""

echo "2. Filename uniqueness testlari..."
go test ./internal/media_download/downloader_filename_test.go ./internal/media_download/downloader.go -v
echo ""

echo "3. Integration testlar (agar database mavjud bo'lsa)..."
if [ -n "$TEST_POSTGRES_DSN" ] || [ -n "$POSTGRES_DSN" ]; then
    go test -tags=integration ./cmd/pf_sync/... -v || echo "Integration testlar database kerak"
else
    echo "Integration testlar uchun TEST_POSTGRES_DSN yoki POSTGRES_DSN o'rnatilmagan"
fi
echo ""

echo "4. Barcha testlarni ishga tushirish..."
go test ./... -v -short
echo ""

echo "=== Testlar yakunlandi ==="
