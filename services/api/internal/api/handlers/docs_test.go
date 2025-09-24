package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocsHandler_ServeDocs(t *testing.T) {
	// Sample OpenAPI spec for testing
	openAPISpec := []byte(`openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0`)

	handler := NewDocsHandler(openAPISpec)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "serve docs page",
			path:           "/docs",
			expectedStatus: http.StatusOK,
			expectedBody:   "KYB Platform API Documentation",
		},
		{
			name:           "serve docs page with trailing slash",
			path:           "/docs/",
			expectedStatus: http.StatusOK,
			expectedBody:   "KYB Platform API Documentation",
		},
		{
			name:           "serve openapi spec",
			path:           "/docs/openapi.yaml",
			expectedStatus: http.StatusOK,
			expectedBody:   "openapi: 3.1.0",
		},
		{
			name:           "redirect unknown path to docs",
			path:           "/docs/unknown",
			expectedStatus: http.StatusMovedPermanently,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeDocs(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := w.Body.String()
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("expected body to contain '%s', got '%s'", tt.expectedBody, body)
				}
			}
		})
	}
}

func TestDocsHandler_ServeDocsPage(t *testing.T) {
	openAPISpec := []byte(`openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0`)

	handler := NewDocsHandler(openAPISpec)

	req := httptest.NewRequest("GET", "/docs", nil)
	w := httptest.NewRecorder()

	handler.ServeDocs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Check for essential HTML elements
	expectedElements := []string{
		"<!DOCTYPE html>",
		"<title>KYB Platform API Documentation</title>",
		"<div id=\"swagger-ui\"></div>",
		"swagger-ui-bundle.js",
		"swagger-ui-standalone-preset.js",
		"url: '/docs/openapi.yaml'",
		"setAuthToken()",
		"clearAuthToken()",
	}

	for _, element := range expectedElements {
		if !strings.Contains(body, element) {
			t.Errorf("expected HTML to contain '%s'", element)
		}
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected content type 'text/html; charset=utf-8', got '%s'", contentType)
	}
}

func TestDocsHandler_ServeOpenAPISpec(t *testing.T) {
	openAPISpec := []byte(`openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
  description: Test API specification`)

	handler := NewDocsHandler(openAPISpec)

	req := httptest.NewRequest("GET", "/docs/openapi.yaml", nil)
	w := httptest.NewRecorder()

	handler.ServeDocs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body != string(openAPISpec) {
		t.Errorf("expected body to match OpenAPI spec, got '%s'", body)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/yaml" {
		t.Errorf("expected content type 'application/yaml', got '%s'", contentType)
	}

	// Check cache headers
	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("expected cache control 'public, max-age=3600', got '%s'", cacheControl)
	}
}

func TestNewDocsHandler(t *testing.T) {
	openAPISpec := []byte("test spec")
	handler := NewDocsHandler(openAPISpec)

	if handler == nil {
		t.Error("expected handler to be created, got nil")
	}

	if string(handler.openAPISpec) != string(openAPISpec) {
		t.Error("expected handler to store the provided OpenAPI spec")
	}
}
