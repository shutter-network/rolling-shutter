package kproapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestIsReadOnlyEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		operation *openapi3.Operation
		want      bool
	}{
		{
			name: "read-only with json.RawMessage true",
			operation: func() *openapi3.Operation {
				op := &openapi3.Operation{}
				op.Extensions = map[string]interface{}{
					"x-read-only": json.RawMessage("true"),
				}
				return op
			}(),
			want: true,
		},
		{
			name: "read-only with json.RawMessage false",
			operation: func() *openapi3.Operation {
				op := &openapi3.Operation{}
				op.Extensions = map[string]interface{}{
					"x-read-only": json.RawMessage("false"),
				}
				return op
			}(),
			want: false,
		},
		{
			name: "read-only with direct boolean true",
			operation: func() *openapi3.Operation {
				op := &openapi3.Operation{}
				op.Extensions = map[string]interface{}{
					"x-read-only": true,
				}
				return op
			}(),
			want: true,
		},
		{
			name: "no read-only extension",
			operation: func() *openapi3.Operation {
				op := &openapi3.Operation{}
				op.Extensions = map[string]interface{}{}
				return op
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReadOnlyEndpoint(tt.operation)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldEnableEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		operation      *openapi3.Operation
		enableWriteOps bool
		want           bool
	}{
		{
			name: "read-only endpoint with write ops disabled",
			operation: func() *openapi3.Operation {
				op := &openapi3.Operation{}
				op.Extensions = map[string]interface{}{
					"x-read-only": json.RawMessage("true"),
				}
				return op
			}(),
			enableWriteOps: false,
			want:           true,
		},
		{
			name: "write endpoint with write ops enabled",
			operation: func() *openapi3.Operation {
				op := &openapi3.Operation{}
				op.Extensions = map[string]interface{}{}
				return op
			}(),
			enableWriteOps: true,
			want:           true,
		},
		{
			name: "write endpoint with write ops disabled",
			operation: func() *openapi3.Operation {
				op := &openapi3.Operation{}
				op.Extensions = map[string]interface{}{}
				return op
			}(),
			enableWriteOps: false,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldEnableEndpoint(tt.operation, tt.enableWriteOps)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindOperation(t *testing.T) {
	// Create a test spec
	spec := &openapi3.T{
		Paths: openapi3.Paths{
			"/test": &openapi3.PathItem{
				Get:    &openapi3.Operation{},
				Post:   &openapi3.Operation{},
				Put:    &openapi3.Operation{},
				Delete: &openapi3.Operation{},
			},
		},
	}

	tests := []struct {
		name   string
		path   string
		method string
		want   *openapi3.Operation
	}{
		{
			name:   "GET operation exists",
			path:   "/test",
			method: http.MethodGet,
			want:   spec.Paths.Find("/test").Get,
		},
		{
			name:   "POST operation exists",
			path:   "/test",
			method: http.MethodPost,
			want:   spec.Paths.Find("/test").Post,
		},
		{
			name:   "PUT operation exists",
			path:   "/test",
			method: http.MethodPut,
			want:   spec.Paths.Find("/test").Put,
		},
		{
			name:   "DELETE operation exists",
			path:   "/test",
			method: http.MethodDelete,
			want:   spec.Paths.Find("/test").Delete,
		},
		{
			name:   "non-existent path",
			path:   "/nonexistent",
			method: http.MethodGet,
			want:   nil,
		},
		{
			name:   "unsupported method",
			path:   "/test",
			method: "PATCH",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findOperation(spec, tt.path, tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfigMiddleware(t *testing.T) {
	// Create a test spec with both read-only and write operations
	spec := &openapi3.T{
		Paths: openapi3.Paths{
			"/read": &openapi3.PathItem{
				Get: func() *openapi3.Operation {
					op := &openapi3.Operation{}
					op.Extensions = map[string]interface{}{
						"x-read-only": json.RawMessage("true"),
					}
					return op
				}(),
			},
			"/write": &openapi3.PathItem{
				Post: &openapi3.Operation{},
			},
		},
	}

	// Create a function that returns our test spec
	getTestSpec := func() (*openapi3.T, error) {
		return spec, nil
	}

	tests := []struct {
		name           string
		enableWriteOps bool
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "read-only endpoint with write ops disabled",
			enableWriteOps: false,
			path:           "/read",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "write endpoint with write ops enabled",
			enableWriteOps: true,
			path:           "/write",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "write endpoint with write ops disabled",
			enableWriteOps: false,
			path:           "/write",
			method:         http.MethodPost,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "non-existent endpoint",
			enableWriteOps: true,
			path:           "/nonexistent",
			method:         http.MethodGet,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that always returns 200 OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Create the middleware with our test spec
			middleware := ConfigMiddlewareWithSpec(tt.enableWriteOps, getTestSpec)
			handler := middleware(nextHandler)

			// Create a test request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Serve the request
			handler.ServeHTTP(w, req)

			// Check the response
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
