#!/usr/bin/env python3
"""
Git Automation Script

This script automates the process of creating a new Git branch,
committing changes, and pushing them to a remote repository.
"""

import subprocess
import sys


def run_git_command(command):
    """
    Execute a Git command and handle errors.
    
    Args:
        command: List of command parts (e.g., ['git', 'checkout', '-b', 'branch-name'])
    
    Returns:
        bool: True if command succeeded, False otherwise
    """
    try:
        result = subprocess.run(
            command,
            check=True,
            capture_output=True,
            text=True
        )
        if result.stdout:
            print(result.stdout)
        return True
    except subprocess.CalledProcessError as e:
        print(f"Error executing Git command: {' '.join(command)}", file=sys.stderr)
        if e.stderr:
            print(f"Error output: {e.stderr}", file=sys.stderr)
        return False


def create_and_push_branch():
    """
    Main function to create a new branch, commit changes, and push to remote.
    """
    # Get user input
    branch_name = input("Enter the new branch name: ").strip()
    commit_message = input("Enter the commit message: ").strip()
    
    # Validate input
    if not branch_name:
        print("Error: Branch name cannot be empty.", file=sys.stderr)
        sys.exit(1)
    
    if not commit_message:
        print("Error: Commit message cannot be empty.", file=sys.stderr)
        sys.exit(1)
    
    # Create and switch to a new branch
    print(f"\nCreating and switching to branch '{branch_name}'...")
    if not run_git_command(["git", "checkout", "-b", branch_name]):
        print("Failed to create branch.", file=sys.stderr)
        sys.exit(1)
    
    # Stage all changes
    print("Staging all changes...")
    if not run_git_command(["git", "add", "."]):
        print("Failed to stage changes.", file=sys.stderr)
        sys.exit(1)
    
    # Commit changes
    print(f"Committing changes with message: '{commit_message}'...")
    if not run_git_command(["git", "commit", "-m", commit_message]):
        print("Failed to commit changes.", file=sys.stderr)
        sys.exit(1)
    
    # Push the new branch to origin
    print(f"Pushing branch '{branch_name}' to origin...")
    if not run_git_command(["git", "push", "-u", "origin", branch_name]):
        print("Failed to push branch to remote.", file=sys.stderr)
        sys.exit(1)
    
    print(f"\n✅ Success! Branch '{branch_name}' created, committed, and pushed successfully.")


if __name__ == "__main__":
    create_and_push_branch()

