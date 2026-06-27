// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package openshell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDialEmptyAddress(t *testing.T) {
	client, err := Dial("")
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, "address must not be empty", err.Error())
}

func TestDialValidAddress(t *testing.T) {
	client, err := Dial("localhost:8080")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClientClose(t *testing.T) {
	client, err := Dial("localhost:8080")
	require.NoError(t, err)
	assert.NoError(t, client.Close())
}
