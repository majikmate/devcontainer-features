#!/bin/sh
set -e

echo "Activating feature 'deno'"
echo "========================"

REQUESTED_VERSION="${VERSION:-lts}"
DL_URL="https://dl.deno.land"
INSTALL_DIR=/usr/local/bin

echo "Requested version: ${REQUESTED_VERSION}"

# Install required packages
missing=""
for cmd in curl unzip; do
    command -v "${cmd}" >/dev/null 2>&1 || missing="${missing} ${cmd}"
done
if [ -n "${missing}" ] || [ ! -f /etc/ssl/certs/ca-certificates.crt ]; then
    apt-get update
    # shellcheck disable=SC2086
    apt-get install -y --no-install-recommends ca-certificates curl unzip
    rm -rf /var/lib/apt/lists/*
fi

# Resolve the version with the same release channel files that 'deno upgrade' uses:
#   lts    -> release-lts-latest.txt (newest LTS release)
#   latest -> release-latest.txt     (newest release)
fetch_channel() {
    curl -fsSL --retry 3 "${DL_URL}/$1" 2>/dev/null | tr -d '[:space:]'
}

case "${REQUESTED_VERSION}" in
    lts)
        DENO_VERSION="$(fetch_channel release-lts-latest.txt || true)"
        if [ -z "${DENO_VERSION}" ]; then
            echo "Warning: Deno publishes no LTS release at the moment. Installing the newest release instead."
            DENO_VERSION="$(fetch_channel release-latest.txt)"
        fi
        ;;
    latest)
        DENO_VERSION="$(fetch_channel release-latest.txt)"
        ;;
    *)
        DENO_VERSION="v${REQUESTED_VERSION#v}"
        ;;
esac

if [ -z "${DENO_VERSION}" ]; then
    echo "Error: could not resolve the Deno version."
    exit 1
fi

case "$(uname -m)" in
    x86_64 | amd64) TARGET="x86_64-unknown-linux-gnu" ;;
    aarch64 | arm64) TARGET="aarch64-unknown-linux-gnu" ;;
    *) echo "Error: unsupported architecture $(uname -m)"; exit 1 ;;
esac

ARCHIVE="deno-${TARGET}.zip"
URL="${DL_URL}/release/${DENO_VERSION}/${ARCHIVE}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Downloading Deno ${DENO_VERSION} for ${TARGET}"
curl -fsSL --retry 3 -o "${TMP_DIR}/${ARCHIVE}" "${URL}"

# Verify the checksum when the release provides one
if curl -fsSL --retry 3 -o "${TMP_DIR}/${ARCHIVE}.sha256sum" "${URL}.sha256sum" 2>/dev/null; then
    EXPECTED="$(grep -oE '[0-9a-fA-F]{64}' "${TMP_DIR}/${ARCHIVE}.sha256sum" | head -n 1)"
    ACTUAL="$(sha256sum "${TMP_DIR}/${ARCHIVE}" | cut -d ' ' -f 1)"
    if [ -n "${EXPECTED}" ] && [ "${EXPECTED}" != "${ACTUAL}" ]; then
        echo "Error: checksum mismatch for ${ARCHIVE}"
        exit 1
    fi
    echo "Checksum verified."
else
    echo "No checksum file published for ${ARCHIVE}; skipping checksum verification."
fi

unzip -o -q "${TMP_DIR}/${ARCHIVE}" -d "${TMP_DIR}"
install -m 0755 "${TMP_DIR}/deno" "${INSTALL_DIR}/deno"

echo "Installed: $("${INSTALL_DIR}/deno" --version | head -n 1)"
echo "Deno feature installation completed."
