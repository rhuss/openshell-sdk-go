// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/charmbracelet/x/term"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// sandboxListMsg carries the result of a sandbox list operation.
type sandboxListMsg struct {
	sandboxes []*types.Sandbox
	err       error
}

// sandboxWatchStartedMsg carries the watcher handle back to the Update
// goroutine, avoiding a data race from assigning st.watcher inside a Cmd.
type sandboxWatchStartedMsg struct {
	watcher v1.WatchInterface[*types.Sandbox]
}

// sandboxWatchEventMsg carries a single watch event for a sandbox.
type sandboxWatchEventMsg struct {
	event types.Event[*types.Sandbox]
}

// sandboxWatchErrMsg signals that the watch stream has broken.
// sandboxWatchReconnectMsg triggers a delayed watch reconnection.
type sandboxWatchReconnectMsg struct{}

type sandboxRefreshTickMsg struct{}

type sandboxWatchErrMsg struct {
	err error
}

// sandboxCreateMsg carries the result of a sandbox create operation.
type sandboxCreateMsg struct {
	sandbox *types.Sandbox
	err     error
}

// sandboxDeleteMsg carries the result of a sandbox delete operation.
type sandboxDeleteMsg struct {
	name string
	err  error
}

// sandboxDetailMsg carries the result of fetching detail info for a sandbox.
type sandboxDetailMsg struct {
	name         string
	policyStatus *types.PolicyStatusResult
	policyErr    error
}

// execResultMsg carries the result of a command execution.
type execResultMsg struct {
	entry ExecEntry
}

// ExecEntry records one command execution with its results.
type ExecEntry struct {
	Sandbox  string
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Err      error
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type execSpinnerTickMsg struct{}

type shellSessionMsg struct {
	session v1.InteractiveSession
	err     error
}

type shellFinishedMsg struct {
	err error
}

// createField tracks which text input has focus in the create dialog.
type createField int

const (
	createFieldName  createField = iota
	createFieldImage createField = iota
)

// SandboxTab displays a table of sandboxes with colored phase indicators
// and supports CRUD operations.
type SandboxTab struct {
	client v1.ClientInterface
	logger *slog.Logger
	table  table.Model

	sandboxes []*types.Sandbox
	loading   bool
	errMsg    string

	// watcher holds the active sandbox watch stream. Nil when no watch is
	// running. The watcher is stopped on cleanup and re-established after
	// stream breaks.
	watcher types.WatchInterface[*types.Sandbox]

	// showCreate indicates whether the create dialog is visible.
	showCreate bool
	// createNameInput is the text input for the sandbox name.
	createNameInput textinput.Model
	// createImageInput is the text input for the container image.
	createImageInput textinput.Model
	// createFocus tracks which field has focus in the create dialog.
	createFocus createField

	// showDelete indicates whether the delete confirmation is visible.
	showDelete bool
	// deleteName is the name of the sandbox pending deletion confirmation.
	deleteName string
	// showDetail indicates whether the detail popup is visible.
	showDetail bool
	detailIdx  int
	detailName string
	detailPolicy  *types.PolicyStatusResult
	detailLoading bool
	detailErr     string

	// showExec indicates whether the exec popup is visible.
	showExec           bool
	execSandboxName    string
	execCommandInput   textinput.Model
	execHistory        []ExecEntry
	execRunning        bool
	execSpinnerFrame   int
	execRunningCommand string
	execScrollOffset   int

	// Sort state for the sandbox table.
	sortColumn int  // 0=Name, 1=Phase, 2=Image, 3=Created
	sortAsc    bool

	// Gateway health cache (from healthCheckMsg broadcast).
	gatewayHealthy *bool
	gatewayVersion string

	width  int
	height int
}

// NewSandboxTab creates a new SandboxTab connected to the SDK client.
func NewSandboxTab(client v1.ClientInterface, logger *slog.Logger) *SandboxTab {
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Phase", Width: 15},
		{Title: "Image", Width: 30},
		{Title: "Created", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorGray).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(colorWhite).
		Background(lipgloss.Color("8")).
		Bold(true)
	t.SetStyles(s)

	nameInput := textinput.New()
	nameInput.Placeholder = "my-sandbox"
	nameInput.CharLimit = 63

	imageInput := textinput.New()
	imageInput.Placeholder = "python:3.12-slim"
	imageInput.CharLimit = 255

	return &SandboxTab{
		client:           client,
		logger:           logger,
		table:            t,
		loading:          true,
		detailIdx:        -1,
		sortAsc:          true,
		createNameInput:  nameInput,
		createImageInput: imageInput,
	}
}

// Title returns the display title for the tab bar.
func (st *SandboxTab) Title() string {
	return "Sandboxes"
}

// Init returns the initial command that loads sandboxes.
func (st *SandboxTab) Init() tea.Cmd {
	return st.listSandboxes()
}

// Update handles messages for the SandboxTab.
func (st *SandboxTab) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case sandboxListMsg:
		st.loading = false
		if msg.err != nil {
			st.errMsg = fmt.Sprintf("Error loading sandboxes: %v", msg.err)
			st.logger.Error("failed to list sandboxes", "error", msg.err)
			return st, NewSDKErrorMsg(msg.err, "sandboxes.list")
		}
		st.errMsg = ""
		st.sandboxes = msg.sandboxes
		st.updateTableRows()
		// Schedule a periodic refresh instead of using Watch, since the
		// gateway requires a specific sandbox name for watch streams.
		return st, st.scheduleRefresh()

	case sandboxWatchStartedMsg:
		st.watcher = msg.watcher
		return st, st.readNextWatchEvent()

	case sandboxWatchEventMsg:
		st.handleWatchEvent(msg.event)
		st.updateTableRows()
		// Continue reading the next event from the watch stream.
		return st, st.readNextWatchEvent()

	case sandboxWatchErrMsg:
		st.logger.Error("sandbox watch stream error, will reconnect after delay", "error", msg.err)
		st.stopWatch()
		// Delay before re-listing to avoid a tight reconnect loop when the
		// gateway is persistently unavailable for watch streams.
		return st, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
			return sandboxWatchReconnectMsg{}
		})

	case sandboxWatchReconnectMsg:
		st.logger.Info("reconnecting sandbox watcher")
		return st, st.listSandboxes()

	case sandboxRefreshTickMsg:
		return st, st.listSandboxes()

	case sandboxCreateMsg:
		if msg.err != nil {
			st.errMsg = fmt.Sprintf("Error creating sandbox: %v", msg.err)
			st.logger.Error("failed to create sandbox", "error", msg.err)
		} else {
			st.logger.Info("sandbox created", "name", msg.sandbox.Name)
		}
		return st, nil

	case sandboxDeleteMsg:
		if msg.err != nil {
			st.errMsg = fmt.Sprintf("Error deleting sandbox: %v", msg.err)
			st.logger.Error("failed to delete sandbox", "error", msg.err, "name", msg.name)
		} else {
			st.logger.Info("sandbox deleted", "name", msg.name)
		}
		return st, nil

	case sandboxDetailMsg:
		// Only apply if the detail is still open for the same sandbox.
		if st.detailIdx >= 0 && st.detailName == msg.name {
			st.detailLoading = false
			if msg.policyErr != nil {
				st.detailErr = fmt.Sprintf("Policy: %v", msg.policyErr)
				st.logger.Error("failed to fetch policy status", "error", msg.policyErr, "sandbox", msg.name)
			} else {
				st.detailErr = ""
				st.detailPolicy = msg.policyStatus
			}
		}
		return st, nil

	case healthCheckMsg:
		if msg.err == nil && msg.result != nil {
			h := msg.result.Healthy
			st.gatewayHealthy = &h
			st.gatewayVersion = msg.result.Version
		} else {
			f := false
			st.gatewayHealthy = &f
			st.gatewayVersion = ""
		}
		return st, nil

	case execResultMsg:
		st.execRunning = false
		st.execRunningCommand = ""
		st.execHistory = append(st.execHistory, msg.entry)
		st.execScrollOffset = 0
		return st, nil

	case execSpinnerTickMsg:
		if st.execRunning {
			st.execSpinnerFrame = (st.execSpinnerFrame + 1) % len(spinnerFrames)
			return st, st.execSpinnerTick()
		}
		return st, nil

	case shellSessionMsg:
		if msg.err != nil {
			st.logger.Error("failed to start shell", "error", msg.err)
			return st, nil
		}
		cmd := &interactiveShellCmd{session: msg.session}
		return st, tea.Exec(cmd, func(err error) tea.Msg {
			return shellFinishedMsg{err: err}
		})

	case shellFinishedMsg:
		if msg.err != nil {
			st.logger.Error("shell session ended with error", "error", msg.err)
		} else {
			st.logger.Info("shell session ended")
		}
		return st, st.listSandboxes()

	case tea.MouseWheelMsg:
		if st.showExec {
			return st.updateExecPopup(msg)
		}
		return st, nil

	case tea.KeyPressMsg:
		if st.showExec {
			return st.updateExecPopup(msg)
		}
		if st.showCreate {
			return st.updateCreateDialog(msg)
		}
		if st.showDelete {
			return st.updateDeleteConfirm(msg)
		}
		if st.showDetail {
			switch msg.String() {
			case "esc":
				st.closeDetail()
				return st, nil
			case "x":
				idx := st.detailIdx
				if idx >= 0 && idx < len(st.sandboxes) && st.sandboxes[idx].Status.Phase == types.SandboxReady {
					sb := st.sandboxes[idx]
					st.closeDetail()
					st.showExec = true
					st.execSandboxName = sb.Name
					st.execCommandInput = textinput.New()
					st.execCommandInput.Prompt = "$ "
					st.execCommandInput.Placeholder = ""
					st.execCommandInput.CharLimit = 1024
					return st, st.execCommandInput.Focus()
				}
				// Not Ready: stay in detail popup.
				return st, nil
			case "l":
				// Keep detail open; log panel updates below it.
				name := st.detailName
				return st, func() tea.Msg {
					return logModeMsg{mode: logModeSandbox, sandboxName: name}
				}
			case "t":
				idx := st.detailIdx
				if idx >= 0 && idx < len(st.sandboxes) && st.sandboxes[idx].Status.Phase == types.SandboxReady {
					sb := st.sandboxes[idx]
					st.closeDetail()
					return st, st.startShell(sb.Name)
				}
				return st, nil
			}
			return st, nil
		}
		if msg.String() == "c" && !st.loading {
			return st.openCreateDialog()
		}
		if msg.String() == "d" && !st.loading && len(st.sandboxes) > 0 {
			row := st.table.SelectedRow()
			if len(row) > 0 {
				st.showDelete = true
				st.deleteName = row[0]
			}
			return st, nil
		}
		if msg.String() == "x" && !st.loading && len(st.sandboxes) > 0 {
			return st.openExecPopup()
		}
		if msg.String() == "t" && !st.loading && len(st.sandboxes) > 0 {
			cursor := st.table.Cursor()
			if cursor >= 0 && cursor < len(st.sandboxes) && st.sandboxes[cursor].Status.Phase == types.SandboxReady {
				return st, st.startShell(st.sandboxes[cursor].Name)
			}
		}
		if msg.String() == "l" && !st.loading && len(st.sandboxes) > 0 {
			cursor := st.table.Cursor()
			if cursor >= 0 && cursor < len(st.sandboxes) {
				name := st.sandboxes[cursor].Name
				return st, func() tea.Msg {
					return logModeMsg{mode: logModeSandbox, sandboxName: name}
				}
			}
		}
		if msg.String() == "s" && !st.loading {
			st.sortColumn = (st.sortColumn + 1) % 4
			st.updateTableRows()
			return st, nil
		}
		if msg.String() == "S" && !st.loading {
			st.sortAsc = !st.sortAsc
			st.updateTableRows()
			return st, nil
		}
		if msg.String() == "enter" && !st.loading && len(st.sandboxes) > 0 {
			cursor := st.table.Cursor()
			if st.detailIdx == cursor && st.showDetail {
				st.closeDetail()
				return st, nil
			}
			return st, st.openDetail(cursor)
		}
		if msg.String() == "esc" && st.detailIdx >= 0 {
			st.closeDetail()
			return st, nil
		}

	case tea.WindowSizeMsg:
		st.width = msg.Width
		st.height = msg.Height
		st.table.SetWidth(msg.Width)
		st.table.SetHeight(msg.Height - 2)
		return st, nil
	}

	// Delegate to table for navigation keys (only when not in create dialog).
	if !st.showCreate {
		var cmd tea.Cmd
		st.table, cmd = st.table.Update(msg)
		return st, cmd
	}

	return st, nil
}

// View renders the sandbox tab content.
func (st *SandboxTab) View(width, height int) string {
	if st.width != width || st.height != height {
		st.width = width
		st.height = height
		st.table.SetWidth(width)
		if height > 2 {
			st.table.SetHeight(height - 2)
		}
	}

	if st.loading {
		return lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Padding(2, 0).
			Render("Loading sandboxes...")
	}

	if st.errMsg != "" {
		return lipgloss.NewStyle().
			Width(width).
			Foreground(colorRed).
			Padding(1, 0).
			Render(st.errMsg)
	}

	if st.showExec {
		return st.renderExecPopup(width, height)
	}

	if st.showCreate {
		return st.renderCreateDialog(width)
	}

	if st.showDelete {
		return st.renderDeleteConfirm(width)
	}

	if st.showDetail {
		return st.renderDetailPopup(width, height)
	}

	if len(st.sandboxes) == 0 {
		return lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Padding(2, 0).
			Foreground(colorGray).
			Render("No sandboxes found. Press 'c' to create one.")
	}

	return st.table.View()
}

// Refresh returns a command that reloads sandbox data.
func (st *SandboxTab) Refresh() tea.Cmd {
	st.loading = true
	return st.listSandboxes()
}

// renderCreateDialog renders the sandbox creation form.
func (st *SandboxTab) renderCreateDialog(width int) string {
	dialogWidth := 50
	if width < dialogWidth+4 {
		dialogWidth = width - 4
	}
	if dialogWidth < 20 {
		dialogWidth = 20
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorCyan).
		Render("Create Sandbox")

	nameLabel := "Name:"
	imageLabel := "Image:"

	var content string
	content += title + "\n\n"
	content += nameLabel + "\n"
	content += st.createNameInput.View() + "\n\n"
	content += imageLabel + "\n"
	content += st.createImageInput.View() + "\n\n"
	content += lipgloss.NewStyle().Foreground(colorGray).
		Render("Enter: create  Tab: next field  Escape: cancel")

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(dialogWidth).
		Render(content)

	// Center the dialog horizontally.
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Padding(2, 0).
		Render(dialog)
}

const sandboxRefreshInterval = 5 * time.Second

func (st *SandboxTab) scheduleRefresh() tea.Cmd {
	return tea.Tick(sandboxRefreshInterval, func(time.Time) tea.Msg {
		return sandboxRefreshTickMsg{}
	})
}

// listSandboxes returns a tea.Cmd that fetches the sandbox list from the SDK.
func (st *SandboxTab) listSandboxes() tea.Cmd {
	client := st.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		sandboxes, err := client.Sandboxes().List(ctx)
		return sandboxListMsg{sandboxes: sandboxes, err: err}
	}
}

// startWatch creates a new sandbox watcher and returns a tea.Cmd that reads
// events from the watch stream. Each event is delivered as a sandboxWatchEventMsg.
// When the stream ends (channel closed), a sandboxWatchErrMsg is sent to trigger
// a re-list and reconnection.
func (st *SandboxTab) startWatch() tea.Cmd {
	client := st.client
	logger := st.logger

	return func() tea.Msg {
		// Watch with empty name to observe all sandboxes.
		watcher, err := client.Sandboxes().Watch(context.Background(), "")
		if err != nil {
			logger.Error("failed to start sandbox watcher", "error", err)
			return sandboxWatchErrMsg{err: err}
		}

		// Return the watcher via a message so it is stored on the main
		// goroutine in Update(), avoiding a data race.
		return sandboxWatchStartedMsg{watcher: watcher}
	}
}

// readNextWatchEvent returns a tea.Cmd that blocks until the next event
// arrives on the watch channel. When the channel is closed, it returns
// a sandboxWatchErrMsg to trigger reconnection.
func (st *SandboxTab) readNextWatchEvent() tea.Cmd {
	watcher := st.watcher
	logger := st.logger

	return func() tea.Msg {
		if watcher == nil {
			return sandboxWatchErrMsg{err: fmt.Errorf("watcher is nil")}
		}

		event, ok := <-watcher.ResultChan()
		if !ok {
			// Channel closed: the watch stream has ended.
			logger.Info("sandbox watch stream closed, will reconnect")
			return sandboxWatchErrMsg{err: fmt.Errorf("watch stream closed")}
		}

		return sandboxWatchEventMsg{event: event}
	}
}

// openCreateDialog initializes and shows the create dialog.
func (st *SandboxTab) openCreateDialog() (TabModel, tea.Cmd) {
	st.showCreate = true
	st.createFocus = createFieldName
	st.createNameInput.SetValue("")
	st.createImageInput.SetValue("")
	st.createImageInput.Blur()
	cmd := st.createNameInput.Focus()
	return st, cmd
}

// closeCreateDialog hides the create dialog and clears inputs.
func (st *SandboxTab) closeCreateDialog() {
	st.showCreate = false
	st.createNameInput.Blur()
	st.createImageInput.Blur()
	st.createNameInput.SetValue("")
	st.createImageInput.SetValue("")
}

// updateCreateDialog handles key presses while the create dialog is open.
func (st *SandboxTab) updateCreateDialog(msg tea.KeyPressMsg) (TabModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		st.closeCreateDialog()
		return st, nil

	case "enter":
		name := st.createNameInput.Value()
		image := st.createImageInput.Value()
		if name == "" {
			return st, nil
		}
		st.closeCreateDialog()
		return st, st.createSandbox(name, image)

	case "tab", "shift+tab":
		// Toggle focus between name and image fields.
		var cmd tea.Cmd
		if st.createFocus == createFieldName {
			st.createFocus = createFieldImage
			st.createNameInput.Blur()
			cmd = st.createImageInput.Focus()
		} else {
			st.createFocus = createFieldName
			st.createImageInput.Blur()
			cmd = st.createNameInput.Focus()
		}
		return st, cmd
	}

	// Delegate to the focused text input.
	var cmd tea.Cmd
	if st.createFocus == createFieldName {
		st.createNameInput, cmd = st.createNameInput.Update(msg)
	} else {
		st.createImageInput, cmd = st.createImageInput.Update(msg)
	}
	return st, cmd
}

// createSandbox returns a tea.Cmd that creates a sandbox via the SDK.
func (st *SandboxTab) createSandbox(name, image string) tea.Cmd {
	client := st.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		spec := &types.SandboxSpec{}
		if image != "" {
			spec.Template = &types.SandboxTemplate{Image: image}
		}

		sandbox, err := client.Sandboxes().Create(ctx, name, spec, nil)
		return sandboxCreateMsg{sandbox: sandbox, err: err}
	}
}

// updateDeleteConfirm handles key presses in the delete confirmation dialog.
func (st *SandboxTab) updateDeleteConfirm(msg tea.KeyPressMsg) (TabModel, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		name := st.deleteName
		st.showDelete = false
		st.deleteName = ""
		return st, st.deleteSandbox(name)

	case "n", "esc":
		st.showDelete = false
		st.deleteName = ""
		return st, nil
	}

	return st, nil
}

// deleteSandbox returns a tea.Cmd that deletes a sandbox via the SDK.
func (st *SandboxTab) deleteSandbox(name string) tea.Cmd {
	client := st.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := client.Sandboxes().Delete(ctx, name)
		return sandboxDeleteMsg{name: name, err: err}
	}
}

// renderDeleteConfirm renders the delete confirmation dialog.
func (st *SandboxTab) renderDeleteConfirm(width int) string {
	dialogWidth := 50
	if width < dialogWidth+4 {
		dialogWidth = width - 4
	}
	if dialogWidth < 20 {
		dialogWidth = 20
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorRed).
		Render("Delete Sandbox")

	question := fmt.Sprintf("Are you sure you want to delete %q?", st.deleteName)

	content := title + "\n\n" + question + "\n\n" +
		lipgloss.NewStyle().Foreground(colorGray).
			Render("y/Enter: confirm  n/Escape: cancel")

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorRed).
		Padding(1, 2).
		Width(dialogWidth).
		Render(content)

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Padding(2, 0).
		Render(dialog)
}

// openDetail expands the detail pane for the sandbox at the given table index
// and starts an async fetch for policy status.
func (st *SandboxTab) openDetail(idx int) tea.Cmd {
	if idx < 0 || idx >= len(st.sandboxes) {
		return nil
	}
	st.showDetail = true
	st.detailIdx = idx
	st.detailName = st.sandboxes[idx].Name
	st.detailPolicy = nil
	st.detailLoading = true
	st.detailErr = ""
	return st.fetchSandboxDetail(st.detailName)
}

// closeDetail collapses the detail pane and clears its state.
func (st *SandboxTab) closeDetail() {
	st.showDetail = false
	st.detailIdx = -1
	st.detailName = ""
	st.detailPolicy = nil
	st.detailLoading = false
	st.detailErr = ""
}

// fetchSandboxDetail returns a tea.Cmd that fetches policy status for a sandbox.
func (st *SandboxTab) fetchSandboxDetail(name string) tea.Cmd {
	client := st.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, err := client.Policy().GetStatus(ctx, name)
		return sandboxDetailMsg{
			name:         name,
			policyStatus: status,
			policyErr:    err,
		}
	}
}

// Cleanup releases resources held by the sandbox tab. It stops the
// active watcher so the gRPC watch stream is closed cleanly before
// the client connection is torn down.
func (st *SandboxTab) Cleanup() {
	st.stopWatch()
}

// stopWatch stops the active watcher if one is running.
func (st *SandboxTab) stopWatch() {
	if st.watcher != nil {
		st.watcher.Stop()
		st.watcher = nil
	}
}

// handleWatchEvent processes a single watch event, updating the local
// sandbox list accordingly.
func (st *SandboxTab) handleWatchEvent(event types.Event[*types.Sandbox]) {
	switch event.Type {
	case types.EventAdded:
		// Add new sandbox if not already present.
		for _, s := range st.sandboxes {
			if s.Name == event.Object.Name {
				return
			}
		}
		st.sandboxes = append(st.sandboxes, event.Object)

	case types.EventModified:
		// Update existing sandbox.
		for i, s := range st.sandboxes {
			if s.Name == event.Object.Name {
				st.sandboxes[i] = event.Object
				return
			}
		}

	case types.EventDeleted:
		// Remove sandbox from list, adjusting the detail pane index.
		for i, s := range st.sandboxes {
			if s.Name == event.Object.Name {
				st.sandboxes = append(st.sandboxes[:i], st.sandboxes[i+1:]...)
				if st.detailIdx >= 0 {
					if i == st.detailIdx {
						st.closeDetail()
					} else if i < st.detailIdx {
						st.detailIdx--
					}
				}
				return
			}
		}
	}
}

func (st *SandboxTab) sortSandboxes() {
	slices.SortStableFunc(st.sandboxes, func(a, b *types.Sandbox) int {
		var result int
		switch st.sortColumn {
		case 0:
			result = cmp.Compare(a.Name, b.Name)
		case 1:
			result = cmp.Compare(string(a.Status.Phase), string(b.Status.Phase))
		case 2:
			ai, bi := "", ""
			if a.Spec.Template != nil {
				ai = a.Spec.Template.Image
			}
			if b.Spec.Template != nil {
				bi = b.Spec.Template.Image
			}
			result = cmp.Compare(ai, bi)
		case 3:
			result = a.CreatedAt.Compare(b.CreatedAt)
		}
		if !st.sortAsc {
			result = -result
		}
		return result
	})
}

func (st *SandboxTab) updateTableRows() {
	st.sortSandboxes()

	titles := []string{"Name", "Phase", "Image", "Created"}
	widths := []int{25, 15, 30, 20}
	indicator := " ▲"
	if !st.sortAsc {
		indicator = " ▼"
	}
	cols := make([]table.Column, len(titles))
	for i, t := range titles {
		title := t
		if i == st.sortColumn {
			title += indicator
		}
		cols[i] = table.Column{Title: title, Width: widths[i]}
	}
	st.table.SetColumns(cols)

	rows := make([]table.Row, 0, len(st.sandboxes))
	for _, s := range st.sandboxes {
		image := ""
		if s.Spec.Template != nil {
			image = s.Spec.Template.Image
		}
		rows = append(rows, table.Row{
			s.Name,
			phaseWithColor(s.Status.Phase),
			image,
			formatTime(s.CreatedAt),
		})
	}
	st.table.SetRows(rows)
}

// phaseWithColor returns the phase string with ANSI color applied based on
// the sandbox lifecycle phase.
func phaseWithColor(phase types.SandboxPhase) string {
	var style lipgloss.Style
	switch phase {
	case types.SandboxReady:
		style = phaseReady
	case types.SandboxProvisioning:
		style = phaseProvisioning
	case types.SandboxError:
		style = phaseError
	case types.SandboxDeleting:
		style = phaseDeleting
	default:
		style = phaseUnknown
	}
	return style.Render(string(phase))
}

// formatTime formats a time.Time as a relative or absolute time string.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	elapsed := time.Since(t)
	switch {
	case elapsed < time.Minute:
		return "just now"
	case elapsed < time.Hour:
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	case elapsed < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	default:
		return t.Format("2006-01-02 15:04")
	}
}

// CapturesInput returns true when a dialog or popup is active.
func (st *SandboxTab) CapturesInput() bool {
	return st.showCreate || st.showDelete || st.showExec || st.showDetail
}

// --- Exec popup ---

func (st *SandboxTab) openExecPopup() (TabModel, tea.Cmd) {
	cursor := st.table.Cursor()
	if cursor < 0 || cursor >= len(st.sandboxes) {
		return st, nil
	}
	sb := st.sandboxes[cursor]
	if sb.Status.Phase != types.SandboxReady {
		return st, nil
	}
	st.showExec = true
	st.execSandboxName = sb.Name
	st.execCommandInput = textinput.New()
	st.execCommandInput.Prompt = "$ "
	st.execCommandInput.Placeholder = ""
	st.execCommandInput.CharLimit = 1024
	cmd := st.execCommandInput.Focus()
	return st, cmd
}

func (st *SandboxTab) closeExecPopup() {
	st.showExec = false
	st.execSandboxName = ""
	st.execCommandInput.Blur()
	st.execRunning = false
	st.execRunningCommand = ""
	st.execScrollOffset = 0
}

func (st *SandboxTab) updateExecPopup(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			st.closeExecPopup()
			return st, nil
		case "enter":
			if !st.execRunning && st.execCommandInput.Value() != "" {
				command := st.execCommandInput.Value()
				st.execCommandInput.SetValue("")
				st.execRunning = true
				st.execRunningCommand = command
				st.execSpinnerFrame = 0
				return st, tea.Batch(st.executeCommand(command), st.execSpinnerTick())
			}
			return st, nil
		case "up":
			st.execScrollUp(1)
			return st, nil
		case "down":
			st.execScrollDown(1)
			return st, nil
		}
		var cmd tea.Cmd
		st.execCommandInput, cmd = st.execCommandInput.Update(msg)
		return st, cmd

	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			st.execScrollUp(3)
		} else if msg.Button == tea.MouseWheelDown {
			st.execScrollDown(3)
		}
		return st, nil
	}

	return st, nil
}

func (st *SandboxTab) execScrollUp(n int) {
	st.execScrollOffset -= n
}

func (st *SandboxTab) execScrollDown(n int) {
	st.execScrollOffset += n
	if st.execScrollOffset > 0 {
		st.execScrollOffset = 0
	}
}

func (st *SandboxTab) executeCommand(command string) tea.Cmd {
	client := st.client
	sandboxName := st.execSandboxName
	return func() tea.Msg {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		args := []string{"/bin/sh", "-c", command}
		result, err := client.Exec().Run(ctx, sandboxName, args)
		duration := time.Since(start)
		entry := ExecEntry{Sandbox: sandboxName, Command: command, Duration: duration}
		if err != nil {
			entry.Err = err
		} else {
			entry.ExitCode = result.ExitCode
			entry.Stdout = string(result.Stdout)
			entry.Stderr = string(result.Stderr)
		}
		return execResultMsg{entry: entry}
	}
}

func (st *SandboxTab) execSpinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return execSpinnerTickMsg{}
	})
}

func (st *SandboxTab) renderExecPopup(width, height int) string {
	popupWidth := width - 8
	if popupWidth > 80 {
		popupWidth = 80
	}
	if popupWidth < 30 {
		popupWidth = 30
	}
	popupHeight := height - 4
	if popupHeight < 10 {
		popupHeight = 10
	}

	// Fixed header: title + command input (4 lines with padding).
	title := lipgloss.NewStyle().Bold(true).Foreground(colorCyan).
		Render(fmt.Sprintf("Exec: %s", st.execSandboxName))
	var inputLine string
	if st.execRunning {
		frame := spinnerFrames[st.execSpinnerFrame%len(spinnerFrames)]
		inputLine = lipgloss.NewStyle().Foreground(colorYellow).
			Render(fmt.Sprintf("%s Running: %s", frame, st.execRunningCommand))
	} else {
		inputLine = st.execCommandInput.View()
	}

	// Fixed footer (1 line).
	footer := lipgloss.NewStyle().Foreground(colorGray).
		Render("Enter: run  Ctrl+D/U: scroll  Esc: close")

	// Build all history lines.
	innerWidth := popupWidth - 6 // border + padding
	var histLines []string
	if len(st.execHistory) == 0 {
		histLines = append(histLines, lipgloss.NewStyle().Foreground(colorGray).
			Render("Type a command and press Enter"))
	} else {
		for i, entry := range st.execHistory {
			if i > 0 {
				histLines = append(histLines, lipgloss.NewStyle().Foreground(colorGray).
					Render(strings.Repeat("-", innerWidth)))
			}
			header := fmt.Sprintf("$ %s", entry.Command)
			if entry.Err != nil {
				header += lipgloss.NewStyle().Foreground(colorRed).
					Render(fmt.Sprintf("  (error: %v)", entry.Err))
			} else {
				exitStyle := lipgloss.NewStyle().Foreground(colorGreen)
				if entry.ExitCode != 0 {
					exitStyle = lipgloss.NewStyle().Foreground(colorRed)
				}
				header += exitStyle.Render(fmt.Sprintf("  [exit %d]", entry.ExitCode))
				header += lipgloss.NewStyle().Foreground(colorGray).
					Render(fmt.Sprintf("  (%s)", entry.Duration.Truncate(time.Millisecond)))
			}
			histLines = append(histLines, lipgloss.NewStyle().Bold(true).Render(header))
			if entry.Stdout != "" {
				for _, l := range strings.Split(strings.TrimRight(entry.Stdout, "\n"), "\n") {
					histLines = append(histLines, l)
				}
			}
			if entry.Stderr != "" {
				for _, l := range strings.Split(strings.TrimRight(entry.Stderr, "\n"), "\n") {
					histLines = append(histLines, lipgloss.NewStyle().Foreground(colorRed).Render(l))
				}
			}
		}
	}

	// Viewport: fixed height, scroll to show latest by default.
	// header=4 lines (title, blank, input, blank), footer=2 lines (blank, footer), border=2
	viewportHeight := popupHeight - 8
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	// Auto-scroll to bottom unless user has scrolled up.
	total := len(histLines)
	start := total - viewportHeight + st.execScrollOffset
	if start > total-viewportHeight {
		start = total - viewportHeight
	}
	if start < 0 {
		start = 0
	}
	end := start + viewportHeight
	if end > total {
		end = total
	}

	visible := histLines[start:end]
	for len(visible) < viewportHeight {
		visible = append(visible, "")
	}

	// Build scrollbar if content exceeds viewport.
	scrollbarLines := buildScrollbar(viewportHeight, total, start)

	// Append scrollbar to each visible line.
	for i, line := range visible {
		visible[i] = line + " " + scrollbarLines[i]
	}

	// Assemble the popup content.
	var content strings.Builder
	content.WriteString(title)
	content.WriteByte('\n')
	content.WriteByte('\n')
	content.WriteString(inputLine)
	content.WriteByte('\n')
	content.WriteByte('\n')
	content.WriteString(strings.Join(visible, "\n"))
	content.WriteByte('\n')
	content.WriteByte('\n')
	content.WriteString(footer)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(popupWidth).
		Height(popupHeight).
		Render(content.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func buildScrollbar(viewportHeight, totalLines, startLine int) []string {
	trackChar := lipgloss.NewStyle().Foreground(lipgloss.Color("237")).Render("│")
	thumbChar := lipgloss.NewStyle().Foreground(colorCyan).Render("┃")

	lines := make([]string, viewportHeight)
	if totalLines <= viewportHeight {
		for i := range lines {
			lines[i] = " "
		}
		return lines
	}

	thumbSize := viewportHeight * viewportHeight / totalLines
	if thumbSize < 1 {
		thumbSize = 1
	}
	thumbPos := startLine * viewportHeight / totalLines

	for i := range lines {
		if i >= thumbPos && i < thumbPos+thumbSize {
			lines[i] = thumbChar
		} else {
			lines[i] = trackChar
		}
	}
	return lines
}

// --- Interactive shell ---

func (st *SandboxTab) startShell(sandboxName string) tea.Cmd {
	client := st.client
	cols, rows := uint32(st.width), uint32(st.height)
	if cols == 0 {
		cols = 80
	}
	if rows == 0 {
		rows = 24
	}
	return func() tea.Msg {
		ctx := context.Background()
		opts := v1.ExecOptions{
			Env: map[string]string{"TERM": "xterm-256color"},
		}
		session, err := client.Exec().Interactive(ctx, sandboxName, []string{"/bin/bash", "-l"}, cols, rows, opts)
		return shellSessionMsg{session: session, err: err}
	}
}

type interactiveShellCmd struct {
	session v1.InteractiveSession
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

func (c *interactiveShellCmd) SetStdin(r io.Reader)  { c.stdin = r }
func (c *interactiveShellCmd) SetStdout(w io.Writer) { c.stdout = w }
func (c *interactiveShellCmd) SetStderr(w io.Writer) { c.stderr = w }

func (c *interactiveShellCmd) Run() error {
	defer c.session.Close()

	// Clear screen and move cursor to top-left before entering the shell.
	fmt.Fprint(c.stdout, "\033[2J\033[H")

	oldState, err := term.MakeRaw(os.Stdin.Fd())
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(os.Stdin.Fd(), oldState)

	done := make(chan error, 2)

	go func() {
		_, err := io.Copy(c.stdout, readerFunc(c.session.Read))
		done <- err
	}()

	go func() {
		_, err := io.Copy(writerFunc(c.session.Write), c.stdin)
		done <- err
	}()

	<-done
	return nil
}

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) { return f(p) }

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// --- Detail popup ---

func (st *SandboxTab) renderDetailPopup(width, height int) string {
	if st.detailIdx < 0 || st.detailIdx >= len(st.sandboxes) {
		return ""
	}
	sb := st.sandboxes[st.detailIdx]

	var lines []string
	header := lipgloss.NewStyle().Bold(true).Foreground(colorCyan).
		Render(fmt.Sprintf("Details: %s", sb.Name))
	lines = append(lines, header, "")

	image := "(none)"
	if sb.Spec.Template != nil && sb.Spec.Template.Image != "" {
		image = sb.Spec.Template.Image
	}
	lines = append(lines, fmt.Sprintf("  Image: %s", image))
	lines = append(lines, fmt.Sprintf("  Phase: %s", phaseWithColor(sb.Status.Phase)))

	if sb.Spec.GPUCount != nil {
		lines = append(lines, fmt.Sprintf("  GPUs:  %d", *sb.Spec.GPUCount))
	}
	if sb.Spec.LogLevel != "" {
		lines = append(lines, fmt.Sprintf("  Log:   %s", sb.Spec.LogLevel))
	}
	if len(sb.Spec.Providers) > 0 {
		lines = append(lines, fmt.Sprintf("  Providers: %d", len(sb.Spec.Providers)))
	}

	if len(sb.Labels) > 0 {
		labelStr := "  Labels: "
		first := true
		for k, v := range sb.Labels {
			if !first {
				labelStr += ", "
			}
			labelStr += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
		lines = append(lines, labelStr)
	}

	// Policy status.
	policyLine := "  Policy: "
	if st.detailLoading {
		policyLine += lipgloss.NewStyle().Foreground(colorGray).Render("loading...")
	} else if st.detailErr != "" {
		policyLine += lipgloss.NewStyle().Foreground(colorYellow).Render("unavailable")
	} else if st.detailPolicy != nil {
		ps := st.detailPolicy
		statusStr := ps.Revision.Status.String()
		if statusStr == "" {
			statusStr = "none"
		}
		policyLine += fmt.Sprintf("v%d (%s) active=v%d", ps.Revision.Version, statusStr, ps.ActiveVersion)
	} else {
		policyLine += lipgloss.NewStyle().Foreground(colorGray).Render("none")
	}
	lines = append(lines, policyLine)

	// Connectivity.
	sshInd := lipgloss.NewStyle().Foreground(colorGreen).Render("SSH: available")
	tcpInd := lipgloss.NewStyle().Foreground(colorGreen).Render("TCP: available")
	if sb.Status.Phase != types.SandboxReady {
		sshInd = lipgloss.NewStyle().Foreground(colorGray).Render("SSH: requires ready")
		tcpInd = lipgloss.NewStyle().Foreground(colorGray).Render("TCP: requires ready")
	}
	lines = append(lines, fmt.Sprintf("  %s  %s", sshInd, tcpInd))

	// Gateway health summary.
	lines = append(lines, "")
	gwLine := "  Gateway: "
	if st.gatewayHealthy == nil {
		gwLine += lipgloss.NewStyle().Foreground(colorGray).Render("unknown")
	} else if *st.gatewayHealthy {
		gwLine += lipgloss.NewStyle().Foreground(colorGreen).Render(healthDotHealthy + " healthy")
		if st.gatewayVersion != "" {
			gwLine += " " + lipgloss.NewStyle().Foreground(colorGray).Render("("+st.gatewayVersion+")")
		}
	} else {
		gwLine += lipgloss.NewStyle().Foreground(colorRed).Render(healthDotUnhealthy + " unhealthy")
	}
	lines = append(lines, gwLine)

	lines = append(lines, "", lipgloss.NewStyle().Foreground(colorGray).
		Render("  x: exec  l: logs  Escape: close"))

	content := strings.Join(lines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(60).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
