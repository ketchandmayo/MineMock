package server

import (
	"fmt"
	"net"

	"MineMock/internal/protocol"
)

func Run(addr string, errorMessage string) error {
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

		go handleConnection(conn, errorMessage)
	}
}

func handleConnection(conn net.Conn, errorMessage string) {
	defer conn.Close()
	fmt.Println("New connection from", conn.RemoteAddr())

	if _, err := protocol.ReadPacket(conn); err != nil {
		fmt.Println("Failed to read handshake:", err)
		return
	}

	if _, err := protocol.ReadPacket(conn); err != nil {
		fmt.Println("Failed to read login start:", err)
		return
	}

	if err := protocol.SendLoginDisconnect(conn, errorMessage); err != nil {
		fmt.Println("Failed to send disconnect:", err)
	}
}
