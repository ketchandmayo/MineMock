package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	port := "25565"
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Listening on 0.0.0.0:" + port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("New connection from", conn.RemoteAddr())

	if _, err := readPacket(conn); err != nil {
		fmt.Println("Failed to read handshake:", err)
		return
	}

	if _, err := readPacket(conn); err != nil {
		fmt.Println("Failed to read login start:", err)
		return
	}

	if err := sendLoginDisconnect(conn, "Сервер временно недоступен"); err != nil {
		fmt.Println("Failed to send disconnect:", err)
	}
}

func readPacket(r io.Reader) ([]byte, error) {
	length, err := readVarInt(r)
	if err != nil {
		return nil, err
	}
	if length <= 0 {
		return nil, fmt.Errorf("invalid packet length: %d", length)
	}

	packet := make([]byte, length)
	if _, err := io.ReadFull(r, packet); err != nil {
		return nil, err
	}

	return packet, nil
}

func sendLoginDisconnect(w io.Writer, message string) error {
	reason := fmt.Sprintf(`{"text":"%s"}`, message)

	payload := make([]byte, 0, 1+len(reason)+5)
	payload = append(payload, 0x00) // Login Disconnect packet id
	payload = append(payload, encodeVarInt(int32(len(reason)))...)
	payload = append(payload, []byte(reason)...)

	packetLen := encodeVarInt(int32(len(payload)))
	packet := append(packetLen, payload...)

	_, err := w.Write(packet)
	return err
}

func readVarInt(r io.Reader) (int32, error) {
	var numRead int32
	var result int32
	var read [1]byte

	for {
		if _, err := io.ReadFull(r, read[:]); err != nil {
			return 0, err
		}

		value := int32(read[0] & 0x7F)
		result |= value << (7 * numRead)

		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("varint is too big")
		}

		if read[0]&0x80 == 0 {
			break
		}
	}

	return result, nil
}

func encodeVarInt(value int32) []byte {
	buf := make([]byte, binary.MaxVarintLen32)
	n := 0

	for {
		if (value & ^0x7F) == 0 {
			buf[n] = byte(value)
			n++
			break
		}

		buf[n] = byte(value&0x7F | 0x80)
		value >>= 7
		n++
	}

	return buf[:n]
}
