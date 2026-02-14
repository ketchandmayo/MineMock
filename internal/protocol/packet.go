package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type StatusResponse struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int32  `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int32 `json:"max"`
		Online int32 `json:"online"`
	} `json:"players"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
}

func ReadPacket(r io.Reader) ([]byte, error) {
	length, err := ReadVarInt(r)
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

func ReadPacketID(packet []byte) (int32, []byte, error) {
	id, n, err := decodeVarIntFromBytes(packet)
	if err != nil {
		return 0, nil, err
	}

	return id, packet[n:], nil
}

func ReadHandshakeNextState(packet []byte) (int32, error) {
	id, payload, err := ReadPacketID(packet)
	if err != nil {
		return 0, fmt.Errorf("read handshake id: %w", err)
	}
	if id != 0x00 {
		return 0, fmt.Errorf("unexpected handshake packet id: %d", id)
	}

	_, n, err := decodeVarIntFromBytes(payload) // protocol version
	if err != nil {
		return 0, fmt.Errorf("read protocol version: %w", err)
	}
	payload = payload[n:]

	hostLen, n, err := decodeVarIntFromBytes(payload)
	if err != nil {
		return 0, fmt.Errorf("read host len: %w", err)
	}
	payload = payload[n:]
	if hostLen < 0 || len(payload) < int(hostLen)+2 {
		return 0, fmt.Errorf("invalid host field")
	}
	payload = payload[hostLen:]

	if len(payload) < 2 {
		return 0, fmt.Errorf("missing port")
	}
	_ = binary.BigEndian.Uint16(payload[:2])
	payload = payload[2:]

	nextState, _, err := decodeVarIntFromBytes(payload)
	if err != nil {
		return 0, fmt.Errorf("read next state: %w", err)
	}

	return nextState, nil
}

func ReadLoginStartUsername(packet []byte) (string, error) {
	id, payload, err := ReadPacketID(packet)
	if err != nil {
		return "", fmt.Errorf("read login start id: %w", err)
	}
	if id != 0x00 {
		return "", fmt.Errorf("unexpected login start packet id: %d", id)
	}

	usernameLen, n, err := decodeVarIntFromBytes(payload)
	if err != nil {
		return "", fmt.Errorf("read username length: %w", err)
	}
	payload = payload[n:]
	if usernameLen <= 0 || len(payload) < int(usernameLen) {
		return "", fmt.Errorf("invalid username length")
	}

	return string(payload[:usernameLen]), nil
}

func SendLoginDisconnect(w io.Writer, message string) error {
	reasonPayload, err := loginDisconnectReasonPayload(message)
	if err != nil {
		return err
	}

	reason := string(reasonPayload)

	payload := make([]byte, 0, 1+len(reason)+5)
	payload = append(payload, 0x00) // Login Disconnect packet id
	payload = append(payload, EncodeVarInt(int32(len(reason)))...)
	payload = append(payload, []byte(reason)...)

	packetLen := EncodeVarInt(int32(len(payload)))
	packet := append(packetLen, payload...)

	_, err = w.Write(packet)
	return err
}

func SendLoginSuccess(w io.Writer, username string) error {
	payload := make([]byte, 0, 1+16+len(username)+16)
	payload = append(payload, 0x02) // Login Success packet id

	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return fmt.Errorf("generate uuid: %w", err)
	}
	payload = append(payload, uuid...)
	payload = append(payload, EncodeVarInt(int32(len(username)))...)
	payload = append(payload, []byte(username)...)
	payload = append(payload, 0x00) // properties count

	packetLen := EncodeVarInt(int32(len(payload)))
	packet := append(packetLen, payload...)

	_, err := w.Write(packet)
	return err
}

func SendPlayDisconnect(w io.Writer, message string) error {
	reasonPayload, err := loginDisconnectReasonPayload(message)
	if err != nil {
		return err
	}

	reason := string(reasonPayload)

	payload := make([]byte, 0, 1+len(reason)+5)
	payload = append(payload, 0x1A) // Play Disconnect packet id (1.20/1.20.1)
	payload = append(payload, EncodeVarInt(int32(len(reason)))...)
	payload = append(payload, []byte(reason)...)

	packetLen := EncodeVarInt(int32(len(payload)))
	packet := append(packetLen, payload...)

	_, err = w.Write(packet)
	return err
}

func loginDisconnectReasonPayload(message string) ([]byte, error) {
	trimmed := strings.TrimSpace(message)
	if trimmed != "" && json.Valid([]byte(trimmed)) {
		return []byte(trimmed), nil
	}

	type disconnectReason struct {
		Text string `json:"text"`
	}

	return json.Marshal(disconnectReason{Text: message})
}

func SendStatusResponse(w io.Writer, version string, protocolVersion int32, motd string, maxPlayers int32, onlinePlayers int32) error {
	status := StatusResponse{}
	status.Version.Name = version
	status.Version.Protocol = protocolVersion
	status.Players.Max = maxPlayers
	status.Players.Online = onlinePlayers
	status.Description.Text = motd

	response, err := json.Marshal(status)
	if err != nil {
		return err
	}

	payload := make([]byte, 0, 1+len(response)+5)
	payload = append(payload, 0x00)
	payload = append(payload, EncodeVarInt(int32(len(response)))...)
	payload = append(payload, response...)

	packetLen := EncodeVarInt(int32(len(payload)))
	packet := append(packetLen, payload...)

	_, err = w.Write(packet)
	return err
}

func SendPong(w io.Writer, pingPayload []byte) error {
	payload := make([]byte, 0, 1+len(pingPayload))
	payload = append(payload, 0x01)
	payload = append(payload, pingPayload...)

	packet := append(EncodeVarInt(int32(len(payload))), payload...)
	_, err := w.Write(packet)
	return err
}

func decodeVarIntFromBytes(data []byte) (int32, int, error) {
	var numRead int
	var result int32

	for {
		if numRead >= len(data) {
			return 0, 0, io.ErrUnexpectedEOF
		}

		value := int32(data[numRead] & 0x7F)
		result |= value << (7 * numRead)

		numRead++
		if numRead > 5 {
			return 0, 0, fmt.Errorf("varint is too big")
		}

		if data[numRead-1]&0x80 == 0 {
			break
		}
	}

	return result, numRead, nil
}
