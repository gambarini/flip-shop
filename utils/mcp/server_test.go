package mcp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testCart struct {
	ID string `json:"id"`
}

type recordedRequest struct {
	Method string
	Path   string
	Body   string
}

func newTestServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*Server, func()) {
	ts := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(ts.Close)
	cfg := Config{BaseURL: ts.URL, Timeout: 3 * time.Second}
	s := NewServer(log.New(testWriter{t}, "", 0), cfg)
	return s, ts.Close
}

type testWriter struct{ t *testing.T }

func (tw testWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestCartCreate_OK(t *testing.T) {
	reqs := make([]recordedRequest, 0, 1)
	srv, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		reqs = append(reqs, recordedRequest{Method: r.Method, Path: r.URL.Path})
		if r.Method != http.MethodPost || r.URL.Path != "/cart" {
			w.WriteHeader(500)
			return
		}
		_ = json.NewEncoder(w).Encode(testCart{ID: "abc-123"})
	})

	res, err := srv.invoke(context.Background(), "cart.create", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cr, ok := res.(cartResponse)
	if !ok {
		t.Fatalf("expected cartResponse, got %T", res)
	}
	b, _ := json.Marshal(cr.Cart)
	if !strings.Contains(string(b), "abc-123") {
		t.Fatalf("expected cart id in response, got %s", string(b))
	}
	if len(reqs) != 1 || reqs[0].Method != http.MethodPost || reqs[0].Path != "/cart" {
		t.Fatalf("unexpected request: %+v", reqs)
	}
}

func TestCartPurchaseAdd_And_Remove_OK(t *testing.T) {
	for _, tc := range []struct{
		name string
		tool string
		method string
		path string
	}{
		{"add", "cart.purchase.add", http.MethodPut, "/cart/xyz/purchase"},
		{"remove", "cart.purchase.remove", http.MethodDelete, "/cart/xyz/purchase"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reqs := make([]recordedRequest, 0, 1)
			srv, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				reqs = append(reqs, recordedRequest{Method: r.Method, Path: r.URL.Path, Body: readBody(t, r)})
				if r.Method != tc.method || r.URL.Path != tc.path {
					w.WriteHeader(500)
					return
				}
				_ = json.NewEncoder(w).Encode(testCart{ID: "xyz"})
			})

			params := map[string]any{"cartID": "xyz", "sku": "120P90", "qty": 2}
			res, err := srv.invoke(context.Background(), tc.tool, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			_ = res.(cartResponse)

			if !strings.Contains(reqs[0].Body, "\"sku\":\"120P90\"") || !strings.Contains(reqs[0].Body, "\"qty\":2") {
				t.Fatalf("expected body to contain sku and qty, got %s", reqs[0].Body)
			}
		})
	}
}

func TestCartSubmit_OK(t *testing.T) {
	reqs := make([]recordedRequest, 0, 1)
	srv, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		reqs = append(reqs, recordedRequest{Method: r.Method, Path: r.URL.Path})
		if r.Method != http.MethodPut || r.URL.Path != "/cart/xyz/status/submitted" {
			w.WriteHeader(500)
			return
		}
		_ = json.NewEncoder(w).Encode(testCart{ID: "xyz"})
	})

	res, err := srv.invoke(context.Background(), "cart.submit", map[string]any{"cartID": "xyz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = res.(cartResponse)
}

func TestErrorMapping_HTTPToMCP(t *testing.T) {
	cases := []struct{
		status int
		code string
	}{
		{status: 404, code: "NOT_FOUND"},
		{status: 422, code: "INVALID_ARGUMENT"},
		{status: 500, code: "INTERNAL"},
	}
	for _, c := range cases {
		srv, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(c.status)
			w.Write([]byte("oops"))
		})
		_, err := srv.invoke(context.Background(), "cart.create", map[string]any{})
		if err == nil {
			t.Fatalf("expected error for status %d", c.status)
		}
		me, ok := err.(*MCPError)
		if !ok {
			t.Fatalf("expected MCPError, got %T", err)
		}
		if me.Code != c.code || me.Status != c.status || !strings.Contains(me.Body, "oops") {
			t.Fatalf("unexpected MCPError: %+v", me)
		}
	}
}

func TestParamValidation(t *testing.T) {
	srv, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(testCart{ID: "x"})
	})
	// qty <= 0
	_, err := srv.invoke(context.Background(), "cart.purchase.add", map[string]any{"cartID": "x", "sku": "120P90", "qty": 0})
	if err == nil || !strings.Contains(err.Error(), "qty > 0") {
		t.Fatalf("expected validation error, got %v", err)
	}
	// missing cartID for submit
	_, err = srv.invoke(context.Background(), "cart.submit", map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "cartID") {
		t.Fatalf("expected cartID validation error, got %v", err)
	}
}

func readBody(t *testing.T, r *http.Request) string {
	b, _ := io.ReadAll(r.Body)
	return string(b)
}
