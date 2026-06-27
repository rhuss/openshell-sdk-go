// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import "context"

// HealthResult holds the result of a health check.
type HealthResult struct {
	Healthy bool
	Version string
}

// HealthInterface defines health check operations.
type HealthInterface interface {
	Check(ctx context.Context) (*HealthResult, error)
}
