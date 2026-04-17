package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const descPath = ".git/description"

const defaultGitDescription = "Unnamed repository; edit this file 'description' to name the repository."

func readDescription() (string, error) {
	data, err := os.ReadFile(descPath)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\n"), nil
}

func isDefaultDescription(s string) bool {
	return strings.TrimSpace(s) == "" || strings.TrimSpace(s) == defaultGitDescription
}

func cmdShow(plainText bool) {
	content, err := readDescription()
	if err != nil || isDefaultDescription(content) {
		fmt.Println("No description set. Run `grd` to set one.")
		return
	}

	if plainText {
		fmt.Println(content)
		return
	}

	lines := strings.Split(content, "\n")

	maxLen := 0
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}

	const title = " Repository Description "
	// Box inner width: at least wide enough for the title plus a couple of dashes.
	innerWidth := maxLen + 2 // 1 space padding each side
	minInner := len(title) + 4
	if innerWidth < minInner {
		innerWidth = minInner
	}

	rightDashes := innerWidth - len(title) - 1
	top := "┌─" + title + strings.Repeat("─", rightDashes) + "┐"
	fmt.Println(top)

	fmt.Println("│" + strings.Repeat(" ", innerWidth) + "│")

	for _, l := range lines {
		padding := innerWidth - len(l) - 1
		fmt.Printf("│ %s%s│\n", l, strings.Repeat(" ", padding))
	}

	fmt.Println("│" + strings.Repeat(" ", innerWidth) + "│")

	bottom := "└" + strings.Repeat("─", innerWidth) + "┘"
	fmt.Println(bottom)
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}
	base := filepath.Base(shell)
	return base
}

func rcFileForShell(shell string) string {
	home, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return filepath.Join(home, ".bashrc")
	}
}

func cmdIntegrate(autoApply bool) {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not determine executable path:", err)
		os.Exit(1)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	shell := detectShell()
	rcFile := rcFileForShell(shell)

	var aliasLine string
	var instructions string

	switch shell {
	case "fish":
		aliasLine = fmt.Sprintf("alias grd '%s'", exe)
		instructions = fmt.Sprintf(`Add this to your %s:

  alias grd '%s'

  # Or as a git subcommand (works in any shell):
  git config --global alias.description '!%s'
`, rcFile, exe, exe)
	default:
		aliasLine = fmt.Sprintf("alias grd='%s'", exe)
		instructions = fmt.Sprintf(`Add this to your %s:

  alias grd='%s'

  # Or as a git subcommand (works in any shell):
  git config --global alias.description '!%s'
`, rcFile, exe, exe)
	}

	if !autoApply {
		fmt.Print(instructions)
		fmt.Println("Run `grd integrate -a` to apply these changes automatically.")
		return
	}

	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not open", rcFile, ":", err)
		os.Exit(1)
	}
	_, werr := fmt.Fprintf(f, "\n# Added by grd integrate\n%s\n", aliasLine)
	f.Close()
	if werr != nil {
		fmt.Fprintln(os.Stderr, "error: could not write to", rcFile, ":", werr)
		os.Exit(1)
	}
	fmt.Printf("Appended alias to %s\n", rcFile)

	gitAlias := fmt.Sprintf("!%s", exe)
	cmd := exec.Command("git", "config", "--global", "alias.description", gitAlias)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not set git global alias:", err)
	} else {
		fmt.Println("Set git global alias: git description ->", exe)
	}

	fmt.Println("\nDone. Restart your shell or source the rc file to use the alias.")
}

type Mode int

const (
	ModeNormal  Mode = iota
	ModeInsert
	ModeCommand
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeCommand:
		return "COMMAND"
	}
	return "NORMAL"
}

type Editor struct {
	screen   tcell.Screen
	lines    []string
	cx, cy   int // cursor col, row
	mode     Mode
	modified bool
	filepath string
	cmdBuf   string
	message  string
	// for dd detection
	lastKey rune
}

func newEditor(filepath string, content string) (*Editor, error) {
	sc, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := sc.Init(); err != nil {
		return nil, err
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	return &Editor{
		screen:   sc,
		lines:    lines,
		filepath: filepath,
	}, nil
}

func (e *Editor) close() {
	e.screen.Fini()
}

func (e *Editor) clampCursor() {
	if e.cy < 0 {
		e.cy = 0
	}
	if e.cy >= len(e.lines) {
		e.cy = len(e.lines) - 1
	}
	lineLen := len(e.lines[e.cy])
	if e.mode == ModeNormal {
		if lineLen == 0 {
			e.cx = 0
		} else if e.cx >= lineLen {
			e.cx = lineLen - 1
		}
	} else {
		if e.cx > lineLen {
			e.cx = lineLen
		}
	}
	if e.cx < 0 {
		e.cx = 0
	}
}

func (e *Editor) draw() {
	e.screen.Clear()
	w, h := e.screen.Size()

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	statusStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	cmdStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorYellow)
	msgStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorRed)

	textRows := h - 1 // last row is the status bar
	for row := 0; row < textRows; row++ {
		if row < len(e.lines) {
			for col, ch := range e.lines[row] {
				e.screen.SetContent(col, row, ch, nil, defStyle)
			}
		}
	}

	modifiedMark := ""
	if e.modified {
		modifiedMark = "[+] "
	}
	statusText := fmt.Sprintf(" %s | %s%s ", e.mode.String(), modifiedMark, e.filepath)
	for len(statusText) < w {
		statusText += " "
	}
	if len(statusText) > w {
		statusText = statusText[:w]
	}
	for col, ch := range statusText {
		e.screen.SetContent(col, h-1, ch, nil, statusStyle)
	}

	if e.mode == ModeCommand {
		cmdLine := ":" + e.cmdBuf
		for col, ch := range cmdLine {
			e.screen.SetContent(col, h-1, ch, nil, cmdStyle)
		}
	} else if e.message != "" {
		for col, ch := range e.message {
			e.screen.SetContent(col, h-1, ch, nil, msgStyle)
		}
	}

	if e.mode == ModeCommand {
		e.screen.ShowCursor(1+len(e.cmdBuf), h-1)
	} else {
		e.screen.ShowCursor(e.cx, e.cy)
	}

	e.screen.Show()
}

func (e *Editor) save() error {
	content := strings.Join(e.lines, "\n")
	if err := os.WriteFile(e.filepath, []byte(content), 0644); err != nil {
		return err
	}
	e.modified = false
	e.message = "Description saved."
	return nil
}

func (e *Editor) insertChar(ch rune) {
	line := e.lines[e.cy]
	newLine := line[:e.cx] + string(ch) + line[e.cx:]
	e.lines[e.cy] = newLine
	e.cx++
	e.modified = true
}

func (e *Editor) deleteCharBefore() {
	if e.cx == 0 {
		if e.cy == 0 {
			return
		}
		prevLine := e.lines[e.cy-1]
		e.cx = len(prevLine)
		e.lines[e.cy-1] = prevLine + e.lines[e.cy]
		e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
		e.cy--
		e.modified = true
		return
	}
	line := e.lines[e.cy]
	e.lines[e.cy] = line[:e.cx-1] + line[e.cx:]
	e.cx--
	e.modified = true
}

func (e *Editor) deleteCharUnder() {
	line := e.lines[e.cy]
	if len(line) == 0 {
		return
	}
	if e.cx >= len(line) {
		e.cx = len(line) - 1
	}
	e.lines[e.cy] = line[:e.cx] + line[e.cx+1:]
	e.modified = true
	e.clampCursor()
}

func (e *Editor) deleteLine() {
	if len(e.lines) == 1 {
		e.lines[0] = ""
		e.cx = 0
		e.modified = true
		return
	}
	e.lines = append(e.lines[:e.cy], e.lines[e.cy+1:]...)
	e.modified = true
	e.clampCursor()
}

func (e *Editor) openLineBelow() {
	newLines := make([]string, len(e.lines)+1)
	copy(newLines, e.lines[:e.cy+1])
	newLines[e.cy+1] = ""
	copy(newLines[e.cy+2:], e.lines[e.cy+1:])
	e.lines = newLines
	e.cy++
	e.cx = 0
	e.mode = ModeInsert
	e.modified = true
}

func (e *Editor) openLineAbove() {
	newLines := make([]string, len(e.lines)+1)
	copy(newLines, e.lines[:e.cy])
	newLines[e.cy] = ""
	copy(newLines[e.cy+1:], e.lines[e.cy:])
	e.lines = newLines
	e.cx = 0
	e.mode = ModeInsert
	e.modified = true
}

func (e *Editor) newlineAtCursor() {
	line := e.lines[e.cy]
	before := line[:e.cx]
	after := line[e.cx:]
	e.lines[e.cy] = before
	newLines := make([]string, len(e.lines)+1)
	copy(newLines, e.lines[:e.cy+1])
	newLines[e.cy+1] = after
	copy(newLines[e.cy+2:], e.lines[e.cy+1:])
	e.lines = newLines
	e.cy++
	e.cx = 0
	e.modified = true
}

func (e *Editor) handleNormal(ev *tcell.EventKey) (quit bool) {
	e.message = ""
	ch := ev.Rune()
	key := ev.Key()

	switch {
	case key == tcell.KeyEscape:
		e.lastKey = 0

	case ch == 'h' || key == tcell.KeyLeft:
		e.cx--
		e.lastKey = 0

	case ch == 'l' || key == tcell.KeyRight:
		e.cx++
		e.lastKey = 0

	case ch == 'k' || key == tcell.KeyUp:
		e.cy--
		e.lastKey = 0

	case ch == 'j' || key == tcell.KeyDown:
		e.cy++
		e.lastKey = 0

	case ch == 'i':
		e.mode = ModeInsert
		e.lastKey = 0

	case ch == 'a':
		if len(e.lines[e.cy]) > 0 {
			e.cx++
		}
		e.mode = ModeInsert
		e.lastKey = 0

	case ch == 'A':
		e.cx = len(e.lines[e.cy])
		e.mode = ModeInsert
		e.lastKey = 0

	case ch == 'I':
		e.cx = 0
		e.mode = ModeInsert
		e.lastKey = 0

	case ch == 'o':
		e.openLineBelow()
		e.lastKey = 0

	case ch == 'O':
		e.openLineAbove()
		e.lastKey = 0

	case ch == 'x':
		e.deleteCharUnder()
		e.lastKey = 0

	case ch == 'd':
		if e.lastKey == 'd' {
			e.deleteLine()
			e.lastKey = 0
		} else {
			e.lastKey = 'd'
		}

	case ch == '0':
		e.cx = 0
		e.lastKey = 0

	case ch == '$':
		if len(e.lines[e.cy]) > 0 {
			e.cx = len(e.lines[e.cy]) - 1
		}
		e.lastKey = 0

	case ch == ':':
		e.mode = ModeCommand
		e.cmdBuf = ""
		e.lastKey = 0

	default:
		e.lastKey = 0
	}

	e.clampCursor()
	return false
}

func (e *Editor) handleInsert(ev *tcell.EventKey) (quit bool) {
	e.message = ""
	key := ev.Key()

	switch key {
	case tcell.KeyEscape:
		e.mode = ModeNormal
		if e.cx > 0 {
			e.cx--
		}
		e.clampCursor()

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.deleteCharBefore()

	case tcell.KeyEnter:
		e.newlineAtCursor()

	case tcell.KeyLeft:
		e.cx--
		e.clampCursor()

	case tcell.KeyRight:
		e.cx++
		e.clampCursor()

	case tcell.KeyUp:
		e.cy--
		e.clampCursor()

	case tcell.KeyDown:
		e.cy++
		e.clampCursor()

	case tcell.KeyTab:
		e.insertChar('\t')

	default:
		if ev.Rune() != 0 {
			e.insertChar(ev.Rune())
		}
	}
	return false
}

func (e *Editor) handleCommand(ev *tcell.EventKey) (quit bool) {
	key := ev.Key()

	switch key {
	case tcell.KeyEscape:
		e.mode = ModeNormal
		e.cmdBuf = ""

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.cmdBuf) > 0 {
			e.cmdBuf = e.cmdBuf[:len(e.cmdBuf)-1]
		} else {
			e.mode = ModeNormal
		}

	case tcell.KeyEnter:
		cmd := strings.TrimSpace(e.cmdBuf)
		e.cmdBuf = ""
		e.mode = ModeNormal
		switch cmd {
		case "w":
			if err := e.save(); err != nil {
				e.message = "Error saving: " + err.Error()
			}
		case "q":
			if e.modified {
				e.message = "Unsaved changes (use :q! to force)"
			} else {
				return true
			}
		case "q!":
			return true
		case "wq", "x", "x!":
			if err := e.save(); err != nil {
				e.message = "Error saving: " + err.Error()
			} else {
				return true
			}
		default:
			e.message = "Unknown command: " + cmd
		}

	default:
		if ev.Rune() != 0 {
			e.cmdBuf += string(ev.Rune())
		}
	}
	return false
}

func (e *Editor) run() {
	for {
		e.draw()
		ev := e.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			var quit bool
			switch e.mode {
			case ModeNormal:
				quit = e.handleNormal(ev)
			case ModeInsert:
				quit = e.handleInsert(ev)
			case ModeCommand:
				quit = e.handleCommand(ev)
			}
			if quit {
				return
			}
		case *tcell.EventResize:
			e.screen.Sync()
		}
	}
}

func requireGitRepo() {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "error: not a git repository (no .git directory found)")
		os.Exit(1)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "show":
			requireGitRepo()
			plainText := len(args) > 1 && args[1] == "-t"
			cmdShow(plainText)
			return

		case "integrate":
			autoApply := len(args) > 1 && args[1] == "-a"
			cmdIntegrate(autoApply)
			return

		case "help", "--help", "-h":
			fmt.Println(`grd — git repository description editor

Usage:
  grd                  Open the vi-like editor for .git/description
  grd show             Display the description in a Unicode box
  grd show -t          Print raw text, no box
  grd integrate        Show instructions for shell alias / git subcommand
  grd integrate -a     Auto-apply alias and git config changes
  grd help             Show this help`)
			return
		}
	}

	requireGitRepo()

	content := ""
	data, err := os.ReadFile(descPath)
	if err == nil {
		content = string(data)
		content = strings.TrimRight(content, "\n")
	}

	editor, err := newEditor(descPath, content)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error initializing editor:", err)
		os.Exit(1)
	}
	defer editor.close()

	editor.run()
}
