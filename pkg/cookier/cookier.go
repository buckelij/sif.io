// example:
// key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 64 hex chars
// ec := cookier.Encode(key, cookier.Cookie{time.Now().Unix() + 3600, "{\token\":\"123\"}"})
// dc, _ := cookier.Decode(key, ec) // ignoring err and expiration err
// json.Unmarshal(dc.Value, &s) // ignoring err
package cookier

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"time"
)

// Cookie is a Value (could be marshalled json) and Expires (unix time)
type Cookie struct {
	Expires int64  `json:"expires"`
	Value   string `json:"value"`
}

// Encodes Cookie c with key k
func Encode(k string, c Cookie) string {
	v, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return encrypt(k, string(v))
}

// Decodes s with key k into Cookie. Returns an error if expired.
func Decode(k string, s string) (Cookie, error) {
	v, err := decrypt(k, []byte(s))
	if err != nil {
		return Cookie{}, err
	}
	c := Cookie{}
	err = json.Unmarshal(v, &c)
	if err != nil {
		return c, err
	}
	if c.Expires < time.Now().Unix() {
		return c, errors.New("expired")
	}
	return c, nil
}

// crypto/cipher#example-NewGCM-Encrypt
func encrypt(k string, s string) string {
	key, err := hex.DecodeString(k)
	if err != nil {
		panic(err)
	}
	cip, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aes, err := cipher.NewGCM(cip)
	if err != nil {
		panic(err)
	}
	nc := make([]byte, aes.NonceSize())
	_, err = io.ReadFull(rand.Reader, nc)
	if err != nil {
		panic(err)
	}
	return string(aes.Seal(nc, nc, []byte(s), nil)) // reuse nc appending ciphertext
}

func decrypt(k string, v []byte) ([]byte, error) {
	key, err := hex.DecodeString(k)
	if err != nil {
		panic(err)
	}
	cip, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aes, err := cipher.NewGCM(cip)
	if err != nil {
		panic(err)
	}
	if len(v) < aes.NonceSize() {
		return v, errors.New("invalid ciphertext")
	}
	return aes.Open(nil, v[:aes.NonceSize()], v[aes.NonceSize():], nil)
}
