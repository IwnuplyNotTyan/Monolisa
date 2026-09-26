
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

## ❄️ Nix
``` bash
nix run github:iwnuplynottyan/monolisa
```

## 🐋 Docker

**Check [Docker Compose](https://github.com/IwnuplyNotTyan/monolisa/blob/main/docker-compose.yaml)!**

## ⛏️ Build from source

> [!NOTE]
> Use `-tags ssh` if needed

```sh
git clone https://github.com/IwnuplyNotTyan/monolisa && cd monolisa
go mod download
go build -o ./bin/monolisa ./cmd/monolisa/main.go
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
