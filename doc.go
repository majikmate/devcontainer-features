// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package features is the collection of distribution-independent layers of
// the Dev Container images. Each layer is a package of its own:
//
//	aliases      the layer aliases: shell aliases ls, ll, grep, vs
//	deno         the layer deno: Deno
//	git          the layer git: system-wide git settings
//	githubcli    the layer github-cli: GitHub CLI (gh)
//	golang       the layer go: Go, the Go tools and golangci-lint
//	node         the layer node: nvm, Node.js and npm
//	prettier     the layer prettier: Prettier with the Tailwind CSS plugin
//	pureprompt   the layer pure-prompt: the Pure prompt for zsh
//
// The release choice of a tool (pinned line and channel) is a pair of
// constants at the top of the file of its layer, for example denoPin and
// denoChannel in deno/deno.go.
//
// A feature downloads, checks and configures; it does not use the package
// manager of the distribution and no distribution-specific paths. What it
// needs from the system, it declares as a needed layer of devcontainer-core
// (for example node needs build-tools). The framework (layer definition,
// helpers, VS Code metadata) is in devcontainer-core; devcontainer-core also
// builds the program devcon with these features.
//
// Every layer package registers its layer in its init function. Importing
// this package registers all of them:
//
//	import _ "github.com/majikmate/devcontainer-features"
package features

import (
	_ "github.com/majikmate/devcontainer-features/aliases"
	_ "github.com/majikmate/devcontainer-features/deno"
	_ "github.com/majikmate/devcontainer-features/git"
	_ "github.com/majikmate/devcontainer-features/githubcli"
	_ "github.com/majikmate/devcontainer-features/golang"
	_ "github.com/majikmate/devcontainer-features/node"
	_ "github.com/majikmate/devcontainer-features/prettier"
	_ "github.com/majikmate/devcontainer-features/pureprompt"
)
