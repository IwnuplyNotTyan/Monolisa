
<div align="center">
  <h1>🌌 ~ Monolisa</h1>
  <p>Gif Screensaver</p>
</div>

<p align="center">
  <a href="https://github.com/IwnuplyNotTyan/monolisa/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/IwnuplyNotTyan/monolisa/ci.yml" alt="Build Status"/>
  </a>
  <img src="https://img.shields.io/github/license/IwnuplyNotTyan/monolisa" alt="License"/>
  <img src="https://img.shields.io/github/stars/IwnuplyNotTyan/monolisa" alt="Stars"/>
  <img src="https://img.shields.io/github/last-commit/IwnuplyNotTyan/monolisa" alt="Last Commit"/>
</p>


<p align="center">
  <img src="https://github.com/IwnuplyNotTyan/Monolisa/blob/main/.github/assets/screenshot.png?raw=true" alt="Screenshot">
</p>

---

```sh
monolisa # Launch random gif from current dir

monolisa ~/Pictures # Use custon folder, and can be selected file!
# Or
MONOLISA_DIR=$HOME/Pictures monolisa
```

---

# 💖 Terminal support

|Terminal|Addition|Status|
|--------|--------|------|
|Kitty| |✅|
|Neovim| |✅|
|Tmux| |✅|
|UXterm|Unicodes can be look ugly|✅|
|IDEA like|Unicodes can be look ugly|✅|

> [!TIP] 
> Want expand this table? Open issue/pull request

---

# 🪻 Install

### ❄️ Nix
``` bash
nix run github:iwnuplynottyan/monolisa
```

### 🐋 Docker

**Check [Docker Compose](https://github.com/IwnuplyNotTyan/monolisa/blob/main/docker-compose.yaml)!**

### ⛏️ Build from source

> [!NOTE]
> Use `-tags ssh` if needed

```sh
git clone https://github.com/IwnuplyNotTyan/monolisa && cd monolisa
go mod download
go build -o ./bin/monolisa ./cmd/monolisa/main.go
```



# ⚙️ Environment variables

|Variable|Default|Description|
|--------|-------|-----------|
|`MONOLISA_DIR`|current dir (`/app/gifs/` in Docker)|Folder with `*.gif` files|
|`MONOLISA_HOST`|`0.0.0.0`|Address the SSH server binds to|
|`MONOLISA_PORT`|`23234`|Port of the SSH server|
|`MONOLISA_PASSWORD`|`monolisa`|Login password. `off`, `0`, `false`, `no`, `none`, `disabled` or empty value disables password auth|
|`MONOLISA_AUTHORIZED_KEYS`|`<empty>`|Comma separated paths to `authorized_keys` files. If empty — `.ssh/authorized_keys` is used when it exists|
|`MONOLISA_SSH_KEYS`|`<empty>`|Inline public keys, separated by newline, `,` or `;`|

> [!WARNING]
> If password auth is disabled and no keys are provided, the server refuses to start — otherwise it would accept anyone without a password.

## 🔑 SSH login

|Method|How to enable|How to connect|
|------|-------------|--------------|
|Password|`MONOLISA_PASSWORD` is set (default `monolisa`)|`ssh -t -p 23234 q@localhost`|
|Key file|`MONOLISA_AUTHORIZED_KEYS` or existing `.ssh/authorized_keys`|`ssh -t -i ~/.ssh/id_ed25519 -p 23234 q@localhost`|
|Inline key|`MONOLISA_SSH_KEYS=ssh-ed25519 AAAA... me@host`|`ssh -t -i ~/.ssh/id_ed25519 -p 23234 q@localhost`|

> [!NOTE]
> `-t` is required — a PTY must be allocated, otherwise the session is closed with `Invalid PTY`.

```sh
# password only
docker compose up -d

# key only, without password
MONOLISA_PASSWORD=off MONOLISA_SSH_KEYS="$(cat ~/.ssh/id_ed25519.pub)" docker compose up -d
```

Mount a key file and point `MONOLISA_AUTHORIZED_KEYS` at it (see the commented volume in `docker-compose.yaml`):

```yaml
services:
  monolisa:
    volumes:
      - ./authorized_keys:/app/authorized_keys:ro
    environment:
      - MONOLISA_AUTHORIZED_KEYS=/app/authorized_keys
```

---

## 🛠️ Libraries Used

- [Wish](charm.land/wish/v2) - SSH
- [Log](github.com/charmbracelet/log) - Pretty logs
- [Term](golang.org/x/term) - Terminal info

### ⛏️ Assets

- [Example gif](https://imgur.com/gallery/beautiful-gifs-1080p-0Slze)

---

## 📄 License
[MIT](https://github.com/IwnuplyNotTyan/Monolisa/blob/main/LICENSE).

<div align="center">
  <h1>Made with ❤️ </h1>
</div>
