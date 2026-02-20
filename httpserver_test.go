package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const testAPIKey = "test-secret-key"

// newTestHTTPServer creates an HTTPServer backed by in-memory SQLite for testing.
func newTestHTTPServer(t *testing.T) *HTTPServer {
	t.Helper()
	tokenStore, err := NewTokenStore(":memory:")
	if err != nil {
		t.Fatalf("NewTokenStore(:memory:) failed: %v", err)
	}
	t.Cleanup(func() { tokenStore.Close() })

	config := Config{
		APIKey:    testAPIKey,
		BaseURL:   "http://localhost:8080",
		MaxTokens: 100000,
	}
	return newHTTPServer(config, tokenStore)
}

func doMCPRequest(t *testing.T, handler http.Handler, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(body))
	req.Header.Set("X-API-Key", testAPIKey)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr.Result()
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return string(data)
}

// --- TempStore tests ---

func TestTempStore_StoreAndLoadAndDelete(t *testing.T) {
	ts := NewTempStore(10)

	entry := TempEntry{Content: "hello", Format: "markdown", URL: "http://x.com", CreatedAt: time.Now()}
	if err := ts.Store("abc", entry); err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	got, ok := ts.LoadAndDelete("abc")
	if !ok {
		t.Fatal("LoadAndDelete() returned false")
	}
	if got.Content != "hello" {
		t.Errorf("Content = %q, want %q", got.Content, "hello")
	}

	// Second load should fail
	_, ok = ts.LoadAndDelete("abc")
	if ok {
		t.Error("second LoadAndDelete() should return false")
	}
}

func TestTempStore_Cleanup_RemovesOld(t *testing.T) {
	ts := NewTempStore(10)

	// Store entry with old timestamp
	entry := TempEntry{Content: "old", CreatedAt: time.Now().Add(-2 * tempStoreTTL)}
	ts.Store("old-id", entry)

	ts.Cleanup()

	_, ok := ts.LoadAndDelete("old-id")
	if ok {
		t.Error("Cleanup() should have removed old entry")
	}
}

func TestTempStore_Cleanup_PreservesRecent(t *testing.T) {
	ts := NewTempStore(10)

	entry := TempEntry{Content: "recent", CreatedAt: time.Now()}
	ts.Store("recent-id", entry)

	ts.Cleanup()

	got, ok := ts.LoadAndDelete("recent-id")
	if !ok {
		t.Fatal("Cleanup() should preserve recent entries")
	}
	if got.Content != "recent" {
		t.Errorf("Content = %q, want %q", got.Content, "recent")
	}
}

func TestTempStore_Store_AtCapacity(t *testing.T) {
	ts := NewTempStore(2)

	ts.Store("a", TempEntry{Content: "1", CreatedAt: time.Now()})
	ts.Store("b", TempEntry{Content: "2", CreatedAt: time.Now()})

	err := ts.Store("c", TempEntry{Content: "3", CreatedAt: time.Now()})
	if err == nil {
		t.Error("Store() should return error at capacity")
	}
}

func TestTempStore_Concurrent(t *testing.T) {
	ts := NewTempStore(1000)
	var wg sync.WaitGroup

	// Concurrent stores
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts.Store(fmt.Sprintf("id-%d", i), TempEntry{Content: "data", CreatedAt: time.Now()})
		}()
	}

	// Concurrent loads
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts.LoadAndDelete(fmt.Sprintf("id-%d", i))
		}()
	}

	// Concurrent cleanup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts.Cleanup()
		}()
	}

	wg.Wait()
}

// --- UUID tests ---

func TestGenerateUUID_Format(t *testing.T) {
	uuid := generateUUID()
	// 8-4-4-4-12 = 36 characters
	if len(uuid) != 36 {
		t.Errorf("UUID length = %d, want 36", len(uuid))
	}
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		t.Errorf("UUID format invalid: %s", uuid)
	}
}

func TestGenerateUUID_Unique(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for range 1000 {
		uuid := generateUUID()
		if seen[uuid] {
			t.Fatalf("duplicate UUID: %s", uuid)
		}
		seen[uuid] = true
	}
}

// --- Auth tests ---

func TestHTTP_Auth_Missing(t *testing.T) {
	srv := newTestHTTPServer(t)
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestHTTP_Auth_Wrong(t *testing.T) {
	srv := newTestHTTPServer(t)
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	req.Header.Set("X-API-Key", "wrong-key")
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestHTTP_Auth_Correct(t *testing.T) {
	srv := newTestHTTPServer(t)
	resp := doMCPRequest(t, srv.mux, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestHTTP_Health_NoAuth(t *testing.T) {
	srv := newTestHTTPServer(t)
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "ok") {
		t.Errorf("health body = %q, want to contain 'ok'", body)
	}
}

// --- Endpoint tests ---

func TestHTTP_Initialize(t *testing.T) {
	srv := newTestHTTPServer(t)
	resp := doMCPRequest(t, srv.mux, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	body := readBody(t, resp)

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal([]byte(body), &rpcResp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if rpcResp.Error != nil {
		t.Errorf("unexpected error: %v", rpcResp.Error)
	}
}

func TestHTTP_ToolsList(t *testing.T) {
	srv := newTestHTTPServer(t)
	resp := doMCPRequest(t, srv.mux, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	body := readBody(t, resp)

	// Verify schema has file_token, result_id, override
	if !strings.Contains(body, "file_token") {
		t.Error("tools/list should include file_token property")
	}
	if !strings.Contains(body, "result_id") {
		t.Error("tools/list should include result_id property")
	}
	if !strings.Contains(body, "override") {
		t.Error("tools/list should include override property")
	}
	// Should NOT have raw "file" parameter (but "file_token" contains "file")
	// Check that no property key is exactly "file"
	var rpcResp struct {
		Result struct {
			Tools []struct {
				InputSchema struct {
					Properties map[string]any `json:"properties"`
				} `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}
	json.Unmarshal([]byte(body), &rpcResp)
	if len(rpcResp.Result.Tools) > 0 {
		props := rpcResp.Result.Tools[0].InputSchema.Properties
		if _, hasFile := props["file"]; hasFile {
			t.Error("HTTP tools/list should NOT have 'file' parameter")
		}
	}
}

func TestHTTP_ToolsCall_URL_UnderLimit(t *testing.T) {
	// Serve small HTML
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><p>Small content</p></body></html>")
	}))
	defer ts.Close()

	srv := newTestHTTPServer(t)
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"%s"}}}`, ts.URL)
	resp := doMCPRequest(t, srv.mux, reqBody)
	body := readBody(t, resp)

	if !strings.Contains(body, "Small content") {
		t.Errorf("response should contain 'Small content', got: %s", body)
	}
}

func TestHTTP_ToolsCall_URL_OverLimit(t *testing.T) {
	// Serve large HTML
	largeContent := "<html><body>"
	for i := 0; i < 100; i++ {
		largeContent += "<p>This is a very long paragraph with content to exceed the token limit.</p>"
	}
	largeContent += "</body></html>"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, largeContent)
	}))
	defer ts.Close()

	srv := newTestHTTPServer(t)
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"%s","max_tokens":50}}}`, ts.URL)
	resp := doMCPRequest(t, srv.mux, reqBody)
	body := readBody(t, resp)

	if !strings.Contains(body, "result_id") {
		t.Errorf("over-limit response should contain result_id, got: %s", body)
	}
	if !strings.Contains(body, "override") {
		t.Errorf("over-limit response should mention override, got: %s", body)
	}
}

func TestHTTP_ToolsCall_Override(t *testing.T) {
	// Serve large HTML
	largeContent := "<html><body>"
	for i := 0; i < 100; i++ {
		largeContent += "<p>Override test paragraph content here.</p>"
	}
	largeContent += "</body></html>"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, largeContent)
	}))
	defer ts.Close()

	srv := newTestHTTPServer(t)

	// First call: trigger over-limit
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"url":"%s","max_tokens":50}}}`, ts.URL)
	resp := doMCPRequest(t, srv.mux, reqBody)
	body := readBody(t, resp)

	// Extract result_id from the over-limit response
	var rpcResp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	json.Unmarshal([]byte(body), &rpcResp)

	var overLimit OverLimitResult
	if len(rpcResp.Result.Content) == 0 {
		t.Fatal("no content in over-limit response")
	}
	json.Unmarshal([]byte(rpcResp.Result.Content[0].Text), &overLimit)

	if overLimit.ResultID == "" {
		t.Fatal("result_id is empty in over-limit response")
	}

	// Second call: override to get full content
	overrideBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"result_id":"%s","override":true}}}`, overLimit.ResultID)
	resp2 := doMCPRequest(t, srv.mux, overrideBody)
	body2 := readBody(t, resp2)

	if !strings.Contains(body2, "Override test paragraph") {
		t.Errorf("override response should contain full content, got: %s", body2[:min(200, len(body2))])
	}
}

func TestHTTP_ToolsCall_Override_Expired(t *testing.T) {
	srv := newTestHTTPServer(t)

	// Try to override a nonexistent/expired result
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"result_id":"nonexistent-id","override":true}}}`
	resp := doMCPRequest(t, srv.mux, reqBody)
	body := readBody(t, resp)

	if !strings.Contains(body, "not found or expired") {
		t.Errorf("should get 'not found or expired' error, got: %s", body)
	}
}

func TestHTTP_ResultDownload(t *testing.T) {
	srv := newTestHTTPServer(t)

	// Manually store a temp entry
	id := generateUUID()
	srv.tempStore.Store(id, TempEntry{
		Content:   "# Download Content\n\nHello world",
		Format:    "markdown",
		URL:       "http://example.com",
		CreatedAt: time.Now(),
	})

	req := httptest.NewRequest("GET", "/results/"+id, nil)
	req.Header.Set("X-API-Key", testAPIKey)
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Download Content") {
		t.Error("response should contain the stored content")
	}
	if !strings.Contains(rr.Header().Get("Content-Disposition"), "attachment") {
		t.Error("should have Content-Disposition: attachment")
	}

	// Entry should be deleted after download
	_, found := srv.tempStore.LoadAndDelete(id)
	if found {
		t.Error("entry should be deleted after download")
	}
}

func TestHTTP_ResultDownload_NotFound(t *testing.T) {
	srv := newTestHTTPServer(t)

	req := httptest.NewRequest("GET", "/results/nonexistent", nil)
	req.Header.Set("X-API-Key", testAPIKey)
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestHTTP_ToolsCall_FileToken(t *testing.T) {
	srv := newTestHTTPServer(t)

	// Create a test file
	tmpfile, err := os.CreateTemp("", "httptest*.html")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.WriteString("<html><body><p>Token file content</p></body></html>")
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Create token via admin endpoint
	adminBody := fmt.Sprintf(`{"file":"%s","expires_minutes":60}`, tmpfile.Name())
	adminReq := httptest.NewRequest("POST", "/admin/tokens", strings.NewReader(adminBody))
	adminReq.Header.Set("X-API-Key", testAPIKey)
	adminReq.Header.Set("Content-Type", "application/json")
	adminRR := httptest.NewRecorder()
	srv.mux.ServeHTTP(adminRR, adminReq)

	if adminRR.Code != http.StatusCreated {
		t.Fatalf("admin token status = %d, want 201, body: %s", adminRR.Code, adminRR.Body.String())
	}

	var tokenResp AdminTokenResponse
	json.Unmarshal(adminRR.Body.Bytes(), &tokenResp)

	if tokenResp.Token == "" {
		t.Fatal("admin response missing token")
	}

	// Use file_token in tools/call
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"file_token":"%s"}}}`, tokenResp.Token)
	resp := doMCPRequest(t, srv.mux, reqBody)
	body := readBody(t, resp)

	if !strings.Contains(body, "Token file content") {
		t.Errorf("file_token response should contain file content, got: %s", body)
	}
}

func TestHTTP_ToolsCall_FileToken_Consumed(t *testing.T) {
	srv := newTestHTTPServer(t)

	tmpfile, err := os.CreateTemp("", "consumed*.html")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.WriteString("<html><body><p>Once only</p></body></html>")
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Create and use token
	token, err := srv.tokenStore.CreateFileToken(tmpfile.Name(), 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	// First use succeeds
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"file_token":"%s"}}}`, token)
	resp := doMCPRequest(t, srv.mux, reqBody)
	body := readBody(t, resp)
	if !strings.Contains(body, "Once only") {
		t.Error("first use should succeed")
	}

	// Second use fails
	resp2 := doMCPRequest(t, srv.mux, reqBody)
	body2 := readBody(t, resp2)
	if !strings.Contains(body2, "token") {
		t.Errorf("second use should fail with token error, got: %s", body2)
	}
}

func TestHTTP_ToolsCall_FileToken_Expired(t *testing.T) {
	srv := newTestHTTPServer(t)

	tmpfile, err := os.CreateTemp("", "expired*.html")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.WriteString("<html><body>expired</body></html>")
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Create token with negative TTL (already expired)
	token, err := srv.tokenStore.CreateFileToken(tmpfile.Name(), -1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"file_token":"%s"}}}`, token)
	resp := doMCPRequest(t, srv.mux, reqBody)
	body := readBody(t, resp)
	if !strings.Contains(body, "token") {
		t.Errorf("expired token should fail, got: %s", body)
	}
}

func TestHTTP_ToolsCall_RawFile_Rejected(t *testing.T) {
	srv := newTestHTTPServer(t)

	reqBody := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"webfetch_clean","arguments":{"file":"/etc/passwd"}}}`
	resp := doMCPRequest(t, srv.mux, reqBody)
	body := readBody(t, resp)

	if !strings.Contains(body, "not available in HTTP mode") {
		t.Errorf("raw file should be rejected, got: %s", body)
	}
}

// --- Admin tests ---

func TestHTTP_AdminTokens_Create(t *testing.T) {
	srv := newTestHTTPServer(t)

	tmpfile, err := os.CreateTemp("", "admin*.html")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.WriteString("<html>test</html>")
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	body := fmt.Sprintf(`{"file":"%s"}`, tmpfile.Name())
	req := httptest.NewRequest("POST", "/admin/tokens", strings.NewReader(body))
	req.Header.Set("X-API-Key", testAPIKey)
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201, body: %s", rr.Code, rr.Body.String())
	}

	var resp AdminTokenResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Token == "" {
		t.Error("response should include token")
	}
}

func TestHTTP_AdminTokens_FileNotFound(t *testing.T) {
	srv := newTestHTTPServer(t)

	body := `{"file":"/nonexistent/path.html"}`
	req := httptest.NewRequest("POST", "/admin/tokens", strings.NewReader(body))
	req.Header.Set("X-API-Key", testAPIKey)
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestHTTP_AdminTokens_CustomExpiry(t *testing.T) {
	srv := newTestHTTPServer(t)

	tmpfile, err := os.CreateTemp("", "expiry*.html")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.WriteString("<html>test</html>")
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	body := fmt.Sprintf(`{"file":"%s","expires_minutes":120}`, tmpfile.Name())
	req := httptest.NewRequest("POST", "/admin/tokens", strings.NewReader(body))
	req.Header.Set("X-API-Key", testAPIKey)
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rr.Code)
	}

	var resp AdminTokenResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)

	// Expiry should be approximately 2 hours from now
	expectedExpiry := time.Now().Add(120 * time.Minute)
	if resp.Expires.Before(expectedExpiry.Add(-1*time.Minute)) || resp.Expires.After(expectedExpiry.Add(1*time.Minute)) {
		t.Errorf("expiry %v not within expected range around %v", resp.Expires, expectedExpiry)
	}
}

// --- Request limit test ---

func TestHTTP_RequestBodyLimit(t *testing.T) {
	srv := newTestHTTPServer(t)

	// Create body larger than MaxScannerBuffer (10MB)
	largeBody := strings.Repeat("x", MaxScannerBuffer+100)
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(largeBody))
	req.Header.Set("X-API-Key", testAPIKey)
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)

	// Should get an error (parse error due to body limit)
	body := rr.Body.String()
	if !strings.Contains(body, "error") && rr.Code == http.StatusOK {
		t.Errorf("oversized request should fail, got status %d body: %s", rr.Code, body[:min(200, len(body))])
	}
}
