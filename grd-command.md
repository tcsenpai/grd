---
name: grd
description: Read, write, or generate .git/description for the current repo or branch
arguments:
  - name: action
    description: "show | set <text> | generate | generate-branch"
    required: false
---

Check that `.git/` exists in the current working directory. If it doesn't, stop and tell the user this only works inside a git repo.

Read `.git/description`. If the content is empty or matches git's default placeholder ("Unnamed repository; edit this file 'description' to name the repository."), treat it as unset.

## What to do based on $ARGUMENTS

**No arguments or `show`:**
Print the current description. If unset, say so and suggest running `/grd generate`.

**`set <text>`:**
Write the provided text to `.git/description`. Confirm what was written.

**`generate`:**
Generate a repo-level description. To do this:
1. Read the README if one exists (README.md, README, README.rst, README.txt)
2. Run `git log --oneline -20` to understand recent activity
3. Glance at the top-level directory structure
4. Check the language/framework from package.json, go.mod, Cargo.toml, pyproject.toml, or similar

From this context, write a short, punchy description (1-2 sentences max). Think of it as what you'd put in the GitHub "About" field. No marketing speak, no "A powerful tool that..." — just say what it does.

Show the generated description to the user and ask if they want to save it. If they confirm, write it to `.git/description`.

**`generate-branch`:**
Generate a branch-specific description. To do this:
1. Get the current branch name with `git branch --show-current`
2. Run `git log main..HEAD --oneline` (or master..HEAD) to see branch-specific commits
3. Run `git diff main --stat` (or master) to see what files changed

From this context, write a short description of what this branch does (1-2 sentences). Focus on the intent, not the mechanics — "Adds fish shell support to the integrate command" not "Modified 3 files in cmd/".

Show the generated description to the user and ask if they want to save it. If they confirm, write it to `.git/description`.

## Rules
- Keep generated descriptions under 100 characters when possible
- Never use words like "robust", "comprehensive", "streamlined", "leverage"
- Write like a dev would write in a commit message, not a press release
- If you can't figure out what the repo does, just say so and ask the user to provide one
