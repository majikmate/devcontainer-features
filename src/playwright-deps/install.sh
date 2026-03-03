#!/bin/sh
set -e

echo "Activating feature 'playwright-deps'"
echo "======================================"
echo "Browsers: ${BROWSERS:-chromium firefox webkit}"

# Verify Node.js and npm are available
if ! command -v npm >/dev/null 2>&1; then
    echo "Error: npm is not installed. Please ensure Node.js is installed before using this feature."
    exit 1
fi

# Resolve browsers option
BROWSERS_TO_INSTALL="${BROWSERS:-chromium firefox webkit}"

# If "all" is specified, install all browsers
if [ "$BROWSERS_TO_INSTALL" = "all" ]; then
    BROWSERS_TO_INSTALL="chromium firefox webkit"
fi

echo "Installing browser dependencies for: $BROWSERS_TO_INSTALL"
echo ""

echo "Step 1: Temporarily installing Playwright globally..."
# Install Playwright globally using npm
if ! npm install -g playwright >/dev/null 2>&1; then
    echo "Error: Failed to install Playwright globally"
    exit 1
fi

echo "Step 2: Installing system dependencies for browsers..."
# Install system dependencies using playwright install-deps
# This installs native OS libraries required by the browsers
if command -v playwright >/dev/null 2>&1; then
    for browser in $BROWSERS_TO_INSTALL; do
        case "$browser" in
            chromium|firefox|webkit)
                echo "  Installing dependencies for $browser..."
                if ! playwright install-deps "$browser" 2>&1 | grep -v "^$"; then
                    echo "  Warning: Some dependencies for $browser may not have installed correctly"
                fi
                ;;
            *)
                echo "  Warning: Unknown browser '$browser'. Skipping."
                ;;
        esac
    done
else
    echo "Error: playwright command not found after installation"
    exit 1
fi

echo ""
echo "Step 3: Removing Playwright package but keeping system dependencies..."
# Uninstall Playwright globally, but the system libraries remain installed
if ! npm uninstall -g playwright >/dev/null 2>&1; then
    echo "Warning: Failed to cleanly uninstall Playwright, but continuing..."
fi

echo ""
echo "✓ Playwright system dependencies installation completed successfully!"
echo "✓ Playwright package itself has been removed"
echo "✓ Native browser libraries remain installed on the system"
echo ""
echo "These dependencies will support Playwright when it's installed in your project."
