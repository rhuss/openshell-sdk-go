// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"charm.land/bubbles/v2/key"
)

// keyMap defines all key bindings for the dashboard.
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Tab1     key.Binding
	Tab2   key.Binding
	Tab3   key.Binding
	Tab4   key.Binding
	Quit   key.Binding
	Create key.Binding
	Delete key.Binding
	Exec   key.Binding
	Logs   key.Binding
	Sort   key.Binding
	Shell  key.Binding
	Enter  key.Binding
	Escape key.Binding
	Retry  key.Binding
	Help   key.Binding
}

// defaultKeyMap returns the default key bindings with vi-style j/k support.
func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "move down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		Tab1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "sandboxes"),
		),
		Tab2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "providers"),
		),
		Tab3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "services"),
		),
		Tab4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "gateway"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Create: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "create"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Exec: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "exec command"),
		),
		Logs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "sandbox logs"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s", "S"),
			key.WithHelp("s/S", "sort/reverse"),
		),
		Shell: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "terminal"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "expand/submit"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "collapse/cancel"),
		),
		Retry: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "retry connection"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ShortHelp returns the short help key bindings shown in the help bar.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Tab, k.Quit, k.Help,
	}
}

// FullHelp returns the full help key bindings shown when help is expanded.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab},
		{k.Tab1, k.Tab2, k.Tab3, k.Tab4},
		{k.Create, k.Delete, k.Exec, k.Logs, k.Enter, k.Escape},
		{k.Retry, k.Quit, k.Help},
	}
}
