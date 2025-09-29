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

| Feature | Name | Description |
|---------|------|-------------|
| [aliases](./src/aliases) | Set aliases | A feature to set custom shell aliases |
| [gh-cli-extensions](./src/gh-cli-extensions) | Setup gh cli extensions | A feature to install gh cli extensions |
| [git](./src/git) | Setup git | A feature to set setup git |
| [locales](./src/locales) | Set locales | A feature to set your preferred locales |
| [update-os](./src/update-os) | Update OS | A feature to update the os |

## Examples

### Basic Usage

```json
{
    "features": {
        "ghcr.io/majikmate/devcontainer-features/aliases:1": {},
        "ghcr.io/majikmate/devcontainer-features/gh-cli-extensions:1": {},
        "ghcr.io/majikmate/devcontainer-features/git:1": {},
        "ghcr.io/majikmate/devcontainer-features/locales:1": {},
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
        "ghcr.io/majikmate/devcontainer-features/update-os:1": {
            "atcreate": "true",
            "atstart": ""
        }
    }
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](./LICENSE).
