// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

const (
	// healthCheckInterval is the time between health checks.
	healthCheckInterval = 10 * time.Second
	// sparklineSize is the number of latency measurements in the sparkline.
	sparklineSize = 30
)

// sparklineBlocks are Unicode block elements used for the latency sparkline,
// ordered from lowest (empty) to tallest (full block).
var sparklineBlocks = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// healthCheckMsg carries the result of a single health check.
type healthCheckMsg struct {
	result  *types.HealthResult
	latency time.Duration
	err     error
}

// healthTickMsg triggers the next background health check.
type healthTickMsg struct{}

// gatewayConfigMsg carries the result of loading gateway configuration.
type gatewayConfigMsg struct {
	config *v1.GatewayConfig
	err    error
}

// HealthStatus summarizes the current gateway health.
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthHealthy
	HealthUnhealthy
)

// GatewayTab displays gateway health status, latency sparkline, and
// gateway configuration summary.
type GatewayTab struct {
	client v1.ClientInterface
	logger *slog.Logger

	// status is the current health status.
	status HealthStatus
	// version is the gateway version from the last successful health check.
	version string
	// lastCheck is the timestamp of the last health check.
	lastCheck time.Time
	// lastErr is the error from the last failed health check.
	lastErr error
	// checkCount tracks the total number of health checks performed.
	checkCount int

	// latencies is a ring buffer of recent latency measurements.
	latencies []time.Duration
	// latencyIdx is the write index into the latencies ring buffer.
	latencyIdx int
	// latencyFull is true once the ring buffer has been filled at least once.
	latencyFull bool

	// gatewayConfig holds the loaded gateway configuration.
	gatewayConfig *v1.GatewayConfig
	// configErr holds any error loading gateway configuration.
	configErr error

	// loading indicates the initial load is in progress.
	loading bool

	width  int
	height int
}

// NewGatewayTab creates a new GatewayTab connected to the SDK client.
func NewGatewayTab(client v1.ClientInterface, logger *slog.Logger) *GatewayTab {
	return &GatewayTab{
		client:    client,
		logger:    logger,
		status:    HealthUnknown,
		latencies: make([]time.Duration, sparklineSize),
		loading:   true,
	}
}

// Title returns the display title for the tab bar.
func (ht *GatewayTab) Title() string {
	return "Gateway"
}

// Init returns the initial commands that start health checking and load config.
func (ht *GatewayTab) Init() tea.Cmd {
	return tea.Batch(ht.runHealthCheck(), ht.loadGatewayConfig())
}

// Update handles messages for the GatewayTab.
func (ht *GatewayTab) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case healthCheckMsg:
		ht.loading = false
		ht.checkCount++
		ht.lastCheck = time.Now()
		if msg.err != nil {
			ht.status = HealthUnhealthy
			ht.lastErr = msg.err
			ht.logger.Error("health check failed", "error", msg.err)
			// Record zero latency for failed checks.
			ht.recordLatency(0)
			// Schedule next tick.
			return ht, tea.Batch(
				ht.scheduleNextCheck(),
				NewSDKErrorMsg(msg.err, "health.check"),
			)
		}
		ht.lastErr = nil
		if msg.result.Healthy {
			ht.status = HealthHealthy
		} else {
			ht.status = HealthUnhealthy
		}
		ht.version = msg.result.Version
		ht.recordLatency(msg.latency)
		ht.logger.Info("health check completed",
			"healthy", msg.result.Healthy,
			"version", msg.result.Version,
			"latency", msg.latency)
		return ht, ht.scheduleNextCheck()

	case healthTickMsg:
		return ht, ht.runHealthCheck()

	case gatewayConfigMsg:
		if msg.err != nil {
			ht.configErr = msg.err
			ht.logger.Error("failed to load gateway config", "error", msg.err)
		} else {
			ht.gatewayConfig = msg.config
			ht.configErr = nil
		}
		return ht, nil

	case tea.WindowSizeMsg:
		ht.width = msg.Width
		ht.height = msg.Height
		return ht, nil
	}

	return ht, nil
}

// View renders the health tab content.
func (ht *GatewayTab) View(width, height int) string {
	if ht.width != width || ht.height != height {
		ht.width = width
		ht.height = height
	}

	if ht.loading && ht.checkCount == 0 {
		return lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center).
			Padding(2, 0).
			Render("Running initial health check...")
	}

	var sections []string

	// Health status section.
	sections = append(sections, ht.renderStatus(width))
	sections = append(sections, "")

	// Latency sparkline section.
	sections = append(sections, ht.renderSparkline(width))
	sections = append(sections, "")

	// Gateway config summary.
	sections = append(sections, ht.renderGatewayConfig(width))

	content := strings.Join(sections, "\n")
	return lipgloss.NewStyle().Padding(1, 2).Width(width).Render(content)
}

// Refresh returns a command that triggers an immediate health check.
func (ht *GatewayTab) Refresh() tea.Cmd {
	return tea.Batch(ht.runHealthCheck(), ht.loadGatewayConfig())
}

// Cleanup is a no-op for the health tab. The tea.Tick-based health
// check loop stops automatically when the Bubble Tea program exits.
func (ht *GatewayTab) Cleanup() {}

// Status returns the current HealthStatus for external consumers (e.g. StatusBar).
func (ht *GatewayTab) Status() HealthStatus {
	return ht.status
}

// renderStatus renders the health status display.
func (ht *GatewayTab) renderStatus(width int) string {
	label := statusKeyStyle.Render("Gateway Health:")

	var statusText string
	switch ht.status {
	case HealthHealthy:
		statusText = lipgloss.NewStyle().Foreground(colorGreen).Bold(true).
			Render(healthDotHealthy + " Healthy")
	case HealthUnhealthy:
		statusText = lipgloss.NewStyle().Foreground(colorRed).Bold(true).
			Render(healthDotUnhealthy + " Unhealthy")
	default:
		statusText = lipgloss.NewStyle().Foreground(colorGray).
			Render("○ Unknown")
	}

	line1 := fmt.Sprintf("%s  %s", label, statusText)

	// Version info.
	var details []string
	if ht.version != "" {
		details = append(details,
			fmt.Sprintf("%s %s",
				statusKeyStyle.Render("Version:"),
				statusValueStyle.Render(ht.version)))
	}

	if !ht.lastCheck.IsZero() {
		ago := time.Since(ht.lastCheck).Truncate(time.Second)
		details = append(details,
			fmt.Sprintf("%s %s",
				statusKeyStyle.Render("Last Check:"),
				statusValueStyle.Render(fmt.Sprintf("%s ago", ago))))
	}

	details = append(details,
		fmt.Sprintf("%s %s",
			statusKeyStyle.Render("Checks:"),
			statusValueStyle.Render(fmt.Sprintf("%d", ht.checkCount))))

	if ht.lastErr != nil {
		details = append(details,
			fmt.Sprintf("%s %s",
				statusKeyStyle.Render("Error:"),
				errorStyle.Render(ht.lastErr.Error())))
	}

	result := line1
	for _, d := range details {
		result += "\n  " + d
	}

	_ = width
	return result
}

// renderSparkline renders the latency sparkline using Unicode block elements.
func (ht *GatewayTab) renderSparkline(width int) string {
	label := statusKeyStyle.Render("Latency (last 30 checks):")

	// Collect valid latency values.
	count := ht.latencyCount()
	if count == 0 {
		return label + "\n  " + lipgloss.NewStyle().Foreground(colorGray).
			Render("No latency data yet")
	}

	values := ht.latencyValues()

	// Find min and max for scaling.
	var minVal, maxVal time.Duration
	minVal = values[0]
	maxVal = values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Build sparkline string.
	var spark strings.Builder
	rang := maxVal - minVal
	for _, v := range values {
		if rang == 0 || v == 0 {
			spark.WriteString(sparklineBlocks[0])
			continue
		}
		// Scale to 0-7 index into sparklineBlocks.
		idx := int(float64(v-minVal) / float64(rang) * 7)
		if idx > 7 {
			idx = 7
		}
		spark.WriteString(sparklineBlocks[idx])
	}

	sparkStyle := lipgloss.NewStyle().Foreground(colorCyan)

	// Show min/max/avg stats.
	var total time.Duration
	validCount := 0
	for _, v := range values {
		if v > 0 {
			total += v
			validCount++
		}
	}
	var avg time.Duration
	if validCount > 0 {
		avg = total / time.Duration(validCount)
	}

	stats := lipgloss.NewStyle().Foreground(colorGray).Render(
		fmt.Sprintf("  min=%s avg=%s max=%s",
			minVal.Truncate(time.Millisecond),
			avg.Truncate(time.Millisecond),
			maxVal.Truncate(time.Millisecond)))

	return label + "\n  " + sparkStyle.Render(spark.String()) + stats
}

// renderGatewayConfig renders the gateway configuration summary.
func (ht *GatewayTab) renderGatewayConfig(width int) string {
	label := statusKeyStyle.Render("Gateway Configuration:")

	if ht.configErr != nil {
		return label + "\n  " + errorStyle.Render(
			fmt.Sprintf("Failed to load: %v", ht.configErr))
	}

	if ht.gatewayConfig == nil {
		return label + "\n  " + lipgloss.NewStyle().Foreground(colorGray).
			Render("Loading...")
	}

	if len(ht.gatewayConfig.Settings) == 0 {
		return label + "\n  " + lipgloss.NewStyle().Foreground(colorGray).
			Render("No settings configured")
	}

	var lines []string
	lines = append(lines, label)
	lines = append(lines,
		fmt.Sprintf("  %s %s",
			statusKeyStyle.Render("Revision:"),
			statusValueStyle.Render(fmt.Sprintf("%d", ht.gatewayConfig.SettingsRevision))))

	settingCount := len(ht.gatewayConfig.Settings)
	lines = append(lines,
		fmt.Sprintf("  %s %s",
			statusKeyStyle.Render("Settings:"),
			statusValueStyle.Render(fmt.Sprintf("%d configured", settingCount))))

	// Collect and sort setting keys for deterministic display order.
	keys := make([]string, 0, settingCount)
	for key := range ht.gatewayConfig.Settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Show up to 10 settings.
	for i, key := range keys {
		if i >= 10 {
			remaining := settingCount - i
			lines = append(lines, lipgloss.NewStyle().Foreground(colorGray).
				Render(fmt.Sprintf("    ... and %d more", remaining)))
			break
		}
		lines = append(lines,
			fmt.Sprintf("    %s",
				lipgloss.NewStyle().Foreground(colorGray).Render(key)))
	}

	_ = width
	return strings.Join(lines, "\n")
}

// runHealthCheck returns a tea.Cmd that performs a health check.
func (ht *GatewayTab) runHealthCheck() tea.Cmd {
	client := ht.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		result, err := client.Health().Check(ctx)
		latency := time.Since(start)

		return healthCheckMsg{result: result, latency: latency, err: err}
	}
}

// scheduleNextCheck returns a tea.Cmd that triggers the next health check
// after the configured interval.
func (ht *GatewayTab) scheduleNextCheck() tea.Cmd {
	return tea.Tick(healthCheckInterval, func(time.Time) tea.Msg {
		return healthTickMsg{}
	})
}

// loadGatewayConfig returns a tea.Cmd that fetches gateway configuration.
func (ht *GatewayTab) loadGatewayConfig() tea.Cmd {
	client := ht.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		config, err := client.Config().GetGateway(ctx)
		return gatewayConfigMsg{config: config, err: err}
	}
}

// recordLatency adds a latency measurement to the ring buffer.
func (ht *GatewayTab) recordLatency(d time.Duration) {
	ht.latencies[ht.latencyIdx] = d
	ht.latencyIdx = (ht.latencyIdx + 1) % sparklineSize
	if ht.latencyIdx == 0 {
		ht.latencyFull = true
	}
}

// latencyCount returns the number of valid latency measurements.
func (ht *GatewayTab) latencyCount() int {
	if ht.latencyFull {
		return sparklineSize
	}
	return ht.latencyIdx
}

// latencyValues returns the latency measurements in chronological order.
func (ht *GatewayTab) latencyValues() []time.Duration {
	count := ht.latencyCount()
	values := make([]time.Duration, count)
	if ht.latencyFull {
		// Ring buffer is full; read from latencyIdx (oldest) forward.
		for i := 0; i < count; i++ {
			values[i] = ht.latencies[(ht.latencyIdx+i)%sparklineSize]
		}
	} else {
		copy(values, ht.latencies[:count])
	}
	return values
}

// CapturesInput returns false; the health tab has no text inputs.
func (ht *GatewayTab) CapturesInput() bool {
	return false
}
