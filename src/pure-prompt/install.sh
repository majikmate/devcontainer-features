#!/bin/sh
set -e

echo "Activating feature 'pure-prompt'"
echo "===================================="
echo "autoUpdate: ${AUTOUPDATE}"

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

# Install Pure prompt for the remote user
install_pure() {
    user_home="$1"
    user_name="$2"
    
    # Create directory for Pure prompt functions
    pure_dir="$user_home/.zsh-pure"
    mkdir -p "$pure_dir"
    
    # Download Pure prompt files
    echo "Installing Pure prompt to $pure_dir"
    
    # Download pure.zsh as prompt_pure_setup (standard prompt function name)
    curl -fsSL https://raw.githubusercontent.com/sindresorhus/pure/main/pure.zsh \
        -o "$pure_dir/prompt_pure_setup"
    
    # Download async.zsh
    curl -fsSL https://raw.githubusercontent.com/sindresorhus/pure/main/async.zsh \
        -o "$pure_dir/async"
    
    # Set proper ownership and permissions
    if [ "$user_name" != "root" ]; then
        chown -R "$user_name:$user_name" "$pure_dir" 2>/dev/null || true
    fi
    chmod 755 "$pure_dir"
    chmod 644 "$pure_dir/prompt_pure_setup"
    chmod 644 "$pure_dir/async"
    
    echo "Pure prompt installed successfully"
}

# Configure Pure prompt in zshrc
configure_zshrc() {
    zshrc_file="$1"
    user_name="$2"
    
    if [ ! -f "$zshrc_file" ]; then
        echo "$zshrc_file not found, creating it"
        touch "$zshrc_file"
        if [ "$user_name" != "root" ]; then
            chown "$user_name:$user_name" "$zshrc_file" 2>/dev/null || true
        fi
    fi
    
    # Check if Pure is already configured
    if grep -q "# Pure prompt configuration" "$zshrc_file"; then
        echo "Pure prompt already configured in $zshrc_file"
        return 0
    fi
    
    echo "Configuring Pure prompt in $zshrc_file"
    
    cat >> "$zshrc_file" <<'EOF'

# ============================================================================
# SHELL PROMPT
# ============================================================================
# Add Pure prompt to fpath
fpath+=("$HOME/.zsh-pure")

# Initialize prompt system
autoload -U promptinit; promptinit

# Pure prompt configuration
zstyle :prompt:pure:user show yes
zstyle :prompt:pure:host show yes
zstyle :prompt:pure:git:stash show yes

# Load Pure prompt
prompt pure
EOF

    # Set proper ownership
    if [ "$user_name" != "root" ]; then
        chown "$user_name:$user_name" "$zshrc_file" 2>/dev/null || true
    fi

    # Add auto-update configuration if enabled
    if [ "$AUTOUPDATE" = "true" ]; then
        cat >> "$zshrc_file" <<'EOF'

# Auto-update Pure prompt
function update_pure_prompt() {
    curl -fsSL https://raw.githubusercontent.com/sindresorhus/pure/main/pure.zsh \
        -o "$HOME/.zsh-pure/prompt_pure_setup"
    curl -fsSL https://raw.githubusercontent.com/sindresorhus/pure/main/async.zsh \
        -o "$HOME/.zsh-pure/async"
}
EOF
    fi
}

# Install Pure for remote user
if [ -n "$REMOTE_USER_HOME" ]; then
    install_pure "$REMOTE_USER_HOME" "$REMOTE_USER"
    configure_zshrc "$REMOTE_USER_HOME/.zshrc" "$REMOTE_USER"
fi

# Also configure for container user if different
if [ "$CONTAINER_USER" != "$REMOTE_USER" ] && [ -n "$CONTAINER_USER_HOME" ]; then
    install_pure "$CONTAINER_USER_HOME" "$CONTAINER_USER"
    configure_zshrc "$CONTAINER_USER_HOME/.zshrc" "$CONTAINER_USER"
fi

echo "Pure prompt feature installation completed."
