#!/bin/bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


# Read JSON input from stdin
input=$(cat)

# Extract values from JSON
model=$(echo "$input" | jq -r '.model.display_name')
cwd=$(echo "$input" | jq -r '.workspace.current_dir')
project_dir=$(echo "$input" | jq -r '.workspace.project_dir')

# Get username and hostname
user=$(whoami)
host=$(hostname -s)

# Determine which directory to use (prefer current_dir, fallback to project_dir)
work_dir="${cwd:-$project_dir}"

# Get relative path from project root if inside project
if [[ -n "$project_dir" && "$work_dir" == "$project_dir"* ]]; then
    # Inside project - show relative path
    if [[ "$work_dir" == "$project_dir" ]]; then
        display_dir="."
    else
        display_dir="${work_dir#$project_dir/}"
    fi
else
    # Outside project or no project - show full path
    display_dir="$work_dir"
fi

# Git information function (skip optional locks for performance)
get_git_info() {
    local dir="$1"

    # Check if we're in a git repository
    if ! git -C "$dir" rev-parse --git-dir > /dev/null 2>&1; then
        echo ""
        return
    fi

    # Get branch name (skip locks)
    local branch=$(git -C "$dir" -c core.fileMode=false rev-parse --abbrev-ref HEAD 2>/dev/null)

    # Get status info (skip locks, untracked files cache)
    local git_status=$(git -C "$dir" -c core.fileMode=false status --porcelain --untracked-files=normal 2>/dev/null)

    # Count changes
    local staged=$(echo "$git_status" | grep -c '^[MADRC]' 2>/dev/null || echo "0")
    local modified=$(echo "$git_status" | grep -c '^.[MD]' 2>/dev/null || echo "0")
    local untracked=$(echo "$git_status" | grep -c '^??' 2>/dev/null || echo "0")

    # Build status indicators
    local indicators=""
    [[ $staged -gt 0 ]] && indicators="${indicators}+${staged}"
    [[ $modified -gt 0 ]] && indicators="${indicators}~${modified}"
    [[ $untracked -gt 0 ]] && indicators="${indicators}?${untracked}"

    # Output git info
    if [[ -n "$indicators" ]]; then
        echo " ($branch $indicators)"
    else
        echo " ($branch)"
    fi
}

# Get git info for current directory
git_info=$(get_git_info "$work_dir")

# Build and output the status line
# The terminal will apply dimming to the colors
printf "%s@%s:%s%s" "$user" "$host" "$display_dir" "$git_info"
