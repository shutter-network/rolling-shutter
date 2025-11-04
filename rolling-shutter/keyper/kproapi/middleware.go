package kproapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// isReadOnlyEndpoint checks if an endpoint is marked as read-only in the OpenAPI spec.
func isReadOnlyEndpoint(operation *openapi3.Operation) bool {
	// Try to get the value directly from the map first.
	if val, exists := operation.Extensions["x-read-only"]; exists {
		// Handle json.RawMessage case.
		if rawMsg, ok := val.(json.RawMessage); ok {
			return string(rawMsg) == "true"
		}

		// Handle direct boolean case.
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

// shouldEnableEndpoint determines if an endpoint should be accessible based on its type and configuration.
func shouldEnableEndpoint(operation *openapi3.Operation, enableWriteOperations bool) bool {
	if isReadOnlyEndpoint(operation) {
		return true
	}
	return enableWriteOperations
}

// findOperation looks up the OpenAPI operation for the given path and method.
func findOperation(spec *openapi3.T, path string, method string) *openapi3.Operation {
	pathItem := spec.Paths.Find(path) // first try to find the path in the spec
	if pathItem == nil {
		for specPath, pItem := range spec.Paths { // fallback for path containing parameters
			rePath := "^" + regexp.QuoteMeta(specPath)
			rePath = strings.ReplaceAll(rePath, `\{`, "{")
			rePath = strings.ReplaceAll(rePath, `\}`, "}")
			rePath = regexp.MustCompile(`\{[^/]+\}`).ReplaceAllString(rePath, `[^/]+`)
			rePath += "$"

			if matched, _ := regexp.MatchString(rePath, path); matched {
				pathItem = pItem
				break
			}
		}
	}

	if pathItem == nil { // if no path is found still, return nil
		return nil
	}

	switch method {
	case http.MethodGet:
		return pathItem.Get
	case http.MethodPost:
		return pathItem.Post
	case http.MethodPut:
		return pathItem.Put
	case http.MethodDelete:
		return pathItem.Delete
	default:
		return nil
	}
}

// ConfigMiddleware creates a middleware that controls endpoint access based on configuration.
func ConfigMiddleware(enableWriteOperations bool) MiddlewareFunc {
	return ConfigMiddlewareWithSpec(enableWriteOperations, GetSwagger)
}

// ConfigMiddlewareWithSpec creates a middleware that controls endpoint access based on configuration.
// This accepts a function to get the spec, making it more testable.
func ConfigMiddlewareWithSpec(enableWriteOperations bool, getSpec func() (*openapi3.T, error)) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Load the OpenAPI specification.
			spec, err := getSpec()
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Find the operation for this request.
			operation := findOperation(spec, r.URL.Path, r.Method)
			if operation == nil {
				http.Error(w, "Endpoint not found", http.StatusNotFound)
				return
			}

			// Check if the endpoint should be accessible.
			if !shouldEnableEndpoint(operation, enableWriteOperations) {
				http.Error(w, "Endpoint not enabled", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
