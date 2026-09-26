//go:build ssh

package remote

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"charm.land/ssh"
)

func sshWireString(s string) []byte {
	var b bytes.Buffer
	_ = binary.Write(&b, binary.BigEndian, uint32(len(s)))
	b.WriteString(s)
	return b.Bytes()
}

func authorizedKey(t *testing.T, seed byte) string {
	t.Helper()
	priv := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, ed25519.SeedSize))
	pub := priv.Public().(ed25519.PublicKey)
	wire := append(sshWireString("ssh-ed25519"), sshWireString(string(pub))...)
	return "ssh-ed25519 " + base64.StdEncoding.EncodeToString(wire) + " monolisa-test"
}

func mustParseKey(t *testing.T, line string) ssh.PublicKey {
	t.Helper()
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
	if err != nil {
		t.Fatalf("parse key %q: %v", line, err)
	}
	return key
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, wasSet := os.LookupEnv(key)
	_ = os.Unsetenv(key)
	t.Cleanup(func() {
		if wasSet {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func applyOptions(t *testing.T, opts []ssh.Option) *ssh.Server {
	t.Helper()
	srv := &ssh.Server{}
	for i, o := range opts {
		if err := o(srv); err != nil {
			t.Fatalf("option %d failed: %v", i, err)
		}
	}
	return srv
}

func TestResolvePassword(t *testing.T) {
	tests := []struct {
		name        string
		set         bool
		value       string
		want        string
		wantEnabled bool
	}{
		{"unset uses default", false, "", "monolisa", true},
		{"custom", true, "secret", "secret", true},
		{"trimmed", true, "  pass  ", "pass", true},
		{"empty disables", true, "", "", false},
		{"zero disables", true, "0", "", false},
		{"off disables", true, "off", "", false},
		{"OFF disables", true, "OFF", "", false},
		{"padded off disables", true, " Off ", "", false},
		{"false disables", true, "false", "", false},
		{"no disables", true, "no", "", false},
		{"none disables", true, "none", "", false},
		{"disabled disables", true, "disabled", "", false},
		{"similar word keeps password", true, "officious", "officious", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.set {
				t.Setenv("MONOLISA_PASSWORD", tc.value)
			} else {
				unsetEnv(t, "MONOLISA_PASSWORD")
			}

			got, enabled := resolvePassword()
			if got != tc.want {
				t.Errorf("password = %q, want %q", got, tc.want)
			}
			if enabled != tc.wantEnabled {
				t.Errorf("enabled = %v, want %v", enabled, tc.wantEnabled)
			}
		})
	}
}

func TestSplitList(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"a,b;c", []string{"a", "b", "c"}},
		{"a\nb\r c", []string{"a", "b", " c"}},
		{"a,,b", []string{"a", "b"}},
	}
	for _, tc := range tests {
		got := splitList(tc.in)
		if len(got) != len(tc.want) {
			t.Errorf("splitList(%q) = %q, want %q", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("splitList(%q) = %q, want %q", tc.in, got, tc.want)
				break
			}
		}
	}
}

func TestKeyFiles(t *testing.T) {
	t.Run("explicit list", func(t *testing.T) {
		t.Setenv("MONOLISA_AUTHORIZED_KEYS", " one.txt , two.txt ")
		want := []string{"one.txt", "two.txt"}
		got := keyFiles()
		if len(got) != len(want) {
			t.Fatalf("keyFiles() = %q, want %q", got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("keyFiles() = %q, want %q", got, want)
			}
		}
	})

	t.Run("default file exists", func(t *testing.T) {
		t.Chdir(t.TempDir())
		t.Setenv("MONOLISA_AUTHORIZED_KEYS", "")
		if err := os.MkdirAll(".ssh", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(defaultAuthorizedKeys, []byte("# keys\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		got := keyFiles()
		if len(got) != 1 || got[0] != defaultAuthorizedKeys {
			t.Errorf("keyFiles() = %q, want [%q]", got, defaultAuthorizedKeys)
		}
	})

	t.Run("no default file", func(t *testing.T) {
		t.Chdir(t.TempDir())
		t.Setenv("MONOLISA_AUTHORIZED_KEYS", "  ")
		if got := keyFiles(); got != nil {
			t.Errorf("keyFiles() = %q, want nil", got)
		}
	})
}

func TestParseKeys(t *testing.T) {
	line1 := authorizedKey(t, 1)
	line2 := authorizedKey(t, 2)
	text := "# комментарий\n\n" + line1 + "\nне-ключ\n   " + line2 + "   \n"

	keys := parseKeys(text)
	if len(keys) != 2 {
		t.Fatalf("parseKeys returned %d keys, want 2", len(keys))
	}
	if !ssh.KeysEqual(keys[0], mustParseKey(t, line1)) {
		t.Error("first key does not match")
	}
	if !ssh.KeysEqual(keys[1], mustParseKey(t, line2)) {
		t.Error("second key does not match")
	}
}

func TestParseKeysNoValidKeys(t *testing.T) {
	for _, in := range []string{"", "\n\n", "# только комментарий\n", "mangled-key-data"} {
		if keys := parseKeys(in); len(keys) != 0 {
			t.Errorf("parseKeys(%q) = %d keys, want 0", in, len(keys))
		}
	}
}

func TestLoadKeys(t *testing.T) {
	t.Chdir(t.TempDir())

	dir := t.TempDir()
	k1, k2, k3 := authorizedKey(t, 1), authorizedKey(t, 2), authorizedKey(t, 3)
	file := filepath.Join(dir, "keys.txt")
	if err := os.WriteFile(file, []byte("# c\n"+k1+"\n"+k2+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MONOLISA_AUTHORIZED_KEYS", file+","+filepath.Join(dir, "missing.txt"))
	t.Setenv("MONOLISA_SSH_KEYS", k3)

	keys := loadKeys()
	if len(keys) != 3 {
		t.Fatalf("loadKeys returned %d keys, want 3", len(keys))
	}
	for i, want := range []string{k1, k2, k3} {
		if !ssh.KeysEqual(keys[i], mustParseKey(t, want)) {
			t.Errorf("key %d does not match", i)
		}
	}
}

func TestLoadKeysInlineList(t *testing.T) {
	t.Chdir(t.TempDir())

	k1, k2 := authorizedKey(t, 4), authorizedKey(t, 5)
	t.Setenv("MONOLISA_AUTHORIZED_KEYS", "")
	t.Setenv("MONOLISA_SSH_KEYS", k1+","+k2)

	keys := loadKeys()
	if len(keys) != 2 {
		t.Fatalf("loadKeys returned %d keys, want 2", len(keys))
	}
	for i, want := range []string{k1, k2} {
		if !ssh.KeysEqual(keys[i], mustParseKey(t, want)) {
			t.Errorf("key %d does not match", i)
		}
	}
}

func TestAuthOptionsPasswordOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("MONOLISA_AUTHORIZED_KEYS", "")
	t.Setenv("MONOLISA_SSH_KEYS", "")

	setPassword(t, "secret", true)

	opts, err := authOptions()
	if err != nil {
		t.Fatalf("authOptions: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("got %d options, want 1", len(opts))
	}

	srv := applyOptions(t, opts)
	if srv.PasswordHandler == nil {
		t.Fatal("password handler is not installed")
	}
	if !srv.PasswordHandler(nil, "secret") {
		t.Error("correct password rejected")
	}
	if srv.PasswordHandler(nil, "wrong") {
		t.Error("wrong password accepted")
	}
	if srv.PublicKeyHandler != nil {
		t.Error("public key handler must be absent")
	}
}

func TestAuthOptionsKeysOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("MONOLISA_AUTHORIZED_KEYS", "")

	key := authorizedKey(t, 7)
	t.Setenv("MONOLISA_SSH_KEYS", key)

	setPassword(t, "", false)

	opts, err := authOptions()
	if err != nil {
		t.Fatalf("authOptions: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("got %d options, want 1", len(opts))
	}

	srv := applyOptions(t, opts)
	if srv.PublicKeyHandler == nil {
		t.Fatal("public key handler is not installed")
	}
	if !srv.PublicKeyHandler(nil, mustParseKey(t, key)) {
		t.Error("authorized key rejected")
	}
	if srv.PublicKeyHandler(nil, mustParseKey(t, authorizedKey(t, 8))) {
		t.Error("unknown key accepted")
	}
	if srv.PasswordHandler != nil {
		t.Error("password handler must be absent")
	}
}

func TestAuthOptionsNoMethodFails(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("MONOLISA_AUTHORIZED_KEYS", "")
	t.Setenv("MONOLISA_SSH_KEYS", "")

	setPassword(t, "", false)

	opts, err := authOptions()
	if err == nil {
		t.Fatal("expected error when password is disabled and no keys given")
	}
	if opts != nil {
		t.Errorf("options = %v, want nil", opts)
	}
}

func TestAuthOptionsBothMethods(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("MONOLISA_AUTHORIZED_KEYS", "")
	t.Setenv("MONOLISA_SSH_KEYS", authorizedKey(t, 9))

	setPassword(t, "secret", true)

	opts, err := authOptions()
	if err != nil {
		t.Fatalf("authOptions: %v", err)
	}
	if len(opts) != 2 {
		t.Fatalf("got %d options, want 2", len(opts))
	}

	srv := applyOptions(t, opts)
	if srv.PasswordHandler == nil || srv.PublicKeyHandler == nil {
		t.Fatal("both auth handlers must be installed")
	}
}

func setPassword(t *testing.T, pw string, enabled bool) {
	t.Helper()
	oldPassword, oldEnabled := password, passEnabled
	t.Cleanup(func() { password, passEnabled = oldPassword, oldEnabled })
	password, passEnabled = pw, enabled
}
