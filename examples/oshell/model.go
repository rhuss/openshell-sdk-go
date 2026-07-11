// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
)

// authLoginMsg is sent when an OIDC login attempt completes.
type authLoginMsg struct {
	err error // nil on success, non-nil on failure.
}

// TabModel is the interface that each tab (Sandboxes, Providers, Services,
// Health, Exec) must implement to participate in the dashboard layout.
type TabModel interface {
	// Init returns the initial command for the tab.
	Init() tea.Cmd

	// Update handles messages for this tab and returns updated model + cmd.
	Update(msg tea.Msg) (TabModel, tea.Cmd)

	// View renders the tab content for the given width and height.
	View(width, height int) string

	// Title returns the display title for the tab bar.
	Title() string

	// Refresh returns a command that reloads the tab's data.
	Refresh() tea.Cmd

	// Cleanup releases resources held by this tab (watchers, goroutines).
	// Called during graceful shutdown before the client is closed.
	Cleanup()

	// CapturesInput returns true when a text input has focus and
	// keystrokes should be routed to the tab instead of global shortcuts.
	CapturesInput() bool
}

// Dashboard is the root Bubble Tea model that composes the tab bar,
// status bar, log panel, and active tab content.
type Dashboard struct {
	// tabs holds all tab sub-models in display order.
	tabs []TabModel
	// activeTab is the index of the currently visible tab.
	activeTab int

	// keys holds the key binding configuration.
	keys keyMap
	// help is the Bubbles help component.
	help help.Model

	// statusBar renders the top status bar.
	statusBar *StatusBar
	// logPanel renders the bottom log panel.
	logPanel *LogPanel

	// client is the SDK client (real or fake).
	client v1.ClientInterface
	// connState tracks the gateway connection state.
	connState *ConnectionManager

	// logger is the structured logger for the dashboard.
	logger *slog.Logger

	// width and height are the terminal dimensions.
	width  int
	height int

	// ctx is the dashboard's lifetime context, cancelled on shutdown.
	ctx context.Context
	// cancelCtx cancels ctx when the dashboard shuts down.
	cancelCtx context.CancelFunc

	// quitting signals that the app is shutting down.
	quitting bool

	// authLoggingIn is true while an OIDC login flow is in progress.
	authLoggingIn bool
}

// NewDashboard creates a new Dashboard model with the given client and logger.
func NewDashboard(client v1.ClientInterface, connMgr *ConnectionManager, logger *slog.Logger, logPanel *LogPanel) *Dashboard {
	km := defaultKeyMap()
	h := help.New()
	h.ShowAll = false

	sb := NewStatusBar()
	if connMgr != nil {
		sb.SetGatewayName(connMgr.GatewayName())
	}

	ctx, cancel := context.WithCancel(context.Background())

	d := &Dashboard{
		tabs:      make([]TabModel, 0, 4),
		activeTab: 0,
		keys:      km,
		help:      h,
		statusBar: sb,
		logPanel:  logPanel,
		client:    client,
		connState: connMgr,
		logger:    logger,
		ctx:       ctx,
		cancelCtx: cancel,
	}

	return d
}

// AddTab appends a tab to the dashboard. Tabs appear in the order added.
func (d *Dashboard) AddTab(tab TabModel) {
	d.tabs = append(d.tabs, tab)
}

// Init initializes all tabs and starts background processes.
func (d *Dashboard) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize each tab.
	for _, tab := range d.tabs {
		cmds = append(cmds, tab.Init())
	}

	// Start connection manager if present.
	if d.connState != nil {
		cmds = append(cmds, d.connState.Init())
	}

	return tea.Batch(cmds...)
}

// Update handles all messages for the dashboard and delegates to sub-models.
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.help.SetWidth(msg.Width)

	case tea.KeyPressMsg:
		// When a tab's text input has focus, skip global shortcuts
		// (except Ctrl+C for emergency quit) so keystrokes go to the input.
		tabCaptures := len(d.tabs) > 0 && d.activeTab < len(d.tabs) && d.tabs[d.activeTab].CapturesInput()
		if !tabCaptures {
			cmd := d.handleKeyPress(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			if d.quitting {
				return d, tea.Batch(cmds...)
			}
		} else if key.Matches(msg, d.keys.Quit) && msg.String() == "ctrl+c" {
			d.quitting = true
			d.cleanup()
			return d, tea.Quit
		}

	case connectionStateMsg:
		if d.connState != nil {
			cmd := d.connState.HandleStateMsg(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			d.statusBar.SetConnectionState(d.connState.State())
		}

	case sdkErrorMsg:
		// Route SDK errors through the connection manager to detect
		// mid-session auth failures (401/Unauthenticated).
		if d.connState != nil {
			d.logger.Error("SDK operation failed",
				"source", msg.source, "error", msg.err)
			cmd := d.connState.HandleSDKError(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			d.statusBar.SetConnectionState(d.connState.State())
		}

	case healthCheckMsg:
		// Wire health check results into the StatusBar health dot.
		if msg.err != nil {
			d.statusBar.SetUnhealthy()
		} else if msg.result != nil && msg.result.Healthy {
			d.statusBar.SetHealthy()
			// Transition to Connected on first successful health check
			// (covers non-OIDC gateways where no login flow fires).
			if d.connState != nil && (d.connState.State() == StateConnecting || d.connState.State() == StateReconnecting) {
				d.connState.state = StateConnected
				d.connState.retryCount = 0
				d.connState.lastError = nil
				d.statusBar.SetConnectionState(StateConnected)
			}
		} else {
			d.statusBar.SetUnhealthy()
		}
		// Fall through to let the HealthTab also process this message.

	case reconnectTickMsg:
		if d.connState != nil {
			cmd := d.connState.HandleReconnectTick()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			d.statusBar.SetConnectionState(d.connState.State())
		}

	case authLoginMsg:
		d.authLoggingIn = false
		if msg.err != nil {
			d.logger.Error("OIDC login failed", "error", msg.err)
			// Stay in AuthRequired so the user can retry.
		} else {
			d.logger.Info("OIDC login successful, reconnecting")
			if d.connState != nil {
				cmds = append(cmds, d.connState.SetConnected())
			}
			// Refresh the active tab to reload data.
			if len(d.tabs) > 0 && d.activeTab < len(d.tabs) {
				cmds = append(cmds, d.tabs[d.activeTab].Refresh())
			}
		}
	}

	// Handle log mode switching.
	if modeMsg, ok := msg.(logModeMsg); ok {
		if modeMsg.mode == logModeSandbox && modeMsg.sandboxName != "" {
			cmds = append(cmds, fetchSandboxLogs(d.client, modeMsg.sandboxName))
		}
	}

	// Update log panel.
	if d.logPanel != nil {
		cmd := d.logPanel.HandleMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Forward non-key messages (ticks, data results) to ALL tabs so that
	// background operations (health ticker, watch streams) continue even
	// when a tab is not active.
	_, isKey := msg.(tea.KeyPressMsg)
	if !isKey {
		for i := range d.tabs {
			newTab, cmd := d.tabs[i].Update(msg)
			d.tabs[i] = newTab
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	} else if len(d.tabs) > 0 && d.activeTab < len(d.tabs) {
		// Key presses go only to the active tab.
		newTab, cmd := d.tabs[d.activeTab].Update(msg)
		d.tabs[d.activeTab] = newTab
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return d, tea.Batch(cmds...)
}

// handleKeyPress processes global key bindings.
func (d *Dashboard) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	// When auth is required, intercept Enter to trigger OIDC login.
	if d.connState != nil && d.connState.State() == StateAuthRequired && !d.authLoggingIn {
		if key.Matches(msg, d.keys.Enter) {
			d.authLoggingIn = true
			d.logger.Info("Starting OIDC login flow")
			return d.startOIDCLogin()
		}
	}

	switch {
	case key.Matches(msg, d.keys.Quit):
		d.quitting = true
		d.cleanup()
		return tea.Quit

	case key.Matches(msg, d.keys.Tab):
		if len(d.tabs) > 0 {
			next := (d.activeTab + 1) % len(d.tabs)
			return d.switchTab(next)
		}
		return nil

	case key.Matches(msg, d.keys.ShiftTab):
		if len(d.tabs) > 0 {
			prev := (d.activeTab - 1 + len(d.tabs)) % len(d.tabs)
			return d.switchTab(prev)
		}
		return nil

	case key.Matches(msg, d.keys.Tab1):
		return d.switchTab(0)
	case key.Matches(msg, d.keys.Tab2):
		return d.switchTab(1)
	case key.Matches(msg, d.keys.Tab3):
		return d.switchTab(2)
	case key.Matches(msg, d.keys.Tab4):
		return d.switchTab(3)

	case key.Matches(msg, d.keys.Retry):
		if d.connState != nil {
			state := d.connState.State()
			if state == StateDisconnected || state == StateReconnecting {
				d.connState.retryCount = 0
				d.statusBar.SetConnectionState(StateConnecting)
				return d.connState.HandleReconnectTick()
			}
		}
		return nil

	case key.Matches(msg, d.keys.Help):
		d.help.ShowAll = !d.help.ShowAll
		return nil
	}

	return nil
}

// switchTab changes the active tab and triggers a refresh.
func (d *Dashboard) switchTab(idx int) tea.Cmd {
	if idx < 0 || idx >= len(d.tabs) {
		return nil
	}
	d.activeTab = idx
	return d.tabs[idx].Refresh()
}

// View renders the entire dashboard layout.
func (d *Dashboard) View() tea.View {
	if d.quitting {
		return tea.NewView("Goodbye!\n")
	}

	if d.width == 0 || d.height == 0 {
		return tea.NewView("Initializing...\n")
	}

	// Layout allocation:
	// Line 1: Status bar (1 line)
	// Line 2: Tab bar (1 line)
	// Bottom: Log panel (1/4 of remaining height, min 3 lines)
	// Middle: Help bar (1 line)
	// Rest: Tab content

	statusBarHeight := 1
	tabBarHeight := 1
	helpBarHeight := 1

	remainingHeight := d.height - statusBarHeight - tabBarHeight - helpBarHeight
	logPanelHeight := remainingHeight / 4
	if logPanelHeight < 3 {
		logPanelHeight = 3
	}
	// Account for log panel border (2 lines).
	logPanelHeight += 2

	contentHeight := remainingHeight - logPanelHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	var view string

	// Status bar.
	view += d.statusBar.View(d.width) + "\n"

	// Tab bar.
	view += d.renderTabBar() + "\n"

	// Tab content or auth prompt overlay.
	if d.connState != nil && d.connState.State() == StateAuthRequired {
		view += d.renderAuthPrompt(d.width, contentHeight)
	} else if len(d.tabs) > 0 && d.activeTab < len(d.tabs) {
		view += d.tabs[d.activeTab].View(d.width, contentHeight)
	}
	view += "\n"

	// Log panel.
	if d.logPanel != nil {
		view += d.logPanel.View(d.width, logPanelHeight-2, false)
	}
	view += "\n"

	// Help bar.
	view += helpStyle.Render(d.help.View(d.keys))

	v := tea.NewView(view)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// renderTabBar renders the tab bar with active/inactive styling.
func (d *Dashboard) renderTabBar() string {
	var bar string
	for i, tab := range d.tabs {
		title := tab.Title()
		if i == d.activeTab {
			bar += activeTabStyle.Render(title)
		} else {
			bar += inactiveTabStyle.Render(title)
		}
		if i < len(d.tabs)-1 {
			bar += " "
		}
	}
	return tabBarStyle.Width(d.width).Render(bar)
}

// renderAuthPrompt renders a centered auth prompt overlay.
func (d *Dashboard) renderAuthPrompt(contentWidth, contentHeight int) string {
	var lines []string

	lines = append(lines, authPromptTitleStyle.Render("Authentication Required"))
	lines = append(lines, "")

	if d.authLoggingIn {
		lines = append(lines, "Opening browser for OIDC login...")
		lines = append(lines, "")
		lines = append(lines, authPromptHintStyle.Render("Complete login in your browser"))
	} else {
		if d.connState != nil && d.connState.LastError() != nil {
			lines = append(lines, errorStyle.Render(
				fmt.Sprintf("Error: %s", d.connState.LastError())))
			lines = append(lines, "")
		}
		lines = append(lines, "Press Enter to login via browser")
		lines = append(lines, "")
		lines = append(lines, authPromptHintStyle.Render("Press q to quit"))
	}

	content := strings.Join(lines, "\n")
	box := authPromptBoxStyle.Render(content)

	// Center the box in the content area.
	return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, box)
}

// cleanup releases resources held by all tabs. It is called during
// graceful shutdown before the SDK client is closed, ensuring that
// protocol-level streams (e.g. watchers) are torn down first.
func (d *Dashboard) cleanup() {
	d.logger.Info("shutting down: cleaning up tab resources")
	for _, tab := range d.tabs {
		tab.Cleanup()
	}
	if d.cancelCtx != nil {
		d.cancelCtx()
	}
}

// startOIDCLogin returns a tea.Cmd that performs the OIDC login flow
// via the OIDCLogin function in connection.go.
func (d *Dashboard) startOIDCLogin() tea.Cmd {
	gatewayName := ""
	if d.connState != nil {
		gatewayName = d.connState.GatewayName()
	}
	return OIDCLogin(d.ctx, gatewayName)
}
