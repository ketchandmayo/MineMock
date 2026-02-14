package server

import (
	"fmt"
	"net"

	"MineMock/internal/protocol"
)

type StatusConfig struct {
	MOTD          string
	VersionName   string
	Protocol      int32
	MaxPlayers    int32
	OnlinePlayers int32
}

func Run(addr string, errorMessage string, forceConnectionLostTitle bool, statusCfg StatusConfig) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer listener.Close()

	fmt.Println("Listening on " + addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}

		go handleConnection(conn, errorMessage, forceConnectionLostTitle, statusCfg)
	}
}

func handleConnection(conn net.Conn, errorMessage string, forceConnectionLostTitle bool, statusCfg StatusConfig) {
	defer conn.Close()
	fmt.Println("New connection from", conn.RemoteAddr())

	handshakePacket, err := protocol.ReadPacket(conn)
	if err != nil {
		fmt.Println("Failed to read handshake:", err)
		return
	}

	nextState, err := protocol.ReadHandshakeNextState(handshakePacket)
	if err != nil {
		fmt.Println("Failed to parse handshake:", err)
		return
	}

	switch nextState {
	case 1:
		handleStatus(conn, statusCfg)
	case 2:
		handleLogin(conn, errorMessage, forceConnectionLostTitle)
	default:
		fmt.Println("Unsupported next state:", nextState)
	}
}

func handleStatus(conn net.Conn, statusCfg StatusConfig) {
	requestPacket, err := protocol.ReadPacket(conn)
	if err != nil {
		fmt.Println("Failed to read status request:", err)
		return
	}

	packetID, _, err := protocol.ReadPacketID(requestPacket)
	if err != nil || packetID != 0x00 {
		fmt.Println("Invalid status request packet")
		return
	}

	if err := protocol.SendStatusResponse(conn, statusCfg.VersionName, statusCfg.Protocol, statusCfg.MOTD, statusCfg.MaxPlayers, statusCfg.OnlinePlayers); err != nil {
		fmt.Println("Failed to send status response:", err)
		return
	}

	pingPacket, err := protocol.ReadPacket(conn)
	if err != nil {
		fmt.Println("Failed to read ping request:", err)
		return
	}

	pingID, pingPayload, err := protocol.ReadPacketID(pingPacket)
	if err != nil || pingID != 0x01 {
		fmt.Println("Invalid ping request packet")
		return
	}

	if err := protocol.SendPong(conn, pingPayload); err != nil {
		fmt.Println("Failed to send pong:", err)
	}
}

func handleLogin(conn net.Conn, errorMessage string, forceConnectionLostTitle bool) {
	loginStartPacket, err := protocol.ReadPacket(conn)
	if err != nil {
		fmt.Println("Failed to read login start:", err)
		return
	}

	if !forceConnectionLostTitle {
		if err := protocol.SendLoginDisconnect(conn, errorMessage); err != nil {
			fmt.Println("Failed to send disconnect:", err)
		}
		return
	}

	username, err := protocol.ReadLoginStartUsername(loginStartPacket)
	if err != nil {
		fmt.Println("Failed to parse login start username:", err)
		return
	}

	if err := protocol.SendLoginSuccess(conn, username); err != nil {
		fmt.Println("Failed to send login success:", err)
		return
	}

	if err := protocol.SendPlayDisconnect(conn, errorMessage); err != nil {
		fmt.Println("Failed to send play disconnect:", err)
	}
}
