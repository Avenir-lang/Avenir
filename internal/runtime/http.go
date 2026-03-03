package runtime

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"avenir/internal/runtime/builtins"
)

type httpService struct {
	nextID   uint64
	mu       sync.Mutex
	servers  map[uint64]net.Listener
	requests map[uint64]*httpRequest
}

type httpRequest struct {
	conn       net.Conn
	method     string
	path       string
	headers    map[string]string
	body       []byte
	protoMajor int
	protoMinor int
}

func newHTTPService() *httpService {
	return &httpService{
		servers:  make(map[uint64]net.Listener),
		requests: make(map[uint64]*httpRequest),
	}
}

func (h *httpService) Request(method string, url string, headers map[string]string, body []byte) (*builtins.HTTPResponseData, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respHeaders := make(map[string]string, len(resp.Header))
	for k, vals := range resp.Header {
		respHeaders[k] = strings.Join(vals, ", ")
	}
	return &builtins.HTTPResponseData{
		Status:  resp.StatusCode,
		Headers: respHeaders,
		Body:    data,
	}, nil
}

func (h *httpService) Listen(host string, port int) ([]byte, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port %d", port)
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	id := h.nextHandle()
	h.mu.Lock()
	h.servers[id] = ln
	h.mu.Unlock()
	return encodeHandle(id), nil
}

func (h *httpService) Accept(serverHandle []byte) (*builtins.HTTPRequestData, error) {
	id, err := decodeHandle(serverHandle)
	if err != nil {
		return nil, err
	}
	h.mu.Lock()
	ln := h.servers[id]
	h.mu.Unlock()
	if ln == nil {
		return nil, fmt.Errorf("invalid server handle")
	}
	conn, err := ln.Accept()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		conn.Close()
		return nil, err
	}
	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		conn.Close()
		return nil, err
	}
	headers := make(map[string]string, len(req.Header))
	for k, vals := range req.Header {
		headers[k] = strings.Join(vals, ", ")
	}
	path := req.RequestURI
	if path == "" && req.URL != nil {
		path = req.URL.Path
	}
	reqID := h.nextHandle()
	h.mu.Lock()
	h.requests[reqID] = &httpRequest{
		conn:       conn,
		method:     req.Method,
		path:       path,
		headers:    headers,
		body:       body,
		protoMajor: req.ProtoMajor,
		protoMinor: req.ProtoMinor,
	}
	h.mu.Unlock()
	return &builtins.HTTPRequestData{
		Handle:  encodeHandle(reqID),
		Method:  req.Method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}, nil
}

func (h *httpService) Respond(reqHandle []byte, status int, headers map[string]string, body []byte) error {
	reqID, err := decodeHandle(reqHandle)
	if err != nil {
		return err
	}
	h.mu.Lock()
	req := h.requests[reqID]
	if req != nil {
		delete(h.requests, reqID)
	}
	h.mu.Unlock()
	if req == nil {
		return fmt.Errorf("invalid request handle")
	}
	respHeaders := http.Header{}
	for k, v := range headers {
		respHeaders.Set(k, v)
	}
	resp := &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		ProtoMajor: req.protoMajor,
		ProtoMinor: req.protoMinor,
		Header:     respHeaders,
		Body:       io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
		Close:         true,
	}
	if err := resp.Write(req.conn); err != nil {
		req.conn.Close()
		return err
	}
	return req.conn.Close()
}

func (h *httpService) nextHandle() uint64 {
	return atomic.AddUint64(&h.nextID, 1)
}
