package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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
	ErrorMessage             string
	ErrorDelay               time.Duration
	ForceConnectionLostTitle bool
	RealServerAddr           string
	IsWhitelisted            func(username string) bool
}

func Run(addr string, statusCfg StatusConfig, loginCfg LoginConfig) error {
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

		go handleConnection(conn, statusCfg, loginCfg)
	}
}

func handleConnection(conn net.Conn, statusCfg StatusConfig, loginCfg LoginConfig) {
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
		handleLogin(conn, handshakePacket, loginCfg)
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

func handleLogin(conn net.Conn, handshakePacket []byte, cfg LoginConfig) {
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
