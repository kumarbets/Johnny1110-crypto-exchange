package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestServer gives each test a fresh, isolated in-memory database
// wired into the real router.
func setupTestServer(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return NewRouter(db)
}

// doRequest fires an HTTP request at the router in-memory and returns the response.
func doRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateProduct(t *testing.T) {
	router := setupTestServer(t)

	rec := doRequest(router, http.MethodPost, "/products",
		`{"name":"Keyboard","price":49.99,"quantity":10}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var p Product
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if p.ID == 0 {
		t.Errorf("expected an assigned id, got 0")
	}
	if p.Name != "Keyboard" {
		t.Errorf("expected name Keyboard, got %q", p.Name)
	}
}

func TestCreateProduct_ValidationError(t *testing.T) {
	router := setupTestServer(t)

	// Missing required "name" -> should be rejected.
	rec := doRequest(router, http.MethodPost, "/products",
		`{"price":10,"quantity":1}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetProduct(t *testing.T) {
	router := setupTestServer(t)

	create := doRequest(router, http.MethodPost, "/products",
		`{"name":"Mouse","price":19.99,"quantity":5}`)
	var created Product
	json.Unmarshal(create.Body.Bytes(), &created)

	rec := doRequest(router, http.MethodGet, "/products/1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var got Product
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got.ID != created.ID || got.Name != "Mouse" {
		t.Errorf("got %+v, want id=%d name=Mouse", got, created.ID)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	router := setupTestServer(t)

	rec := doRequest(router, http.MethodGet, "/products/999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestListProducts(t *testing.T) {
	router := setupTestServer(t)

	doRequest(router, http.MethodPost, "/products", `{"name":"A","price":1,"quantity":1}`)
	doRequest(router, http.MethodPost, "/products", `{"name":"B","price":2,"quantity":2}`)

	rec := doRequest(router, http.MethodGet, "/products", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var list []Product
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("bad JSON: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 products, got %d", len(list))
	}
}

func TestUpdateProduct(t *testing.T) {
	router := setupTestServer(t)

	doRequest(router, http.MethodPost, "/products", `{"name":"Old","price":1,"quantity":1}`)

	rec := doRequest(router, http.MethodPut, "/products/1",
		`{"name":"New","price":99.5,"quantity":3}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var updated Product
	json.Unmarshal(rec.Body.Bytes(), &updated)
	if updated.Name != "New" || updated.Price != 99.5 {
		t.Errorf("update not applied: %+v", updated)
	}

	// Confirm it persisted.
	get := doRequest(router, http.MethodGet, "/products/1", "")
	var got Product
	json.Unmarshal(get.Body.Bytes(), &got)
	if got.Name != "New" {
		t.Errorf("expected persisted name New, got %q", got.Name)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	router := setupTestServer(t)

	rec := doRequest(router, http.MethodPut, "/products/999",
		`{"name":"X","price":1,"quantity":1}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteProduct(t *testing.T) {
	router := setupTestServer(t)

	doRequest(router, http.MethodPost, "/products", `{"name":"Temp","price":1,"quantity":1}`)

	rec := doRequest(router, http.MethodDelete, "/products/1", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// It should now be gone.
	get := doRequest(router, http.MethodGet, "/products/1", "")
	if get.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", get.Code)
	}
}

func TestDeleteProduct_NotFound(t *testing.T) {
	router := setupTestServer(t)

	rec := doRequest(router, http.MethodDelete, "/products/999", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
