package sock

import "testing"

func TestFakeSNI_NotEmpty(t *testing.T) {
	if len(FakeSNI) == 0 {
		t.Error("FakeSNI should not be empty")
	}
}

func TestFakeSNI_StartsWithTLSRecord(t *testing.T) {
	// TLS record should start with 0x16 (handshake)
	if FakeSNI[0] != 0x16 {
		t.Errorf("FakeSNI should start with TLS handshake byte 0x16, got 0x%02x", FakeSNI[0])
	}
}
