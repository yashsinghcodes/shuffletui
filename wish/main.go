package main

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yashsinghcodes/shuffle-tui/shuffletea"
	"github.com/yashsinghcodes/shuffle-tui/sqlc/queries"
)

const (
    host = "192.168.29.215"
    port = "42069"
    url = "https://charm.sh/"
)

//go:embed schema.sql
var ddl string;
var sshkey []byte;

func dbInt() error {
    db, err := sql.Open("sqlite3", "sessions.db")
    if err != nil {
        log.Error("Failed to open database", err)
        return err
    }

    if _, err := db.ExecContext(context.Background(), ddl); err != nil {
        log.Error("Error: ", err)
        return err
    }

    shuffletea.Queries = shuffletui.New(db)
    return nil
}

func main() {
    dbInt()
    s, err := wish.NewServer(wish.WithAddress(net.JoinHostPort(host, port)),
        wish.WithHostKeyPath("/Users/yashsingh/.ssh/id_ed25519"),
        wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
            sshkey = key.Marshal()
            _, err := shuffletea.Queries.InsertSession(ctx, shuffletui.InsertSessionParams{
                Sshkey: key.Marshal(),
                Username: ctx.User(),
            })
            if err != nil {
                log.Info("Failed to create session for user ", key.Marshal(), ctx.User(), "with error",err)
                shuffletea.DB.Close()
                return false
            }

            return true
        }),
        wish.WithMiddleware(
            bubbletea.Middleware(teaHandler),
            activeterm.Middleware(),
            logging.Middleware(),
            ),
        )
    if err != nil {
        log.Error("Could not start the sever", "error", err)
    }

    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    log.Info("Starting SSH server", "host", host, "port", port)
    go func() {
        if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
            log.Error("Could not start the server")
            done <- nil
        }
    }()

    <-done
    log.Info("Stopping SSH Sever")
    ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
    defer func() { cancel() }()
    if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
        log.Error("Could not stop sever", "error", err)
    }
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
        ctx := context.Background()
        var m tea.Model

        // FIXME: Can be a same query
        data, err := shuffletea.Queries.GetUser(context.Background(), sshkey)
        if err != nil {
            log.Error("Failed to get the user", err)
        }

        apikey, err := shuffletea.Queries.GetApiKey(ctx, sshkey)
        if err != nil {
            log.Info("Failed to get info")
        }

        if apikey.Valid != true {
            var input textinput.Model
            input = textinput.New()
            input.Placeholder = "<Your API Key>"
            input.Focus()

            m = shuffletea.ApiModel{
            Input: input,
            Username: data.Username,
            SSHkey: sshkey,
           }
            return m, []tea.ProgramOption{tea.WithAltScreen()}
        }

        return shuffletea.Model{SSHkey: sshkey, Username: data.Username, ApiKey: apikey.String}, []tea.ProgramOption{tea.WithAltScreen()}
}
