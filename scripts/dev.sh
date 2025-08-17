#!/usr/bin/env bash
set -euo pipefail
export $(grep -v '^#' .env.example | xargs) || true
go run ./cmd/server
