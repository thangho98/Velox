package cloudstorage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/thawng/velox/internal/cloudstorage"
)

// stubDriver implements cloudstorage.Driver for testing the Registry.
type stubDriver struct {
	typeName    string
	displayName string
	flow        cloudstorage.AuthFlow
}

func (d *stubDriver) Type() string                    { return d.typeName }
func (d *stubDriver) DisplayName() string             { return d.displayName }
func (d *stubDriver) AuthFlow() cloudstorage.AuthFlow { return d.flow }
func (d *stubDriver) NewProvider(*cloudstorage.Credentials) (cloudstorage.Provider, error) {
	return nil, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := cloudstorage.NewRegistry()

	d := &stubDriver{typeName: "fshare", displayName: "Fshare", flow: cloudstorage.AuthFlowPassword}
	if err := r.Register(d); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, err := r.Get("fshare")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.DisplayName() != "Fshare" {
		t.Errorf("got %q; want Fshare", got.DisplayName())
	}
}

func TestRegistry_DuplicateRegister(t *testing.T) {
	r := cloudstorage.NewRegistry()
	d := &stubDriver{typeName: "fshare"}
	_ = r.Register(d)
	err := r.Register(d)
	if !errors.Is(err, cloudstorage.ErrDuplicateDriver) {
		t.Errorf("err = %v; want ErrDuplicateDriver", err)
	}
}

func TestRegistry_UnknownProvider(t *testing.T) {
	r := cloudstorage.NewRegistry()
	_, err := r.Get("missing")
	if !errors.Is(err, cloudstorage.ErrUnknownProvider) {
		t.Errorf("err = %v; want ErrUnknownProvider", err)
	}
}

func TestRegistry_ListSorted(t *testing.T) {
	r := cloudstorage.NewRegistry()
	_ = r.Register(&stubDriver{typeName: "onedrive", displayName: "OneDrive", flow: cloudstorage.AuthFlowOAuth})
	_ = r.Register(&stubDriver{typeName: "gdrive", displayName: "Google Drive", flow: cloudstorage.AuthFlowOAuth})
	_ = r.Register(&stubDriver{typeName: "fshare", displayName: "Fshare", flow: cloudstorage.AuthFlowPassword})

	list := r.List()
	if len(list) != 3 {
		t.Fatalf("len = %d; want 3", len(list))
	}
	// Sorted by display name: Fshare, Google Drive, OneDrive.
	want := []string{"Fshare", "Google Drive", "OneDrive"}
	for i, d := range list {
		if d.DisplayName != want[i] {
			t.Errorf("list[%d] = %q; want %q", i, d.DisplayName, want[i])
		}
	}
	if list[0].AuthFlow != "password" {
		t.Errorf("fshare auth flow = %q; want password", list[0].AuthFlow)
	}
	if list[1].AuthFlow != "oauth" {
		t.Errorf("gdrive auth flow = %q; want oauth", list[1].AuthFlow)
	}
}

func TestCredentials_EncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	original := &cloudstorage.Credentials{
		ProviderType: "fshare",
		Email:        "user@example.com",
		Password:     "secret-password",
		AccessToken:  "tok-abc",
		SessionID:    "sid-xyz",
		Meta:         map[string]string{"account_type": "Vip"},
	}

	ct, err := cloudstorage.EncryptCredentials(key, original)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	decrypted, err := cloudstorage.DecryptCredentials(key, ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if decrypted.Email != original.Email || decrypted.Password != original.Password {
		t.Errorf("roundtrip lost plain fields: got %+v", decrypted)
	}
	if decrypted.Meta["account_type"] != "Vip" {
		t.Errorf("Meta lost: %+v", decrypted.Meta)
	}
}

func TestCredentials_WrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1

	ct, err := cloudstorage.EncryptCredentials(key1, &cloudstorage.Credentials{Email: "a@b"})
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	_, err = cloudstorage.DecryptCredentials(key2, ct)
	if err == nil {
		t.Error("decrypt with wrong key should fail")
	}
}

// Assert our test import is used; keeps the compiler happy on empty test files.
var _ = context.Background
