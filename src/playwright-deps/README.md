
# Playwright Browser Dependencies (playwright-deps)

Install Playwright browser dependencies without installing Playwright itself. Installs native libraries required for Chromium, Firefox, and WebKit browsers.

## Example Usage

```json
"features": {
    "ghcr.io/majikmate/devcontainer-features/playwright-deps:1": {}
}
```

## Options

| Options Id | Description | Type | Default Value |
|-----|-----|-----|-----|
| browsers | Space-separated list of browsers to install dependencies for (chromium, firefox, webkit, or 'all') | string | chromium firefox webkit |



---

_Note: This file was auto-generated from the [devcontainer-feature.json](https://github.com/majikmate/devcontainer-features/blob/main/src/playwright-deps/devcontainer-feature.json).  Add additional notes to a `NOTES.md`._
