#!/usr/bin/env bash

## Helper to install the asdf plugins required by this project

set -euo pipefail

YELLOW='\033[1;33m'
DEFAULT='\033[0m'

BASE_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

INSTALLED_PLUGINS=$(asdf plugin list --urls)

while read -r plugin_name plugin_url; do
    { [[ -z $plugin_name ]] || [[ $plugin_name == \#* ]]; } && continue

    {
        installed_plugin_source=$(echo -e "$INSTALLED_PLUGINS" | grep -e "^$plugin_name")
        is_plugin_installed=$?
    } || :
    if [[ $is_plugin_installed -eq 0 ]]; then
        # Plugin is installed, check if repo url matches if given in plugin config file
        read -r _ installed_tool_url <<<"$installed_plugin_source"
        if [[ -n $plugin_url ]] && [[ $installed_tool_url != "$plugin_url" ]]; then
            echo -e "asdf plugin ${YELLOW}${plugin_name}${DEFAULT} is installed but using the wrong source - reinstalling"
            asdf plugin remove "$plugin_name"
            asdf plugin add "$plugin_name" "$plugin_url"
        fi
    else
        echo -e "asdf plugin ${YELLOW}${plugin_name}${DEFAULT} is missing, installing"
        asdf plugin add "$plugin_name" "$plugin_url"
    fi
done <"${BASE_DIR}/../.asdf-plugins"
