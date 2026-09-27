# devcontainer-features

The collection of **distribution-independent features** of the Dev Container
images. Each feature is one layer that the program `devcon` installs with
`RUN devcon install <feature>` in a Dockerfile.

This repository is a Go library (module `github.com/majikmate/devcontainer-features`).
It publishes no image. [devcontainer-core](https://github.com/majikmate/devcontainer-core)
provides the framework, the Debian-bound layers and the core image, and builds
`devcon` with the newest version of this library.

## Features

| Feature | Content | Tools and version source | Needs |
| ------- | ------- | ------------------------ | ----- |
| `git` | system-wide git settings (rebase on pull, auto stash) | — | os |
| `aliases` | shell aliases: ls, ll, grep, vs | — | user |
| `pure-prompt` | Pure prompt for zsh | pure: GitHub releases | user |
| `go` | Go, gopls, dlv, staticcheck, govulncheck, golangci-lint | Go: go.dev; tools: Go module proxy; golangci-lint: GitHub releases | user |
| `node` | nvm, the newest Node.js LTS release, npm | nvm: GitHub releases; Node.js: release index (LTS) | user, build-tools |
| `deno` | Deno | Deno LTS channel | user |
| `prettier` | Prettier with the Tailwind CSS plugin and a global configuration `/.prettierrc.json` | npm registry | node |
| `github-cli` | GitHub CLI (`gh`) from the GitHub release archive | GitHub releases | — |

`os`, `user` and `build-tools` are layers of devcontainer-core. The build
arguments, VS Code settings and tests of all layers are listed in
[docs/layers.md of devcontainer-core](https://github.com/majikmate/devcontainer-core/blob/main/docs/layers.md).

## Rules for a feature

- **Distribution-independent:** a feature downloads, checks and configures.
  It does not use the package manager of the distribution (no `pkg/debian`
  of the framework) and no distribution-specific paths. A test enforces this.
- **What a feature needs from the system**, it declares as a needed layer of
  devcontainer-core (`Needs`), for example `node` needs `build-tools`.
- **Every download is checked** against a checksum from the publisher
  (`sys.VerifySHA256`, `sys.ChecksumFor`).
- **Every feature has a test** that runs in the built image and records the
  installed versions (`t.Version`).
- **Versions:** a feature declares its tools with a function that returns the
  newest version (`Tools`). The release workflow of every image asks for these
  versions every night, passes them as build arguments, and rebuilds the image
  when a version changes. The feature itself does not change for a new tool
  version.

A feature uses the framework of devcontainer-core: `pkg/layer` (definition,
tests), `pkg/sys` (commands, downloads, archives, users, files),
`pkg/shellrc` (shell settings), `pkg/state`, `pkg/versions` and
`pkg/devcontainer` (VS Code settings).

## Adding or changing a feature

1. Add or change one file in this repository (the file registers the layer in
   its `init` function).
2. Open a pull request. CI checks the format, runs `go vet` and the unit tests.
3. After the merge, the Release workflow creates the next version tag
   (patch; a manual run can choose minor or major).
4. devcontainer-core finds the new version in its nightly check (23:17 UTC),
   builds `devcon` with it and releases; the other images follow. A manual
   chain build of an image does the same at once.
5. To use a new feature in an image, add `RUN devcon install <feature>` to its
   Dockerfile.

## Development

```sh
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go test ./...
```

## History

Until version 1 of the images, this repository contained Dev Container
features (Bash scripts). They are replaced by this library. The feature
packages already published under `ghcr.io/majikmate/devcontainer-features/…`
get no more updates.

## License

MIT
