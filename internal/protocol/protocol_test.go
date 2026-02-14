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
	if err := SendLoginDisconnect(&out, "", "test"); err != nil {
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

	if err := SendLoginDisconnect(&out, "", rawReason); err != nil {
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

func TestSendLoginDisconnect_WithTitle(t *testing.T) {
	var out bytes.Buffer

	if err := SendLoginDisconnect(&out, "Соединение потеряно", "Internal Exception: io.netty.channel.unix.Errors$NativeIoException"); err != nil {
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

	var reason map[string]any
	if err := json.Unmarshal(reasonPayload, &reason); err != nil {
		t.Fatalf("reason json unmarshal failed: %v", err)
	}
	if reason["translate"] != "disconnect.genericReason" {
		t.Fatalf("unexpected translate value: %v", reason["translate"])
	}
	with, ok := reason["with"].([]any)
	if !ok || len(with) != 2 {
		t.Fatalf("unexpected with payload: %#v", reason["with"])
	}
	if with[0] != "Соединение потеряно" {
		t.Fatalf("unexpected title value: %v", with[0])
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
