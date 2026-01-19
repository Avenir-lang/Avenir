package runtime

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

type netService struct {
	nextID    uint64
	mu        sync.Mutex
	conns     map[uint64]net.Conn
	listeners map[uint64]net.Listener
}

func newNetService() *netService {
	return &netService{
		conns:     make(map[uint64]net.Conn),
		listeners: make(map[uint64]net.Listener),
	}
}

func (n *netService) Connect(host string, port int) ([]byte, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port %d", port)
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	id := n.nextHandle()
	n.mu.Lock()
	n.conns[id] = conn
	n.mu.Unlock()
	return encodeHandle(id), nil
}

func (n *netService) Listen(host string, port int) ([]byte, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port %d", port)
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	id := n.nextHandle()
	n.mu.Lock()
	n.listeners[id] = ln
	n.mu.Unlock()
	return encodeHandle(id), nil
}

func (n *netService) Accept(serverHandle []byte) ([]byte, error) {
	id, err := decodeHandle(serverHandle)
	if err != nil {
		return nil, err
	}
	n.mu.Lock()
	ln := n.listeners[id]
	n.mu.Unlock()
	if ln == nil {
		return nil, fmt.Errorf("invalid server handle")
	}
	conn, err := ln.Accept()
	if err != nil {
		return nil, err
	}
	connID := n.nextHandle()
	n.mu.Lock()
	n.conns[connID] = conn
	n.mu.Unlock()
	return encodeHandle(connID), nil
}

func (n *netService) Read(sockHandle []byte, count int) ([]byte, error) {
	if count < 0 {
		return nil, fmt.Errorf("invalid read size %d", count)
	}
	id, err := decodeHandle(sockHandle)
	if err != nil {
		return nil, err
	}
	n.mu.Lock()
	conn := n.conns[id]
	n.mu.Unlock()
	if conn == nil {
		return nil, fmt.Errorf("invalid socket handle")
	}
	if count == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, count)
	nread, err := conn.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return buf[:nread], nil
}

func (n *netService) Write(sockHandle []byte, data []byte) (int, error) {
	id, err := decodeHandle(sockHandle)
	if err != nil {
		return 0, err
	}
	n.mu.Lock()
	conn := n.conns[id]
	n.mu.Unlock()
	if conn == nil {
		return 0, fmt.Errorf("invalid socket handle")
	}
	return conn.Write(data)
}

func (n *netService) Close(handle []byte) error {
	id, err := decodeHandle(handle)
	if err != nil {
		return err
	}
	n.mu.Lock()
	if conn, ok := n.conns[id]; ok {
		delete(n.conns, id)
		n.mu.Unlock()
		return conn.Close()
	}
	if ln, ok := n.listeners[id]; ok {
		delete(n.listeners, id)
		n.mu.Unlock()
		return ln.Close()
	}
	n.mu.Unlock()
	return fmt.Errorf("invalid handle")
}

func (n *netService) nextHandle() uint64 {
	return atomic.AddUint64(&n.nextID, 1)
}
