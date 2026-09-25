#!/bin/sh
set -e

echo "Activating feature 'prettier'"
echo "============================"

PRETTIER_VERSION="${VERSION:-latest}"
TAILWINDCSS="${TAILWINDCSS:-true}"
TAILWINDCSS_PLUGIN_VERSION="${TAILWINDCSSPLUGINVERSION:-latest}"

echo "Prettier version: ${PRETTIER_VERSION}"
echo "Tailwind CSS plugin: ${TAILWINDCSS} (version ${TAILWINDCSS_PLUGIN_VERSION})"

# Prettier needs Node.js and npm (for example from ghcr.io/devcontainers/features/node)
if ! command -v npm >/dev/null 2>&1 || ! command -v node >/dev/null 2>&1; then
    echo "Error: Node.js and npm are required. Add the feature ghcr.io/devcontainers/features/node before this feature."
    exit 1
fi

# Install into /usr/local (and not into the nvm directory), so the installation does
# not depend on the Node.js version that is selected with nvm.
PREFIX=/usr/local
MODULES="${PREFIX}/lib/node_modules"

PACKAGES="prettier@${PRETTIER_VERSION}"
if [ "${TAILWINDCSS}" = "true" ]; then
    PACKAGES="${PACKAGES} prettier-plugin-tailwindcss@${TAILWINDCSS_PLUGIN_VERSION}"
fi

echo "Installing ${PACKAGES} into ${PREFIX}"
# shellcheck disable=SC2086
npm install --global --prefix "${PREFIX}" --no-fund --no-audit --no-update-notifier ${PACKAGES}
npm cache clean --force >/dev/null 2>&1 || true

# Global fallback configuration. Prettier searches for a configuration from the folder
# of the formatted file up to "/". A project without its own Prettier configuration
# therefore uses this file. It contains no style options (standard Prettier style),
# only the Tailwind CSS plugin, which sorts the classes into the standard order.
CONFIG=/.prettierrc.json
if [ "${TAILWINDCSS}" = "true" ]; then
    PLUGIN_ENTRY="$(node -e "console.log(require.resolve('prettier-plugin-tailwindcss', { paths: ['${MODULES}'] }))")"
    printf '{\n  "plugins": ["%s"]\n}\n' "${PLUGIN_ENTRY}" > "${CONFIG}"
    chmod 0644 "${CONFIG}"
    echo "Wrote ${CONFIG}:"
    cat "${CONFIG}"
elif [ -f "${CONFIG}" ]; then
    echo "Removing ${CONFIG} (Tailwind CSS plugin disabled)"
    rm -f "${CONFIG}"
fi

echo "Installed: prettier $("${PREFIX}/bin/prettier" --version)"

# Self-test: format a sample file with the global configuration
if [ "${TAILWINDCSS}" = "true" ]; then
    SAMPLE_DIR="$(mktemp -d)"
    printf '<div class="p-4 flex m-2"></div>\n' > "${SAMPLE_DIR}/sample.html"
    RESULT="$("${PREFIX}/bin/prettier" "${SAMPLE_DIR}/sample.html")"
    rm -rf "${SAMPLE_DIR}"
    case "${RESULT}" in
        *'class="m-2 flex p-4"'*) echo "Tailwind CSS class sorting works." ;;
        *) echo "Error: Tailwind CSS class sorting did not work. Output: ${RESULT}"; exit 1 ;;
    esac
fi

echo "Prettier feature installation completed."
