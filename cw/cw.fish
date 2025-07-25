function cw --description "Manage git worktrees and Claude Code sessions"
    set -l subcommand $argv[1]
    
    switch $subcommand
        case make create new
            _cw_make
        case list ls
            _cw_list
        case remove delete rm
            _cw_remove
        case prune clean
            _cw_prune
        case help -h --help
            _cw_help
        case ''
            _cw_make  # Default to make if no subcommand
        case '*'
            echo "Unknown subcommand: $subcommand"
            echo "Run 'cw help' for usage"
            return 1
    end
end

function _cw_make --description "Create a git worktree and start Claude Code"
    # Check if we're in a git repository
    if not git rev-parse --git-dir > /dev/null 2>&1
        echo "Error: Not in a git repository"
        return 1
    end

    # Get the current directory name for the base path
    set -l repo_name (basename (pwd))
    set -l parent_dir (dirname (pwd))

    # Prompt for directory name
    read -l -P "Enter directory name for the worktree: " dir_name
    if test -z "$dir_name"
        echo "Error: Directory name cannot be empty"
        return 1
    end

    # Prompt for branch name
    read -l -P "Enter branch name (e.g., feature/my-feature): " branch_name
    if test -z "$branch_name"
        echo "Error: Branch name cannot be empty"
        return 1
    end

    # Prompt for base branch (default to current branch)
    set -l current_branch (git branch --show-current)
    read -l -P "Base branch (default: $current_branch): " base_branch
    if test -z "$base_branch"
        set base_branch $current_branch
    end

    # Construct the full path for the worktree
    set -l worktree_path "$parent_dir/$repo_name-$dir_name"

    # Check if directory already exists
    if test -d "$worktree_path"
        echo "Error: Directory $worktree_path already exists"
        return 1
    end

    # Create the worktree
    echo "Creating worktree at: $worktree_path"
    echo "Branch: $branch_name (based on $base_branch)"
    
    if not git worktree add "$worktree_path" -b "$branch_name" "$base_branch"
        echo "Error: Failed to create worktree"
        return 1
    end

    # Navigate to the new worktree
    cd "$worktree_path"
    echo "Successfully created and moved to worktree!"

    # Optional: Install dependencies if package.json exists
    if test -f package.json
        read -l -P "Install npm dependencies? (y/N): " install_deps
        if test "$install_deps" = "y" -o "$install_deps" = "Y"
            echo "Installing dependencies..."
            npm install
        end
    end

    # Optional: Copy .env file if it exists in the original repo
    set -l original_env "$parent_dir/$repo_name/.env"
    if test -f "$original_env"
        read -l -P "Copy .env file from main repository? (y/N): " copy_env
        if test "$copy_env" = "y" -o "$copy_env" = "Y"
            cp "$original_env" .env
            echo "Copied .env file"
        end
    end

    # Start Claude Code
    echo ""
    echo "Starting Claude Code in: $worktree_path"
    echo "Branch: $branch_name"
    echo ""
    
    # Optional: Add initial context for Claude
    read -l -P "Enter initial prompt for Claude (optional): " claude_prompt
    
    if test -n "$claude_prompt"
        claude "$claude_prompt"
    else
        claude
    end
end

function _cw_list --description "List all git worktrees"
    if not git rev-parse --git-dir > /dev/null 2>&1
        echo "Error: Not in a git repository"
        return 1
    end
    
    echo "Current worktrees:"
    git worktree list
end

function _cw_remove --description "Remove a git worktree"
    if not git rev-parse --git-dir > /dev/null 2>&1
        echo "Error: Not in a git repository"
        return 1
    end
    
    # List current worktrees
    echo "Current worktrees:"
    git worktree list
    echo ""
    
    # Prompt for worktree to remove
    read -l -P "Enter path of worktree to remove: " worktree_path
    
    if test -z "$worktree_path"
        echo "Error: No path provided"
        return 1
    end
    
    # Confirm removal
    read -l -P "Are you sure you want to remove $worktree_path? (y/N): " confirm
    if test "$confirm" = "y" -o "$confirm" = "Y"
        if git worktree remove "$worktree_path"
            echo "Worktree removed successfully"
        else
            echo "Failed to remove worktree"
            return 1
        end
    else
        echo "Removal cancelled"
    end
end

function _cw_prune --description "Clean up stale worktree references"
    if not git rev-parse --git-dir > /dev/null 2>&1
        echo "Error: Not in a git repository"
        return 1
    end
    
    git worktree prune
    echo "Pruned stale worktree references"
end

function _cw_help --description "Show help for cw"
    echo "cw - Claude Worktree: Manage git worktrees and Claude Code sessions"
    echo ""
    echo "Usage:"
    echo "  cw [make|create|new]    Create a new worktree and start Claude"
    echo "  cw [list|ls]            List all worktrees"
    echo "  cw [remove|delete|rm]   Remove a worktree"
    echo "  cw [prune|clean]        Clean up stale worktree references"
    echo "  cw [help|-h|--help]     Show this help"
    echo ""
    echo "Examples:"
    echo "  cw make                 # Create new worktree (default)"
    echo "  cw                      # Same as 'make'"
    echo "  cw list                 # List all worktrees"
    echo "  cw remove               # Remove a worktree"
    echo "  cw prune                # Clean up references"
end