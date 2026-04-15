package httpapi

import (
	"bufio"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultRedisPort = "6379"

type Checker interface {
	Check(context.Context) error
}

type CheckFunc func(context.Context) error

func (fn CheckFunc) Check(ctx context.Context) error {
	return fn(ctx)
}

type sqlChecker struct {
	db *sql.DB
}

type redisChecker struct {
	network     string
	address     string
	tls         bool
	username    string
	password    string
	hasPassword bool
}

type componentStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type livenessResponse struct {
	Status string `json:"status"`
}

type readinessResponse struct {
	Status   string          `json:"status"`
	Database componentStatus `json:"database"`
	Redis    componentStatus `json:"redis"`
}

func NewLivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, livenessResponse{Status: "ok"})
	})
}

func NewReadinessHandler(timeout time.Duration, database Checker, redis Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		response := readinessResponse{
			Status: "ok",
			Database: componentStatus{
				Status: "ok",
			},
			Redis: componentStatus{
				Status: "ok",
			},
		}

		statusCode := http.StatusOK

		if err := database.Check(ctx); err != nil {
			response.Status = "error"
			response.Database.Status = "error"
			response.Database.Error = err.Error()
			statusCode = http.StatusServiceUnavailable
		}

		if err := redis.Check(ctx); err != nil {
			response.Status = "error"
			response.Redis.Status = "error"
			response.Redis.Error = err.Error()
			statusCode = http.StatusServiceUnavailable
		}

		writeJSON(w, statusCode, response)
	})
}

func NewSQLChecker(db *sql.DB) Checker {
	return sqlChecker{db: db}
}

func NewRedisChecker(rawURL string) (Checker, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	if parsed.Host == "" {
		return nil, fmt.Errorf("parse redis url: missing host")
	}

	host := parsed.Hostname()
	if host == "" {
		return nil, fmt.Errorf("parse redis url: missing host")
	}

	port := parsed.Port()
	if port == "" {
		port = defaultRedisPort
	}

	checker := redisChecker{
		network: "tcp",
		address: net.JoinHostPort(host, port),
		tls:     parsed.Scheme == "rediss",
	}
	if parsed.User != nil {
		checker.username = parsed.User.Username()
		checker.password, checker.hasPassword = parsed.User.Password()
	}

	return checker, nil
}

func (c sqlChecker) Check(ctx context.Context) error {
	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

func (c redisChecker) Check(ctx context.Context) error {
	var (
		conn net.Conn
		err  error
	)

	if c.tls {
		dialer := tls.Dialer{
			NetDialer: &net.Dialer{},
		}
		conn, err = dialer.DialContext(ctx, c.network, c.address)
	} else {
		dialer := net.Dialer{}
		conn, err = dialer.DialContext(ctx, c.network, c.address)
	}
	if err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("set redis deadline: %w", err)
		}
	}

	if err := c.writeAUTH(conn); err != nil {
		return err
	}

	if err := writeRedisCommand(conn, "PING"); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	line, err := readRedisLine(conn)
	if err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	if line != "+PONG" {
		return fmt.Errorf("ping redis: unexpected response %q", line)
	}

	return nil
}

func (c redisChecker) writeAUTH(conn net.Conn) error {
	if !c.hasPassword {
		return nil
	}

	args := []string{"AUTH"}
	if c.username != "" {
		args = append(args, c.username)
	}
	args = append(args, c.password)

	if err := writeRedisCommand(conn, args...); err != nil {
		return fmt.Errorf("auth redis: %w", err)
	}

	line, err := readRedisLine(conn)
	if err != nil {
		return fmt.Errorf("auth redis: %w", err)
	}

	if line != "+OK" {
		return fmt.Errorf("auth redis: unexpected response %q", line)
	}

	return nil
}

func writeRedisCommand(conn net.Conn, args ...string) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&builder, "$%d\r\n%s\r\n", len(arg), arg)
	}

	if _, err := conn.Write([]byte(builder.String())); err != nil {
		return err
	}

	return nil
}

func readRedisLine(conn net.Conn) (string, error) {
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func writeJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}
