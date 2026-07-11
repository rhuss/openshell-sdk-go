// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// serviceListMsg carries the result of listing service endpoints.
type serviceListMsg struct {
	endpoints []*types.ServiceEndpoint
	err       error
}

// serviceCopiedClearMsg clears the "copied" notification after a delay.
type serviceCopiedClearMsg struct{}

// ServiceTab displays service endpoints in a table.
type ServiceTab struct {
	client v1.ClientInterface
	logger *slog.Logger

	table     table.Model
	endpoints []*types.ServiceEndpoint
	loading   bool
	errMsg    string

	// copiedURL holds the URL that was last copied, shown briefly as feedback.
	copiedURL string

	width  int
	height int
}

// NewServiceTab creates a new ServiceTab connected to the SDK client.
func NewServiceTab(client v1.ClientInterface, logger *slog.Logger) *ServiceTab {
	columns := []table.Column{
		{Title: "Sandbox", Width: 20},
		{Title: "Service", Width: 20},
		{Title: "Port", Width: 8},
		{Title: "URL", Width: 40},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(colorWhite).
		Background(lipgloss.Color("#3C3C3C")).
		Bold(false)
	t.SetStyles(s)

	return &ServiceTab{
		client:  client,
		logger:  logger,
		table:   t,
		loading: true,
	}
}

// Title returns the display title for the tab bar.
func (st *ServiceTab) Title() string {
	return "Services"
}

// Init returns the initial command that loads service endpoints.
func (st *ServiceTab) Init() tea.Cmd {
	return st.loadServices()
}

// Update handles messages for the ServiceTab.
func (st *ServiceTab) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case serviceListMsg:
		st.loading = false
		if msg.err != nil {
			st.errMsg = fmt.Sprintf("Error loading services: %v", msg.err)
			st.logger.Error("service tab: failed to list endpoints", "error", msg.err)
			return st, NewSDKErrorMsg(msg.err, "service.list")
		}
		st.errMsg = ""
		st.endpoints = msg.endpoints
		st.updateTableRows()
		return st, nil

	case tea.WindowSizeMsg:
		st.width = msg.Width
		st.height = msg.Height
		st.table.SetWidth(msg.Width)
		st.table.SetHeight(max(msg.Height-2, 1))
		return st, nil

	case serviceCopiedClearMsg:
		st.copiedURL = ""
		return st, nil

	case tea.KeyPressMsg:
		if msg.String() == "enter" {
			cmd := st.copySelectedURL()
			if cmd != nil {
				return st, cmd
			}
		}
		var cmd tea.Cmd
		st.table, cmd = st.table.Update(msg)
		return st, cmd
	}

	var cmd tea.Cmd
	st.table, cmd = st.table.Update(msg)
	return st, cmd
}

// View renders the service tab content.
func (st *ServiceTab) View(width, height int) string {
	if st.width != width || st.height != height {
		st.width = width
		st.height = height
		st.table.SetWidth(width)
		st.table.SetHeight(height - 2)
	}

	if st.loading {
		return lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Padding(2, 0).
			Render("Loading service endpoints...")
	}

	if st.errMsg != "" {
		return lipgloss.NewStyle().
			Width(width).
			Foreground(colorRed).
			Padding(1, 0).
			Render(st.errMsg)
	}

	if len(st.endpoints) == 0 {
		return emptyStateStyle.Width(width).Render("No service endpoints found.")
	}

	tableView := lipgloss.NewStyle().Padding(0, 1).Render(st.table.View())

	// Show "Copied!" feedback when a URL was recently copied.
	if st.copiedURL != "" {
		hint := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).
			Render(fmt.Sprintf("  Copied: %s", st.copiedURL))
		tableView += "\n" + lipgloss.NewStyle().Padding(0, 1).Render(hint)
	} else if len(st.endpoints) > 0 {
		hint := lipgloss.NewStyle().Foreground(colorGray).
			Render("  Press Enter to copy URL to clipboard")
		tableView += "\n" + lipgloss.NewStyle().Padding(0, 1).Render(hint)
	}

	return tableView
}

// Refresh returns a command that reloads service data.
func (st *ServiceTab) Refresh() tea.Cmd {
	st.loading = true
	return st.loadServices()
}

// Cleanup is a no-op for the service tab (no long-lived resources).
func (st *ServiceTab) Cleanup() {}

// loadServices returns a tea.Cmd that fetches service endpoints.
// Passing an empty sandbox name returns endpoints across all sandboxes.
func (st *ServiceTab) loadServices() tea.Cmd {
	client := st.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		endpoints, err := client.Services().List(ctx, "")
		return serviceListMsg{endpoints: endpoints, err: err}
	}
}

// updateTableRows rebuilds the table rows from the current endpoints.
func (st *ServiceTab) updateTableRows() {
	rows := make([]table.Row, 0, len(st.endpoints))
	for _, ep := range st.endpoints {
		rows = append(rows, table.Row{
			ep.SandboxName,
			ep.ServiceName,
			fmt.Sprintf("%d", ep.TargetPort),
			ep.URL,
		})
	}
	st.table.SetRows(rows)
}

// copySelectedURL copies the URL of the selected service endpoint to the
// system clipboard using the OSC 52 terminal escape sequence. This works
// in terminals that support OSC 52 (most modern terminals).
func (st *ServiceTab) copySelectedURL() tea.Cmd {
	if len(st.endpoints) == 0 {
		return nil
	}
	cursor := st.table.Cursor()
	if cursor < 0 || cursor >= len(st.endpoints) {
		return nil
	}
	ep := st.endpoints[cursor]
	if ep.URL == "" {
		return nil
	}

	url := ep.URL
	st.copiedURL = url
	st.logger.Info("copied service URL to clipboard",
		"sandbox", ep.SandboxName, "service", ep.ServiceName)

	// Send OSC 52 escape sequence and schedule clear after 2 seconds.
	return tea.Batch(
		osc52Copy(url),
		tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return serviceCopiedClearMsg{}
		}),
	)
}

// osc52Copy returns a tea.Cmd that writes the OSC 52 escape sequence to
// copy text to the system clipboard via the terminal. The sequence is:
// ESC ] 52 ; c ; <base64-encoded text> ESC \
func osc52Copy(text string) tea.Cmd {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	seq := fmt.Sprintf("\x1b]52;c;%s\x1b\\", encoded)
	return tea.Printf("%s", seq)
}

// CapturesInput returns false; the service tab has no text inputs.
func (st *ServiceTab) CapturesInput() bool {
	return false
}
