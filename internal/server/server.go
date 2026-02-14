package server

import (
	"fmt"
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

func Run(addr string, errorMessage string, errorDelay time.Duration, forceConnectionLostTitle bool, statusCfg StatusConfig) error {
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

		go handleConnection(conn, errorMessage, errorDelay, forceConnectionLostTitle, statusCfg)
	}
}

func handleConnection(conn net.Conn, errorMessage string, errorDelay time.Duration, forceConnectionLostTitle bool, statusCfg StatusConfig) {
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
		handleLogin(conn, errorMessage, errorDelay, forceConnectionLostTitle)
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

func handleLogin(conn net.Conn, errorMessage string, errorDelay time.Duration, forceConnectionLostTitle bool) {
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

	if errorDelay > 0 {
		time.Sleep(errorDelay)
	}

	if !forceConnectionLostTitle {
		if err := protocol.SendLoginDisconnect(conn, errorMessage); err != nil {
			log.Println("Failed to send disconnect:", err)
		}
		return
	}

	if err := protocol.SendLoginSuccess(conn, username); err != nil {
		log.Println("Failed to send login success:", err)
		return
	}

	if err := protocol.SendPlayDisconnect(conn, errorMessage); err != nil {
		log.Println("Failed to send play disconnect:", err)
	}
}
