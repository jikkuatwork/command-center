package hosting

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// RunServerless executes JavaScript if main.js exists in the site directory
// Returns true if serverless was executed, false if should fall back to static
func RunServerless(w http.ResponseWriter, r *http.Request, siteDir string, db *sql.DB, siteID string) bool {
	mainJS := filepath.Join(siteDir, "main.js")

	// Check if main.js exists
	code, err := os.ReadFile(mainJS)
	if err != nil {
		return false // No main.js, serve static files
	}

	// Create JavaScript runtime
	vm := goja.New()

	// Create response object
	response := &jsResponse{
		w:           w,
		headers:     make(map[string]string),
		statusCode:  200,
		bodyWritten: false,
	}

	// Create request object (limit body to 1MB)
	limitedReader := io.LimitReader(r.Body, 1<<20) // 1MB
	bodyBytes, _ := io.ReadAll(limitedReader)
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	reqObj := map[string]interface{}{
		"method":  r.Method,
		"path":    r.URL.Path,
		"query":   r.URL.RawQuery,
		"headers": headers,
		"body":    string(bodyBytes),
	}

	// Inject objects into VM
	vm.Set("req", reqObj)
	vm.Set("res", map[string]interface{}{
		"send": func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				response.Send(call.Arguments[0].String())
			}
			return goja.Undefined()
		},
		"json": func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				response.JSON(call.Arguments[0].Export())
			}
			return goja.Undefined()
		},
		"status": func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				response.statusCode = int(call.Arguments[0].ToInteger())
			}
			return goja.Undefined()
		},
		"header": func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) >= 2 {
				response.headers[call.Arguments[0].String()] = call.Arguments[1].String()
			}
			return goja.Undefined()
		},
	})

	// Inject console for debugging
	vm.Set("console", map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value {
			args := make([]string, len(call.Arguments))
			for i, arg := range call.Arguments {
				args[i] = arg.String()
			}
			fmt.Printf("[JS:%s] %s\n", siteID, strings.Join(args, " "))
			return goja.Undefined()
		},
	})

	// Load environment variables for this site
	envVars := make(map[string]interface{})
	if db != nil {
		rows, err := db.Query("SELECT name, value FROM env_vars WHERE site_id = ?", siteID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var name, value string
				if rows.Scan(&name, &value) == nil {
					envVars[name] = value
				}
			}
		}
	}

	// Inject process.env
	vm.Set("process", map[string]interface{}{
		"env": envVars,
	})

	// Inject socket object for WebSocket broadcast
	hub := GetHub(siteID)
	vm.Set("socket", map[string]interface{}{
		"broadcast": func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				hub.Broadcast(call.Arguments[0].String())
			}
			return goja.Undefined()
		},
		"clients": func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(hub.ClientCount())
		},
	})

	// Inject db object for KV store
	vm.Set("db", map[string]interface{}{
		"get": func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 || db == nil {
				return goja.Null()
			}
			key := call.Arguments[0].String()
			var value string
			err := db.QueryRow("SELECT value FROM kv_store WHERE site_id = ? AND key = ?", siteID, key).Scan(&value)
			if err != nil {
				return goja.Null()
			}
			// Try to parse as JSON
			var parsed interface{}
			if json.Unmarshal([]byte(value), &parsed) == nil {
				return vm.ToValue(parsed)
			}
			return vm.ToValue(value)
		},
		"set": func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) < 2 || db == nil {
				return goja.Undefined()
			}
			key := call.Arguments[0].String()
			val := call.Arguments[1].Export()
			var valueStr string
			switch v := val.(type) {
			case string:
				valueStr = v
			default:
				jsonBytes, _ := json.Marshal(v)
				valueStr = string(jsonBytes)
			}
			db.Exec(`
				INSERT INTO kv_store (site_id, key, value, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
				ON CONFLICT(site_id, key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP
			`, siteID, key, valueStr, valueStr)
			return goja.Undefined()
		},
		"delete": func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 || db == nil {
				return goja.Undefined()
			}
			key := call.Arguments[0].String()
			db.Exec("DELETE FROM kv_store WHERE site_id = ? AND key = ?", siteID, key)
			return goja.Undefined()
		},
	})

	// Inject fetch function for HTTP requests
	vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(map[string]interface{}{
				"error": "URL required",
			})
		}

		fetchURL := call.Arguments[0].String()

		// Parse and validate URL
		parsedURL, err := url.Parse(fetchURL)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"error": "Invalid URL: " + err.Error(),
			})
		}

		// SSRF protection: block localhost and internal IPs
		host := parsedURL.Hostname()
		if isInternalHost(host) {
			return vm.ToValue(map[string]interface{}{
				"error": "Blocked: internal/localhost URLs not allowed",
			})
		}

		// Get options
		method := "GET"
		var reqBody io.Reader
		headers := make(map[string]string)

		if len(call.Arguments) > 1 {
			opts := call.Arguments[1].Export()
			if optsMap, ok := opts.(map[string]interface{}); ok {
				if m, ok := optsMap["method"].(string); ok {
					method = strings.ToUpper(m)
				}
				if h, ok := optsMap["headers"].(map[string]interface{}); ok {
					for k, v := range h {
						headers[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
					}
				}
				if body, ok := optsMap["body"].(string); ok {
					reqBody = strings.NewReader(body)
				}
			}
		}

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		req, err := http.NewRequest(method, fetchURL, reqBody)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"error": "Request error: " + err.Error(),
			})
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"error": "Fetch error: " + err.Error(),
			})
		}
		defer resp.Body.Close()

		// Limit response body to 1MB
		limitedResp := io.LimitReader(resp.Body, 1<<20)
		bodyBytes, err := io.ReadAll(limitedResp)
		if err != nil {
			return vm.ToValue(map[string]interface{}{
				"error": "Read error: " + err.Error(),
			})
		}

		// Build response headers
		respHeaders := make(map[string]string)
		for k, v := range resp.Header {
			if len(v) > 0 {
				respHeaders[k] = v[0]
			}
		}

		return vm.ToValue(map[string]interface{}{
			"status":  resp.StatusCode,
			"headers": respHeaders,
			"body":    string(bodyBytes),
		})
	})

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		_, err := vm.RunString(string(code))
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			response.Error(fmt.Sprintf("JavaScript error: %v", err))
		}
	case <-time.After(100 * time.Millisecond):
		vm.Interrupt("script timeout")
		response.Error("Script execution timed out (100ms limit)")
	}

	// Write response if not already written
	if !response.bodyWritten {
		response.Send("")
	}

	return true
}

// jsResponse handles the HTTP response from JavaScript
type jsResponse struct {
	w           http.ResponseWriter
	headers     map[string]string
	statusCode  int
	bodyWritten bool
}

func (r *jsResponse) writeHeaders() {
	if r.bodyWritten {
		return
	}
	for k, v := range r.headers {
		r.w.Header().Set(k, v)
	}
	r.w.WriteHeader(r.statusCode)
	r.bodyWritten = true
}

func (r *jsResponse) Send(body string) {
	if r.w.Header().Get("Content-Type") == "" {
		r.w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	r.writeHeaders()
	r.w.Write([]byte(body))
}

func (r *jsResponse) JSON(data interface{}) {
	r.w.Header().Set("Content-Type", "application/json")
	r.writeHeaders()
	json.NewEncoder(r.w).Encode(data)
}

func (r *jsResponse) Error(msg string) {
	r.statusCode = 500
	r.w.Header().Set("Content-Type", "text/plain")
	r.writeHeaders()
	r.w.Write([]byte(msg))
}

// HasServerless checks if a site has a main.js file
func HasServerless(siteDir string) bool {
	mainJS := filepath.Join(siteDir, "main.js")
	_, err := os.Stat(mainJS)
	return err == nil
}

// isInternalHost checks if a host is localhost or an internal IP (SSRF protection)
func isInternalHost(host string) bool {
	// Check for localhost variations
	host = strings.ToLower(host)
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "0.0.0.0" {
		return true
	}

	// Parse IP and check for internal ranges
	ip := net.ParseIP(host)
	if ip == nil {
		// Not an IP, try to resolve it
		ips, err := net.LookupIP(host)
		if err == nil && len(ips) > 0 {
			ip = ips[0]
		}
	}

	if ip != nil {
		// Check for private/internal IP ranges
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
	}

	return false
}
