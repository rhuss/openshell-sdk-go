// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"net"
	"sync"
	"testing"

	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// --- Mock server for TCP forwarding ---

// mockTCPServer implements the ForwardTcp bidi stream. It records the init
// frame and echoes every data frame back to the client.
type mockTCPServer struct {
	pb.UnimplementedOpenShellServer
	mu       sync.Mutex
	lastInit *pb.TcpForwardInit
	err      error // if non-nil, return this error immediately on stream open
}

func newMockTCPServer() *mockTCPServer {
	return &mockTCPServer{}
}

func (s *mockTCPServer) ForwardTcp(stream grpc.BidiStreamingServer[pb.TcpForwardFrame, pb.TcpForwardFrame]) error { //nolint:revive // proto-generated method name
	s.mu.Lock()
	earlyErr := s.err
	s.mu.Unlock()
	if earlyErr != nil {
		return earlyErr
	}

	// First frame must be init.
	frame, err := stream.Recv()
	if err != nil {
		return err
	}
	init := frame.GetInit()
	if init == nil {
		return status.Errorf(codes.InvalidArgument, "first frame must be init")
	}

	s.mu.Lock()
	s.lastInit = init
	s.mu.Unlock()

	// Echo loop: every data frame is sent back verbatim.
	for {
		frame, err = stream.Recv()
		if err != nil {
			return err
		}
		data := frame.GetData()
		if data == nil {
			continue
		}
		if err := stream.Send(&pb.TcpForwardFrame{
			Payload: &pb.TcpForwardFrame_Data{Data: data},
		}); err != nil {
			return err
		}
	}
}

// --- Test setup ---

func setupTCPTest(t *testing.T, mock *mockTCPServer) (*tcpClient, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	pb.RegisterOpenShellServer(srv, mock)
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient("passthrough:///bufconn",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	return newTCPClient(conn), func() {
		_ = conn.Close()
		srv.Stop()
	}
}

// --- Tests ---

func TestTCPForward_InitFrame(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	rwc, err := client.Forward(context.Background(), "my-sandbox", 8080)
	require.NoError(t, err)
	require.NotNil(t, rwc)
	defer func() { _ = rwc.Close() }()

	// Write something to trigger the init frame to be sent (init is sent
	// on Forward, before any Write — but we need a brief moment for the
	// server to process it).
	_, err = rwc.Write([]byte("ping"))
	require.NoError(t, err)

	// Read back the echo.
	buf := make([]byte, 64)
	n, err := rwc.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "ping", string(buf[:n]))

	// Verify the init frame the server received.
	mock.mu.Lock()
	init := mock.lastInit
	mock.mu.Unlock()

	require.NotNil(t, init)
	assert.Equal(t, "my-sandbox", init.GetSandboxId())
	assert.Empty(t, init.GetServiceId(), "service_id should be empty per FR-007a")
	assert.Empty(t, init.GetAuthorizationToken())

	tcp := init.GetTcp()
	require.NotNil(t, tcp, "target should be TcpRelayTarget")
	assert.Equal(t, "127.0.0.1", tcp.GetHost())
	assert.Equal(t, uint32(8080), tcp.GetPort())
}

func TestTCPForward_ReadWrite(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	rwc, err := client.Forward(context.Background(), "test-sandbox", 3000)
	require.NoError(t, err)
	defer func() { _ = rwc.Close() }()

	// Write data and read the echo back.
	payload := []byte("hello, sandbox!")
	_, err = rwc.Write(payload)
	require.NoError(t, err)

	buf := make([]byte, 64)
	n, err := rwc.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, payload, buf[:n])

	// Second round-trip.
	_, err = rwc.Write([]byte("round2"))
	require.NoError(t, err)

	n, err = rwc.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "round2", string(buf[:n]))
}

func TestTCPForward_Close(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	rwc, err := client.Forward(context.Background(), "my-sandbox", 5432)
	require.NoError(t, err)

	err = rwc.Close()
	require.NoError(t, err)

	// Subsequent writes should fail.
	_, err = rwc.Write([]byte("should fail"))
	assert.Error(t, err)

	// Subsequent reads should also fail.
	buf := make([]byte, 64)
	_, err = rwc.Read(buf)
	assert.Error(t, err)
}

func TestTCPForward_PartialRead(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	rwc, err := client.Forward(context.Background(), "my-sandbox", 8080)
	require.NoError(t, err)
	defer func() { _ = rwc.Close() }()

	// Write a payload larger than the read buffer.
	payload := []byte("abcdefghijklmnopqrstuvwxyz")
	_, err = rwc.Write(payload)
	require.NoError(t, err)

	// Read with a small buffer — should get partial data and buffer the rest.
	var collected []byte
	buf := make([]byte, 10)
	for len(collected) < len(payload) {
		n, readErr := rwc.Read(buf)
		require.NoError(t, readErr)
		collected = append(collected, buf[:n]...)
	}
	assert.Equal(t, payload, collected)
}

func TestTCPForward_ConcurrentReadWrite(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	rwc, err := client.Forward(context.Background(), "my-sandbox", 8080)
	require.NoError(t, err)
	defer func() { _ = rwc.Close() }()

	const iterations = 50
	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	wg.Add(2)

	go func() {
		defer wg.Done()
		for range iterations {
			_, writeErr := rwc.Write([]byte("ping"))
			if writeErr != nil {
				errCh <- writeErr
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		buf := make([]byte, 64)
		for range iterations {
			_, readErr := rwc.Read(buf)
			if readErr != nil {
				errCh <- readErr
				return
			}
		}
	}()

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent goroutine failed: %v", err)
	}
}

func TestTCPForward_PortValidation(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	tests := []struct {
		name string
		port uint32
	}{
		{"port zero", 0},
		{"port too high", 65536},
		{"port way too high", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rwc, err := client.Forward(context.Background(), "my-sandbox", tt.port)
			assert.Nil(t, rwc)
			require.Error(t, err)
			assert.True(t, IsInvalidArgument(err), "expected InvalidArgument, got: %v", err)
		})
	}

	// Valid boundary ports should not get client-side rejection.
	for _, port := range []uint32{1, 65535} {
		rwc, err := client.Forward(context.Background(), "my-sandbox", port)
		require.NoError(t, err, "port %d should be valid", port)
		require.NotNil(t, rwc)
		_ = rwc.Close()
	}
}

func TestTCPForward_ContextCancellation(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	rwc, err := client.Forward(ctx, "my-sandbox", 8080)
	require.NoError(t, err)
	require.NotNil(t, rwc)

	// Cancel the context.
	cancel()

	// Reads should return an error (context cancelled propagates through the gRPC stream).
	buf := make([]byte, 64)
	_, err = rwc.Read(buf)
	assert.Error(t, err)

	// Writes should also fail after context cancellation.
	_, err = rwc.Write([]byte("should fail"))
	assert.Error(t, err)
}

func TestTCPForward_WithServiceID(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	rwc, err := client.Forward(context.Background(), "my-sandbox", 8080, WithForwardServiceID("audit-svc"))
	require.NoError(t, err)
	require.NotNil(t, rwc)
	defer func() { _ = rwc.Close() }()

	// Trigger a round-trip so the server has processed the init frame.
	_, err = rwc.Write([]byte("ping"))
	require.NoError(t, err)
	buf := make([]byte, 64)
	_, err = rwc.Read(buf)
	require.NoError(t, err)

	mock.mu.Lock()
	init := mock.lastInit
	mock.mu.Unlock()

	require.NotNil(t, init)
	assert.Equal(t, "audit-svc", init.GetServiceId())
	assert.Equal(t, "my-sandbox", init.GetSandboxId())
}

func TestTCPForward_WithoutOptions_BackwardCompat(t *testing.T) {
	mock := newMockTCPServer()
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	rwc, err := client.Forward(context.Background(), "my-sandbox", 8080)
	require.NoError(t, err)
	require.NotNil(t, rwc)
	defer func() { _ = rwc.Close() }()

	_, err = rwc.Write([]byte("ping"))
	require.NoError(t, err)
	buf := make([]byte, 64)
	_, err = rwc.Read(buf)
	require.NoError(t, err)

	mock.mu.Lock()
	init := mock.lastInit
	mock.mu.Unlock()

	require.NotNil(t, init)
	assert.Empty(t, init.GetServiceId(), "service_id should be empty when no option provided")
}

func TestTCPForward_ServerError(t *testing.T) {
	mock := newMockTCPServer()
	mock.err = status.Errorf(codes.NotFound, "sandbox not found")
	client, cleanup := setupTCPTest(t, mock)
	defer cleanup()

	rwc, err := client.Forward(context.Background(), "missing", 8080)

	// The stream opens successfully (gRPC bidi streams don't fail on open),
	// but the first write or read should surface the server error.
	if err != nil {
		// When Send(initFrame) races with the server returning NotFound,
		// the client may get the server status or a transport-level error.
		assert.Nil(t, rwc)
		require.Error(t, err)
		return
	}

	// If stream opened, the error surfaces on Read (the server returns it
	// immediately, which closes the recv side).
	require.NotNil(t, rwc)
	defer func() { _ = rwc.Close() }()

	buf := make([]byte, 64)
	_, err = rwc.Read(buf)
	assert.Error(t, err)
}
