// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"fmt"
	"io"
	"sync"

	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"google.golang.org/grpc"
)

type execClient struct {
	client pb.OpenShellClient
}

func newExecClient(conn grpc.ClientConnInterface) *execClient {
	return &execClient{client: pb.NewOpenShellClient(conn)}
}

func (e *execClient) Run(ctx context.Context, sandboxID string, command []string, opts ...ExecOptions) (*ExecResult, error) {
	var opt *ExecOptions
	if len(opts) > 0 {
		opt = &opts[0]
	}
	req := execRequestToProto(sandboxID, command, opt)

	stream, err := e.client.ExecSandbox(ctx, req)
	if err != nil {
		return nil, fromGRPCError(err)
	}

	var events []*pb.ExecSandboxEvent
	for {
		ev, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return nil, fromGRPCError(recvErr)
		}
		events = append(events, ev)
	}

	return execResultFromEvents(events)
}

func (e *execClient) Stream(ctx context.Context, sandboxID string, command []string, opts ...ExecOptions) (ExecStream, error) {
	var opt *ExecOptions
	if len(opts) > 0 {
		opt = &opts[0]
	}
	req := execRequestToProto(sandboxID, command, opt)

	stream, err := e.client.ExecSandbox(ctx, req)
	if err != nil {
		return nil, fromGRPCError(err)
	}

	return &execStream{stream: stream}, nil
}

func (e *execClient) Interactive(ctx context.Context, sandboxID string, command []string, cols, rows uint32, opts ...ExecOptions) (InteractiveSession, error) {
	var opt *ExecOptions
	if len(opts) > 0 {
		opt = &opts[0]
	}

	stream, err := e.client.ExecSandboxInteractive(ctx)
	if err != nil {
		return nil, fromGRPCError(err)
	}

	startReq := execInteractiveRequestToProto(sandboxID, command, cols, rows, opt)
	if sendErr := stream.Send(&pb.ExecSandboxInput{
		Payload: &pb.ExecSandboxInput_Start{Start: startReq},
	}); sendErr != nil {
		return nil, fromGRPCError(sendErr)
	}

	return newInteractiveSession(stream), nil
}

// execStream wraps a server-streaming RPC into the ExecStream interface.
type execStream struct {
	stream   grpc.ServerStreamingClient[pb.ExecSandboxEvent]
	exitCode int
	exited   bool
	hasExit  bool
}

func (s *execStream) Next() (*ExecChunk, error) {
	for {
		ev, err := s.stream.Recv()
		if err == io.EOF {
			return nil, io.EOF
		}
		if err != nil {
			return nil, fromGRPCError(err)
		}

		chunk, code, isExit, convErr := execChunkFromEvent(ev)
		if convErr != nil {
			return nil, convErr
		}
		if chunk != nil {
			return chunk, nil
		}
		if isExit {
			s.exitCode = code
			s.exited = true
			s.hasExit = true
			return nil, io.EOF
		}
	}
}

func (s *execStream) ExitCode() (int, error) {
	if !s.exited {
		for {
			_, err := s.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return -1, err
			}
		}
	}
	if !s.hasExit {
		return -1, &StatusError{Code: ErrorInternal, Message: "stream ended without exit event"}
	}
	return s.exitCode, nil
}

func (s *execStream) Close() error {
	return nil
}

// interactiveSession wraps a bidirectional streaming RPC into the InteractiveSession interface.
type interactiveSession struct {
	stream   grpc.BidiStreamingClient[pb.ExecSandboxInput, pb.ExecSandboxEvent]
	mu       sync.Mutex
	sendMu   sync.Mutex
	buf      []byte
	exitCode int
	exited   bool
	closed   bool
}

func newInteractiveSession(stream grpc.BidiStreamingClient[pb.ExecSandboxInput, pb.ExecSandboxEvent]) *interactiveSession {
	return &interactiveSession{
		stream:   stream,
		exitCode: -1,
	}
}

func (s *interactiveSession) Read(p []byte) (int, error) {
	s.mu.Lock()
	if len(s.buf) > 0 {
		n := copy(p, s.buf)
		s.buf = s.buf[n:]
		s.mu.Unlock()
		return n, nil
	}
	if s.exited {
		s.mu.Unlock()
		return 0, io.EOF
	}
	s.mu.Unlock()

	ev, err := s.stream.Recv()
	if err == io.EOF {
		return 0, io.EOF
	}
	if err != nil {
		return 0, fromGRPCError(err)
	}

	chunk, code, isExit, convErr := execChunkFromEvent(ev)
	if convErr != nil {
		return 0, convErr
	}

	s.mu.Lock()
	if isExit {
		s.exitCode = code
		s.exited = true
		s.mu.Unlock()
		return 0, io.EOF
	}
	s.mu.Unlock()

	if chunk == nil {
		return s.Read(p)
	}

	s.mu.Lock()
	n := copy(p, chunk.Data)
	if n < len(chunk.Data) {
		s.buf = append(s.buf, chunk.Data[n:]...)
	}
	s.mu.Unlock()
	return n, nil
}

func (s *interactiveSession) Write(p []byte) (int, error) {
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	err := s.stream.Send(&pb.ExecSandboxInput{
		Payload: &pb.ExecSandboxInput_Stdin{Stdin: p},
	})
	if err != nil {
		return 0, fromGRPCError(err)
	}
	return len(p), nil
}

func (s *interactiveSession) Resize(cols, rows uint32) error {
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return s.stream.Send(&pb.ExecSandboxInput{
		Payload: &pb.ExecSandboxInput_Resize{
			Resize: &pb.ExecSandboxWindowResize{
				Cols: cols,
				Rows: rows,
			},
		},
	})
}

func (s *interactiveSession) ExitCode() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.exited {
		return s.exitCode, nil
	}

	for {
		ev, err := s.stream.Recv()
		if err == io.EOF {
			if !s.exited {
				return -1, &StatusError{Code: ErrorInternal, Message: "stream ended without exit event"}
			}
			return s.exitCode, nil
		}
		if err != nil {
			return -1, fromGRPCError(err)
		}

		_, code, isExit, convErr := execChunkFromEvent(ev)
		if convErr != nil {
			return -1, convErr
		}
		if isExit {
			s.exitCode = code
			s.exited = true
			return s.exitCode, nil
		}
	}
}

func (s *interactiveSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	return s.stream.CloseSend()
}

// Package-level converter functions to avoid circular imports
// (internal/converter imports v1 for domain types).

func execChunkFromEvent(event *pb.ExecSandboxEvent) (chunk *ExecChunk, exitCode int, isExit bool, err error) {
	if event == nil {
		return nil, -1, false, fmt.Errorf("nil exec event")
	}

	switch p := event.Payload.(type) {
	case *pb.ExecSandboxEvent_Stdout:
		return &ExecChunk{
			Stream: StreamStdout,
			Data:   p.Stdout.GetData(),
		}, -1, false, nil
	case *pb.ExecSandboxEvent_Stderr:
		return &ExecChunk{
			Stream: StreamStderr,
			Data:   p.Stderr.GetData(),
		}, -1, false, nil
	case *pb.ExecSandboxEvent_Exit:
		return nil, int(p.Exit.GetExitCode()), true, nil
	default:
		return nil, -1, false, nil
	}
}

func execRequestToProto(sandboxID string, command []string, opts *ExecOptions) *pb.ExecSandboxRequest {
	req := &pb.ExecSandboxRequest{
		SandboxId: sandboxID,
		Command:   command,
	}
	if opts != nil {
		req.Workdir = opts.WorkDir
		req.Environment = opts.Env
	}
	return req
}

func execInteractiveRequestToProto(sandboxID string, command []string, cols, rows uint32, opts *ExecOptions) *pb.ExecSandboxRequest {
	req := execRequestToProto(sandboxID, command, opts)
	req.Tty = true
	req.Cols = cols
	req.Rows = rows
	return req
}

func execResultFromEvents(events []*pb.ExecSandboxEvent) (*ExecResult, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("no exec events received")
	}

	var stdout, stderr []byte
	exitCode := -1

	for _, event := range events {
		chunk, code, isExit, err := execChunkFromEvent(event)
		if err != nil {
			return nil, err
		}
		if isExit {
			exitCode = code
		} else if chunk != nil {
			switch chunk.Stream {
			case StreamStdout:
				stdout = append(stdout, chunk.Data...)
			case StreamStderr:
				stderr = append(stderr, chunk.Data...)
			}
		}
	}

	if exitCode == -1 {
		return nil, fmt.Errorf("no exit event received")
	}

	return &ExecResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}, nil
}
