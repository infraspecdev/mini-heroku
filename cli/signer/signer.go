package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

const (
	HeaderTimestamp = "X-Timestamp"
	HeaderSignature = "X-Signature"
)

func Sign(secret, method, urlPath string, timestamp int64) string {
	msg := fmt.Sprintf("%s\n%s\n%d", method, urlPath, timestamp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

func Headers(secret, method, urlPath string) map[string]string {
	ts := time.Now().Unix()
	return map[string]string{
		HeaderTimestamp: strconv.FormatInt(ts, 10),
		HeaderSignature: Sign(secret, method, urlPath, ts),
	}
}
