package testpkg

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestGenerateSignature(t *testing.T) {

	secret := "supersecret1"
	payload := `{"ref": "refs/heads/main", "repository": {"html_url": "https://github.com/khaledibrahim1015/goflowdotnet.git"}}`

	signature := GenerateSignature(secret, payload)

	t.Log(signature)
	fmt.Println(signature)

}

func GenerateSignature(secret, payload string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(payload))
	return "sha1=" + hex.EncodeToString(mac.Sum(nil))
}
