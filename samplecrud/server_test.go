package main

// server_test.go — REAL end-to-end tests.
// TestMain builds the actual server binary, launches it as a separate process,
// and every test hits it over real HTTP. Run with:  go test -v

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const baseURL = "http://localhost:8090"

var client = &http.Client{Timeout: 5 * time.Second}

type product struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// TestMain builds and launches the real server, waits until it serves,
// runs every test against it over the network, then tears it down.
func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "samplecrud-e2e-*")
	if err != nil {
		fmt.Println("mkdir temp:", err)
		os.Exit(1)
	}

	bin := filepath.Join(tmp, "server.exe")

	// Build the server binary from THIS module (current directory).
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		fmt.Println("build server:", err)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	// Launch it against an isolated temp DB on a dedicated port.
	srv := exec.Command(bin)
	srv.Env = append(os.Environ(),
		"PORT=8090",
		"DB_DSN="+filepath.Join(tmp, "e2e.db"),
	)
	srv.Stdout, srv.Stderr = os.Stdout, os.Stderr
	if err := srv.Start(); err != nil {
		fmt.Println("start server:", err)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	if err := waitForServer(baseURL+"/products", 10*time.Second); err != nil {
		fmt.Println("server never became ready:", err)
		srv.Process.Kill()
		srv.Process.Wait()
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	// Run tests, then ALWAYS clean up (os.Exit skips defers, so do it manually).
	code := m.Run()

	srv.Process.Kill()
	srv.Process.Wait()
	os.RemoveAll(tmp)
	os.Exit(code)
}

func waitForServer(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if resp, err := client.Get(url); err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s", url)
}

// do makes a real HTTP request to the running server and returns status + body.
func do(t *testing.T, method, path, body string) (int, []byte) {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = bytes.NewReader([]byte(body))
	}
	req, err := http.NewRequest(method, baseURL+path, r)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

// TestProductLifecycle exercises the full CRUD flow against the live server.
func TestProductLifecycle(t *testing.T) {
	// CREATE
	code, body := do(t, http.MethodPost, "/products",
		`{"name":"Keyboard","price":49.99,"quantity":10}`)
	if code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d (%s)", code, body)
	}
	var created product
	if err := json.Unmarshal(body, &created); err != nil {
		t.Fatalf("create json: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected assigned id, got 0")
	}
	path := fmt.Sprintf("/products/%d", created.ID)

	// GET
	code, body = do(t, http.MethodGet, path, "")
	if code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d (%s)", code, body)
	}
	var got product
	json.Unmarshal(body, &got)
	if got.Name != "Keyboard" {
		t.Errorf("get name = %q, want Keyboard", got.Name)
	}

	// LIST should contain it
	code, body = do(t, http.MethodGet, "/products", "")
	if code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", code)
	}
	var list []product
	json.Unmarshal(body, &list)
	found := false
	for _, p := range list {
		if p.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("created product %d not found in list", created.ID)
	}

	// UPDATE
	code, body = do(t, http.MethodPut, path,
		`{"name":"Mechanical","price":89.99,"quantity":7}`)
	if code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d (%s)", code, body)
	}

	// GET confirms it persisted
	code, body = do(t, http.MethodGet, path, "")
	json.Unmarshal(body, &got)
	if got.Name != "Mechanical" || got.Price != 89.99 {
		t.Errorf("update not persisted: %+v", got)
	}

	// DELETE
	code, _ = do(t, http.MethodDelete, path, "")
	if code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", code)
	}

	// GET now 404
	code, _ = do(t, http.MethodGet, path, "")
	if code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", code)
	}
}

func TestValidationError(t *testing.T) {
	code, _ := do(t, http.MethodPost, "/products", `{"price":10,"quantity":1}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", code)
	}
}

func TestGetNotFound(t *testing.T) {
	code, _ := do(t, http.MethodGet, "/products/999999", "")
	if code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", code)
	}
}

func TestUpdateNotFound(t *testing.T) {
	code, _ := do(t, http.MethodPut, "/products/999999",
		`{"name":"X","price":1,"quantity":1}`)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", code)
	}
}

func TestDeleteNotFound(t *testing.T) {
	code, _ := do(t, http.MethodDelete, "/products/999999", "")
	if code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", code)
	}
}
