package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestLivenessHandlerReportsOK(t *testing.T) {
	handler := NewLivenessHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body livenessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("status body = %q, want %q", body.Status, "ok")
	}
}

func TestReadinessHandlerReportsHealthyDependencies(t *testing.T) {
	handler := NewReadinessHandler(
		time.Second,
		CheckFunc(func(context.Context) error { return nil }),
		CheckFunc(func(context.Context) error { return nil }),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body readinessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("body status = %q, want %q", body.Status, "ok")
	}

	if body.Database.Status != "ok" {
		t.Fatalf("database status = %q, want %q", body.Database.Status, "ok")
	}

	if body.Redis.Status != "ok" {
		t.Fatalf("redis status = %q, want %q", body.Redis.Status, "ok")
	}
}

func TestReadinessHandlerReportsDatabaseFailure(t *testing.T) {
	handler := NewReadinessHandler(
		time.Second,
		CheckFunc(func(context.Context) error { return errors.New("db down") }),
		CheckFunc(func(context.Context) error { return nil }),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body readinessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Database.Status != "error" {
		t.Fatalf("database status = %q, want %q", body.Database.Status, "error")
	}

	if body.Database.Error != "db down" {
		t.Fatalf("database error = %q, want %q", body.Database.Error, "db down")
	}

	if body.Redis.Status != "ok" {
		t.Fatalf("redis status = %q, want %q", body.Redis.Status, "ok")
	}
}

func TestReadinessHandlerReportsRedisFailure(t *testing.T) {
	handler := NewReadinessHandler(
		time.Second,
		CheckFunc(func(context.Context) error { return nil }),
		CheckFunc(func(context.Context) error { return errors.New("redis down") }),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body readinessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Redis.Status != "error" {
		t.Fatalf("redis status = %q, want %q", body.Redis.Status, "error")
	}

	if body.Redis.Error != "redis down" {
		t.Fatalf("redis error = %q, want %q", body.Redis.Error, "redis down")
	}

	if body.Database.Status != "ok" {
		t.Fatalf("database status = %q, want %q", body.Database.Status, "ok")
	}
}

func TestRedisCheckerPingsWithoutAuthWhenURLHasNoCredentials(t *testing.T) {
	addr, wait := startFakeRedisServer(t, func(conn net.Conn) error {
		command, err := readRedisCommand(conn)
		if err != nil {
			return err
		}
		if len(command) != 1 || command[0] != "PING" {
			return fmt.Errorf("first command = %q, want %q", command, []string{"PING"})
		}

		if err := writeRedisSimpleString(conn, "PONG"); err != nil {
			return err
		}

		return nil
	})

	checker, err := NewRedisChecker("redis://" + addr)
	if err != nil {
		t.Fatalf("NewRedisChecker() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := checker.Check(ctx); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	wait()
}

func TestRedisCheckerAuthenticatesWithPasswordBeforePing(t *testing.T) {
	addr, wait := startFakeRedisServer(t, func(conn net.Conn) error {
		command, err := readRedisCommand(conn)
		if err != nil {
			return err
		}
		if len(command) != 2 || command[0] != "AUTH" || command[1] != "secret" {
			return fmt.Errorf("first command = %q, want %q", command, []string{"AUTH", "secret"})
		}

		if err := writeRedisSimpleString(conn, "OK"); err != nil {
			return err
		}

		command, err = readRedisCommand(conn)
		if err != nil {
			return err
		}
		if len(command) != 1 || command[0] != "PING" {
			return fmt.Errorf("second command = %q, want %q", command, []string{"PING"})
		}

		if err := writeRedisSimpleString(conn, "PONG"); err != nil {
			return err
		}

		return nil
	})

	checker, err := NewRedisChecker("redis://:secret@" + addr)
	if err != nil {
		t.Fatalf("NewRedisChecker() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := checker.Check(ctx); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	wait()
}

func TestRedisCheckerAuthenticatesWithUsernameAndPasswordBeforePing(t *testing.T) {
	addr, wait := startFakeRedisServer(t, func(conn net.Conn) error {
		command, err := readRedisCommand(conn)
		if err != nil {
			return err
		}
		if len(command) != 3 || command[0] != "AUTH" || command[1] != "tenant-user" || command[2] != "secret" {
			return fmt.Errorf("first command = %q, want %q", command, []string{"AUTH", "tenant-user", "secret"})
		}

		if err := writeRedisSimpleString(conn, "OK"); err != nil {
			return err
		}

		command, err = readRedisCommand(conn)
		if err != nil {
			return err
		}
		if len(command) != 1 || command[0] != "PING" {
			return fmt.Errorf("second command = %q, want %q", command, []string{"PING"})
		}

		if err := writeRedisSimpleString(conn, "PONG"); err != nil {
			return err
		}

		return nil
	})

	checker, err := NewRedisChecker("redis://tenant-user:secret@" + addr)
	if err != nil {
		t.Fatalf("NewRedisChecker() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := checker.Check(ctx); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	wait()
}

func TestRedisCheckerReturnsAuthFailure(t *testing.T) {
	addr, wait := startFakeRedisServer(t, func(conn net.Conn) error {
		command, err := readRedisCommand(conn)
		if err != nil {
			return err
		}
		if len(command) != 2 || command[0] != "AUTH" || command[1] != "secret" {
			return fmt.Errorf("first command = %q, want %q", command, []string{"AUTH", "secret"})
		}

		if _, err := conn.Write([]byte("-NOAUTH invalid password\r\n")); err != nil {
			return err
		}

		return nil
	})

	checker, err := NewRedisChecker("redis://:secret@" + addr)
	if err != nil {
		t.Fatalf("NewRedisChecker() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = checker.Check(ctx)
	if err == nil {
		t.Fatal("Check() error = nil, want error")
	}
	if err.Error() != `auth redis: unexpected response "-NOAUTH invalid password"` {
		t.Fatalf("Check() error = %q, want %q", err.Error(), `auth redis: unexpected response "-NOAUTH invalid password"`)
	}

	wait()
}

func startFakeRedisServer(t *testing.T, handler func(net.Conn) error) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	errCh := make(chan error, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			return
		}
		defer conn.Close()

		errCh <- handler(conn)
	}()

	return listener.Addr().String(), func() {
		t.Helper()
		defer listener.Close()

		if err := <-errCh; err != nil {
			t.Fatalf("fake redis server: %v", err)
		}
	}
}

func readRedisCommand(conn net.Conn) ([]string, error) {
	reader := bufio.NewReader(conn)

	header, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if len(header) < 2 || header[0] != '*' {
		return nil, fmt.Errorf("unexpected array header %q", header)
	}

	count, err := strconv.Atoi(header[1 : len(header)-2])
	if err != nil {
		return nil, fmt.Errorf("parse array length %q: %w", header, err)
	}

	command := make([]string, 0, count)
	for i := 0; i < count; i++ {
		bulkHeader, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if len(bulkHeader) < 2 || bulkHeader[0] != '$' {
			return nil, fmt.Errorf("unexpected bulk header %q", bulkHeader)
		}

		size, err := strconv.Atoi(bulkHeader[1 : len(bulkHeader)-2])
		if err != nil {
			return nil, fmt.Errorf("parse bulk length %q: %w", bulkHeader, err)
		}

		payload := make([]byte, size+2)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return nil, err
		}

		command = append(command, string(payload[:size]))
	}

	return command, nil
}

func writeRedisSimpleString(conn net.Conn, value string) error {
	if _, err := conn.Write([]byte("+" + value + "\r\n")); err != nil {
		return err
	}

	return nil
}
