# grd

A tiny editor for `.git/description`.

I kept forgetting the path to that file, so I wrote this. It opens a vi-like editor right there, and can also print the description or set up a `git description` alias for you.

## Install

**From source:**

```sh
go install github.com/tcsenpai/grd@latest
```

Or clone and build manually:

```sh
go build -o grd .
```

**Pre-built binary (after `go build`):**

```sh
./install.sh          # installs to ~/.local/bin
./sudo-install.sh     # installs to /usr/local/bin (system-wide)
```

**Cross-compile for all platforms:**

```sh
./build.sh            # outputs to dist/
```

## Usage

```
grd                  Open the vi-like editor for .git/description
grd show             Display the description in a Unicode box
grd show -t          Just the text, no box
grd integrate        Show instructions for shell alias / git subcommand
grd integrate -a     Auto-apply alias and git config changes
```

Run `grd` inside any git repo to edit the description. `:wq` to save, `:q!` to bail.

### Editor keybindings

The editor is modal, like vi. It has Normal, Insert, and Command modes.

**Normal mode**

| Key | Action |
|-----|--------|
| `h j k l` | Move cursor (also arrow keys) |
| `i` | Insert before cursor |
| `a` / `A` | Append after cursor / EOL |
| `I` | Insert at start of line |
| `o` / `O` | New line below / above |
| `x` | Delete char under cursor |
| `dd` | Delete current line |
| `0` / `$` | Start / end of line |
| `:` | Enter command mode |

**Command mode** (after `:`)

| Command | Action |
|---------|--------|
| `:w` | Save |
| `:q` | Quit (fails if unsaved) |
| `:q!` | Force quit, discard changes |
| `:wq` / `:x` | Save and quit |
| `Esc` | Back to normal mode |

## Shell integration

`grd integrate` tells you what to put in your rc file. `grd integrate -a` just does it for you (appends to bashrc/zshrc/config.fish and sets `git config --global alias.description`). After that you can use `git description` from anywhere.

## Claude Code slash command

If you use [Claude Code](https://github.com/anthropics/claude-code), there's a standalone slash command included (`grd-command.md`) that does the same thing without the binary — plus it can generate descriptions from your repo context.

Copy it to your commands directory:

```sh
cp grd-command.md ~/.claude/commands/grd.md    # global
# or
cp grd-command.md .claude/commands/grd.md      # per-project
```

Then use `/grd show`, `/grd generate`, `/grd generate-branch`, or `/grd set my project does X`.

## License

MIT
