#!/usr/bin/env bash
set -euo pipefail

IMAGE="atrabilis/modbus-exporter"
BUILDER_NAME="multiarch-builder"
PLATFORMS="linux/amd64,linux/arm64"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <version-tag> [also-tag-latest]"
  exit 1
fi

VERSION="$1"
TAG_LATEST="${2:-}"

echo "==> Releasing ${IMAGE}:${VERSION}"
echo "Platforms: ${PLATFORMS}"

# ------------------------------------------------------------
# 1) Verify Docker Engine local
# ------------------------------------------------------------
if ! docker info >/dev/null 2>&1; then
  echo "ERROR: Docker Engine is not running or not accessible."
  exit 1
fi

# Force local socket (ignore weird envs)
export DOCKER_HOST=unix:///var/run/docker.sock

# ------------------------------------------------------------
# 2) Remove any broken builder with same name
# ------------------------------------------------------------
if docker buildx inspect "${BUILDER_NAME}" >/dev/null 2>&1; then
  echo "==> Removing existing builder (${BUILDER_NAME}) to avoid corruption"
  docker buildx rm "${BUILDER_NAME}" || true
fi

# ------------------------------------------------------------
# 3) Install QEMU/binfmt if needed
# ------------------------------------------------------------
echo "==> Ensuring multi-arch emulation support"
docker run --privileged --rm tonistiigi/binfmt --install all >/dev/null

# ------------------------------------------------------------
# 4) Create clean builder
# ------------------------------------------------------------
echo "==> Creating clean buildx builder"
docker buildx create \
  --name "${BUILDER_NAME}" \
  --driver docker-container \
  --use >/dev/null

echo "==> Bootstrapping builder"
docker buildx inspect --bootstrap >/dev/null

# ------------------------------------------------------------
# 5) Build & push
# ------------------------------------------------------------
CMD=(
  docker buildx build
  --platform "${PLATFORMS}"
  -t "${IMAGE}:${VERSION}"
  --push
)

if [ "${TAG_LATEST}" = "latest" ]; then
  CMD+=(-t "${IMAGE}:latest")
fi

CMD+=(.)

echo "==> Running:"
printf ' %q' "${CMD[@]}"
echo

"${CMD[@]}"

echo
echo "==> Done."
echo "Published:"
echo "  - ${IMAGE}:${VERSION}"

if [ "${TAG_LATEST}" = "latest" ]; then
  echo "  - ${IMAGE}:latest"
fi