#!/bin/bash
set -e

VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
APP_NAME="gallery"
DIST_DIR="dist"

ARCHITECTURES=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm/7"
)

echo "Building release version: $VERSION"

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

for arch in "${ARCHITECTURES[@]}"; do
    GOOS="${arch%%/*}"
    GOARCH="${arch#*/}"
    GOARM=""
    
    if [[ "$GOARCH" == arm/* ]]; then
        GOARM="${GOARCH#*/}"
        GOARCH="arm"
    fi
    
    output_name="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    [ -n "$GOARM" ] && output_name="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}v${GOARM}"
    
    echo "Building for $GOOS/$GOARCH${GOARM:+v$GOARM}..."
    
    env GOOS="$GOOS" GOARCH="$GOARCH" ${GOARM:+GOARM=$GOARM} go build -ldflags "-s -w -X main.version=$VERSION" -o "$DIST_DIR/$output_name" ./cmd/gallery
    
    tar -czf "$DIST_DIR/${output_name}.tar.gz" -C "$DIST_DIR" "$output_name"
    rm "$DIST_DIR/$output_name"
    
    echo "Created: $DIST_DIR/${output_name}.tar.gz"
done

echo ""
echo "Release artifacts created in $DIST_DIR/:"
ls -la "$DIST_DIR/"
