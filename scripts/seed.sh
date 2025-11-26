#!/usr/bin/env bash
set -euo pipefail

# Usage: PINECONE_API_KEY=... PINECONE_INDEX_NAME=... OPENAI_API_KEY=... OPEN_TRIP_MAP_KEY=... ./scripts/seed.sh

export SEED_INDEX=true

go run ./main.go
