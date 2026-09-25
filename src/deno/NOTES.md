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
