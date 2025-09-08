# Classroom Codespace Features

This repository contains a collection of
[Dev Container features](https://containers.dev/implementors/features/) for use
in classroom environments.

## Usage

To use a feature, add it to your `devcontainer.json` file:

```json
{
    "features": {
        "ghcr.io/majikmate/classroom-codespace-feature/locales:1": {},
        "ghcr.io/majikmate/classroom-codespace-feature/git:1": {},
        "ghcr.io/majikmate/classroom-codespace-feature/gh-cli-extensions:1": {},
        "ghcr.io/majikmate/classroom-codespace-feature/update-os:1": {}
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

For more information, see the [locales feature README](./src/locales/README.md).

### Git

This feature allows you to configure git settings including pull rebase
preferences.

For more information, see the [git feature README](./src/git/README.md).

### GitHub CLI Extensions

This feature allows you to install GitHub CLI extensions for enhanced
functionality in your development environment.

For more information, see the
[gh-cli-extensions feature README](./src/gh-cli-extensions/README.md).

### Update OS

This feature allows you to update the OS when a container is created or started.

For more information, see the
[update-os feature README](./src/update-os/README.md).

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](./LICENSE).
