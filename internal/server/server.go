package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"MineMock/internal/protocol"
)

type StatusConfig struct {
	MOTD          string
	VersionName   string
	Protocol      int32
	MaxPlayers    int32
	OnlinePlayers int32
}

type LoginConfig struct {
	ErrorMessage               string
	ErrorDelay                 time.Duration
	ForceConnectionLostTitle   bool
	RealServerAddr             string
	IsWhitelisted              func(username string) bool
	SimpleVoicechatListenAddr  string
	SimpleVoicechatBackendAddr string
}

func Run(addr string, statusCfg StatusConfig, loginCfg LoginConfig) error {
	voicechatProxy, err := newUDPProxy(loginCfg.SimpleVoicechatListenAddr, loginCfg.SimpleVoicechatBackendAddr)
	if err != nil {
		return fmt.Errorf("start UDP voice chat proxy: %w", err)
	}
	if voicechatProxy != nil {
		defer voicechatProxy.Close()
		go voicechatProxy.Run()
	} else {
		log.Println("UDP voice chat proxy disabled")
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer listener.Close()

	log.Println("Listening on " + addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		go handleConnection(conn, statusCfg, loginCfg, voicechatProxy)
	}
}

func handleConnection(conn net.Conn, statusCfg StatusConfig, loginCfg LoginConfig, voicechatProxy *udpProxy) {
	defer conn.Close()
	log.Println("New connection from", conn.RemoteAddr())

	handshakePacket, err := protocol.ReadPacket(conn)
	if err != nil {
		log.Println("Failed to read handshake:", err)
		return
	}

	nextState, err := protocol.ReadHandshakeNextState(handshakePacket)
	if err != nil {
		log.Println("Failed to parse handshake:", err)
		return
	}

	switch nextState {
	case 1:
		handleStatus(conn, statusCfg)
	case 2:
		handleLogin(conn, handshakePacket, loginCfg, voicechatProxy)
	default:
		log.Println("Unsupported next state:", nextState)
	}
}

func handleStatus(conn net.Conn, statusCfg StatusConfig) {
	requestPacket, err := protocol.ReadPacket(conn)
	if err != nil {
		log.Println("Failed to read status request:", err)
		return
	}

	packetID, _, err := protocol.ReadPacketID(requestPacket)
	if err != nil || packetID != 0x00 {
		log.Println("Invalid status request packet")
		return
	}

	if err := protocol.SendStatusResponse(conn, statusCfg.VersionName, statusCfg.Protocol, statusCfg.MOTD, statusCfg.MaxPlayers, statusCfg.OnlinePlayers); err != nil {
		log.Println("Failed to send status response:", err)
		return
	}

	pingPacket, err := protocol.ReadPacket(conn)
	if err != nil {
		log.Println("Failed to read ping request:", err)
		return
	}

	pingID, pingPayload, err := protocol.ReadPacketID(pingPacket)
	if err != nil || pingID != 0x01 {
		log.Println("Invalid ping request packet")
		return
	}

	if err := protocol.SendPong(conn, pingPayload); err != nil {
		log.Println("Failed to send pong:", err)
	}
}

func handleLogin(conn net.Conn, handshakePacket []byte, cfg LoginConfig, voicechatProxy *udpProxy) {
	loginStartPacket, err := protocol.ReadPacket(conn)
	if err != nil {
		log.Println("Failed to read login start:", err)
		return
	}

	username, err := protocol.ReadLoginStartUsername(loginStartPacket)
	if err != nil {
		log.Println("Failed to parse login start username:", err)
		return
	}

	remoteIP := conn.RemoteAddr().String()
	if host, _, err := net.SplitHostPort(remoteIP); err == nil {
		remoteIP = host
	}
	log.Printf("Login attempt: username=%q ip=%s", username, remoteIP)

	if shouldProxyPlayer(username, cfg) {
		if voicechatProxy != nil {
			voicechatProxy.AuthorizeIP(remoteIP)
		}
		if err := proxyToRealServer(conn, cfg.RealServerAddr, handshakePacket, loginStartPacket, username); err != nil {
			log.Printf("Proxy error for %q: %v", username, err)
			if sendErr := protocol.SendLoginDisconnect(conn, cfg.ErrorMessage); sendErr != nil {
				log.Println("Failed to send disconnect after proxy error:", sendErr)
			}
		}
		return
	}

	if cfg.ErrorDelay > 0 {
		time.Sleep(cfg.ErrorDelay)
	}

	if !cfg.ForceConnectionLostTitle {
		if err := protocol.SendLoginDisconnect(conn, cfg.ErrorMessage); err != nil {
			log.Println("Failed to send disconnect:", err)
		}
		return
	}

	if err := protocol.SendLoginSuccess(conn, username); err != nil {
		log.Println("Failed to send login success:", err)
		return
	}

	if err := protocol.SendPlayDisconnect(conn, cfg.ErrorMessage); err != nil {
		log.Println("Failed to send play disconnect:", err)
	}
}

func shouldProxyPlayer(username string, cfg LoginConfig) bool {
	if cfg.RealServerAddr == "" || cfg.IsWhitelisted == nil {
		return false
	}

	return cfg.IsWhitelisted(username)
}

func proxyToRealServer(clientConn net.Conn, serverAddr string, handshakePacket []byte, loginStartPacket []byte, username string) error {
	backendConn, err := net.DialTimeout("tcp", serverAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("connect to real server %s: %w", serverAddr, err)
	}
	defer backendConn.Close()

	if _, err := backendConn.Write(protocol.WrapPacket(handshakePacket)); err != nil {
		return fmt.Errorf("forward handshake: %w", err)
	}
	if _, err := backendConn.Write(protocol.WrapPacket(loginStartPacket)); err != nil {
		return fmt.Errorf("forward login start: %w", err)
	}

	log.Printf("Proxy enabled for username=%q -> %s", username, serverAddr)

	errCh := make(chan error, 2)
	go relayTraffic(backendConn, clientConn, errCh)
	go relayTraffic(clientConn, backendConn, errCh)

	firstErr := <-errCh
	secondErr := <-errCh

	if !isRelayClosed(firstErr) {
		return firstErr
	}
	if !isRelayClosed(secondErr) {
		return secondErr
	}

	return nil
}

func relayTraffic(dst net.Conn, src net.Conn, errCh chan<- error) {
	_, err := io.Copy(dst, src)

	if tcpConn, ok := dst.(*net.TCPConn); ok {
		_ = tcpConn.CloseWrite()
	}

	errCh <- err
}

func isRelayClosed(err error) bool {
	return err == nil || errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)
}

const (
	udpAuthorizationTTL = 10 * time.Minute
	udpSessionTTL       = 10 * time.Minute
	udpCleanupInterval  = time.Minute
	udpReadBufferSize   = 65535
)

type udpProxy struct {
	listener     *net.UDPConn
	backendAddr  *net.UDPAddr
	sessions     map[string]*udpSession
	authorizedIP map[string]time.Time
	mu           sync.Mutex
	done         chan struct{}
}

type udpSession struct {
	backendConn *net.UDPConn
	clientAddr  *net.UDPAddr
	lastSeen    time.Time
}

func newUDPProxy(listenAddr string, backendAddr string) (*udpProxy, error) {
	if listenAddr == "" || backendAddr == "" {
		return nil, nil
	}

	resolvedListenAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("resolve listen address %q: %w", listenAddr, err)
	}

	resolvedBackendAddr, err := net.ResolveUDPAddr("udp", backendAddr)
	if err != nil {
		return nil, fmt.Errorf("resolve backend address %q: %w", backendAddr, err)
	}

	listener, err := net.ListenUDP("udp", resolvedListenAddr)
	if err != nil {
		return nil, fmt.Errorf("listen UDP %q: %w", listenAddr, err)
	}

	return &udpProxy{
		listener:     listener,
		backendAddr:  resolvedBackendAddr,
		sessions:     map[string]*udpSession{},
		authorizedIP: map[string]time.Time{},
		done:         make(chan struct{}),
	}, nil
}

func (p *udpProxy) Run() {
	log.Printf("UDP voice chat proxy listening on %s -> %s", p.listener.LocalAddr(), p.backendAddr)

	go p.cleanupLoop()

	buffer := make([]byte, udpReadBufferSize)
	for {
		n, clientAddr, err := p.listener.ReadFromUDP(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("UDP proxy read error: %v", err)
			continue
		}

		payload := append([]byte(nil), buffer[:n]...)
		p.handleClientPacket(clientAddr, payload)
	}
}

func (p *udpProxy) Close() {
	select {
	case <-p.done:
		return
	default:
		close(p.done)
	}

	_ = p.listener.Close()

	p.mu.Lock()
	defer p.mu.Unlock()

	for key, session := range p.sessions {
		_ = session.backendConn.Close()
		delete(p.sessions, key)
	}
}

func (p *udpProxy) AuthorizeIP(clientIP string) {
	if clientIP == "" {
		return
	}

	p.mu.Lock()
	p.authorizedIP[clientIP] = time.Now().Add(udpAuthorizationTTL)
	p.mu.Unlock()
}

func (p *udpProxy) handleClientPacket(clientAddr *net.UDPAddr, payload []byte) {
	if !p.isAuthorized(clientAddr.IP.String()) {
		return
	}

	session, err := p.getOrCreateSession(clientAddr)
	if err != nil {
		log.Printf("UDP proxy session error for %s: %v", clientAddr, err)
		return
	}

	if _, err := session.backendConn.Write(payload); err != nil {
		log.Printf("UDP proxy forward error for %s: %v", clientAddr, err)
		p.removeSession(clientAddr.String())
		return
	}

	p.touchSession(clientAddr.String())
}

func (p *udpProxy) getOrCreateSession(clientAddr *net.UDPAddr) (*udpSession, error) {
	key := clientAddr.String()

	p.mu.Lock()
	if existing, ok := p.sessions[key]; ok {
		existing.lastSeen = time.Now()
		p.mu.Unlock()
		return existing, nil
	}
	p.mu.Unlock()

	backendConn, err := net.DialUDP("udp", nil, p.backendAddr)
	if err != nil {
		return nil, fmt.Errorf("dial backend %s: %w", p.backendAddr, err)
	}

	session := &udpSession{
		backendConn: backendConn,
		clientAddr:  clientAddr,
		lastSeen:    time.Now(),
	}

	p.mu.Lock()
	p.sessions[key] = session
	p.mu.Unlock()

	go p.relayBackendToClient(key, session)

	return session, nil
}

func (p *udpProxy) relayBackendToClient(sessionKey string, session *udpSession) {
	buffer := make([]byte, udpReadBufferSize)
	for {
		n, err := session.backendConn.Read(buffer)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("UDP proxy backend read error for %s: %v", session.clientAddr, err)
			}
			p.removeSession(sessionKey)
			return
		}

		if _, err := p.listener.WriteToUDP(buffer[:n], session.clientAddr); err != nil {
			log.Printf("UDP proxy write to client error for %s: %v", session.clientAddr, err)
			p.removeSession(sessionKey)
			return
		}

		p.touchSession(sessionKey)
	}
}

func (p *udpProxy) cleanupLoop() {
	ticker := time.NewTicker(udpCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.cleanupExpiredEntries()
		}
	}
}

func (p *udpProxy) cleanupExpiredEntries() {
	now := time.Now()
	var toClose []*net.UDPConn

	p.mu.Lock()
	for ip, expiresAt := range p.authorizedIP {
		if now.After(expiresAt) {
			delete(p.authorizedIP, ip)
		}
	}

	for key, session := range p.sessions {
		if now.Sub(session.lastSeen) <= udpSessionTTL {
			continue
		}
		toClose = append(toClose, session.backendConn)
		delete(p.sessions, key)
	}
	p.mu.Unlock()

	for _, conn := range toClose {
		_ = conn.Close()
	}
}

func (p *udpProxy) touchSession(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, ok := p.sessions[key]
	if !ok {
		return
	}

	session.lastSeen = time.Now()
}

func (p *udpProxy) removeSession(key string) {
	p.mu.Lock()
	session, ok := p.sessions[key]
	if ok {
		delete(p.sessions, key)
	}
	p.mu.Unlock()

	if ok {
		_ = session.backendConn.Close()
	}
}

func (p *udpProxy) isAuthorized(ip string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	expiresAt, ok := p.authorizedIP[ip]
	if !ok {
		return false
	}

	if time.Now().After(expiresAt) {
		delete(p.authorizedIP, ip)
		return false
	}

	return true
}
