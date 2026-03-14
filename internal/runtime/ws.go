package runtime

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"avenir/internal/runtime/builtins"
)

const (
	wsGUID = "258EAFA5-E914-47DA-95CA-5AB9FC6B9332"

	wsOpContinuation = 0
	wsOpText         = 1
	wsOpBinary       = 2
	wsOpClose        = 8
	wsOpPing         = 9
	wsOpPong         = 10

	wsDefaultMaxMessageSize = 1 << 20 // 1MB
	wsDefaultWriteDeadline  = 30 * time.Second
	wsDefaultSendQueueSize  = 256
)

type wsService struct {
	nextID      uint64
	mu          sync.Mutex
	conns       map[uint64]*wsConn
	httpService *httpService
}

type wsConn struct {
	id           uint64
	conn         net.Conn
	reader       *bufio.Reader
	writeMu      sync.Mutex
	closed       int32
	maxMsgSize   int64
	path         string
	query        string
	remoteAddr   string
	headers      map[string]string
	subprotocol  string
}

func newWSService(httpSvc *httpService) *wsService {
	return &wsService{
		conns:       make(map[uint64]*wsConn),
		httpService: httpSvc,
	}
}

func (ws *wsService) nextHandle() uint64 {
	return atomic.AddUint64(&ws.nextID, 1)
}

func (ws *wsService) Upgrade(reqHandle []byte, protocols []string, extraHeaders map[string]string) (*builtins.WSUpgradeResult, error) {
	reqID, err := decodeHandle(reqHandle)
	if err != nil {
		return nil, fmt.Errorf("ws upgrade: invalid request handle: %w", err)
	}

	ws.httpService.mu.Lock()
	req := ws.httpService.requests[reqID]
	if req != nil {
		delete(ws.httpService.requests, reqID)
	}
	ws.httpService.mu.Unlock()

	if req == nil {
		return nil, fmt.Errorf("ws upgrade: invalid request handle")
	}

	upgradeHeader := ""
	connectionHeader := ""
	wsKey := ""
	wsVersion := ""
	wsProtocol := ""
	for k, v := range req.headers {
		lk := strings.ToLower(k)
		switch lk {
		case "upgrade":
			upgradeHeader = v
		case "connection":
			connectionHeader = v
		case "sec-websocket-key":
			wsKey = v
		case "sec-websocket-version":
			wsVersion = v
		case "sec-websocket-protocol":
			wsProtocol = v
		}
	}

	if !strings.EqualFold(upgradeHeader, "websocket") {
		req.conn.Close()
		return nil, fmt.Errorf("ws upgrade: missing or invalid Upgrade header")
	}
	if !containsTokenCaseInsensitive(connectionHeader, "upgrade") {
		req.conn.Close()
		return nil, fmt.Errorf("ws upgrade: missing or invalid Connection header")
	}
	if wsKey == "" {
		req.conn.Close()
		return nil, fmt.Errorf("ws upgrade: missing Sec-WebSocket-Key")
	}
	if wsVersion != "13" {
		req.conn.Close()
		return nil, fmt.Errorf("ws upgrade: unsupported version %q (expected 13)", wsVersion)
	}

	acceptKey := computeAcceptKey(wsKey)

	negotiatedProtocol := ""
	if len(protocols) > 0 && wsProtocol != "" {
		clientProtocols := parseCSV(wsProtocol)
		for _, sp := range protocols {
			for _, cp := range clientProtocols {
				if strings.EqualFold(sp, cp) {
					negotiatedProtocol = sp
					break
				}
			}
			if negotiatedProtocol != "" {
				break
			}
		}
	}

	var resp strings.Builder
	resp.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	resp.WriteString("Upgrade: websocket\r\n")
	resp.WriteString("Connection: Upgrade\r\n")
	resp.WriteString("Sec-WebSocket-Accept: " + acceptKey + "\r\n")
	if negotiatedProtocol != "" {
		resp.WriteString("Sec-WebSocket-Protocol: " + negotiatedProtocol + "\r\n")
	}
	for k, v := range extraHeaders {
		resp.WriteString(k + ": " + v + "\r\n")
	}
	resp.WriteString("\r\n")

	if _, err := req.conn.Write([]byte(resp.String())); err != nil {
		req.conn.Close()
		return nil, fmt.Errorf("ws upgrade: failed to write handshake response: %w", err)
	}

	connID := ws.nextHandle()
	wsc := &wsConn{
		id:          connID,
		conn:        req.conn,
		reader:      bufio.NewReaderSize(req.conn, 4096),
		maxMsgSize:  wsDefaultMaxMessageSize,
		path:        req.path,
		remoteAddr:  req.remoteAddr,
		headers:     req.headers,
		subprotocol: negotiatedProtocol,
	}

	pathStr := req.path
	queryStr := ""
	if idx := strings.IndexByte(pathStr, '?'); idx >= 0 {
		queryStr = pathStr[idx+1:]
		pathStr = pathStr[:idx]
	}
	wsc.path = pathStr
	wsc.query = queryStr

	ws.mu.Lock()
	ws.conns[connID] = wsc
	ws.mu.Unlock()

	return &builtins.WSUpgradeResult{
		Handle:   encodeHandle(connID),
		Protocol: negotiatedProtocol,
	}, nil
}

func (ws *wsService) getConn(handle []byte) (*wsConn, error) {
	id, err := decodeHandle(handle)
	if err != nil {
		return nil, fmt.Errorf("ws: invalid handle: %w", err)
	}
	ws.mu.Lock()
	wsc := ws.conns[id]
	ws.mu.Unlock()
	if wsc == nil {
		return nil, fmt.Errorf("ws: invalid connection handle")
	}
	if atomic.LoadInt32(&wsc.closed) != 0 {
		return nil, fmt.Errorf("ws: connection is closed")
	}
	return wsc, nil
}

func (ws *wsService) SendText(handle []byte, text string) error {
	wsc, err := ws.getConn(handle)
	if err != nil {
		return err
	}
	return wsc.writeMessage(wsOpText, []byte(text))
}

func (ws *wsService) SendBytes(handle []byte, data []byte) error {
	wsc, err := ws.getConn(handle)
	if err != nil {
		return err
	}
	return wsc.writeMessage(wsOpBinary, data)
}

func (ws *wsService) SendPing(handle []byte, data []byte) error {
	wsc, err := ws.getConn(handle)
	if err != nil {
		return err
	}
	if len(data) > 125 {
		data = data[:125]
	}
	return wsc.writeMessage(wsOpPing, data)
}

func (ws *wsService) Receive(handle []byte) (*builtins.WSMessageData, error) {
	wsc, err := ws.getConn(handle)
	if err != nil {
		return nil, err
	}
	return wsc.readMessage()
}

func (ws *wsService) Close(handle []byte, code int, reason string) error {
	wsc, err := ws.getConn(handle)
	if err != nil {
		return err
	}

	if !atomic.CompareAndSwapInt32(&wsc.closed, 0, 1) {
		return nil
	}

	payload := make([]byte, 2+len(reason))
	binary.BigEndian.PutUint16(payload, uint16(code))
	copy(payload[2:], reason)
	if len(payload) > 125 {
		payload = payload[:125]
	}

	_ = wsc.writeFrame(true, wsOpClose, payload)

	wsc.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 512)
	for {
		_, err := wsc.conn.Read(buf)
		if err != nil {
			break
		}
	}

	ws.mu.Lock()
	delete(ws.conns, wsc.id)
	ws.mu.Unlock()

	return wsc.conn.Close()
}

func (ws *wsService) SetReadLimit(handle []byte, limit int64) error {
	wsc, err := ws.getConn(handle)
	if err != nil {
		return err
	}
	atomic.StoreInt64(&wsc.maxMsgSize, limit)
	return nil
}

func (ws *wsService) GetInfo(handle []byte) (*builtins.WSInfoData, error) {
	wsc, err := ws.getConn(handle)
	if err != nil {
		return nil, err
	}
	headersCopy := make(map[string]string, len(wsc.headers))
	for k, v := range wsc.headers {
		headersCopy[k] = v
	}
	return &builtins.WSInfoData{
		ID:         fmt.Sprintf("ws-%d", wsc.id),
		Path:       wsc.path,
		Headers:    headersCopy,
		Query:      wsc.query,
		RemoteAddr: wsc.remoteAddr,
	}, nil
}

// --- Frame-level I/O ---

func (wsc *wsConn) writeMessage(opcode int, data []byte) error {
	return wsc.writeFrame(true, opcode, data)
}

func (wsc *wsConn) writeFrame(fin bool, opcode int, payload []byte) error {
	wsc.writeMu.Lock()
	defer wsc.writeMu.Unlock()

	wsc.conn.SetWriteDeadline(time.Now().Add(wsDefaultWriteDeadline))

	var header [10]byte
	headerLen := 2

	b0 := byte(opcode)
	if fin {
		b0 |= 0x80
	}
	header[0] = b0

	payloadLen := len(payload)
	switch {
	case payloadLen <= 125:
		header[1] = byte(payloadLen)
	case payloadLen <= 65535:
		header[1] = 126
		binary.BigEndian.PutUint16(header[2:4], uint16(payloadLen))
		headerLen = 4
	default:
		header[1] = 127
		binary.BigEndian.PutUint64(header[2:10], uint64(payloadLen))
		headerLen = 10
	}

	if _, err := wsc.conn.Write(header[:headerLen]); err != nil {
		return fmt.Errorf("ws write: %w", err)
	}
	if len(payload) > 0 {
		if _, err := wsc.conn.Write(payload); err != nil {
			return fmt.Errorf("ws write: %w", err)
		}
	}
	return nil
}

func (wsc *wsConn) readMessage() (*builtins.WSMessageData, error) {
	var msgBuf []byte
	var msgOpcode int
	fragmented := false

	for {
		fin, opcode, payload, err := wsc.readFrame()
		if err != nil {
			if atomic.LoadInt32(&wsc.closed) != 0 {
				return &builtins.WSMessageData{Type: wsOpClose, Code: 1000}, nil
			}
			return nil, fmt.Errorf("ws read: %w", err)
		}

		if opcode >= 0x8 {
			switch opcode {
			case wsOpPing:
				if len(payload) > 125 {
					payload = payload[:125]
				}
				_ = wsc.writeFrame(true, wsOpPong, payload)
				continue
			case wsOpPong:
				continue
			case wsOpClose:
				code := 1000
				reason := ""
				if len(payload) >= 2 {
					code = int(binary.BigEndian.Uint16(payload[:2]))
					if len(payload) > 2 {
						reason = string(payload[2:])
					}
				}
				atomic.StoreInt32(&wsc.closed, 1)
				closePayload := make([]byte, 2)
				binary.BigEndian.PutUint16(closePayload, uint16(code))
				_ = wsc.writeFrame(true, wsOpClose, closePayload)
				return &builtins.WSMessageData{
					Type: wsOpClose,
					Code: code,
					Data: []byte(reason),
				}, nil
			}
			continue
		}

		if opcode == wsOpContinuation {
			if !fragmented {
				return nil, fmt.Errorf("ws: unexpected continuation frame")
			}
			msgBuf = append(msgBuf, payload...)
		} else {
			if fragmented {
				return nil, fmt.Errorf("ws: new message started before previous completed")
			}
			msgOpcode = opcode
			msgBuf = append(msgBuf[:0], payload...)
			if !fin {
				fragmented = true
			}
		}

		maxSize := atomic.LoadInt64(&wsc.maxMsgSize)
		if int64(len(msgBuf)) > maxSize {
			return nil, fmt.Errorf("ws: message size %d exceeds limit %d", len(msgBuf), maxSize)
		}

		if fin {
			return &builtins.WSMessageData{
				Type: msgOpcode,
				Data: msgBuf,
			}, nil
		}
	}
}

func (wsc *wsConn) readFrame() (fin bool, opcode int, payload []byte, err error) {
	header := make([]byte, 2)
	if _, err = io.ReadFull(wsc.reader, header); err != nil {
		return false, 0, nil, err
	}

	fin = header[0]&0x80 != 0
	opcode = int(header[0] & 0x0F)
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7F)

	switch length {
	case 126:
		buf := make([]byte, 2)
		if _, err = io.ReadFull(wsc.reader, buf); err != nil {
			return false, 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(buf))
	case 127:
		buf := make([]byte, 8)
		if _, err = io.ReadFull(wsc.reader, buf); err != nil {
			return false, 0, nil, err
		}
		length = binary.BigEndian.Uint64(buf)
	}

	maxSize := atomic.LoadInt64(&wsc.maxMsgSize)
	if int64(length) > maxSize {
		return false, 0, nil, fmt.Errorf("ws: frame payload %d exceeds limit %d", length, maxSize)
	}

	var maskKey [4]byte
	if masked {
		if _, err = io.ReadFull(wsc.reader, maskKey[:]); err != nil {
			return false, 0, nil, err
		}
	}

	payload = make([]byte, length)
	if length > 0 {
		if _, err = io.ReadFull(wsc.reader, payload); err != nil {
			return false, 0, nil, err
		}
		if masked {
			for i := range payload {
				payload[i] ^= maskKey[i%4]
			}
		}
	}

	return fin, opcode, payload, nil
}

// --- Helpers ---

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte(wsGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func containsTokenCaseInsensitive(header, token string) bool {
	for _, part := range strings.Split(header, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}

func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
