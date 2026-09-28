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

Each layer is a Go package of its own. The package file with the name of the
folder (for example `deno/deno.go`) declares the layer; a file `versions.go`
holds the version source of the layer when it has its own.

| Layer | Package | Content | Version source | Needs |
| ----- | ------- | ------- | -------------- | ----- |
| `git` | [`git`](git) | system-wide git settings (rebase on pull, auto stash) | — | os |
| `aliases` | [`aliases`](aliases) | shell aliases: ls, ll, grep, vs | — | user |
| `pure-prompt` | [`pureprompt`](pureprompt) | Pure prompt for zsh | GitHub releases | user |
| `go` | [`golang`](golang) | Go, gopls, dlv, staticcheck, govulncheck, golangci-lint | Go: go.dev; Go tools: Go module proxy; golangci-lint: GitHub releases | user |
| `node` | [`node`](node) | nvm, Node.js, npm | nvm: GitHub releases; Node.js: release index | user, build-tools |
| `deno` | [`deno`](deno) | Deno | Deno releases and release files | user |
| `prettier` | [`prettier`](prettier) | Prettier with the Tailwind CSS plugin, global configuration `/.prettierrc.json` | npm registry | node |
| `github-cli` | [`githubcli`](githubcli) | GitHub CLI (`gh`) from the GitHub release archive | GitHub releases | — |

The root package imports all layer packages, so
`import _ "github.com/majikmate/devcontainer-features"` registers all layers
(devcontainer-core builds `devcon` this way). A single layer is registered
with the import of its package, for example
`import _ "github.com/majikmate/devcontainer-features/deno"`.

`os`, `user` and `build-tools` are layers of devcontainer-core. Build
arguments, VS Code settings and tests of all layers:
[docs/layers.md](https://github.com/majikmate/devcontainer-core/blob/main/docs/layers.md).

## Versions

**The feature decides the version, never the Dockerfile.** Every tool
declares its source (`layer.Source`) and the release choice of the feature
(`Tool.Version`: pinned line and channel). The release choice is a pair of
constants at the top of the layer file, directly after the imports, for
example in [`deno/deno.go`](deno/deno.go):

```go
const (
	denoPin     = "2"
	denoChannel = "lts"
)
```

An empty pin means no pin (the newest release); an empty channel means the
default channel of the source. A test checks that every layer file with
tools starts with these constants. The general rule of
devcontainer-core chooses the newest release in the channel and in the line
([Versions in devcontainer-core](https://github.com/majikmate/devcontainer-core#versions)).
The `devcontainer.json` of an image can override the choice
(`customizations.devcon.<tool>`); no image does this today.

| Tool | Pin | Channel | End of life of the line | Source |
| ---- | --- | ------- | ----------------------- | ------ |
| `go` | `1.27` | — | when Go 1.(N+2) is released | [Go release policy](https://go.dev/doc/devel/release#policy) |
| `node` | `24` | `lts` | on the end date of the line | [Node.js release schedule](https://github.com/nodejs/Release#release-schedule) |
| `deno` | `2` | `lts` | when Deno publishes a newer major release | [Deno releases](https://github.com/denoland/deno/releases) |
| all other tools | none | — | — | the newest release |

- **Channels.** Node.js: `lts` (the releases of an LTS line) or `current`
  (all releases). Deno: `lts` or `stable` (all releases).
- **Deno LTS is one release, not a minor line.** The channel `lts` is exactly
  the release in [release-lts-latest.txt](https://dl.deno.land/release-lts-latest.txt),
  the same as `deno upgrade lts`. Deno promotes a release of the LTS line (for
  example v2.9.3) to LTS and builds it again with the channel "long term
  support". Newer patch releases of the same line (for example v2.9.7) are
  stable releases until Deno promotes one of them
  ([Deno stability and releases](https://docs.deno.com/runtime/fundamentals/stability_and_releases/)).
- **End of life.** At the end of life of a pinned line, the release check and
  the build fail with a message that names the line, the reason, the source
  and the supported lines. There is no warning before.
- **The Go tools follow Go.** gopls, dlv, staticcheck and govulncheck get the
  newest release whose `go.mod` accepts the installed Go version.
- **Release notes.** They show `tool/<name>`, `pin/<name>` and
  `channel/<name>`, and the installed `deno-channel` (from `deno --version`,
  for example `long term support`) and `node-channel` (for example
  `LTS Krypton`).
- **No upgrade messages.** The feature decides the version, so the layers set
  `DENO_NO_UPDATE_CHECK=1` (Deno CLI and the language server in VS Code) and
  `NPM_CONFIG_UPDATE_NOTIFIER=false` (npm) as `containerEnv`.

The sources and their rules are in [`golang/versions.go`](golang/versions.go),
[`node/versions.go`](node/versions.go) and [`deno/versions.go`](deno/versions.go);
the general sources (GitHub releases, npm, Go module proxy, Go releases) are
in `pkg/versions` of devcontainer-core.

## Rules for a layer

- **Distribution-independent:** a layer downloads, checks and configures. It
  does not use the package manager of the distribution (no `pkg/debian`). A
  test enforces this.
- **System needs** are declared as needed layers of devcontainer-core
  (`Needs`), for example `node` needs `build-tools`.
- **Every download is checked** against a checksum of the publisher.
- **Every layer has a test** that runs in the built image and records the
  installed versions.
- **Versions:** a layer declares its tools with a source and its release
  choice (see [Versions](#versions)). The release workflow of every image passes these versions
  as build arguments and releases a new image when one changes.

## Change a layer

1. Add or change one layer package (its file registers the layer in its
   `init` function; the release choice is at the top of the file). A new
   package is also added to the imports of [`doc.go`](doc.go) and to the
   list in [`features_test.go`](features_test.go). Open a pull request; CI
   checks the format, `go vet` and the unit tests.
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
