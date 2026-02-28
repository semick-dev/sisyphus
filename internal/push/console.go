package push

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxRecentLogLines = 5
const codexAnimationHeight = 5
const codexContextHeight = 40

var codexSpinnerFrames = []string{
	`  _o
 /|\
 / \
     O
____/`,
	` \o_
  |\
 / \
    O
___/`,
	`  _o/
 /|
 / \
   O
__/`,
	` _o_
  |\
 / \
  O
_/`,
	` \o
  |\
 / \
 O
/`,
	`  o/
 /|
 / \
O
`,
}

type consoleUI struct {
	mu        sync.Mutex
	repoPath  string
	buildURL  string
	logs      []string
	context   []string
	ctxTitle  string
	success   string
	enabled   bool
	rendered  int
	animating bool
	frameIdx  int
	done      chan struct{}
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func newConsoleUI(repoPath string, buildURL string) *consoleUI {
	ui := &consoleUI{
		repoPath: repoPath,
		buildURL: buildURL,
		logs:     make([]string, 0, maxRecentLogLines),
		context:  make([]string, 0, codexContextHeight),
		ctxTitle: "Codex context:",
		enabled:  isInteractiveTerminal(),
		done:     make(chan struct{}),
	}
	if ui.enabled {
		ui.renderLocked()
	}
	return ui
}

func (c *consoleUI) Close() {
	c.mu.Lock()
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	c.animating = false
	if c.enabled {
		c.renderLocked()
	}
	c.mu.Unlock()
	if c.enabled {
		fmt.Fprintln(os.Stdout)
	}
}

func (c *consoleUI) Log(line string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if !c.enabled {
		fmt.Fprintln(os.Stdout, line)
		return
	}
	c.logs = append(c.logs, line)
	if len(c.logs) > maxRecentLogLines {
		c.logs = c.logs[len(c.logs)-maxRecentLogLines:]
	}
	c.renderLocked()
}

func (c *consoleUI) StartCodexAnimation() {
	c.mu.Lock()
	if !c.enabled || c.animating {
		c.mu.Unlock()
		return
	}
	c.animating = true
	c.renderLocked()
	c.mu.Unlock()

	go func() {
		t := time.NewTicker(120 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-c.done:
				return
			case <-t.C:
				c.mu.Lock()
				if !c.animating {
					c.mu.Unlock()
					return
				}
				c.frameIdx = (c.frameIdx + 1) % len(codexSpinnerFrames)
				c.renderLocked()
				c.mu.Unlock()
			}
		}
	}()
}

func (c *consoleUI) StopCodexAnimation() {
	c.mu.Lock()
	if !c.enabled {
		c.mu.Unlock()
		return
	}
	c.animating = false
	c.renderLocked()
	c.mu.Unlock()
}

func (c *consoleUI) SetContext(title string, text string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if strings.TrimSpace(title) == "" {
		title = "Codex context:"
	}
	c.ctxTitle = title
	c.context = tailLines(text, codexContextHeight)
	if c.enabled {
		c.renderLocked()
	}
}

func (c *consoleUI) ShowSuccess(buildID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.animating = false
	c.logs = make([]string, 0, maxRecentLogLines)
	c.context = make([]string, 0, codexContextHeight)
	c.ctxTitle = "Codex context:"
	c.success = fmt.Sprintf("BUILD %d SUCCEEDED. WOKR COMPLETED ON BUILD %d.", buildID, buildID)
	if c.enabled {
		c.renderLocked()
	}
}

func (c *consoleUI) renderLocked() {
	lines := make([]string, 0, 12)
	lines = append(lines, "Sisyphus Console")
	lines = append(lines, fmt.Sprintf("Running in: %s", c.repoPath))
	lines = append(lines, fmt.Sprintf("Against   : %s", c.buildURL))
	lines = append(lines, "----------------------------------------")
	lines = append(lines, "Recent Events:")
	if c.success != "" {
		lines = append(lines, "")
		lines = append(lines, c.success)
		for i := 0; i < maxRecentLogLines-2; i++ {
			lines = append(lines, "")
		}
		lines = append(lines, "----------------------------------------")
		lines = append(lines, "")
		for i := 0; i < codexAnimationHeight; i++ {
			lines = append(lines, "")
		}
		lines = append(lines, "----------------------------------------")
		lines = append(lines, "")
		for i := 0; i < codexContextHeight; i++ {
			lines = append(lines, "")
		}
	} else {
		for i := 0; i < maxRecentLogLines; i++ {
			if i < len(c.logs) {
				lines = append(lines, c.logs[i])
			} else {
				lines = append(lines, "")
			}
		}
		lines = append(lines, "----------------------------------------")
		if c.animating {
			lines = append(lines, "Codex:")
			frameLines := strings.Split(codexSpinnerFrames[c.frameIdx], "\n")
			for i := 0; i < codexAnimationHeight; i++ {
				if i < len(frameLines) {
					lines = append(lines, frameLines[i])
				} else {
					lines = append(lines, "")
				}
			}
		} else {
			lines = append(lines, "Codex: idle")
			for i := 0; i < codexAnimationHeight; i++ {
				lines = append(lines, "")
			}
		}
		lines = append(lines, "----------------------------------------")
		lines = append(lines, c.ctxTitle)
		for i := 0; i < codexContextHeight; i++ {
			if i < len(c.context) {
				lines = append(lines, c.context[i])
			} else {
				lines = append(lines, "")
			}
		}
	}

	if c.rendered > 0 {
		fmt.Fprintf(os.Stdout, "\033[%dA", c.rendered)
	}
	width := terminalWidth()
	for _, line := range lines {
		fitted := fitLine(line, width)
		if c.success != "" && line == c.success {
			fmt.Fprintf(os.Stdout, "\r\033[K\033[32m%s\033[0m\n", fitted)
			continue
		}
		if strings.HasPrefix(line, "[sisyphus] ") {
			if color := colorForLogLine(line); color != "" {
				fmt.Fprintf(os.Stdout, "\r\033[K%s%s\033[0m\n", color, fitted)
				continue
			}
		}
		fmt.Fprintf(os.Stdout, "\r\033[K%s\n", fitted)
	}
	c.rendered = len(lines)
}

func isInteractiveTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if fi.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	return term != "" && term != "dumb"
}

func tailLines(text string, n int) []string {
	if n <= 0 {
		return []string{}
	}
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	parts := strings.Split(normalized, "\n")
	for len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) > n {
		parts = parts[len(parts)-n:]
	}
	return parts
}

func terminalWidth() int {
	const defaultWidth = 120
	raw := strings.TrimSpace(os.Getenv("COLUMNS"))
	if raw == "" {
		return defaultWidth
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 40 {
		return defaultWidth
	}
	return v
}

func fitLine(line string, width int) string {
	line = strings.ReplaceAll(line, "\t", "    ")
	line = ansiRegexp.ReplaceAllString(line, "")
	if width <= 0 {
		return line
	}
	r := []rune(line)
	if len(r) <= width-1 {
		return line
	}
	if width <= 2 {
		return string(r[:width])
	}
	return string(r[:width-2]) + "…"
}

func colorForLogLine(line string) string {
	l := strings.ToLower(line)

	if strings.Contains(l, "failed") ||
		strings.Contains(l, "error") ||
		strings.Contains(l, "command failed") {
		return "\033[31m" // red
	}

	if strings.Contains(l, "sleeping ") {
		return "\033[90m" // dark gray
	}

	if strings.Contains(l, "status=running") ||
		strings.Contains(l, "status=inprogress") {
		return "\033[34m" // blue
	}

	if strings.Contains(l, "succeeded") ||
		strings.Contains(l, "completed") ||
		strings.Contains(l, "produced changes") ||
		strings.Contains(l, "pushed commit") {
		return "\033[32m" // green
	}

	if strings.Contains(l, "pending") ||
		strings.Contains(l, "waiting") ||
		strings.Contains(l, "queueing") ||
		strings.Contains(l, "queued") ||
		strings.Contains(l, "status=") ||
		strings.Contains(l, "running command") ||
		strings.Contains(l, "invoking codex") ||
		strings.Contains(l, "starting run") {
		return "\033[33m" // yellow
	}

	return ""
}
