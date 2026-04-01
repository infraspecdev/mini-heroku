package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// ---- helpers ---------------------------------------------------------------

func makeSignedRequest(t *testing.T, method, path, secret string, ts int64) *http.Request {
	t.Helper()
	msg := fmt.Sprintf("%s\n%s\n%d", method, path, ts)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))

	r := httptest.NewRequest(method, path, nil)
	r.Header.Set("X-Timestamp", strconv.FormatInt(ts, 10))
	r.Header.Set("X-Signature", sig)
	return r
}

// ---- ValidateTimestamp -----------------------------------------------------

func TestValidateTimestamp_Valid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.Header.Set("X-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))

	if err := ValidateTimestamp(r); err != nil {
		t.Errorf("ValidateTimestamp() unexpected error: %v", err)
	}
}

func TestValidateTimestamp_Missing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	if err := ValidateTimestamp(r); err == nil {
		t.Error("ValidateTimestamp() should fail when header is absent")
	}
}

func TestValidateTimestamp_TooOld(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	old := time.Now().Add(-10 * time.Minute).Unix()
	r.Header.Set("X-Timestamp", strconv.FormatInt(old, 10))

	if err := ValidateTimestamp(r); err == nil {
		t.Error("ValidateTimestamp() should fail for timestamp more than 5 minutes old")
	}
}

func TestValidateTimestamp_TooFarFuture(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	future := time.Now().Add(10 * time.Minute).Unix()
	r.Header.Set("X-Timestamp", strconv.FormatInt(future, 10))

	if err := ValidateTimestamp(r); err == nil {
		t.Error("ValidateTimestamp() should fail for timestamp more than 5 minutes in the future")
	}
}

func TestValidateTimestamp_NotAnInteger(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.Header.Set("X-Timestamp", "not-a-number")

	if err := ValidateTimestamp(r); err == nil {
		t.Error("ValidateTimestamp() should fail for non-integer timestamp")
	}
}

// ---- ValidateHMAC ----------------------------------------------------------

func TestValidateHMAC_Valid(t *testing.T) {
	const secret = "correct-secret"
	r := makeSignedRequest(t, http.MethodPost, "/upload", secret, time.Now().Unix())

	if err := ValidateHMAC(r, secret); err != nil {
		t.Errorf("ValidateHMAC() unexpected error: %v", err)
	}
}

func TestValidateHMAC_WrongSecret(t *testing.T) {
	r := makeSignedRequest(t, http.MethodPost, "/upload", "actual-secret", time.Now().Unix())

	if err := ValidateHMAC(r, "wrong-secret"); err == nil {
		t.Error("ValidateHMAC() should fail when secret does not match")
	}
}

func TestValidateHMAC_TamperedSignature(t *testing.T) {
	r := makeSignedRequest(t, http.MethodPost, "/upload", "secret", time.Now().Unix())
	r.Header.Set("X-Signature", "00000000deadbeef")

	if err := ValidateHMAC(r, "secret"); err == nil {
		t.Error("ValidateHMAC() should fail for tampered signature")
	}
}

func TestValidateHMAC_MissingSignatureHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.Header.Set("X-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))

	if err := ValidateHMAC(r, "secret"); err == nil {
		t.Error("ValidateHMAC() should fail when X-Signature header is absent")
	}
}

func TestValidateHMAC_MissingTimestampHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.Header.Set("X-Signature", "somesig")

	if err := ValidateHMAC(r, "secret"); err == nil {
		t.Error("ValidateHMAC() should fail when X-Timestamp header is absent")
	}
}
