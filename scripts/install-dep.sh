#!/bin/sh

set -eo pipefail

go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

echo
echo "========================================="
echo "Dependencies installed!"
echo "========================================="
echo
echo "Installed:"
echo "  ✓ Make"
echo "  ✓ sqlc"
echo "  ✓ golang-migrate"