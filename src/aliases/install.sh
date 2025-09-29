#!/bin/sh
set -e

echo "Activating feature 'aliases'"
echo "============================"
echo "aliases: ${ALIASES}"

# Resolve users
REMOTE_USER="${_REMOTE_USER:-$(id -un 2>/dev/null || true)}"
CONTAINER_USER="${_CONTAINER_USER:-root}"

# Try to resolve remote user home from the env var, passwd, or a sane default
resolve_home() {
    u="$1"
    h="${2:-}"
    if [ -n "$h" ]; then
        printf '%s' "$h"; return 0
    fi
    # If user exists in /etc/passwd, use its home
    if getent passwd "$u" >/dev/null 2>&1; then
        getent passwd "$u" | cut -d: -f6
        return 0
    fi
    # Common default
    if [ -d "/home/$u" ]; then
        printf '/home/%s' "$u"; return 0
    fi
    # Last resort (build-time before user exists)
    printf '%s' "${_CONTAINER_USER_HOME:-/root}"
}

REMOTE_USER_HOME="$(resolve_home "$REMOTE_USER" "${_REMOTE_USER_HOME:-}")"
CONTAINER_USER_HOME="$(resolve_home "$CONTAINER_USER" "${_CONTAINER_USER_HOME:-}")"

echo "The effective dev container remoteUser is '${REMOTE_USER}'"
echo "The effective dev container remoteUser's home directory is '${REMOTE_USER_HOME}'"
echo "The effective dev container containerUser is '${CONTAINER_USER}'"
echo "The effective dev container containerUser's home directory is '${CONTAINER_USER_HOME}'"

# Parse aliases
parse_aliases() {
    [ -z "$ALIASES" ] && { echo "No aliases provided, skipping."; return 0; }
    echo "$ALIASES" | tr ',' '\n' | while IFS='=' read -r alias_name alias_command; do
        alias_name=$(printf '%s' "$alias_name" | sed 's/^[[:blank:]]*//;s/[[:blank:]]*$//')
        alias_command=$(printf '%s' "$alias_command" | sed 's/^[[:blank:]]*//;s/[[:blank:]]*$//')
        [ -n "$alias_name" ] && [ -n "$alias_command" ] && \
          printf 'alias %s="%s"\n' "$alias_name" "$alias_command"
    done
}

ALIAS_CONFIG=""
if [ -n "$ALIASES" ]; then
    ALIAS_CONFIG="# Custom aliases\n$(parse_aliases)"
fi

append_if_exists() {
    target="$1"
    if [ -f "$target" ]; then
        echo "Adding aliases to $target"
        [ -n "$ALIAS_CONFIG" ] && printf "%b\n" "$ALIAS_CONFIG" >> "$target"
    else
        echo "$target not found, skipping"
    fi
}

# Prefer drop-in files that work even before the user exists
# Bash (Debian/Ubuntu): /etc/profile.d
if [ -n "$ALIAS_CONFIG" ]; then
    echo "Installing shell-agnostic alias snippets"
    printf "%b\n" "$ALIAS_CONFIG" > /etc/profile.d/99-aliases.sh
    chmod 0644 /etc/profile.d/99-aliases.sh

    # Zsh global config locations (use ones that exist)
    if [ -d /etc/zsh/zshrc.d ]; then
        printf "%b\n" "$ALIAS_CONFIG" > /etc/zsh/zshrc.d/99-aliases.zsh
        chmod 0644 /etc/zsh/zshrc.d/99-aliases.zsh
    elif [ -f /etc/zsh/zshrc ]; then
        printf "\n%b\n" "$ALIAS_CONFIG" >> /etc/zsh/zshrc
    fi
fi

# Also try per-user shells if the home exists at this point
append_if_exists "${REMOTE_USER_HOME}/.bashrc"
append_if_exists "${REMOTE_USER_HOME}/.zshrc"

echo "Aliases feature installation completed."