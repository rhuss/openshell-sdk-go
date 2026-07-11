// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"charm.land/lipgloss/v2"
)

// ANSI 256 color constants for broad terminal compatibility.
var (
	colorGreen  = lipgloss.Color("2")
	colorYellow = lipgloss.Color("3")
	colorRed    = lipgloss.Color("1")
	colorGray   = lipgloss.Color("8")
	colorWhite  = lipgloss.Color("15")
	colorCyan   = lipgloss.Color("6")
)

// Status bar styles.
var (
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(colorWhite).
			Padding(0, 1)

	statusKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	statusValueStyle = lipgloss.NewStyle().
				Foreground(colorWhite)
)

// Tab bar styles.
var (
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorGray).
				Padding(0, 1)

	tabBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236"))
)

// Log panel styles.
var (
	logPanelBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorGray)

	logPanelFocusedBorderStyle = lipgloss.NewStyle().
					Border(lipgloss.NormalBorder()).
					BorderForeground(colorCyan)

	logLevelDebug = lipgloss.NewStyle().Foreground(colorGray)
	logLevelInfo  = lipgloss.NewStyle().Foreground(colorWhite)
	logLevelWarn  = lipgloss.NewStyle().Foreground(colorYellow)
	logLevelError = lipgloss.NewStyle().Foreground(colorRed)
)

// Phase color styles for sandbox status indicators.
var (
	phaseReady        = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	phaseProvisioning = lipgloss.NewStyle().Foreground(colorYellow)
	phaseError        = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	phaseDeleting     = lipgloss.NewStyle().Foreground(colorGray)
	phaseUnknown      = lipgloss.NewStyle().Foreground(colorGray)
)

// Health dot characters.
const (
	healthDotHealthy  = "●" // filled circle
	healthDotDegraded = "●"
	healthDotUnhealthy = "●"
)

// General content styles.
var (
	helpStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	emptyStateStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Italic(true).
			Padding(1, 2)
)

// Auth prompt styles.
var (
	authPromptBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.DoubleBorder()).
				BorderForeground(colorYellow).
				Padding(1, 3).
				Align(lipgloss.Center)

	authPromptTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorYellow)

	authPromptHintStyle = lipgloss.NewStyle().
				Foreground(colorGray).
				Italic(true)
)

