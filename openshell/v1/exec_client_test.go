// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"io"
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

type mockExecServer struct {
	pb.UnimplementedOpenShellServer
	mu              sync.Mutex
	execEvents      []*pb.ExecSandboxEvent
	execErr         error
	lastExecRequest *pb.ExecSandboxRequest

	interactiveEvents []*pb.ExecSandboxEvent
	interactiveErr    error
	receivedInputs    []*pb.ExecSandboxInput
}

func newMockExecServer() *mockExecServer {
	return &mockExecServer{}
}

func (s *mockExecServer) ExecSandbox(req *pb.ExecSandboxRequest, stream grpc.ServerStreamingServer[pb.ExecSandboxEvent]) error {
	s.mu.Lock()
	s.lastExecRequest = req
	events := make([]*pb.ExecSandboxEvent, len(s.execEvents))
	copy(events, s.execEvents)
	execErr := s.execErr
	s.mu.Unlock()

	if execErr != nil {
		return execErr
	}

	for _, ev := range events {
		if err := stream.Send(ev); err != nil {
			return err
		}
	}
	return nil
}

func (s *mockExecServer) ExecSandboxInteractive(stream grpc.BidiStreamingServer[pb.ExecSandboxInput, pb.ExecSandboxEvent]) error {
	s.mu.Lock()
	interactiveErr := s.interactiveErr
	events := make([]*pb.ExecSandboxEvent, len(s.interactiveEvents))
	copy(events, s.interactiveEvents)
	s.mu.Unlock()

	if interactiveErr != nil {
		return interactiveErr
	}

	// Read the start message
	startMsg, err := stream.Recv()
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.receivedInputs = append(s.receivedInputs, startMsg)
	s.mu.Unlock()

	// Read subsequent messages until client closes, collecting them
	go func() {
		for {
			msg, recvErr := stream.Recv()
			if recvErr != nil {
				return
			}
			s.mu.Lock()
			s.receivedInputs = append(s.receivedInputs, msg)
			s.mu.Unlock()
		}
	}()

	// Send canned events
	for _, ev := range events {
		if err := stream.Send(ev); err != nil {
			return err
		}
	}
	return nil
}

func setupExecTest(t *testing.T, mock *mockExecServer) (*execClient, func()) {
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

	return newExecClient(conn), func() {
		_ = conn.Close()
		srv.Stop()
	}
}

// --- T043: Run and Stream tests ---

func TestExecRun(t *testing.T) {
	mock := newMockExecServer()
	mock.execEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("hello ")}}},
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("world\n")}}},
		{Payload: &pb.ExecSandboxEvent_Stderr{Stderr: &pb.ExecSandboxStderr{Data: []byte("warn\n")}}},
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 0}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	result, err := client.Run(context.Background(), "sb-1", []string{"echo", "hello", "world"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, []byte("hello world\n"), result.Stdout)
	assert.Equal(t, []byte("warn\n"), result.Stderr)
}

func TestExecRun_WithOptions(t *testing.T) {
	mock := newMockExecServer()
	mock.execEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 0}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	opts := ExecOptions{
		Env:     map[string]string{"FOO": "bar"},
		WorkDir: "/tmp",
	}
	result, err := client.Run(context.Background(), "sb-1", []string{"ls"}, opts)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ExitCode)

	mock.mu.Lock()
	defer mock.mu.Unlock()
	assert.Equal(t, "sb-1", mock.lastExecRequest.GetSandboxId())
	assert.Equal(t, []string{"ls"}, mock.lastExecRequest.GetCommand())
	assert.Equal(t, "/tmp", mock.lastExecRequest.GetWorkdir())
	assert.Equal(t, map[string]string{"FOO": "bar"}, mock.lastExecRequest.GetEnvironment())
}

func TestExecRun_NonZeroExit(t *testing.T) {
	mock := newMockExecServer()
	mock.execEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Stderr{Stderr: &pb.ExecSandboxStderr{Data: []byte("fail\n")}}},
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 1}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	result, err := client.Run(context.Background(), "sb-1", []string{"false"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ExitCode)
	assert.Empty(t, result.Stdout)
	assert.Equal(t, []byte("fail\n"), result.Stderr)
}

func TestExecRun_ServerError(t *testing.T) {
	mock := newMockExecServer()
	mock.execErr = status.Error(codes.NotFound, "sandbox not found")
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	_, err := client.Run(context.Background(), "sb-missing", []string{"ls"})

	require.Error(t, err)
	assert.True(t, IsNotFound(err))
}

func TestExecStream(t *testing.T) {
	mock := newMockExecServer()
	mock.execEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("line1\n")}}},
		{Payload: &pb.ExecSandboxEvent_Stderr{Stderr: &pb.ExecSandboxStderr{Data: []byte("err1\n")}}},
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("line2\n")}}},
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 42}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	stream, err := client.Stream(context.Background(), "sb-1", []string{"cat"})
	require.NoError(t, err)
	require.NotNil(t, stream)
	defer func() { _ = stream.Close() }()

	chunk1, err := stream.Next()
	require.NoError(t, err)
	assert.Equal(t, StreamStdout, chunk1.Stream)
	assert.Equal(t, []byte("line1\n"), chunk1.Data)

	chunk2, err := stream.Next()
	require.NoError(t, err)
	assert.Equal(t, StreamStderr, chunk2.Stream)
	assert.Equal(t, []byte("err1\n"), chunk2.Data)

	chunk3, err := stream.Next()
	require.NoError(t, err)
	assert.Equal(t, StreamStdout, chunk3.Stream)
	assert.Equal(t, []byte("line2\n"), chunk3.Data)

	// Next call after exit should return io.EOF
	_, err = stream.Next()
	assert.ErrorIs(t, err, io.EOF)

	exitCode, err := stream.ExitCode()
	require.NoError(t, err)
	assert.Equal(t, 42, exitCode)
}

func TestExecStream_ServerError(t *testing.T) {
	mock := newMockExecServer()
	mock.execErr = status.Error(codes.Internal, "internal error")
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	stream, err := client.Stream(context.Background(), "sb-1", []string{"ls"})
	if err != nil {
		return
	}
	_, err = stream.Next()
	require.Error(t, err)
}

func TestExecStream_EmptyOutput(t *testing.T) {
	mock := newMockExecServer()
	mock.execEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 0}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	stream, err := client.Stream(context.Background(), "sb-1", []string{"true"})
	require.NoError(t, err)
	defer func() { _ = stream.Close() }()

	_, err = stream.Next()
	assert.ErrorIs(t, err, io.EOF)

	exitCode, err := stream.ExitCode()
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
}

// --- T044: Interactive session tests ---

func TestExecInteractive(t *testing.T) {
	mock := newMockExecServer()
	mock.interactiveEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("$ ")}}},
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("output\n")}}},
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 0}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	session, err := client.Interactive(context.Background(), "sb-1", []string{"/bin/bash"}, 80, 24)
	require.NoError(t, err)
	require.NotNil(t, session)
	defer func() { _ = session.Close() }()

	// Read output
	buf := make([]byte, 1024)
	n, err := session.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "$ ", string(buf[:n]))

	// Verify start message was received
	mock.mu.Lock()
	require.GreaterOrEqual(t, len(mock.receivedInputs), 1)
	startInput := mock.receivedInputs[0]
	mock.mu.Unlock()

	startReq := startInput.GetStart()
	require.NotNil(t, startReq)
	assert.Equal(t, "sb-1", startReq.GetSandboxId())
	assert.Equal(t, []string{"/bin/bash"}, startReq.GetCommand())
	assert.True(t, startReq.GetTty())
	assert.Equal(t, uint32(80), startReq.GetCols())
	assert.Equal(t, uint32(24), startReq.GetRows())
}

func TestExecInteractive_Write(t *testing.T) {
	mock := newMockExecServer()
	mock.interactiveEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("$ ")}}},
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 0}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	session, err := client.Interactive(context.Background(), "sb-1", []string{"/bin/sh"}, 80, 24)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	n, err := session.Write([]byte("ls\n"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)
}

func TestExecInteractive_Resize(t *testing.T) {
	mock := newMockExecServer()
	mock.interactiveEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("$ ")}}},
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 0}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	session, err := client.Interactive(context.Background(), "sb-1", []string{"/bin/sh"}, 80, 24)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	err = session.Resize(120, 40)
	require.NoError(t, err)
}

func TestExecInteractive_ServerError(t *testing.T) {
	mock := newMockExecServer()
	mock.interactiveErr = status.Error(codes.PermissionDenied, "not allowed")
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	session, err := client.Interactive(context.Background(), "sb-1", []string{"/bin/sh"}, 80, 24)
	if err != nil {
		return
	}
	buf := make([]byte, 1024)
	_, err = session.Read(buf)
	require.Error(t, err)
}

func TestExecInteractive_ExitCode(t *testing.T) {
	mock := newMockExecServer()
	mock.interactiveEvents = []*pb.ExecSandboxEvent{
		{Payload: &pb.ExecSandboxEvent_Stdout{Stdout: &pb.ExecSandboxStdout{Data: []byte("done\n")}}},
		{Payload: &pb.ExecSandboxEvent_Exit{Exit: &pb.ExecSandboxExit{ExitCode: 130}}},
	}
	client, cleanup := setupExecTest(t, mock)
	defer cleanup()

	session, err := client.Interactive(context.Background(), "sb-1", []string{"/bin/sh"}, 80, 24)
	require.NoError(t, err)

	// Drain output
	buf := make([]byte, 1024)
	for {
		_, readErr := session.Read(buf)
		if readErr != nil {
			break
		}
	}

	exitCode, err := session.ExitCode()
	require.NoError(t, err)
	assert.Equal(t, 130, exitCode)

	_ = session.Close()
}
