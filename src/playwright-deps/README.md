# Playwright Browser Dependencies (playwright-deps)

Install Playwright browser dependencies without installing Playwright itself.
Installs native libraries required for Chromium, Firefox, and WebKit browsers.

## Example Usage

```json
"features": {
    "ghcr.io/majikmate/devcontainer-features/playwright-deps:1": {}
}
```

## Options

| Options Id | Description                                                                                        | Type   | Default Value           |
| ---------- | -------------------------------------------------------------------------------------------------- | ------ | ----------------------- |
| browsers   | Space-separated list of browsers to install dependencies for (chromium, firefox, webkit, or 'all') | string | chromium firefox webkit |

## How It Works

This feature uses a clever approach to install system dependencies for
Playwright browsers without keeping Playwright itself installed:

1. **Temporarily installs** Playwright globally via npm
2. **Runs** `playwright install-deps` to install native OS libraries for the
   specified browsers
3. **Uninstalls** Playwright, leaving only the system dependencies

This ensures your development container has all the necessary native libraries
for running Playwright tests without the overhead of maintaining Playwright in
the feature layer.

## Examples

### Install dependencies for all browsers

```json
"features": {
    "ghcr.io/majikmate/devcontainer-features/playwright-deps:1": {
        "browsers": "all"
    }
}
```

### Install dependencies for Chromium only

```json
"features": {
    "ghcr.io/majikmate/devcontainer-features/playwright-deps:1": {
        "browsers": "chromium"
    }
}
```

### Install dependencies for specific browsers

```json
"features": {
    "ghcr.io/majikmate/devcontainer-features/playwright-deps:1": {
        "browsers": "chromium firefox"
    }
}
```

---

_Note: This file was auto-generated from the
[devcontainer-feature.json](https://github.com/majikmate/devcontainer-features/blob/main/src/playwright-deps/devcontainer-feature.json).
Add additional notes to a `NOTES.md`._
