// cli/keychain/keychain_test.go
package keychain_test

import (
	kclib "mini-heroku/cli/keychain"
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestMain swaps in the in-memory mock so tests run without a real OS keychain.
func TestMain(m *testing.M) {
	keyring.MockInit()
	os.Exit(m.Run())
}

func TestSet_StoresKey(t *testing.T) {
	const want = "super-secret-key"

	if err := kclib.Set(want); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := kclib.Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != want {
		t.Errorf("Get() = %q, want %q", got, want)
	}

	t.Cleanup(func() { _ = kclib.Delete() })
}

func TestSet_EmptyKeyReturnsError(t *testing.T) {
	if err := kclib.Set(""); err == nil {
		t.Error("Set(\"\") should return an error")
	}
}

func TestGet_NotFound(t *testing.T) {
	_ = kclib.Delete() // ensure clean state

	_, err := kclib.Get()
	if err == nil {
		t.Error("Get() with no stored key should return an error")
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	_ = kclib.Set("temp-key")
	_ = kclib.Delete()

	_, err := kclib.Get()
	if err == nil {
		t.Error("Get() after Delete() should return an error")
	}
}
