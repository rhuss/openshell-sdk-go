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

	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := e.client.ExecSandbox(streamCtx, req)
	if err != nil {
		cancel()
		return nil, fromGRPCError(err)
	}

	return &execStream{stream: stream, cancel: cancel}, nil
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
	cancel   context.CancelFunc
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
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

// interactiveSession wraps a bidirectional streaming RPC into the InteractiveSession interface.
// A background goroutine owns the Recv loop and routes events to dataCh (for Read)
// and exitCh (for ExitCode), preventing concurrent Recv calls on the stream.
type interactiveSession struct {
	stream  grpc.BidiStreamingClient[pb.ExecSandboxInput, pb.ExecSandboxEvent]
	sendMu  sync.Mutex
	dataCh  chan []byte
	exitCh  chan int
	done    chan struct{}
	errOnce sync.Once
	err     error
	buf     []byte
}

func newInteractiveSession(stream grpc.BidiStreamingClient[pb.ExecSandboxInput, pb.ExecSandboxEvent]) *interactiveSession {
	s := &interactiveSession{
		stream: stream,
		dataCh: make(chan []byte, 64),
		exitCh: make(chan int, 1),
		done:   make(chan struct{}),
	}
	go s.readLoop()
	return s
}

func (s *interactiveSession) setErr(err error) {
	s.errOnce.Do(func() { s.err = err })
}

func (s *interactiveSession) readLoop() {
	defer close(s.dataCh)
	defer close(s.done)
	for {
		ev, err := s.stream.Recv()
		if err != nil {
			if err != io.EOF {
				s.setErr(fromGRPCError(err))
			}
			return
		}

		chunk, code, isExit, convErr := execChunkFromEvent(ev)
		if convErr != nil {
			s.setErr(convErr)
			return
		}
		if isExit {
			select {
			case s.exitCh <- code:
			default:
			}
			return
		}
		if chunk != nil {
			select {
			case s.dataCh <- chunk.Data:
			case <-s.done:
				return
			}
		}
	}
}

func (s *interactiveSession) Read(p []byte) (int, error) {
	if len(s.buf) > 0 {
		n := copy(p, s.buf)
		s.buf = s.buf[n:]
		return n, nil
	}

	data, ok := <-s.dataCh
	if !ok {
		if s.err != nil {
			return 0, s.err
		}
		return 0, io.EOF
	}
	n := copy(p, data)
	if n < len(data) {
		s.buf = append(s.buf, data[n:]...)
	}
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
	err := s.stream.Send(&pb.ExecSandboxInput{
		Payload: &pb.ExecSandboxInput_Resize{
			Resize: &pb.ExecSandboxWindowResize{
				Cols: cols,
				Rows: rows,
			},
		},
	})
	if err != nil {
		return fromGRPCError(err)
	}
	return nil
}

func (s *interactiveSession) ExitCode() (int, error) {
	select {
	case code := <-s.exitCh:
		return code, nil
	case <-s.done:
		select {
		case code := <-s.exitCh:
			return code, nil
		default:
			if s.err != nil {
				return -1, s.err
			}
			return -1, &StatusError{Code: ErrorInternal, Message: "stream ended without exit event"}
		}
	}
}

func (s *interactiveSession) Close() error {
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
