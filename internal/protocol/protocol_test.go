package protocol

import (
	"bytes"
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
}
