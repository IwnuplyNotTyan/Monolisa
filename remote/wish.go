//go:build ssh

package remote

import (
	"context"
	"errors"
	"fmt"
	"monolisa/render"
	"monolisa/utils"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	"charm.land/ssh"
	"charm.land/wish/v2"
	"charm.land/wish/v2/elapsed"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/log"
)

const (
	host = "localhost"
	port = "23234"
)

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
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithBannerHandler(func(ctx ssh.Context) string {
			return fmt.Sprintf(banner, ctx.User())
		}),
		wish.WithPasswordAuth(func(ctx ssh.Context, password string) bool {
			return password == "monolisa"
		}),
		wish.WithMiddleware(
			gifMiddleware,
			logging.Middleware(),
			elapsed.Middleware(),
		),
	)
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
