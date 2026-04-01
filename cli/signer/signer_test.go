package signer_test

import (
	"mini-heroku/cli/signer"
	"strconv"
	"testing"
	"time"
)

func TestSign_Deterministic(t *testing.T) {
	sig1 := signer.Sign("secret", "POST", "/upload", 1700000000)
	sig2 := signer.Sign("secret", "POST", "/upload", 1700000000)
	if sig1 != sig2 {
		t.Error("Sign() must be deterministic for identical inputs")
	}
}

func TestSign_DiffersOnMethod(t *testing.T) {
	ts := int64(1700000000)
	if signer.Sign("s", "POST", "/upload", ts) == signer.Sign("s", "GET", "/upload", ts) {
		t.Error("Sign() must differ when method changes")
	}
}

func TestSign_DiffersOnPath(t *testing.T) {
	ts := int64(1700000000)
	if signer.Sign("s", "POST", "/upload", ts) == signer.Sign("s", "POST", "/health", ts) {
		t.Error("Sign() must differ when path changes")
	}
}

func TestSign_DiffersOnTimestamp(t *testing.T) {
	if signer.Sign("s", "POST", "/upload", 1000) == signer.Sign("s", "POST", "/upload", 1001) {
		t.Error("Sign() must differ when timestamp changes")
	}
}

func TestSign_DiffersOnSecret(t *testing.T) {
	ts := int64(1700000000)
	if signer.Sign("secret-a", "POST", "/upload", ts) == signer.Sign("secret-b", "POST", "/upload", ts) {
		t.Error("Sign() must differ when secret changes")
	}
}

func TestHeaders_ContainsRequiredFields(t *testing.T) {
	h := signer.Headers("secret", "POST", "/upload")
	if _, ok := h[signer.HeaderTimestamp]; !ok {
		t.Errorf("Headers() is missing %s", signer.HeaderTimestamp)
	}
	if _, ok := h[signer.HeaderSignature]; !ok {
		t.Errorf("Headers() is missing %s", signer.HeaderSignature)
	}
}

func TestHeaders_TimestampIsRecent(t *testing.T) {
	h := signer.Headers("secret", "POST", "/upload")
	ts, err := strconv.ParseInt(h[signer.HeaderTimestamp], 10, 64)
	if err != nil {
		t.Fatalf("timestamp header is not a valid int64: %v", err)
	}
	skew := time.Now().Unix() - ts
	if skew < 0 {
		skew = -skew
	}
	if skew > 5 {
		t.Errorf("timestamp is %d seconds off from now — too large", skew)
	}
}

func TestHeaders_SignatureMatchesManualSign(t *testing.T) {
	const secret = "my-key"
	h := signer.Headers(secret, "DELETE", "/apps/foo")

	ts, _ := strconv.ParseInt(h[signer.HeaderTimestamp], 10, 64)
	expected := signer.Sign(secret, "DELETE", "/apps/foo", ts)

	if h[signer.HeaderSignature] != expected {
		t.Errorf("Headers() signature %q != Sign() %q", h[signer.HeaderSignature], expected)
	}
}
