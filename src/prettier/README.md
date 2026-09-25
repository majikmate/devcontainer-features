# Prettier (prettier)

Installs the Prettier CLI system-wide and a global fallback configuration that sorts Tailwind CSS classes

## Example Usage

```json
"features": {
    "ghcr.io/majikmate/devcontainer-features/prettier:1": {}
}
```

## Options

| Options Id               | Description                                                                                     | Type    | Default Value |
| ------------------------ | ----------------------------------------------------------------------------------------------- | ------- | ------------- |
| version                  | Prettier version to install (npm version or dist-tag)                                           | string  | latest        |
| tailwindcss              | Install prettier-plugin-tailwindcss and enable it in the global configuration /.prettierrc.json | boolean | true          |
| tailwindcssPluginVersion | prettier-plugin-tailwindcss version to install (npm version or dist-tag)                        | string  | latest        |

## Customizations

### VS Code Extensions

- `esbenp.prettier-vscode`

## What the feature installs

- The Prettier CLI and (by default) the plugin
  [`prettier-plugin-tailwindcss`](https://github.com/tailwindlabs/prettier-plugin-tailwindcss),
  both in `/usr/local/lib/node_modules`. The `prettier` command is in
  `/usr/local/bin`. The installation does not depend on the Node.js version that
  is selected with `nvm`.
- The VS Code extension `esbenp.prettier-vscode`, configured with
  `prettier.prettierPath` so that the editor and the terminal use the same
  Prettier version.
- The global fallback configuration `/.prettierrc.json`.

## Global fallback configuration

Prettier searches for a configuration file from the folder of the formatted file
upward to the root folder `/`. A project without its own Prettier configuration
therefore uses `/.prettierrc.json`. This file contains no style options, so the
standard Prettier style applies. It only loads the Tailwind CSS plugin, which
sorts Tailwind CSS classes (in `class`, `className` and `@apply`) into the
standard order. The project does not need to install Tailwind CSS for this.

A project with its own Prettier configuration (for example `.prettierrc.json`)
uses its own configuration, as usual with Prettier. To keep the Tailwind CSS
class sorting in such a project, add the plugin to that configuration.

## Requirements

Node.js and npm must be installed before this feature, for example with
`ghcr.io/devcontainers/features/node`.

---

_Note: This file was auto-generated from the [devcontainer-feature.json](https://github.com/majikmate/devcontainer-features/blob/main/src/prettier/devcontainer-feature.json). Add additional notes to a `NOTES.md`._
