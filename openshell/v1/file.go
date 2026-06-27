// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import "context"

// FileInterface defines file transfer operations on sandboxes.
type FileInterface interface {
	Upload(ctx context.Context, sandboxID string, localPath string, remotePath string) error
	Download(ctx context.Context, sandboxID string, remotePath string, localPath string) error
}
