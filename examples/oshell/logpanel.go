// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// ringBufferSize is the maximum number of log entries kept in memory.
const ringBufferSize = 200

// logEntry holds a formatted log record for display.
type logEntry struct {
	time    time.Time
	level   slog.Level
	message string
	attrs   string
}

// logEntryMsg is sent to the Bubble Tea update loop when a new log entry
// arrives in the ring buffer.
type logEntryMsg struct {
	entry logEntry
}

func fetchSandboxLogs(client v1.ClientInterface, sandboxName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result, err := client.Sandboxes().GetLogs(ctx, sandboxName, v1.WithLogLines(100))
		if err != nil {
			return sandboxLogMsg{sandboxName: sandboxName, err: err}
		}
		return sandboxLogMsg{sandboxName: sandboxName, lines: result.Lines}
	}
}

type logMode int

const (
	logModeGlobal  logMode = iota
	logModeSandbox
)

// logModeMsg tells the log panel to switch modes.
type logModeMsg struct {
	mode        logMode
	sandboxName string
}

// sandboxLogMsg carries the result of a GetLogs call.
type sandboxLogMsg struct {
	sandboxName string
	lines       []types.LogLine
	err         error
}

// ringBuffer is a fixed-size circular buffer of log entries with thread-safe
// access. When the buffer is full, the oldest entry is overwritten.
type ringBuffer struct {
	mu      sync.Mutex
	entries []logEntry
	cursor  int
	count   int
}

// newRingBuffer creates a ring buffer with the given capacity.
func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{
		entries: make([]logEntry, size),
	}
}

// push adds a log entry to the ring buffer.
func (rb *ringBuffer) push(e logEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.entries[rb.cursor] = e
	rb.cursor = (rb.cursor + 1) % len(rb.entries)
	if rb.count < len(rb.entries) {
		rb.count++
	}
}

// snapshot returns a copy of all entries in chronological order.
func (rb *ringBuffer) snapshot() []logEntry {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	result := make([]logEntry, 0, rb.count)
	if rb.count < len(rb.entries) {
		// Buffer not yet full, entries are at indices [0..count).
		result = append(result, rb.entries[:rb.count]...)
	} else {
		// Buffer full, oldest is at cursor, wrap around.
		result = append(result, rb.entries[rb.cursor:]...)
		result = append(result, rb.entries[:rb.cursor]...)
	}
	return result
}

// teeHandler is a slog.Handler that writes log records to a ring buffer
// and optionally to a file-based JSON handler. It sends a tea.Msg for each
// new entry so the TUI can update.
type teeHandler struct {
	ring     *ringBuffer
	fileH    slog.Handler // optional, may be nil
	program  *tea.Program
	mu       sync.Mutex
	attrs    []slog.Attr
	groups   []string
}

// newTeeHandler creates a tee handler writing to the ring buffer and
// optionally to a file handler. The program field can be set later via
// SetProgram once the Bubble Tea program is running.
func newTeeHandler(ring *ringBuffer, fileHandler slog.Handler) *teeHandler {
	return &teeHandler{
		ring:  ring,
		fileH: fileHandler,
	}
}

// SetProgram sets the Bubble Tea program reference so the handler can
// send messages to the TUI update loop.
func (h *teeHandler) SetProgram(p *tea.Program) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.program = p
}

// Enabled reports whether the handler handles records at the given level.
func (h *teeHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle processes a log record by adding it to the ring buffer and
// optionally writing to the file handler.
func (h *teeHandler) Handle(_ context.Context, r slog.Record) error {
	// Build attribute string from record and handler attrs.
	var attrParts []string
	for _, a := range h.attrs {
		attrParts = append(attrParts, fmt.Sprintf("%s=%v", a.Key, a.Value))
	}
	r.Attrs(func(a slog.Attr) bool {
		attrParts = append(attrParts, fmt.Sprintf("%s=%v", a.Key, a.Value))
		return true
	})

	entry := logEntry{
		time:    r.Time,
		level:   r.Level,
		message: r.Message,
		attrs:   strings.Join(attrParts, " "),
	}

	h.ring.push(entry)

	// Send to TUI if program is set.
	h.mu.Lock()
	p := h.program
	h.mu.Unlock()
	if p != nil {
		go p.Send(logEntryMsg{entry: entry})
	}

	// Forward to file handler if present.
	if h.fileH != nil {
		// Ignore file write errors gracefully (per spec).
		_ = h.fileH.Handle(context.Background(), r)
	}

	return nil
}

// WithAttrs returns a new handler with the given attributes added.
func (h *teeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	var newFileH slog.Handler
	if h.fileH != nil {
		newFileH = h.fileH.WithAttrs(attrs)
	}

	h.mu.Lock()
	prog := h.program
	h.mu.Unlock()

	return &teeHandler{
		ring:    h.ring,
		fileH:   newFileH,
		program: prog,
		attrs:   newAttrs,
		groups:  h.groups,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *teeHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	var newFileH slog.Handler
	if h.fileH != nil {
		newFileH = h.fileH.WithGroup(name)
	}

	h.mu.Lock()
	prog := h.program
	h.mu.Unlock()

	return &teeHandler{
		ring:    h.ring,
		fileH:   newFileH,
		program: prog,
		attrs:   h.attrs,
		groups:  newGroups,
	}
}

// LogPanel is a Bubble Tea component that displays log entries from the
// ring buffer in a scrollable viewport with color-coded log levels.
type LogPanel struct {
	ring       *ringBuffer
	entries    []logEntry
	scrollPos  int
	autoScroll bool

	mode           logMode
	sandboxName    string
	sandboxLogs    []types.LogLine
	sandboxErr     string
	sandboxLoading bool
}

// NewLogPanel creates a new LogPanel connected to the given ring buffer.
func NewLogPanel(ring *ringBuffer) *LogPanel {
	return &LogPanel{
		ring:       ring,
		entries:    nil,
		scrollPos:  0,
		autoScroll: true,
	}
}

// Init returns the initial command for the log panel.
func (lp *LogPanel) Init() tea.Cmd {
	// Load any existing entries from the ring buffer.
	lp.entries = lp.ring.snapshot()
	return nil
}

// HandleMsg handles messages for the log panel. LogPanel is not a tea.Model;
// it is an internal component managed by the Dashboard.
func (lp *LogPanel) HandleMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case logEntryMsg:
		if lp.mode == logModeGlobal {
			lp.entries = lp.ring.snapshot()
			if lp.autoScroll {
				lp.scrollPos = len(lp.entries)
			}
		}
	case logModeMsg:
		lp.mode = msg.mode
		lp.sandboxName = msg.sandboxName
		if msg.mode == logModeGlobal {
			lp.sandboxLogs = nil
			lp.sandboxErr = ""
			lp.sandboxLoading = false
			lp.entries = lp.ring.snapshot()
		} else {
			lp.sandboxLoading = true
			lp.sandboxLogs = nil
			lp.sandboxErr = ""
		}
	case sandboxLogMsg:
		if lp.mode == logModeSandbox && lp.sandboxName == msg.sandboxName {
			lp.sandboxLoading = false
			if msg.err != nil {
				lp.sandboxErr = msg.err.Error()
			} else {
				lp.sandboxErr = ""
				lp.sandboxLogs = msg.lines
			}
		}
	}
	return nil
}

// View renders the log panel content with borders and level coloring.
func (lp *LogPanel) View(width, height int, focused bool) string {
	if height < 1 {
		height = 1
	}

	borderStyle := logPanelBorderStyle
	if focused {
		borderStyle = logPanelFocusedBorderStyle
	}

	if lp.mode == logModeSandbox {
		return lp.viewSandboxLogs(width, height, borderStyle)
	}

	visibleCount := height
	startIdx := len(lp.entries) - visibleCount
	if startIdx < 0 {
		startIdx = 0
	}
	if lp.scrollPos < len(lp.entries) && !lp.autoScroll {
		startIdx = lp.scrollPos
		if startIdx+visibleCount > len(lp.entries) {
			startIdx = len(lp.entries) - visibleCount
		}
		if startIdx < 0 {
			startIdx = 0
		}
	}

	endIdx := startIdx + visibleCount
	if endIdx > len(lp.entries) {
		endIdx = len(lp.entries)
	}

	var lines []string
	for i := startIdx; i < endIdx; i++ {
		e := lp.entries[i]
		lines = append(lines, lp.formatEntry(e, width-4))
	}

	for len(lines) < visibleCount {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return borderStyle.Width(width - 2).Render(content)
}

func (lp *LogPanel) viewSandboxLogs(width, height int, borderStyle lipgloss.Style) string {
	modeLabel := lipgloss.NewStyle().Foreground(colorCyan).Bold(true).
		Render(fmt.Sprintf("Logs: %s", lp.sandboxName))

	var lines []string
	lines = append(lines, modeLabel)

	if lp.sandboxLoading {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorGray).Render("Loading..."))
	} else if lp.sandboxErr != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorRed).Render(lp.sandboxErr))
	} else if len(lp.sandboxLogs) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorGray).Render("No log entries"))
	} else {
		maxLines := height - 1
		start := len(lp.sandboxLogs) - maxLines
		if start < 0 {
			start = 0
		}
		for _, ll := range lp.sandboxLogs[start:] {
			lines = append(lines, lp.formatLogLine(ll, width-4))
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return borderStyle.Width(width - 2).Render(content)
}

func (lp *LogPanel) formatLogLine(ll types.LogLine, maxWidth int) string {
	timeStr := ll.Timestamp.Format("15:04:05")
	var levelStyle lipgloss.Style
	switch strings.ToUpper(ll.Level) {
	case "ERROR":
		levelStyle = logLevelError
	case "WARN", "WARNING":
		levelStyle = logLevelWarn
	case "INFO":
		levelStyle = logLevelInfo
	default:
		levelStyle = logLevelDebug
	}

	line := fmt.Sprintf("%s %s [%s] %s",
		logLevelDebug.Render(timeStr),
		levelStyle.Render(fmt.Sprintf("%-5s", ll.Level)),
		lipgloss.NewStyle().Foreground(colorGray).Render(ll.Target),
		ll.Message,
	)

	if maxWidth > 0 && ansi.StringWidth(line) > maxWidth {
		line = ansi.Truncate(line, maxWidth, "")
	}
	return line
}

// formatEntry formats a single log entry with level coloring.
func (lp *LogPanel) formatEntry(e logEntry, maxWidth int) string {
	timeStr := e.time.Format("15:04:05")
	levelStr := e.level.String()

	var levelStyle lipgloss.Style
	switch {
	case e.level >= slog.LevelError:
		levelStyle = logLevelError
	case e.level >= slog.LevelWarn:
		levelStyle = logLevelWarn
	case e.level >= slog.LevelInfo:
		levelStyle = logLevelInfo
	default:
		levelStyle = logLevelDebug
	}

	line := fmt.Sprintf("%s %s %s",
		logLevelDebug.Render(timeStr),
		levelStyle.Render(fmt.Sprintf("%-5s", levelStr)),
		e.message,
	)

	if e.attrs != "" {
		line += " " + logLevelDebug.Render(e.attrs)
	}

	// Truncate to available display width using ANSI-aware truncation
	// to avoid cutting escape sequences mid-sequence.
	if maxWidth > 0 && ansi.StringWidth(line) > maxWidth {
		line = ansi.Truncate(line, maxWidth, "")
	}

	return line
}

// ScrollUp scrolls the log panel up by one line.
func (lp *LogPanel) ScrollUp() {
	lp.autoScroll = false
	if lp.scrollPos > 0 {
		lp.scrollPos--
	}
}

// ScrollDown scrolls the log panel down by one line.
func (lp *LogPanel) ScrollDown() {
	if lp.scrollPos < len(lp.entries)-1 {
		lp.scrollPos++
	}
	// Re-enable auto-scroll if at the bottom.
	if lp.scrollPos >= len(lp.entries)-1 {
		lp.autoScroll = true
	}
}
