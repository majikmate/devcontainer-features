#!/bin/sh
set -e

echo "Activating feature 'aliases'"
echo "============================"
echo "aliases: ${ALIASES}"

echo "The effective dev container remoteUser is '$_REMOTE_USER'"
echo "The effective dev container remoteUser's home directory is '$_REMOTE_USER_HOME'"

echo "The effective dev container containerUser is '$_CONTAINER_USER'"
echo "The effective dev container containerUser's home directory is '$_CONTAINER_USER_HOME'"


# Parse the comma-separated aliases string and create alias commands
parse_aliases() {
    if [ -z "$ALIASES" ]; then
        echo "No aliases provided, skipping."
        return 0
    fi
    
    # Split the aliases string by comma and process each alias
    echo "$ALIASES" | tr ',' '\n' | while IFS='=' read -r alias_name alias_command; do
        if [ -n "$alias_name" ] && [ -n "$alias_command" ]; then
            echo "alias $alias_name=\"$alias_command\""
        fi
    done
}

# Generate the alias configuration
ALIAS_CONFIG=""
if [ -n "$ALIASES" ]; then
    ALIAS_CONFIG="# Custom aliases
$(parse_aliases)"
fi

# Add aliases to .bashrc if it exists
if [ -f "${_REMOTE_USER_HOME}/.bashrc" ]; then
    echo "Adding aliases to ${_REMOTE_USER_HOME}/.bashrc"
    if [ -n "$ALIAS_CONFIG" ]; then
        echo "$ALIAS_CONFIG" >> "${_REMOTE_USER_HOME}/.bashrc"
    fi
else
    echo "${_REMOTE_USER_HOME}/.bashrc not found, skipping"
fi

# Add aliases to .zshrc if it exists
if [ -f "${_REMOTE_USER_HOME}/.zshrc" ]; then
    echo "Adding aliases to ${_REMOTE_USER_HOME}/.zshrc"
    if [ -n "$ALIAS_CONFIG" ]; then
        echo "$ALIAS_CONFIG" >> "${_REMOTE_USER_HOME}/.zshrc"
    fi
else
    echo "${_REMOTE_USER_HOME}/.zshrc not found, skipping"
fi

echo "Aliases feature installation completed."