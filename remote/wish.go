//go:build ssh

package remote

import (
	"context"
	"errors"
	"monolisa/render"
	"monolisa/utils"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "embed"

	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/elapsed"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/log"
)

var (
	host = envDefault("MONOLISA_HOST", "0.0.0.0")
	port = envDefault("MONOLISA_PORT", "23234")

	password, passEnabled = resolvePassword()
)

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

//go:embed banner.txt
var banner string

func gifMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		pty, winCh, isPty := sess.Pty()
		if !isPty {
			wish.Fatalln(sess, "Invalid PTY: Run `ssh -t ...`")
			return
		}

		ctx, cancel := context.WithCancel(sess.Context())
		defer cancel()

		resize := make(chan render.Size, 1)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case win, ok := <-winCh:
					if !ok {
						return
					}
					sz := render.Size{Cols: win.Width, Rows: win.Height}
					select {
					case <-resize:
					default:
					}
					resize <- sz
				}
			}
		}()

		go func() {
			buf := make([]byte, 1)
			for {
				if _, err := sess.Read(buf); err != nil {
					cancel()
					return
				}
				if buf[0] == 3 || buf[0] == 'q' {
					cancel()
					return
				}
			}
		}()

		size := render.Size{Cols: pty.Window.Width, Rows: pty.Window.Height}
		if err := render.Gif(ctx, sess, utils.GifFind(), size, resize); err != nil {
			wish.Println(sess, "Error: ", err)
		}
		next(sess)
	}
}

func Init() {
	opts := []ssh.Option{
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithBannerHandler(func(ctx ssh.Context) string {
			b := strings.ReplaceAll(banner, "%s", ctx.User())
			if !strings.HasSuffix(b, "\n") {
				b += "\n"
			}
			return b
		}),
	}
	auth, err := authOptions()
	if err != nil {
		log.Error("Could not start server", "error", err)
		return
	}
	opts = append(opts, auth...)
	opts = append(opts, wish.WithMiddleware(
		gifMiddleware,
		logging.Middleware(),
		elapsed.Middleware(),
	))

	s, err := wish.NewServer(opts...)
	if err != nil {
		log.Error("Could not start server", "error", err)
		return
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
