package cookier

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEncodeDecode(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 64 hex chars
	ec := Encode(key, Cookie{time.Now().Unix() + 3600, "{\"token\":\"123\"}"})
	dc, err := Decode(key, ec)
	if err != nil {
		t.Error("Decode error", err)
	}
	s := struct{ Token string }{}
	err = json.Unmarshal([]byte(dc.Value), &s)
	if s.Token != "123" {
		t.Error("encode decode failed", dc, err)
	}
}

func TestEncodeDecodeExpired(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 64 hex chars
	ec := Encode(key, Cookie{time.Now().Unix() - 100, "{\"token\":\"123\"}"})
	_, err := Decode(key, ec)
	if err == nil {
		t.Error("expired decoded")
	}
}

func TestDecodeGarbage(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 64 hex chars
	_, err := Decode(key, "junk")
	if err == nil {
		t.Error("junk did not error")
	}
}

func TestEncodeBadKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	key := "01f"
	Encode(key, Cookie{time.Now().Unix() + 3600, "stuff"})
}

func TestDecodeBadKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	key := "01f"
	Decode(key, "whatever")
}
