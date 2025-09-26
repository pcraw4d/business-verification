package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIDocumentation represents API documentation
type APIDocumentation struct {
	Title       string                  `json:"title"`
	Version     string                  `json:"version"`
	Description string                  `json:"description"`
	BaseURL     string                  `json:"base_url"`
	Endpoints   []EndpointDocumentation `json:"endpoints"`
	Schemas     map[string]interface{}  `json:"schemas"`
}

// EndpointDocumentation represents documentation for an API endpoint
type EndpointDocumentation struct {
	Path        string                           `json:"path"`
	Method      string                           `json:"method"`
	Summary     string                           `json:"summary"`
	Description string                           `json:"description"`
	Parameters  []ParameterDocumentation         `json:"parameters"`
	RequestBody RequestBodyDocumentation         `json:"request_body"`
	Responses   map[string]ResponseDocumentation `json:"responses"`
	Tags        []string                         `json:"tags"`
}

// ParameterDocumentation represents documentation for a parameter
type ParameterDocumentation struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBodyDocumentation represents documentation for request body
type RequestBodyDocumentation struct {
	Required bool                   `json:"required"`
	Content  map[string]interface{} `json:"content"`
	Example  interface{}            `json:"example,omitempty"`
}

// ResponseDocumentation represents documentation for a response
type ResponseDocumentation struct {
	Description string                 `json:"description"`
	Content     map[string]interface{} `json:"content"`
	Example     interface{}            `json:"example,omitempty"`
}

// DocumentationGenerator generates API documentation
type DocumentationGenerator struct {
	documentation APIDocumentation
}

// NewDocumentationGenerator creates a new documentation generator
func NewDocumentationGenerator() *DocumentationGenerator {
	return &DocumentationGenerator{
		documentation: APIDocumentation{
			Title:       "KYB Platform API",
			Version:     "4.0.0",
			Description: "Know Your Business verification platform API",
			BaseURL:     "https://kyb-api-gateway-production.up.railway.app",
			Endpoints:   make([]EndpointDocumentation, 0),
			Schemas:     make(map[string]interface{}),
		},
	}
}

// AddEndpoint adds an endpoint to the documentation
func (dg *DocumentationGenerator) AddEndpoint(endpoint EndpointDocumentation) {
	dg.documentation.Endpoints = append(dg.documentation.Endpoints, endpoint)
}

// AddSchema adds a schema to the documentation
func (dg *DocumentationGenerator) AddSchema(name string, schema interface{}) {
	dg.documentation.Schemas[name] = schema
}

// GenerateDocumentation generates the complete API documentation
func (dg *DocumentationGenerator) GenerateDocumentation() APIDocumentation {
	return dg.documentation
}

// GenerateOpenAPISpec generates OpenAPI 3.0 specification
func (dg *DocumentationGenerator) GenerateOpenAPISpec() map[string]interface{} {
	openAPI := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       dg.documentation.Title,
			"version":     dg.documentation.Version,
			"description": dg.documentation.Description,
		},
		"servers": []map[string]interface{}{
			{
				"url":         dg.documentation.BaseURL,
				"description": "Production server",
			},
		},
		"paths": dg.generatePaths(),
		"components": map[string]interface{}{
			"schemas": dg.documentation.Schemas,
		},
	}

	return openAPI
}

// generatePaths generates the paths section for OpenAPI spec
func (dg *DocumentationGenerator) generatePaths() map[string]interface{} {
	paths := make(map[string]interface{})

	for _, endpoint := range dg.documentation.Endpoints {
		path := endpoint.Path
		if paths[path] == nil {
			paths[path] = make(map[string]interface{})
		}

		method := strings.ToLower(endpoint.Method)
		paths[path].(map[string]interface{})[method] = map[string]interface{}{
			"summary":     endpoint.Summary,
			"description": endpoint.Description,
			"parameters":  dg.generateParameters(endpoint.Parameters),
			"requestBody": dg.generateRequestBody(endpoint.RequestBody),
			"responses":   dg.generateResponses(endpoint.Responses),
			"tags":        endpoint.Tags,
		}
	}

	return paths
}

// generateParameters generates parameters for OpenAPI spec
func (dg *DocumentationGenerator) generateParameters(params []ParameterDocumentation) []map[string]interface{} {
	parameters := make([]map[string]interface{}, 0)

	for _, param := range params {
		parameter := map[string]interface{}{
			"name":        param.Name,
			"in":          "query", // Default to query parameter
			"required":    param.Required,
			"description": param.Description,
			"schema": map[string]interface{}{
				"type": param.Type,
			},
		}

		if param.Example != nil {
			parameter["example"] = param.Example
		}

		parameters = append(parameters, parameter)
	}

	return parameters
}

// generateRequestBody generates request body for OpenAPI spec
func (dg *DocumentationGenerator) generateRequestBody(body RequestBodyDocumentation) map[string]interface{} {
	if !body.Required {
		return nil
	}

	requestBody := map[string]interface{}{
		"required": body.Required,
		"content":  body.Content,
	}

	if body.Example != nil {
		requestBody["example"] = body.Example
	}

	return requestBody
}

// generateResponses generates responses for OpenAPI spec
func (dg *DocumentationGenerator) generateResponses(responses map[string]ResponseDocumentation) map[string]interface{} {
	openAPIResponses := make(map[string]interface{})

	for status, response := range responses {
		openAPIResponse := map[string]interface{}{
			"description": response.Description,
			"content":     response.Content,
		}

		if response.Example != nil {
			openAPIResponse["example"] = response.Example
		}

		openAPIResponses[status] = openAPIResponse
	}

	return openAPIResponses
}

// DocumentationHandler provides HTTP handler for API documentation
func (dg *DocumentationGenerator) DocumentationHandler(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	switch format {
	case "openapi":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dg.GenerateOpenAPISpec())
	case "swagger":
		// Return Swagger UI HTML
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, dg.generateSwaggerUI())
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dg.GenerateDocumentation())
	}
}

// generateSwaggerUI generates Swagger UI HTML
func (dg *DocumentationGenerator) generateSwaggerUI() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>%s API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '%s/docs?format=openapi',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>
`, dg.documentation.Title, dg.documentation.BaseURL)
}

// InitializeDefaultDocumentation initializes default API documentation
func (dg *DocumentationGenerator) InitializeDefaultDocumentation() {
	// Add V1 classification endpoint
	dg.AddEndpoint(EndpointDocumentation{
		Path:        "/v1/classify",
		Method:      "POST",
		Summary:     "Classify Business (V1)",
		Description: "Performs business classification and risk assessment with basic features",
		Parameters: []ParameterDocumentation{
			{
				Name:        "business_name",
				Type:        "string",
				Required:    true,
				Description: "Name of the business to classify",
				Example:     "Acme Corporation",
			},
			{
				Name:        "description",
				Type:        "string",
				Required:    true,
				Description: "Description of the business",
				Example:     "Technology company specializing in software development",
			},
		},
		RequestBody: RequestBodyDocumentation{
			Required: true,
			Content: map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"business_name": map[string]interface{}{
								"type": "string",
							},
							"description": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		},
		Responses: map[string]ResponseDocumentation{
			"200": {
				Description: "Successful classification",
				Content: map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"classification": map[string]interface{}{
									"type": "object",
								},
								"risk_assessment": map[string]interface{}{
									"type": "object",
								},
							},
						},
					},
				},
			},
		},
		Tags: []string{"classification", "v1"},
	})

	// Add V2 classification endpoint
	dg.AddEndpoint(EndpointDocumentation{
		Path:        "/v2/classify",
		Method:      "POST",
		Summary:     "Classify Business (V2)",
		Description: "Enhanced business classification with analytics, risk assessment, and compliance checking",
		Parameters: []ParameterDocumentation{
			{
				Name:        "business_name",
				Type:        "string",
				Required:    true,
				Description: "Name of the business to classify",
				Example:     "Acme Corporation",
			},
			{
				Name:        "description",
				Type:        "string",
				Required:    true,
				Description: "Description of the business",
				Example:     "Technology company specializing in software development",
			},
		},
		RequestBody: RequestBodyDocumentation{
			Required: true,
			Content: map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"business_name": map[string]interface{}{
								"type": "string",
							},
							"description": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		},
		Responses: map[string]ResponseDocumentation{
			"200": {
				Description: "Successful classification with enhanced features",
				Content: map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"classification": map[string]interface{}{
									"type": "object",
								},
								"risk_assessment": map[string]interface{}{
									"type": "object",
								},
								"analytics": map[string]interface{}{
									"type": "object",
								},
								"compliance_check": map[string]interface{}{
									"type": "object",
								},
							},
						},
					},
				},
			},
		},
		Tags: []string{"classification", "v2", "enhanced"},
	})

	// Add health endpoint
	dg.AddEndpoint(EndpointDocumentation{
		Path:        "/health",
		Method:      "GET",
		Summary:     "Health Check",
		Description: "Returns the health status of the service",
		Responses: map[string]ResponseDocumentation{
			"200": {
				Description: "Service is healthy",
				Content: map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"status": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
		},
		Tags: []string{"health"},
	})
}
