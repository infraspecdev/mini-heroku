package auth_test

import (
	"testing"
	"mini-heroku/controller/internal/auth"
)

func TestValidate_ValidKey(t *testing.T) {
	svc := auth.New("secret-key")
	if !svc.Validate("secret-key") {
		t.Error("expected valid key to pass")
	}
}

func TestValidate_InvalidKey(t *testing.T) {
	svc := auth.New("secret-key")
	if svc.Validate("wrong-key") {
		t.Error("expected wrong key to fail")
	}
}

func TestValidate_EmptyKey(t *testing.T) {
	svc := auth.New("secret-key")
	if svc.Validate("") {
		t.Error("expected empty key to fail")
	}
}

func TestValidate_EmptyServiceKey(t *testing.T) {
	svc := auth.New("")
	if svc.Validate("any-key") {
		t.Error("expected empty service key to always fail")
	}
}
