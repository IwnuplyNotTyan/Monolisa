//go:build ssh

package remote

import (
	"errors"
	"os"
	"strings"

	"charm.land/ssh"
	"charm.land/wish/v2"
	"github.com/charmbracelet/log"
)

// defaultAuthorizedKeys подхватывается, если MONOLISA_AUTHORIZED_KEYS не задан.
const defaultAuthorizedKeys = ".ssh/authorized_keys"

// disabledValues — значения MONOLISA_PASSWORD, отключающие парольную авторизацию.
var disabledValues = map[string]bool{
	"": true, "0": true, "off": true, "no": true,
	"false": true, "none": true, "disabled": true,
}

// resolvePassword возвращает пароль и признак его включения.
// Если MONOLISA_PASSWORD не задан — используется пароль по умолчанию.
func resolvePassword() (string, bool) {
	v, ok := os.LookupEnv("MONOLISA_PASSWORD")
	if !ok {
		return "monolisa", true
	}
	v = strings.TrimSpace(v)
	if disabledValues[strings.ToLower(v)] {
		return "", false
	}
	return v, true
}

// authOptions собирает опции авторизации: пароль (если не выключен)
// и публичные ключи из файлов/переменной окружения.
// Если не включён ни один метод — возвращает ошибку: без обработчиков
// сервер charm.land/ssh включает NoClientAuth и пускает всех без пароля.
func authOptions() ([]ssh.Option, error) {
	keys := loadKeys()
	opts := make([]ssh.Option, 0, 2)

	if passEnabled {
		opts = append(opts, wish.WithPasswordAuth(func(_ ssh.Context, pass string) bool {
			return pass == password
		}))
	}
	if len(keys) > 0 {
		opts = append(opts, wish.WithPublicKeyAuth(func(_ ssh.Context, key ssh.PublicKey) bool {
			for _, k := range keys {
				if ssh.KeysEqual(key, k) {
					return true
				}
			}
			return false
		}))
	}

	log.Info("Authentication", "password", passEnabled, "keys", len(keys))
	if len(opts) == 0 {
		return nil, errors.New("no auth method enabled: password is disabled and no SSH keys provided")
	}
	return opts, nil
}

// loadKeys читает ключи из файлов MONOLISA_AUTHORIZED_KEYS (список через запятую),
// из .ssh/authorized_keys (если MONOLISA_AUTHORIZED_KEYS не задан и файл есть)
// и из строк MONOLISA_SSH_KEYS (через перевод строки или запятую).
func loadKeys() []ssh.PublicKey {
	var keys []ssh.PublicKey
	for _, path := range keyFiles() {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Warn("Could not read authorized keys", "path", path, "error", err)
			continue
		}
		keys = append(keys, parseKeys(string(data))...)
	}
	if v := os.Getenv("MONOLISA_SSH_KEYS"); v != "" {
		for _, line := range splitList(v) {
			keys = append(keys, parseKeys(line)...)
		}
	}
	return keys
}

// splitList разбивает список через запятую, точку с запятой или перевод строки.
func splitList(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ',' || r == ';'
	})
}

// keyFiles возвращает пути к файлам с ключами.
func keyFiles() []string {
	v := os.Getenv("MONOLISA_AUTHORIZED_KEYS")
	if strings.TrimSpace(v) == "" {
		if _, err := os.Stat(defaultAuthorizedKeys); err != nil {
			return nil
		}
		return []string{defaultAuthorizedKeys}
	}

	var paths []string
	for _, p := range splitList(v) {
		if p = strings.TrimSpace(p); p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// parseKeys разбирает строки формата authorized_keys, пропуская пустые и комментарии.
func parseKeys(text string) []ssh.PublicKey {
	var keys []ssh.PublicKey
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
		if err != nil {
			log.Warn("Skipping invalid SSH key", "error", err)
			continue
		}
		keys = append(keys, key)
	}
	return keys
}
