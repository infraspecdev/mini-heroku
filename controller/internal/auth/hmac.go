package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	// TimestampHeader is the Unix timestamp of the request (seconds).
	TimestampHeader = "X-Timestamp"
	// SignatureHeader is the HMAC-SHA256 hex signature.
	SignatureHeader = "X-Signature"

	maxClockSkew = 5 * time.Minute
)

func ValidateTimestamp(r *http.Request) error {
	tsStr := r.Header.Get(TimestampHeader)
	if tsStr == "" {
		return fmt.Errorf("missing %s header", TimestampHeader)
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s value: %w", TimestampHeader, err)
	}

	skew := time.Since(time.Unix(ts, 0))
	if skew < 0 {
		skew = -skew
	}
	if skew > maxClockSkew {
		return fmt.Errorf("timestamp rejected: clock skew %v exceeds ±%v", skew, maxClockSkew)
	}
	return nil
}

func ValidateHMAC(r *http.Request, secret string) error {
	tsStr := r.Header.Get(TimestampHeader)
	if tsStr == "" {
		return fmt.Errorf("missing %s header", TimestampHeader)
	}

	gotSig := r.Header.Get(SignatureHeader)
	if gotSig == "" {
		return fmt.Errorf("missing %s header", SignatureHeader)
	}

	msg := fmt.Sprintf("%s\n%s\n%s", r.Method, r.URL.Path, tsStr)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	wantSig := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison to prevent timing attacks.
	if !hmac.Equal([]byte(gotSig), []byte(wantSig)) {
		return fmt.Errorf("HMAC signature mismatch")
	}
	return nil
}
