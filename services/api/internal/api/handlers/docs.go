package handlers

import (
	"html/template"
	"net/http"
)

// DocsHandler handles API documentation requests
type DocsHandler struct {
	openAPISpec []byte
}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler(openAPISpec []byte) *DocsHandler {
	return &DocsHandler{
		openAPISpec: openAPISpec,
	}
}

// ServeDocs serves the interactive API documentation
func (h *DocsHandler) ServeDocs(w http.ResponseWriter, r *http.Request) {
	// Serve the main documentation page
	if r.URL.Path == "/docs" || r.URL.Path == "/docs/" {
		h.serveDocsPage(w, r)
		return
	}

	// Serve the OpenAPI specification
	if r.URL.Path == "/docs/openapi.yaml" {
		h.serveOpenAPISpec(w, r)
		return
	}

	// Redirect to main docs page
	http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
}

// serveDocsPage serves the main documentation HTML page
func (h *DocsHandler) serveDocsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// HTML template for the documentation page
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KYB Platform API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #2c3e50;
        }
        .swagger-ui .topbar .download-url-wrapper .select-label {
            color: #fff;
        }
        .swagger-ui .info .title {
            color: #2c3e50;
        }
        .swagger-ui .scheme-container {
            background: #f8f9fa;
            margin: 0 0 20px;
            padding: 20px 0;
            box-shadow: 0 1px 2px 0 rgba(0,0,0,.15);
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/docs/openapi.yaml',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                docExpansion: "list",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true,
                tryItOutEnabled: true,
                requestInterceptor: function(request) {
                    const token = localStorage.getItem('kyb_api_token');
                    if (token) {
                        request.headers['Authorization'] = 'Bearer ' + token;
                    }
                    return request;
                },
                onComplete: function() {
                    const authContainer = document.createElement('div');
                    authContainer.innerHTML = '<div style="padding: 20px; background: #f8f9fa; border-bottom: 1px solid #dee2e6;"><h3 style="margin: 0 0 10px 0; color: #2c3e50;">Authentication</h3><div style="display: flex; gap: 10px; align-items: center;"><input type="text" id="api-token" placeholder="Enter your API token" style="flex: 1; padding: 8px; border: 1px solid #ddd; border-radius: 4px;"><button onclick="setAuthToken()" style="padding: 8px 16px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">Set Token</button><button onclick="clearAuthToken()" style="padding: 8px 16px; background: #6c757d; color: white; border: none; border-radius: 4px; cursor: pointer;">Clear</button></div><p style="margin: 10px 0 0 0; font-size: 14px; color: #6c757d;">Set your API token to test authenticated endpoints. Get your token from the login endpoint.</p></div>';
                    document.querySelector('.swagger-ui .topbar').appendChild(authContainer);
                    
                    const savedToken = localStorage.getItem('kyb_api_token');
                    if (savedToken) {
                        document.getElementById('api-token').value = savedToken;
                    }
                }
            });
            
            window.ui = ui;
        };
        
        function setAuthToken() {
            const token = document.getElementById('api-token').value;
            if (token) {
                localStorage.setItem('kyb_api_token', token);
                if (window.ui) {
                    window.ui.preauthorizeApiKey('BearerAuth', token);
                }
                alert('Token set successfully!');
            } else {
                alert('Please enter a valid token.');
            }
        }
        
        function clearAuthToken() {
            localStorage.removeItem('kyb_api_token');
            document.getElementById('api-token').value = '';
            if (window.ui) {
                window.ui.preauthorizeApiKey('BearerAuth', '');
            }
            alert('Token cleared!');
        }
    </script>
</body>
</html>`

	tmpl, err := template.New("docs").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// serveOpenAPISpec serves the OpenAPI specification
func (h *DocsHandler) serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Write(h.openAPISpec)
}
