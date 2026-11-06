#!/usr/bin/env bash
set -euo pipefail

output_dir="${1:-dist}"
go_bin="${NETRISE_GO:-go}"
if ! command -v "$go_bin" >/dev/null 2>&1; then
  for candidate in "/c/Program Files/Go/bin/go.exe" "/usr/local/go/bin/go"; do
    if [[ -x "$candidate" ]]; then
      go_bin="$candidate"
      break
    fi
  done
fi
command -v "$go_bin" >/dev/null 2>&1 || { echo "Go executable not found. Set NETRISE_GO or add go to PATH." >&2; exit 1; }

package_name="${2:-netrise-$("$go_bin" env GOOS)-$("$go_bin" env GOARCH)}"
output_dir="$(cd "$output_dir" 2>/dev/null && pwd || (mkdir -p "$output_dir" && cd "$output_dir" && pwd))"
staging="$output_dir/${package_name}-bundle"
archive="$output_dir/${package_name}.tar.gz"
checksum="$archive.sha256"
verify_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$staging" "$verify_dir"
}
trap cleanup EXIT

rm -rf "$staging" "$archive" "$checksum"
mkdir -p "$staging/assets/content"

goos="$("$go_bin" env GOOS)"
goarch="$("$go_bin" env GOARCH)"
binary="$staging/netrise"
if [[ "$goos" == "windows" ]]; then
  binary="$staging/netrise.exe"
fi
GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 "$go_bin" build -trimpath -ldflags='-s -w' -o "$binary" ./cmd/netrise
cp assets/content/catalog.json "$staging/assets/content/catalog.json"

tar -C "$output_dir" -czf "$archive" "${package_name}-bundle"
if command -v sha256sum >/dev/null 2>&1; then
  sha256sum "$archive" > "$checksum"
else
  shasum -a 256 "$archive" > "$checksum"
fi
recorded_hash="$(awk '{print $1}' "$checksum")"
if command -v sha256sum >/dev/null 2>&1; then
  actual_hash="$(sha256sum "$archive" | awk '{print $1}')"
else
  actual_hash="$(shasum -a 256 "$archive" | awk '{print $1}')"
fi
test "$recorded_hash" = "$actual_hash"

tar -xzf "$archive" -C "$verify_dir"
verified_binary="$verify_dir/${package_name}-bundle/$(basename "$binary")"
test -s "$verified_binary"
test -s "$verify_dir/${package_name}-bundle/assets/content/catalog.json"

host_os="$("$go_bin" env GOHOSTOS)"
host_arch="$("$go_bin" env GOHOSTARCH)"
if [[ "$goos" == "$host_os" && "$goarch" == "$host_arch" ]]; then
  "$verified_binary" --version >/dev/null
  "$verified_binary" --help >/dev/null
  "$verified_binary" corp --help >/dev/null
  "$verified_binary" validate-content
  "$verified_binary" playtest --faction runner --seed 17 >/dev/null
  "$verified_binary" playtest --faction corp --seed 17 >/dev/null
  "$verified_binary" balance --seed-from 1 --seed-to 2 >/dev/null
  "$verified_binary" simulate --faction runner --seed 17 --ticks 600 >/dev/null
  "$verified_binary" simulate --faction corp --seed 17 --ticks 600 >/dev/null
  if "$verified_binary" unknown-command >/dev/null 2>&1; then
    echo "Packaged unknown command unexpectedly succeeded." >&2
    exit 1
  fi
fi

echo "NETRISE Unix package created and verified: $archive"
echo "SHA256: $checksum"
