# DevContainer Features

This repository contains a collection of
[Dev Container features](https://containers.dev/implementors/features/) for
enhanced development environment setup and configuration.

## Usage

To use a feature, add it to your `devcontainer.json` file:

### Basic Usage

```json
{
    "features": {
        "ghcr.io/majikmate/devcontainer-features/locales:1": {},
        "ghcr.io/majikmate/devcontainer-features/git:1": {},
        "ghcr.io/majikmate/devcontainer-features/gh-cli-extensions:1": {},
        "ghcr.io/majikmate/devcontainer-features/update-os:1": {}
    }
}
```

### Advanced Configuration

You can customize each feature with specific options:

```json
{
    "features": {
        "ghcr.io/majikmate/devcontainer-features/locales:1": {
            "timezone": "America/New_York",
            "lang": "en_US.UTF-8",
            "time": "en_US.UTF-8"
        },
        "ghcr.io/majikmate/devcontainer-features/git:1": {
            "pull-rebase": false,
            "rebase-autostash": false
        },
        "ghcr.io/majikmate/devcontainer-features/gh-cli-extensions:1": {
            "gh-mmc": true
        },
        "ghcr.io/majikmate/devcontainer-features/update-os:1": {
            "atcreate": true,
            "atstart": true
        }
    }
}
```

## Features

| Feature                                      | Description                                                                                      |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| [locales](./src/locales)                     | Set your preferred system locales for language, time, numeric, monetary, and measurement formats |
| [git](./src/git)                             | Configure git settings including pull rebase preferences                                         |
| [gh-cli-extensions](./src/gh-cli-extensions) | Install GitHub CLI extensions for enhanced functionality                                         |
| [update-os](./src/update-os)                 | Update the OS when a container is created or started                                             |

### Locales

This feature allows you to set your preferred system locales for language, time,
numeric, monetary, and measurement formats in your development container.

**Configuration options:**

- `timezone`: Your preferred timezone (default: `Europe/Vienna`)
- `lang`: Your preferred language (default: `en_GB.UTF-8`)
- `time`: Your preferred time format (default: `de_AT.UTF-8`)
- `numeric`: Your preferred numeric format (default: `de_AT.UTF-8`)
- `monetary`: Your preferred monetary format (default: `de_AT.UTF-8`)
- `measurement`: Your preferred measurement format (default: `de_AT.UTF-8`)

### Git

This feature configures git settings including pull rebase preferences and
autostash behavior.

**Configuration options:**

- `pull-rebase`: Enable pull rebase by default (default: `true`)
- `rebase-autostash`: Enable rebase autostash by default (default: `true`)

### GitHub CLI Extensions

This feature installs GitHub CLI extensions for enhanced functionality in your
development environment.

**Configuration options:**

- `gh-mmc`: Install the gh-mmc CLI extension (default: `true`)

### Update OS

This feature updates the operating system packages when a container is created
or started.

**Configuration options:**

- `atcreate`: Update OS when container is created or prebuilt (default: `true`)
- `atstart`: Update OS when container is started (default: `false`)

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](./LICENSE).
