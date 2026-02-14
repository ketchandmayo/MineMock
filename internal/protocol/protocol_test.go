package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestVarIntRoundTrip(t *testing.T) {
	values := []int32{0, 1, 127, 128, 255, 2147483647}

	for _, v := range values {
		encoded := EncodeVarInt(v)
		decoded, err := ReadVarInt(bytes.NewReader(encoded))
		if err != nil {
			t.Fatalf("ReadVarInt failed for %d: %v", v, err)
		}
		if decoded != v {
			t.Fatalf("expected %d, got %d", v, decoded)
		}
	}
}

func TestSendLoginDisconnect(t *testing.T) {
	var out bytes.Buffer
	if err := SendLoginDisconnect(&out, "test"); err != nil {
		t.Fatalf("SendLoginDisconnect failed: %v", err)
	}

	packet, err := ReadPacket(&out)
	if err != nil {
		t.Fatalf("ReadPacket failed: %v", err)
	}

	if len(packet) == 0 {
		t.Fatal("packet is empty")
	}
	if packet[0] != 0x00 {
		t.Fatalf("unexpected packet id: %d", packet[0])
	}

	_, payload, err := ReadPacketID(packet)
	if err != nil {
		t.Fatalf("ReadPacketID failed: %v", err)
	}

	jsonLen, n, err := decodeVarIntFromBytes(payload)
	if err != nil {
		t.Fatalf("decode reason length failed: %v", err)
	}
	reasonPayload := payload[n:]
	if int(jsonLen) != len(reasonPayload) {
		t.Fatalf("reason payload length mismatch: declared %d, got %d", jsonLen, len(reasonPayload))
	}

	var reason map[string]string
	if err := json.Unmarshal(reasonPayload, &reason); err != nil {
		t.Fatalf("reason json unmarshal failed: %v", err)
	}
	if reason["text"] != "test" {
		t.Fatalf("unexpected reason text: %q", reason["text"])
	}
}

func TestSendLoginDisconnect_AllowsRawJSONComponent(t *testing.T) {
	var out bytes.Buffer
	rawReason := `{"text":"Соединение потеряно"}`

	if err := SendLoginDisconnect(&out, rawReason); err != nil {
		t.Fatalf("SendLoginDisconnect failed: %v", err)
	}

	packet, err := ReadPacket(&out)
	if err != nil {
		t.Fatalf("ReadPacket failed: %v", err)
	}

	_, payload, err := ReadPacketID(packet)
	if err != nil {
		t.Fatalf("ReadPacketID failed: %v", err)
	}

	jsonLen, n, err := decodeVarIntFromBytes(payload)
	if err != nil {
		t.Fatalf("decode reason length failed: %v", err)
	}
	reasonPayload := payload[n:]
	if int(jsonLen) != len(reasonPayload) {
		t.Fatalf("reason payload length mismatch: declared %d, got %d", jsonLen, len(reasonPayload))
	}

	if string(reasonPayload) != rawReason {
		t.Fatalf("expected raw json component to be preserved, got %s", string(reasonPayload))
	}
}

func TestReadLoginStartUsername(t *testing.T) {
	payload := make([]byte, 0)
	payload = append(payload, EncodeVarInt(0x00)...)
	payload = append(payload, EncodeVarInt(int32(len("Steve")))...)
	payload = append(payload, []byte("Steve")...)

	username, err := ReadLoginStartUsername(payload)
	if err != nil {
		t.Fatalf("ReadLoginStartUsername failed: %v", err)
	}
	if username != "Steve" {
		t.Fatalf("unexpected username: %s", username)
	}
}

func TestSendLoginSuccess(t *testing.T) {
	var out bytes.Buffer
	if err := SendLoginSuccess(&out, "Steve"); err != nil {
		t.Fatalf("SendLoginSuccess failed: %v", err)
	}

	packet, err := ReadPacket(&out)
	if err != nil {
		t.Fatalf("ReadPacket failed: %v", err)
	}
	packetID, payload, err := ReadPacketID(packet)
	if err != nil {
		t.Fatalf("ReadPacketID failed: %v", err)
	}
	if packetID != 0x02 {
		t.Fatalf("unexpected login success packet id: %d", packetID)
	}
	if len(payload) < 16 {
		t.Fatalf("payload too short for uuid: %d", len(payload))
	}
	payload = payload[16:]

	usernameLen, n, err := decodeVarIntFromBytes(payload)
	if err != nil {
		t.Fatalf("decode username length failed: %v", err)
	}
	payload = payload[n:]
	if len(payload) < int(usernameLen)+1 {
		t.Fatalf("payload too short for username/properties")
	}
	if string(payload[:usernameLen]) != "Steve" {
		t.Fatalf("unexpected username in login success: %s", string(payload[:usernameLen]))
	}
	if payload[usernameLen] != 0x00 {
		t.Fatalf("expected zero properties, got %d", payload[usernameLen])
	}
}

func TestReadHandshakeNextState(t *testing.T) {
	handshake := make([]byte, 0)
	handshake = append(handshake, EncodeVarInt(0x00)...)
	handshake = append(handshake, EncodeVarInt(760)...)
	handshake = append(handshake, EncodeVarInt(int32(len("localhost")))...)
	handshake = append(handshake, []byte("localhost")...)
	handshake = append(handshake, 0x63, 0xDD) // 25565
	handshake = append(handshake, EncodeVarInt(1)...)

	nextState, err := ReadHandshakeNextState(handshake)
	if err != nil {
		t.Fatalf("ReadHandshakeNextState failed: %v", err)
	}
	if nextState != 1 {
		t.Fatalf("expected status state 1, got %d", nextState)
	}
}

func TestSendStatusResponse(t *testing.T) {
	var out bytes.Buffer
	if err := SendStatusResponse(&out, "1.19.4", 760, "MineMock", 20, 5); err != nil {
		t.Fatalf("SendStatusResponse failed: %v", err)
	}

	packet, err := ReadPacket(&out)
	if err != nil {
		t.Fatalf("ReadPacket failed: %v", err)
	}

	packetID, payload, err := ReadPacketID(packet)
	if err != nil {
		t.Fatalf("ReadPacketID failed: %v", err)
	}
	if packetID != 0x00 {
		t.Fatalf("unexpected packet id: %d", packetID)
	}

	jsonLen, n, err := decodeVarIntFromBytes(payload)
	if err != nil {
		t.Fatalf("decode status length failed: %v", err)
	}
	payload = payload[n:]
	if int(jsonLen) != len(payload) {
		t.Fatalf("payload length mismatch: declared %d, got %d", jsonLen, len(payload))
	}

	var status StatusResponse
	if err := json.Unmarshal(payload, &status); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if status.Description.Text != "MineMock" {
		t.Fatalf("unexpected motd: %s", status.Description.Text)
	}
	if status.Players.Max != 20 || status.Players.Online != 5 {
		t.Fatalf("unexpected player stats: %+v", status.Players)
	}
}
