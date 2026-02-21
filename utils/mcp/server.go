package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"
)

// ToolHandler handles a tool invocation with raw JSON params and returns a structured result.
type ToolHandler func(ctx context.Context, params json.RawMessage) (any, error)

// Server implements a minimal in-process registry of MCP tools and HTTP-backed handlers
// that proxy requests to the flip-shop HTTP server. Transport (JSON-RPC/MCP framing)
// will be added in later tasks.
type Server struct {
	logger *log.Logger
	config Config
	http   *http.Client
	tools  map[string]ToolHandler
}

// NewServer creates a new MCP server instance with registered tools.
func NewServer(logger *log.Logger, cfg Config) *Server {
	s := &Server{logger: logger, config: cfg, http: &http.Client{Timeout: cfg.Timeout}, tools: map[string]ToolHandler{}}
	// Register tools for Phase 1 task 7
	s.tools["cart.create"] = s.handleCartCreate
	s.tools["cart.purchase.add"] = s.handleCartPurchaseAdd
	s.tools["cart.purchase.remove"] = s.handleCartPurchaseRemove
	s.tools["cart.submit"] = s.handleCartSubmit
	return s
}

// Start logs configuration and available tools. Transport server added later.
func (s *Server) Start(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if s.logger != nil {
		s.logger.Println("flipshop-mcp: server start (tool registry active)")
		s.logger.Println(fmt.Sprintf("flipshop-mcp: config BaseURL=%s Timeout=%s", s.config.BaseURL, s.config.Timeout))
		for name := range s.tools {
			s.logger.Println("flipshop-mcp: tool registered:", name)
		}
	}
	return nil
}

// Stop performs graceful shutdown logging (transport wiring to be added in later phases).
func (s *Server) Stop(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if s.logger != nil {
		s.logger.Println("flipshop-mcp: server stop")
	}
	return nil
}

// ToolNames returns a list of registered tool names.
func (s *Server) ToolNames() []string {
	n := make([]string, 0, len(s.tools))
	for k := range s.tools {
		n = append(n, k)
	}
	return n
}

// invoke is a helper to call a tool handler directly (useful for tests before transport exists).
func (s *Server) invoke(ctx context.Context, name string, params any) (any, error) {
	h, ok := s.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	var raw json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	start := time.Now()
	if s.logger != nil {
		s.logger.Println("flipshop-mcp: invoke start", name)
	}
	res, err := h(ctx, raw)
	dur := time.Since(start)
	if s.logger != nil {
		if err != nil {
			s.logger.Println("flipshop-mcp: invoke error", name, "duration=", dur)
		} else {
			s.logger.Println("flipshop-mcp: invoke ok", name, "duration=", dur)
		}
	}
	return res, err
}

// Common param structs

type cartIDParam struct {
	CartID string `json:"cartID"`
}

type purchaseParams struct {
	CartID string `json:"cartID"`
	SKU    string `json:"sku"`
	Qty    int    `json:"qty"`
}

// Response wrapper: MCP plan returns { cart: Cart }
type cartResponse struct {
	Cart any `json:"cart"`
}

// Handlers

func (s *Server) handleCartCreate(ctx context.Context, _ json.RawMessage) (any, error) {
 body, status, err := s.doJSON(ctx, http.MethodPost, "/cart", nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, mapHTTPToMCPError(status, body)
	}
	var cart any
	if err := json.Unmarshal(body, &cart); err != nil {
		return nil, fmt.Errorf("decode cart: %w", err)
	}
	return cartResponse{Cart: cart}, nil
}

func (s *Server) handleCartPurchaseAdd(ctx context.Context, raw json.RawMessage) (any, error) {
	var p purchaseParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, errors.New("invalid params: expected {cartID, sku, qty}")
	}
	if p.CartID == "" || p.SKU == "" || p.Qty <= 0 {
		return nil, errors.New("invalid params: cartID, sku must be non-empty and qty > 0")
	}
	reqBody := map[string]any{"sku": p.SKU, "qty": p.Qty}
 body, status, err := s.doJSON(ctx, http.MethodPut, path.Join("/cart", p.CartID, "purchase"), reqBody)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, mapHTTPToMCPError(status, body)
	}
	var cart any
	if err := json.Unmarshal(body, &cart); err != nil {
		return nil, fmt.Errorf("decode cart: %w", err)
	}
	return cartResponse{Cart: cart}, nil
}

func (s *Server) handleCartPurchaseRemove(ctx context.Context, raw json.RawMessage) (any, error) {
	var p purchaseParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, errors.New("invalid params: expected {cartID, sku, qty}")
	}
	if p.CartID == "" || p.SKU == "" || p.Qty <= 0 {
		return nil, errors.New("invalid params: cartID, sku must be non-empty and qty > 0")
	}
	reqBody := map[string]any{"sku": p.SKU, "qty": p.Qty}
 body, status, err := s.doJSON(ctx, http.MethodDelete, path.Join("/cart", p.CartID, "purchase"), reqBody)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, mapHTTPToMCPError(status, body)
	}
	var cart any
	if err := json.Unmarshal(body, &cart); err != nil {
		return nil, fmt.Errorf("decode cart: %w", err)
	}
	return cartResponse{Cart: cart}, nil
}

func (s *Server) handleCartSubmit(ctx context.Context, raw json.RawMessage) (any, error) {
	var p cartIDParam
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, errors.New("invalid params: expected {cartID}")
	}
	if p.CartID == "" {
		return nil, errors.New("invalid params: cartID must be non-empty")
	}
 body, status, err := s.doJSON(ctx, http.MethodPut, path.Join("/cart", p.CartID, "status", "submitted"), nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, mapHTTPToMCPError(status, body)
	}
	var cart any
	if err := json.Unmarshal(body, &cart); err != nil {
		return nil, fmt.Errorf("decode cart: %w", err)
	}
	return cartResponse{Cart: cart}, nil
}

// doJSON performs an HTTP request to flip-shop, sending/receiving JSON.
func (s *Server) doJSON(ctx context.Context, method, relativePath string, payload any) ([]byte, int, error) {
	u, err := url.Parse(s.config.BaseURL)
	if err != nil {
		return nil, 0, err
	}
	u.Path = path.Join(u.Path, relativePath)

	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		body = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, 0, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return respBody, resp.StatusCode, nil
}

// MCPError represents a structured error for MCP tools, mapped from HTTP responses.
type MCPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Body    string `json:"body,omitempty"`
}

func (e *MCPError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s (status=%d): %s", e.Code, e.Status, e.Message)
}

// mapHTTPToMCPError maps flip-shop HTTP status codes to MCP error categories.
func mapHTTPToMCPError(status int, body []byte) error {
	msg := string(body)
	switch status {
	case http.StatusNotFound:
		return &MCPError{Code: "NOT_FOUND", Message: msg, Status: status, Body: msg}
	case http.StatusUnprocessableEntity:
		return &MCPError{Code: "INVALID_ARGUMENT", Message: msg, Status: status, Body: msg}
	default:
		return &MCPError{Code: "INTERNAL", Message: msg, Status: status, Body: msg}
	}
}

// ToolDeclaration describes an MCP tool including JSON Schemas and examples.
type ToolDeclaration struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	InputSchema  map[string]any         `json:"inputSchema"`
	OutputSchema map[string]any         `json:"outputSchema"`
	Examples     []map[string]any       `json:"examples,omitempty"`
}

// Declarations returns the JSON Schema declarations for the registered tools.
// Note: Schemas are intentionally permissive in Phase 1; Phase 4 will tighten.
func (s *Server) Declarations() []ToolDeclaration {
	uuidPattern := "^[0-9a-fA-F-]{36}$"

	cartSchema := map[string]any{
		"type": "object",
	}

	emptyParams := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}

	cartIDParam := map[string]any{
		"type":     "object",
		"required": []string{"cartID"},
		"properties": map[string]any{
			"cartID": map[string]any{"type": "string", "pattern": uuidPattern},
		},
	}

	purchaseParam := map[string]any{
		"type":     "object",
		"required": []string{"cartID", "sku", "qty"},
		"properties": map[string]any{
			"cartID": map[string]any{"type": "string", "pattern": uuidPattern},
			"sku":    map[string]any{"type": "string", "minLength": 1},
			"qty":    map[string]any{"type": "integer", "minimum": 1},
		},
	}

	cartOutput := map[string]any{
		"type":     "object",
		"required": []string{"cart"},
		"properties": map[string]any{
			"cart": cartSchema,
		},
	}

	decls := []ToolDeclaration{
		{
			Name:        "cart.create",
			Description: "Create a new shopping cart",
			InputSchema: emptyParams,
			OutputSchema: cartOutput,
			Examples: []map[string]any{
				{"params": map[string]any{}},
			},
		},
		{
			Name:        "cart.purchase.add",
			Description: "Add a purchase (sku, qty) to the given cart",
			InputSchema: purchaseParam,
			OutputSchema: cartOutput,
			Examples: []map[string]any{
				{"params": map[string]any{"cartID": "123e4567-e89b-12d3-a456-426614174000", "sku": "120P90", "qty": 3}},
			},
		},
		{
			Name:        "cart.purchase.remove",
			Description: "Remove a purchase (sku, qty) from the given cart",
			InputSchema: purchaseParam,
			OutputSchema: cartOutput,
			Examples: []map[string]any{
				{"params": map[string]any{"cartID": "123e4567-e89b-12d3-a456-426614174000", "sku": "120P90", "qty": 1}},
			},
		},
		{
			Name:        "cart.submit",
			Description: "Submit the cart to apply promotions and finalize totals",
			InputSchema: cartIDParam,
			OutputSchema: cartOutput,
			Examples: []map[string]any{
				{"params": map[string]any{"cartID": "123e4567-e89b-12d3-a456-426614174000"}},
			},
		},
	}

	return decls
}
