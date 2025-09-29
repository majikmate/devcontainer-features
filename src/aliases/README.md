# Set aliases (aliases)

A feature to set custom shell aliases

## Example Usage

```json
"features": {
    "ghcr.io/majikmate/devcontainer-features/aliases:1": {}
}
```

## Options

| Options Id | Description                                               | Type   | Default Value                                |
| ---------- | --------------------------------------------------------- | ------ | -------------------------------------------- |
| aliases    | Comma-separated list of aliases in format 'alias=command' | string | ls=ls --color,ll=ls --color -la,vs=code -r . |

---

_Note: This file was auto-generated from the
[devcontainer-feature.json](https://github.com/majikmate/devcontainer-features/blob/main/src/aliases/devcontainer-feature.json).
Add additional notes to a `NOTES.md`._
