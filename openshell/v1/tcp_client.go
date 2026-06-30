// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/rhuss/openshell-sdk-go/openshell/v1/internal/converter"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"google.golang.org/grpc"
)

type tcpClient struct {
	client    pb.OpenShellClient
	sandboxes SandboxInterface
}

func newTCPClient(conn grpc.ClientConnInterface, sandboxes SandboxInterface) *tcpClient {
	return &tcpClient{client: pb.NewOpenShellClient(conn), sandboxes: sandboxes}
}

func (t *tcpClient) Forward(ctx context.Context, sandboxName string, port uint32, opts ...ForwardOption) (io.ReadWriteCloser, error) {
	if sandboxName == "" {
		return nil, &StatusError{Code: ErrorInvalidArgument, Message: "sandbox name must not be empty"}
	}
	if port == 0 || port > 65535 {
		return nil, &StatusError{
			Code:    ErrorInvalidArgument,
			Message: fmt.Sprintf("port must be in range 1-65535, got %d", port),
		}
	}

	// Resolve sandbox name to ID — the proto RPC takes SandboxId, not name.
	sb, err := t.sandboxes.Get(ctx, sandboxName)
	if err != nil {
		return nil, err
	}

	var cfg forwardConfig
	for _, o := range opts {
		o(&cfg)
	}

	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := t.client.ForwardTcp(streamCtx)
	if err != nil {
		cancel()
		return nil, converter.FromGRPCError(err)
	}

	initFrame := &pb.TcpForwardFrame{
		Payload: &pb.TcpForwardFrame_Init{
			Init: &pb.TcpForwardInit{
				SandboxId: sb.ID,
				ServiceId: cfg.serviceID,
				Target: &pb.TcpForwardInit_Tcp{
					Tcp: &pb.TcpRelayTarget{
						Host: "127.0.0.1",
						Port: port,
					},
				},
			},
		},
	}

	if err := stream.Send(initFrame); err != nil {
		cancel()
		return nil, converter.FromGRPCError(err)
	}

	conn := &tcpForwardConn{
		stream:    stream,
		streamCtx: streamCtx,
		cancel:    cancel,
		dataCh:    make(chan []byte, 64),
		done:      make(chan struct{}),
	}
	go conn.readLoop()
	return conn, nil
}

// tcpForwardConn wraps a bidirectional TcpForwardFrame stream into an
// io.ReadWriteCloser. A background goroutine owns the Recv loop and routes
// data frames to dataCh. Read and Write may be called from different
// goroutines, but multiple concurrent Read callers are not supported.
type tcpForwardConn struct {
	stream    grpc.BidiStreamingClient[pb.TcpForwardFrame, pb.TcpForwardFrame]
	streamCtx context.Context
	cancel    context.CancelFunc
	sendMu    sync.Mutex
	dataCh    chan []byte
	done      chan struct{}
	errOnce   sync.Once
	err       error
	buf       []byte
}

func (c *tcpForwardConn) setErr(err error) {
	c.errOnce.Do(func() { c.err = err })
}

func (c *tcpForwardConn) readLoop() {
	defer close(c.dataCh)
	defer close(c.done)
	for {
		frame, err := c.stream.Recv()
		if err != nil {
			if err != io.EOF {
				c.setErr(converter.FromGRPCError(err))
			}
			return
		}
		data := frame.GetData()
		if data == nil {
			continue
		}
		dataCopy := make([]byte, len(data))
		copy(dataCopy, data)
		select {
		case c.dataCh <- dataCopy:
		case <-c.streamCtx.Done():
			return
		}
	}
}

func (c *tcpForwardConn) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(c.buf) > 0 {
		n := copy(p, c.buf)
		c.buf = c.buf[n:]
		return n, nil
	}

	data, ok := <-c.dataCh
	if !ok {
		if c.err != nil {
			return 0, c.err
		}
		return 0, io.EOF
	}
	n := copy(p, data)
	if n < len(data) {
		c.buf = append(c.buf, data[n:]...)
	}
	return n, nil
}

func (c *tcpForwardConn) Write(p []byte) (int, error) {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	err := c.stream.Send(&pb.TcpForwardFrame{
		Payload: &pb.TcpForwardFrame_Data{Data: p},
	})
	if err != nil {
		return 0, converter.FromGRPCError(err)
	}
	return len(p), nil
}

func (c *tcpForwardConn) Close() error {
	c.sendMu.Lock()
	err := c.stream.CloseSend()
	c.sendMu.Unlock()
	c.cancel()
	<-c.done
	return err
}
