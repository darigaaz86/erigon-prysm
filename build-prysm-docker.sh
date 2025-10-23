#!/bin/bash
set -e

echo "🔨 Building Prysm from source using Docker..."
echo ""
echo "This builds Prysm binaries inside a Docker container for Linux compatibility"
echo ""

# Create a multi-stage Dockerfile that builds Prysm
cat > Dockerfile.prysm-builder <<'EOF'
# Stage 1: Build
FROM golang:1.25-bookworm AS builder

WORKDIR /build

# Copy Prysm source
COPY prysm/ ./

# Build beacon-chain
RUN cd cmd/beacon-chain && \
    CGO_ENABLED=1 go build -o /build/bin/beacon-chain .

# Build validator
RUN cd cmd/validator && \
    CGO_ENABLED=1 go build -o /build/bin/validator .

# Stage 2: Beacon Chain Runtime
FROM debian:bookworm-slim AS beacon-chain

RUN apt-get update && \
    apt-get install -y ca-certificates libssl3 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/bin/beacon-chain /usr/local/bin/beacon-chain

ENTRYPOINT ["/usr/local/bin/beacon-chain"]

# Stage 3: Validator Runtime
FROM debian:bookworm-slim AS validator

RUN apt-get update && \
    apt-get install -y ca-certificates libssl3 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/bin/validator /usr/local/bin/validator

ENTRYPOINT ["/usr/local/bin/validator"]
EOF

echo "📦 Building beacon-chain image..."
docker build --no-cache --target beacon-chain -t prysm-beacon-chain:local -f Dockerfile.prysm-builder .

echo ""
echo "📦 Building validator image..."
docker build --no-cache --target validator -t prysm-validator:local -f Dockerfile.prysm-builder .

# Cleanup
rm Dockerfile.prysm-builder

echo ""
echo "✅ Docker images created successfully!"
echo ""
echo "Images:"
echo "  - prysm-beacon-chain:local"
echo "  - prysm-validator:local"
echo ""
echo "Next steps:"
echo "  1. Start the network:"
echo "     docker compose -f docker-compose-prysm-local.yml up -d"
echo ""
echo "  2. Watch logs:"
echo "     docker logs -f prysm-beacon-local"
echo ""
echo "  3. To rebuild after changes:"
echo "     ./build-prysm-docker.sh"
echo "     docker compose -f docker-compose-prysm-local.yml restart beacon-chain validator"
