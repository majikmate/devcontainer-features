# devcontainer-features

The **distribution-independent layers** of the Dev Container images, as a Go
library (module `github.com/majikmate/devcontainer-features`). devcontainer-core
compiles them into the layer tool `devcon`; a Dockerfile installs a layer with
`RUN devcon install <layer>`. This repository publishes no image.

## Dependencies

```text
                                               Nightly Content
devcontainer-features                                  Go library of layers, compiled into devcon
  ▼
devcontainer-core:1                            23:17   Debian 13, devcon, user dev, zsh, SSH server
├── devcontainer-base:2                        01:17   + go, build-tools, node, deno, prettier
│   ├── devcontainer-dev:2                     03:37   + github-cli
│   ├── devcontainer-classroom-web:2           03:47   classroom settings, AI off
│   └── devcontainer-classroom-web-advanced:2  03:57   + playwright-deps, AI on
└── devcontainer-classroom-exam-ts:2           01:27   + deno, AI and coding assistance off
```

This repository: **devcontainer-features**. Nightly checks in UTC.
Repositories: [core](https://github.com/majikmate/devcontainer-core) ·
[features](https://github.com/majikmate/devcontainer-features) ·
[base](https://github.com/majikmate/devcontainer-base) ·
[dev](https://github.com/majikmate/devcontainer-dev) ·
[classroom-web](https://github.com/majikmate/devcontainer-classroom-web) ·
[classroom-web-advanced](https://github.com/majikmate/devcontainer-classroom-web-advanced) ·
[classroom-exam-ts](https://github.com/majikmate/devcontainer-classroom-exam-ts)

## Layers

| Layer | Content | Version source | Needs |
| ----- | ------- | -------------- | ----- |
| `git` | system-wide git settings (rebase on pull, auto stash) | — | os |
| `aliases` | shell aliases: ls, ll, grep, vs | — | user |
| `pure-prompt` | Pure prompt for zsh | GitHub releases | user |
| `go` | Go, gopls, dlv, staticcheck, govulncheck, golangci-lint | Go: go.dev; Go tools: Go module proxy; golangci-lint: GitHub releases | user |
| `node` | nvm, Node.js, npm | nvm: GitHub releases; Node.js: release index | user, build-tools |
| `deno` | Deno | Deno releases | user |
| `prettier` | Prettier with the Tailwind CSS plugin, global configuration `/.prettierrc.json` | npm registry | node |
| `github-cli` | GitHub CLI (`gh`) from the GitHub release archive | GitHub releases | — |

`os`, `user` and `build-tools` are layers of devcontainer-core. Build
arguments, VS Code settings and tests of all layers:
[docs/layers.md](https://github.com/majikmate/devcontainer-core/blob/main/docs/layers.md).

## Pinned release lines

The Dockerfile that installs a layer can pin a release line, for example
`ARG GO_PIN=1.27` before `RUN devcon install go`. Without a pin, the layer
installs the newest release.

| Layer | Build argument | A line is | End of life | Source |
| ----- | -------------- | --------- | ----------- | ------ |
| `go` | `GO_PIN` (for example `1.27`) | a major release 1.N | when Go 1.(N+2) is released | [Go release policy](https://go.dev/doc/devel/release#policy) |
| `node` | `NODE_PIN` (for example `24`) | a major release | on the end date of the line | [Node.js release schedule](https://github.com/nodejs/Release#release-schedule) |
| `deno` | `DENO_PIN` (for example `2`) | a major release | when Deno publishes a newer major release | [Deno releases](https://github.com/denoland/deno/releases) |

- With a pin, the layer installs the newest release inside the line.
- At the end of life, the release check and the build fail with a message that
  names the line, the reason, the source and the supported lines. There is no
  warning before.
- **The Go tools follow Go:** gopls, dlv, staticcheck and govulncheck get the
  newest release whose `go.mod` accepts the installed Go version.
  golangci-lint stays on the newest release.
- All other tools (nvm, Prettier and its plugin, GitHub CLI, Pure) stay on the
  newest release.

A layer declares a pinnable tool with `layer.Pin` (see `golang_pin.go`,
`node_pin.go`, `deno_pin.go`).

## Rules for a layer

- **Distribution-independent:** a layer downloads, checks and configures. It
  does not use the package manager of the distribution (no `pkg/debian`). A
  test enforces this.
- **System needs** are declared as needed layers of devcontainer-core
  (`Needs`), for example `node` needs `build-tools`.
- **Every download is checked** against a checksum of the publisher.
- **Every layer has a test** that runs in the built image and records the
  installed versions.
- **Versions:** a layer declares its tools with a function that returns the
  newest version. The release workflow of every image passes these versions
  as build arguments and releases a new image when one changes.

## Change a layer

1. Add or change one file (the file registers the layer in its `init`
   function). Open a pull request; CI checks the format, `go vet` and the
   unit tests.
2. After the merge, the Release workflow creates the next version tag.
3. devcontainer-core uses the new version in its nightly check (23:17 UTC) and
   releases; the other images follow.
4. To use a new layer in an image, add `RUN devcon install <layer>` to its
   Dockerfile.

## Workflow runs

Every Sunday, the workflow [Prune](.github/workflows/prune.yml) deletes the
finished workflow runs (any result) older than 90 days. **Actions → Prune**
lists or deletes them at once; the scope `all-but-newest` keeps only the
newest run of each workflow.

## Development

```sh
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go test ./...
```

## License

MIT
