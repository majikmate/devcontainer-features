# Deno (deno)

Installs the Deno runtime from the official Deno release channels (LTS by default)

## Example Usage

```json
"features": {
    "ghcr.io/majikmate/devcontainer-features/deno:1": {}
}
```

## Options

| Options Id | Description                                                                                      | Type   | Default Value |
| ---------- | ------------------------------------------------------------------------------------------------ | ------ | ------------- |
| version    | Deno version: 'lts' (newest LTS release), 'latest' (newest release) or a version such as '2.5.6' | string | lts           |

## Customizations

### VS Code Extensions

- `denoland.vscode-deno`

## Versions

The feature resolves the version with the same release channel files that
`deno upgrade` uses:

| Option value      | Installed version    | Source                                        |
| ----------------- | -------------------- | --------------------------------------------- |
| `lts` (default)   | newest LTS release   | `https://dl.deno.land/release-lts-latest.txt` |
| `latest`          | newest release       | `https://dl.deno.land/release-latest.txt`     |
| `2.5.6` (example) | exactly this version | –                                             |

If Deno publishes no LTS release at build time, `lts` installs the newest
release and prints a warning.

The `deno` binary is installed to `/usr/local/bin/deno`. The feature also adds
the VS Code extension `denoland.vscode-deno`.

---

_Note: This file was auto-generated from the [devcontainer-feature.json](https://github.com/majikmate/devcontainer-features/blob/main/src/deno/devcontainer-feature.json). Add additional notes to a `NOTES.md`._
