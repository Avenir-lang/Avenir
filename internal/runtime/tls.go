package runtime

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/acme/autocert"

	"avenir/internal/runtime/builtins"
)

type tlsService struct {
	nextID    uint64
	mu        sync.Mutex
	conns     map[uint64]*tls.Conn
	listeners map[uint64]net.Listener
	certs     map[uint64]*tls.Certificate
}

func newTLSService() *tlsService {
	return &tlsService{
		conns:     make(map[uint64]*tls.Conn),
		listeners: make(map[uint64]net.Listener),
		certs:     make(map[uint64]*tls.Certificate),
	}
}

func (t *tlsService) secureDefaults() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{"h2", "http/1.1"},
	}
}

func (t *tlsService) buildTLSConfig(cfg *builtins.TLSConfigData) (*tls.Config, error) {
	tlsCfg := t.secureDefaults()

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("tls: failed to load certificate: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	if cfg.MinVersion != "" {
		v, err := parseTLSVersion(cfg.MinVersion)
		if err != nil {
			return nil, err
		}
		tlsCfg.MinVersion = v
	}
	if cfg.MaxVersion != "" {
		v, err := parseTLSVersion(cfg.MaxVersion)
		if err != nil {
			return nil, err
		}
		tlsCfg.MaxVersion = v
	}

	if len(cfg.ALPNProtocols) > 0 {
		tlsCfg.NextProtos = cfg.ALPNProtocols
	}

	if cfg.ServerName != "" {
		tlsCfg.ServerName = cfg.ServerName
	}

	if cfg.InsecureSkipVerify {
		tlsCfg.InsecureSkipVerify = true
	}

	if cfg.ClientAuth != "" {
		ca, err := parseClientAuth(cfg.ClientAuth)
		if err != nil {
			return nil, err
		}
		tlsCfg.ClientAuth = ca
	}

	if len(cfg.ClientCAs) > 0 {
		pool := x509.NewCertPool()
		for _, caFile := range cfg.ClientCAs {
			pem, err := os.ReadFile(caFile)
			if err != nil {
				return nil, fmt.Errorf("tls: failed to read CA file %s: %w", caFile, err)
			}
			if !pool.AppendCertsFromPEM(pem) {
				return nil, fmt.Errorf("tls: failed to parse CA certificate from %s", caFile)
			}
		}
		tlsCfg.ClientCAs = pool
	}

	return tlsCfg, nil
}

func (t *tlsService) Connect(host string, port int, serverName string) ([]byte, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("tls: invalid port %d", port)
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	cfg := t.secureDefaults()
	if serverName != "" {
		cfg.ServerName = serverName
	} else {
		cfg.ServerName = host
	}
	conn, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("tls: connect failed: %w", err)
	}
	id := t.nextHandle()
	t.mu.Lock()
	t.conns[id] = conn
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) ConnectConfig(host string, port int, cfg *builtins.TLSConfigData) ([]byte, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("tls: invalid port %d", port)
	}
	tlsCfg, err := t.buildTLSConfig(cfg)
	if err != nil {
		return nil, err
	}
	if tlsCfg.ServerName == "" {
		tlsCfg.ServerName = host
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("tls: connect failed: %w", err)
	}
	id := t.nextHandle()
	t.mu.Lock()
	t.conns[id] = conn
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) Listen(host string, port int, certFile, keyFile string) ([]byte, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("tls: invalid port %d", port)
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: failed to load certificate: %w", err)
	}
	cfg := t.secureDefaults()
	cfg.Certificates = []tls.Certificate{cert}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	ln, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		return nil, fmt.Errorf("tls: listen failed: %w", err)
	}
	id := t.nextHandle()
	t.mu.Lock()
	t.listeners[id] = ln
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) ListenConfig(host string, port int, cfg *builtins.TLSConfigData) ([]byte, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("tls: invalid port %d", port)
	}
	tlsCfg, err := t.buildTLSConfig(cfg)
	if err != nil {
		return nil, err
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	ln, err := tls.Listen("tcp", addr, tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("tls: listen failed: %w", err)
	}
	id := t.nextHandle()
	t.mu.Lock()
	t.listeners[id] = ln
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) ListenAutoTLS(host string, port int, domain, email string) ([]byte, error) {
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("tls: invalid port %d", port)
	}
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache("avenir-certs"),
		Email:      email,
	}
	tlsCfg := m.TLSConfig()
	tlsCfg.MinVersion = tls.VersionTLS12
	tlsCfg.NextProtos = append(tlsCfg.NextProtos, "h2", "http/1.1")

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	ln, err := tls.Listen("tcp", addr, tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("tls: auto-tls listen failed: %w", err)
	}
	go http.ListenAndServe(":80", m.HTTPHandler(nil))

	id := t.nextHandle()
	t.mu.Lock()
	t.listeners[id] = ln
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) Accept(listenerHandle []byte) ([]byte, error) {
	lid, err := decodeHandle(listenerHandle)
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	ln := t.listeners[lid]
	t.mu.Unlock()
	if ln == nil {
		return nil, fmt.Errorf("tls: invalid listener handle")
	}
	conn, err := ln.Accept()
	if err != nil {
		return nil, fmt.Errorf("tls: accept failed: %w", err)
	}
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		conn.Close()
		return nil, fmt.Errorf("tls: accepted connection is not TLS")
	}
	id := t.nextHandle()
	t.mu.Lock()
	t.conns[id] = tlsConn
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) Read(connHandle []byte, n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("tls: invalid read size %d", n)
	}
	id, err := decodeHandle(connHandle)
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	conn := t.conns[id]
	t.mu.Unlock()
	if conn == nil {
		return nil, fmt.Errorf("tls: invalid connection handle")
	}
	if n == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, n)
	nread, err := conn.Read(buf)
	if err != nil && nread == 0 {
		return nil, fmt.Errorf("tls: read failed: %w", err)
	}
	return buf[:nread], nil
}

func (t *tlsService) Write(connHandle []byte, data []byte) (int, error) {
	id, err := decodeHandle(connHandle)
	if err != nil {
		return 0, err
	}
	t.mu.Lock()
	conn := t.conns[id]
	t.mu.Unlock()
	if conn == nil {
		return 0, fmt.Errorf("tls: invalid connection handle")
	}
	return conn.Write(data)
}

func (t *tlsService) Close(handle []byte) error {
	id, err := decodeHandle(handle)
	if err != nil {
		return err
	}
	t.mu.Lock()
	if conn, ok := t.conns[id]; ok {
		delete(t.conns, id)
		t.mu.Unlock()
		return conn.Close()
	}
	if ln, ok := t.listeners[id]; ok {
		delete(t.listeners, id)
		t.mu.Unlock()
		return ln.Close()
	}
	t.mu.Unlock()
	return fmt.Errorf("tls: invalid handle")
}

func (t *tlsService) CloseListener(handle []byte) error {
	id, err := decodeHandle(handle)
	if err != nil {
		return err
	}
	t.mu.Lock()
	ln, ok := t.listeners[id]
	if ok {
		delete(t.listeners, id)
	}
	t.mu.Unlock()
	if !ok {
		return fmt.Errorf("tls: invalid listener handle")
	}
	return ln.Close()
}

func (t *tlsService) LoadCert(certFile, keyFile string) ([]byte, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: failed to load certificate: %w", err)
	}
	id := t.nextHandle()
	t.mu.Lock()
	t.certs[id] = &cert
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) LoadCertChain(files []string) ([]byte, error) {
	if len(files) < 2 {
		return nil, fmt.Errorf("tls: loadCertChain requires at least 2 files (cert + key)")
	}
	certFile := files[0]
	keyFile := files[1]
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: failed to load certificate chain: %w", err)
	}
	id := t.nextHandle()
	t.mu.Lock()
	t.certs[id] = &cert
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) LoadCertPEM(certPEM, keyPEM []byte) ([]byte, error) {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("tls: failed to parse PEM certificate: %w", err)
	}
	id := t.nextHandle()
	t.mu.Lock()
	t.certs[id] = &cert
	t.mu.Unlock()
	return encodeHandle(id), nil
}

func (t *tlsService) CertInfo(certHandle []byte) (map[string]interface{}, error) {
	id, err := decodeHandle(certHandle)
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	cert := t.certs[id]
	t.mu.Unlock()
	if cert == nil {
		return nil, fmt.Errorf("tls: invalid certificate handle")
	}
	if cert.Leaf == nil && len(cert.Certificate) > 0 {
		parsed, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return nil, fmt.Errorf("tls: failed to parse certificate: %w", err)
		}
		cert.Leaf = parsed
	}
	if cert.Leaf == nil {
		return nil, fmt.Errorf("tls: certificate has no leaf")
	}
	info := map[string]interface{}{
		"subject":   cert.Leaf.Subject.String(),
		"issuer":    cert.Leaf.Issuer.String(),
		"notBefore": cert.Leaf.NotBefore.Unix(),
		"notAfter":  cert.Leaf.NotAfter.Unix(),
		"isCA":      cert.Leaf.IsCA,
		"dnsNames":  cert.Leaf.DNSNames,
	}
	return info, nil
}

func (t *tlsService) PeerCerts(connHandle []byte) ([]map[string]interface{}, error) {
	id, err := decodeHandle(connHandle)
	if err != nil {
		return nil, err
	}
	t.mu.Lock()
	conn := t.conns[id]
	t.mu.Unlock()
	if conn == nil {
		return nil, fmt.Errorf("tls: invalid connection handle")
	}
	state := conn.ConnectionState()
	var result []map[string]interface{}
	for _, cert := range state.PeerCertificates {
		info := map[string]interface{}{
			"subject":   cert.Subject.String(),
			"issuer":    cert.Issuer.String(),
			"notBefore": cert.NotBefore.Unix(),
			"notAfter":  cert.NotAfter.Unix(),
			"isCA":      cert.IsCA,
			"dnsNames":  cert.DNSNames,
		}
		result = append(result, info)
	}
	return result, nil
}

func (t *tlsService) NegotiatedProtocol(connHandle []byte) (string, error) {
	id, err := decodeHandle(connHandle)
	if err != nil {
		return "", err
	}
	t.mu.Lock()
	conn := t.conns[id]
	t.mu.Unlock()
	if conn == nil {
		return "", fmt.Errorf("tls: invalid connection handle")
	}
	return conn.ConnectionState().NegotiatedProtocol, nil
}

func (t *tlsService) TLSVersion(connHandle []byte) (string, error) {
	id, err := decodeHandle(connHandle)
	if err != nil {
		return "", err
	}
	t.mu.Lock()
	conn := t.conns[id]
	t.mu.Unlock()
	if conn == nil {
		return "", fmt.Errorf("tls: invalid connection handle")
	}
	return formatTLSVersion(conn.ConnectionState().Version), nil
}

func (t *tlsService) HTTPSRequest(method, url string, headers map[string]string, body []byte, cfg *builtins.TLSConfigData) (*builtins.HTTPResponseData, error) {
	tlsCfg, err := t.buildTLSConfig(cfg)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}
	client := &http.Client{Transport: transport}
	return doHTTPRequest(client, method, url, headers, body)
}

func (t *tlsService) HTTPSListen(host string, port int, certFile, keyFile string) ([]byte, error) {
	return t.Listen(host, port, certFile, keyFile)
}

func (t *tlsService) HTTPSListenConfig(host string, port int, cfg *builtins.TLSConfigData) ([]byte, error) {
	return t.ListenConfig(host, port, cfg)
}

func (t *tlsService) HTTPSListenAuto(host string, port int, domain, email string) ([]byte, error) {
	return t.ListenAutoTLS(host, port, domain, email)
}

func (t *tlsService) nextHandle() uint64 {
	return atomic.AddUint64(&t.nextID, 1)
}

func parseTLSVersion(v string) (uint16, error) {
	switch v {
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.2":
		return tls.VersionTLS12, nil
	case "1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("tls: unknown version %q (valid: 1.0, 1.1, 1.2, 1.3)", v)
	}
}

func parseClientAuth(s string) (tls.ClientAuthType, error) {
	switch s {
	case "", "none":
		return tls.NoClientCert, nil
	case "request":
		return tls.RequestClientCert, nil
	case "require":
		return tls.RequireAnyClientCert, nil
	case "verify":
		return tls.VerifyClientCertIfGiven, nil
	case "requireAndVerify":
		return tls.RequireAndVerifyClientCert, nil
	default:
		return 0, fmt.Errorf("tls: unknown clientAuth %q", s)
	}
}

func formatTLSVersion(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return "unknown"
	}
}

func doHTTPRequest(client *http.Client, method, url string, headers map[string]string, body []byte) (*builtins.HTTPResponseData, error) {
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
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respHeaders := make(map[string]string, len(resp.Header))
	for k, vals := range resp.Header {
		if len(vals) > 0 {
			respHeaders[k] = vals[0]
		}
	}
	return &builtins.HTTPResponseData{
		Status:  resp.StatusCode,
		Headers: respHeaders,
		Body:    respBody,
	}, nil
}
