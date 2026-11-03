package engine

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSendEmailUsesConfiguredRelayPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	delivered := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		_, _ = io.WriteString(conn, "220 localhost SMTP\r\n")
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			switch {
			case strings.HasPrefix(line, "DATA"):
				_, _ = io.WriteString(conn, "354 Send message\r\n")
				var body strings.Builder
				for {
					part, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if part == ".\r\n" {
						break
					}
					body.WriteString(part)
				}
				delivered <- body.String()
				_, _ = io.WriteString(conn, "250 Queued\r\n")
			case strings.HasPrefix(line, "QUIT"):
				_, _ = io.WriteString(conn, "221 Bye\r\n")
				return
			default:
				_, _ = io.WriteString(conn, "250 OK\r\n")
			}
		}
	}()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	portNumber, err := strconv.Atoi(port)
	require.NoError(t, err)
	sender := NewSender(SenderConfig{
		Hostname: "mailit.local", HeloDomain: "mailit.local",
		RelayMode: "relay", RelayHost: "127.0.0.1", RelayPort: portNumber,
	}, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	result, err := sender.SendEmail(context.Background(), &OutgoingMessage{
		From: "sender@mailit.local", To: []string{"recipient@example.test"},
		Subject: "Local relay smoke test", TextBody: "Captured locally",
	})
	require.NoError(t, err)
	require.Equal(t, "sent", result.Recipients["recipient@example.test"].Status)
	select {
	case message := <-delivered:
		require.Contains(t, message, "Captured locally")
	case <-time.After(5 * time.Second):
		t.Fatal("relay did not receive the message")
	}
}

func TestUnavailableRelayDefersEmail(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	sender := NewSender(SenderConfig{
		Hostname: "mailit.local", HeloDomain: "mailit.local",
		RelayMode: "relay", RelayHost: "127.0.0.1", RelayPort: port,
		ConnectTimeout: time.Second,
	}, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	result, err := sender.SendEmail(context.Background(), &OutgoingMessage{
		From: "sender@mailit.local", To: []string{"recipient@example.test"}, TextBody: "Retry later",
	})
	require.NoError(t, err)
	require.Equal(t, "deferred", result.Recipients["recipient@example.test"].Status)
	require.False(t, result.Recipients["recipient@example.test"].Permanent)
}
