# DevContainer Features

This repository contains a collection of
[Dev Container features](https://containers.dev/implementors/features/) for enhanced
development environment setup and configuration.

## Usage

To use a feature, add it to your `devcontainer.json` file:

```json
{
  "features": {
    "ghcr.io/majikmate/devcontainer-features/feature-name:1": {}
  }
}
```

## Available Features

| Feature                                      | Name                            | Description                                                                                                                                                  |
| -------------------------------------------- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [aliases](./src/aliases)                     | Set aliases                     | A feature to set custom shell aliases                                                                                                                        |
| [deno](./src/deno)                           | Deno                            | Installs the Deno runtime from the official Deno release channels (LTS by default)                                                                           |
| [gh-cli-extensions](./src/gh-cli-extensions) | Setup gh cli extensions         | A feature to install gh cli extensions                                                                                                                       |
| [git](./src/git)                             | Setup git                       | A feature to set setup git                                                                                                                                   |
| [locales](./src/locales)                     | Set locales                     | A feature to set your preferred locales                                                                                                                      |
| [playwright-deps](./src/playwright-deps)     | Playwright Browser Dependencies | Install Playwright browser dependencies without installing Playwright itself. Installs native libraries required for Chromium, Firefox, and WebKit browsers. |
| [prettier](./src/prettier)                   | Prettier                        | Installs the Prettier CLI system-wide and a global fallback configuration that sorts Tailwind CSS classes                                                    |
| [pure-prompt](./src/pure-prompt)             | Pure Prompt                     | Install Pure prompt for zsh - a minimal and elegant prompt                                                                                                   |
| [update-os](./src/update-os)                 | Update OS                       | A feature to update the os                                                                                                                                   |

## Examples

### Basic Usage

```json
{
  "features": {
    "ghcr.io/majikmate/devcontainer-features/aliases:1": {},
    "ghcr.io/majikmate/devcontainer-features/deno:1": {},
    "ghcr.io/majikmate/devcontainer-features/gh-cli-extensions:1": {},
    "ghcr.io/majikmate/devcontainer-features/git:1": {},
    "ghcr.io/majikmate/devcontainer-features/locales:1": {},
    "ghcr.io/majikmate/devcontainer-features/playwright-deps:1": {},
    "ghcr.io/majikmate/devcontainer-features/prettier:1": {},
    "ghcr.io/majikmate/devcontainer-features/pure-prompt:1": {},
    "ghcr.io/majikmate/devcontainer-features/update-os:1": {}
  }
}
```

### Advanced Configuration

```json
{
  "features": {
    "ghcr.io/majikmate/devcontainer-features/aliases:1": {
      "aliases": "ls=ls --color,ll=ls --color -la,vs=code -r ."
    },
    "ghcr.io/majikmate/devcontainer-features/deno:1": {
      "version": "lts"
    },
    "ghcr.io/majikmate/devcontainer-features/gh-cli-extensions:1": {
      "gh-mmc": "true"
    },
    "ghcr.io/majikmate/devcontainer-features/git:1": {
      "pull-rebase": "true",
      "rebase-autostash": "true"
    },
    "ghcr.io/majikmate/devcontainer-features/locales:1": {
      "lang": "en_GB.UTF-8",
      "measurement": "de_AT.UTF-8"
    },
    "ghcr.io/majikmate/devcontainer-features/playwright-deps:1": {
      "browsers": "chromium firefox webkit"
    },
    "ghcr.io/majikmate/devcontainer-features/prettier:1": {
      "tailwindcss": "true",
      "tailwindcssPluginVersion": "latest"
    },
    "ghcr.io/majikmate/devcontainer-features/pure-prompt:1": {
      "autoUpdate": "true"
    },
    "ghcr.io/majikmate/devcontainer-features/update-os:1": {
      "atcreate": "true",
      "atstart": ""
    }
  }
}
```

## Versions and releases

- Reference a feature by its **major version** (for example `:1`). New minor
  and patch versions then reach your image automatically at the next build;
  a new major version (breaking change) needs a change of the reference.
- A feature is published when its `version` in `devcontainer-feature.json`
  changes and the change reaches the `main` branch
  (`.github/workflows/release.yaml`). The workflow also creates the git tag
  `feature_<id>_<version>`.
- The images in the `majikmate/devcontainer-*` repositories check every hour
  whether a feature they use has a new version, and then rebuild and release
  themselves. No further action is needed here.

## Development

- Pull requests validate all `devcontainer-feature.json` files and run the
  feature tests in `test/<feature>/` on Debian 13 (trixie)
  (`.github/workflows/validate.yml`). Run a test locally with
  `devcontainer features test --skip-autogenerated --features <feature> .`
- Regenerate this README and the feature READMEs with `./generate-readme.sh`.
  Text that is not generated belongs into `src/<feature>/NOTES.md`.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](./LICENSE).
