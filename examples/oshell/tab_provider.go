// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// providerListMsg carries the result of listing provider profiles.
type providerListMsg struct {
	profiles []*types.ProviderProfile
	err      error
}

// ProviderTab displays provider profiles in a table.
type ProviderTab struct {
	client v1.ClientInterface
	logger *slog.Logger

	table    table.Model
	profiles []*types.ProviderProfile
	loading  bool
	errMsg   string

	width  int
	height int
}

// NewProviderTab creates a new ProviderTab connected to the SDK client.
func NewProviderTab(client v1.ClientInterface, logger *slog.Logger) *ProviderTab {
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Category", Width: 15},
		{Title: "Description", Width: 40},
		{Title: "Inference", Width: 10},
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

	return &ProviderTab{
		client:  client,
		logger:  logger,
		table:   t,
		loading: true,
	}
}

// Title returns the display title for the tab bar.
func (pt *ProviderTab) Title() string {
	return "Providers"
}

// Init returns the initial command that loads provider profiles.
func (pt *ProviderTab) Init() tea.Cmd {
	return pt.loadProfiles()
}

// Update handles messages for the ProviderTab.
func (pt *ProviderTab) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case providerListMsg:
		pt.loading = false
		if msg.err != nil {
			pt.errMsg = fmt.Sprintf("Error loading providers: %v", msg.err)
			pt.logger.Error("provider tab: failed to list profiles", "error", msg.err)
			return pt, NewSDKErrorMsg(msg.err, "provider.profiles.list")
		}
		pt.errMsg = ""
		pt.profiles = msg.profiles
		pt.updateTableRows()
		return pt, nil

	case tea.WindowSizeMsg:
		pt.width = msg.Width
		pt.height = msg.Height
		pt.table.SetWidth(msg.Width)
		pt.table.SetHeight(max(msg.Height-2, 1))
		return pt, nil

	case tea.KeyPressMsg:
		// Let the table handle navigation keys.
		var cmd tea.Cmd
		pt.table, cmd = pt.table.Update(msg)
		return pt, cmd
	}

	var cmd tea.Cmd
	pt.table, cmd = pt.table.Update(msg)
	return pt, cmd
}

// View renders the provider tab content.
func (pt *ProviderTab) View(width, height int) string {
	if pt.width != width || pt.height != height {
		pt.width = width
		pt.height = height
		pt.table.SetWidth(width)
		pt.table.SetHeight(height - 2)
	}

	if pt.loading {
		return lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Padding(2, 0).
			Render("Loading provider profiles...")
	}

	if pt.errMsg != "" {
		return lipgloss.NewStyle().
			Width(width).
			Foreground(colorRed).
			Padding(1, 0).
			Render(pt.errMsg)
	}

	if len(pt.profiles) == 0 {
		return emptyStateStyle.Width(width).Render("No provider profiles found.")
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(pt.table.View())
}

// Refresh returns a command that reloads provider data.
func (pt *ProviderTab) Refresh() tea.Cmd {
	pt.loading = true
	return pt.loadProfiles()
}

// Cleanup is a no-op for the provider tab (no long-lived resources).
func (pt *ProviderTab) Cleanup() {}

// loadProfiles returns a tea.Cmd that fetches provider profiles.
func (pt *ProviderTab) loadProfiles() tea.Cmd {
	client := pt.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		profiles, err := client.Providers().Profiles().List(ctx)
		return providerListMsg{profiles: profiles, err: err}
	}
}

// updateTableRows rebuilds the table rows from the current profiles.
func (pt *ProviderTab) updateTableRows() {
	rows := make([]table.Row, 0, len(pt.profiles))
	for _, p := range pt.profiles {
		inference := "No"
		if p.InferenceCapable {
			inference = "Yes"
		}
		rows = append(rows, table.Row{
			p.DisplayName,
			string(p.Category),
			p.Description,
			inference,
		})
	}
	pt.table.SetRows(rows)
}

// CapturesInput returns false; the provider tab has no text inputs.
func (pt *ProviderTab) CapturesInput() bool {
	return false
}
