package protocol

import (
	"fmt"
	"io"
)

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

func SendLoginDisconnect(w io.Writer, message string) error {
	reason := fmt.Sprintf(`{"text":"%s"}`, message)

	payload := make([]byte, 0, 1+len(reason)+5)
	payload = append(payload, 0x00) // Login Disconnect packet id
	payload = append(payload, EncodeVarInt(int32(len(reason)))...)
	payload = append(payload, []byte(reason)...)

	packetLen := EncodeVarInt(int32(len(payload)))
	packet := append(packetLen, payload...)

	_, err := w.Write(packet)
	return err
}
