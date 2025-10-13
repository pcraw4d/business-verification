package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// GrafanaDashboard represents a Grafana dashboard configuration
type GrafanaDashboard struct {
	ID          int                    `json:"id"`
	UID         string                 `json:"uid"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	TimeZone    string                 `json:"timezone"`
	Panels      []GrafanaPanel         `json:"panels"`
	Time        GrafanaTimeRange       `json:"time"`
	Refresh     string                 `json:"refresh"`
	SchemaVersion int                  `json:"schemaVersion"`
	Version     int                    `json:"version"`
	Links       []GrafanaLink          `json:"links"`
	Annotations GrafanaAnnotations     `json:"annotations"`
	Templating  GrafanaTemplating      `json:"templating"`
	Editable    bool                   `json:"editable"`
	HideControls bool                  `json:"hideControls"`
	SharedCrosshair bool               `json:"sharedCrosshair"`
	GraphTooltip int                   `json:"graphTooltip"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GrafanaPanel represents a Grafana panel
type GrafanaPanel struct {
	ID          int                    `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	GridPos     GrafanaGridPos         `json:"gridPos"`
	Targets     []GrafanaTarget        `json:"targets"`
	YAxes       []GrafanaYAxis         `json:"yAxes"`
	XAxis       GrafanaXAxis           `json:"xAxis"`
	Legend      GrafanaLegend          `json:"legend"`
	Thresholds  []GrafanaThreshold     `json:"thresholds"`
	Options     map[string]interface{} `json:"options"`
	FieldConfig GrafanaFieldConfig     `json:"fieldConfig"`
	Transparent bool                   `json:"transparent"`
	Datasource  string                 `json:"datasource"`
	Description string                 `json:"description"`
	Links       []GrafanaLink          `json:"links"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GrafanaGridPos represents panel grid position
type GrafanaGridPos struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

// GrafanaTarget represents a query target
type GrafanaTarget struct {
	Expr         string                 `json:"expr"`
	Format       string                 `json:"format"`
	Interval     string                 `json:"interval"`
	LegendFormat string                 `json:"legendFormat"`
	RefID        string                 `json:"refId"`
	Step         int                    `json:"step"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// GrafanaYAxis represents Y-axis configuration
type GrafanaYAxis struct {
	Label  string  `json:"label"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Unit   string  `json:"unit"`
	LogBase int    `json:"logBase"`
}

// GrafanaXAxis represents X-axis configuration
type GrafanaXAxis struct {
	Mode  string `json:"mode"`
	Name  string `json:"name"`
	Show  bool   `json:"show"`
	Values []string `json:"values"`
}

// GrafanaLegend represents legend configuration
type GrafanaLegend struct {
	Avg     bool   `json:"avg"`
	Current bool   `json:"current"`
	Max     bool   `json:"max"`
	Min     bool   `json:"min"`
	Show    bool   `json:"show"`
	Total   bool   `json:"total"`
	Values  bool   `json:"values"`
	AsTable bool   `json:"asTable"`
	AsRightSide bool `json:"asRightSide"`
}

// GrafanaThreshold represents threshold configuration
type GrafanaThreshold struct {
	ColorMode string  `json:"colorMode"`
	Fill     bool    `json:"fill"`
	Line     bool    `json:"line"`
	Op       string  `json:"op"`
	Value    float64 `json:"value"`
	YAxis    int     `json:"yAxis"`
}

// GrafanaFieldConfig represents field configuration
type GrafanaFieldConfig struct {
	Defaults GrafanaFieldConfigDefaults `json:"defaults"`
	Overrides []GrafanaFieldConfigOverride `json:"overrides"`
}

// GrafanaFieldConfigDefaults represents field config defaults
type GrafanaFieldConfigDefaults struct {
	Color      GrafanaColor `json:"color"`
	Custom     map[string]interface{} `json:"custom"`
	Decimals   int          `json:"decimals"`
	DisplayName string      `json:"displayName"`
	Filterable bool         `json:"filterable"`
	Links      []GrafanaLink `json:"links"`
	Mappings   []interface{} `json:"mappings"`
	Max        float64      `json:"max"`
	Min        float64      `json:"min"`
	NoValue    string       `json:"noValue"`
	NullValueMode string    `json:"nullValueMode"`
	Path       string       `json:"path"`
	Thresholds GrafanaThresholds `json:"thresholds"`
	Type       string       `json:"type"`
	Unit       string       `json:"unit"`
}

// GrafanaFieldConfigOverride represents field config override
type GrafanaFieldConfigOverride struct {
	Matcher GrafanaMatcher `json:"matcher"`
	Properties []GrafanaProperty `json:"properties"`
}

// GrafanaMatcher represents a matcher
type GrafanaMatcher struct {
	ID      string `json:"id"`
	Options string `json:"options"`
}

// GrafanaProperty represents a property
type GrafanaProperty struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// GrafanaColor represents color configuration
type GrafanaColor struct {
	Mode  string `json:"mode"`
	FixedColor string `json:"fixedColor,omitempty"`
}

// GrafanaThresholds represents thresholds configuration
type GrafanaThresholds struct {
	Mode  string           `json:"mode"`
	Steps []GrafanaThreshold `json:"steps"`
}

// GrafanaTimeRange represents time range configuration
type GrafanaTimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// GrafanaLink represents a dashboard link
type GrafanaLink struct {
	AsDropdown bool   `json:"asDropdown"`
	Icon       string `json:"icon"`
	IncludeVars bool  `json:"includeVars"`
	KeepTime   bool   `json:"keepTime"`
	Tags       []string `json:"tags"`
	TargetBlank bool   `json:"targetBlank"`
	Title      string `json:"title"`
	Tooltip    string `json:"tooltip"`
	Type       string `json:"type"`
	URL        string `json:"url"`
}

// GrafanaAnnotations represents annotations configuration
type GrafanaAnnotations struct {
	List []GrafanaAnnotation `json:"list"`
}

// GrafanaAnnotation represents an annotation
type GrafanaAnnotation struct {
	BuiltIn    int    `json:"builtIn"`
	Datasource string `json:"datasource"`
	Enable     bool   `json:"enable"`
	Hide       bool   `json:"hide"`
	IconColor  string `json:"iconColor"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}

// GrafanaTemplating represents templating configuration
type GrafanaTemplating struct {
	List []GrafanaTemplate `json:"list"`
}

// GrafanaTemplate represents a template variable
type GrafanaTemplate struct {
	AllValue   string                 `json:"allValue"`
	Current    GrafanaCurrent         `json:"current"`
	Datasource string                 `json:"datasource"`
	Definition string                 `json:"definition"`
	Hide       int                    `json:"hide"`
	IncludeAll bool                   `json:"includeAll"`
	Label      string                 `json:"label"`
	Multi      bool                   `json:"multi"`
	Name       string                 `json:"name"`
	Options    []GrafanaOption        `json:"options"`
	Query      string                 `json:"query"`
	Refresh    int                    `json:"refresh"`
	Regex      string                 `json:"regex"`
	SkipURLSync bool                  `json:"skipUrlSync"`
	Sort       int                    `json:"sort"`
	TagValuesQuery string             `json:"tagValuesQuery"`
	Tags       []string               `json:"tags"`
	TagsQuery  string                 `json:"tagsQuery"`
	Type       string                 `json:"type"`
	UseTags    bool                   `json:"useTags"`
}

// GrafanaCurrent represents current template value
type GrafanaCurrent struct {
	Selected bool     `json:"selected"`
	Text     string   `json:"text"`
	Value    string   `json:"value"`
}

// GrafanaOption represents a template option
type GrafanaOption struct {
	Selected bool   `json:"selected"`
	Text     string `json:"text"`
	Value    string `json:"value"`
}

// GrafanaClient represents a Grafana API client
type GrafanaClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// GrafanaConfig represents Grafana configuration
type GrafanaConfig struct {
	BaseURL    string        `json:"base_url"`
	APIKey     string        `json:"api_key"`
	Timeout    time.Duration `json:"timeout"`
	Username   string        `json:"username"`
	Password   string        `json:"password"`
}

// NewGrafanaClient creates a new Grafana client
func NewGrafanaClient(config GrafanaConfig, logger *zap.Logger) *GrafanaClient {
	return &GrafanaClient{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// CreateDashboard creates a Grafana dashboard
func (gc *GrafanaClient) CreateDashboard(ctx context.Context, dashboard *GrafanaDashboard) error {
	url := fmt.Sprintf("%s/api/dashboards/db", gc.baseURL)
	
	payload := map[string]interface{}{
		"dashboard": dashboard,
		"overwrite": true,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gc.apiKey))
	
	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create dashboard: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create dashboard: status %d", resp.StatusCode)
	}
	
	gc.logger.Info("Created Grafana dashboard", zap.String("title", dashboard.Title))
	return nil
}

// GetDashboard retrieves a Grafana dashboard
func (gc *GrafanaClient) GetDashboard(ctx context.Context, uid string) (*GrafanaDashboard, error) {
	url := fmt.Sprintf("%s/api/dashboards/uid/%s", gc.baseURL, uid)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gc.apiKey))
	
	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get dashboard: status %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	dashboardData, ok := result["dashboard"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid dashboard data")
	}
	
	jsonData, err := json.Marshal(dashboardData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dashboard data: %w", err)
	}
	
	var dashboard GrafanaDashboard
	if err := json.Unmarshal(jsonData, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dashboard: %w", err)
	}
	
	return &dashboard, nil
}

// DeleteDashboard deletes a Grafana dashboard
func (gc *GrafanaClient) DeleteDashboard(ctx context.Context, uid string) error {
	url := fmt.Sprintf("%s/api/dashboards/uid/%s", gc.baseURL, uid)
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gc.apiKey))
	
	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete dashboard: status %d", resp.StatusCode)
	}
	
	gc.logger.Info("Deleted Grafana dashboard", zap.String("uid", uid))
	return nil
}

// CreateRiskAssessmentDashboard creates a comprehensive risk assessment dashboard
func (gc *GrafanaClient) CreateRiskAssessmentDashboard(ctx context.Context) error {
	dashboard := &GrafanaDashboard{
		ID:          1,
		UID:         "risk-assessment-overview",
		Title:       "Risk Assessment Service Overview",
		Description: "Comprehensive monitoring dashboard for the Risk Assessment Service",
		Tags:        []string{"risk-assessment", "monitoring", "enterprise"},
		TimeZone:    "browser",
		Refresh:     "30s",
		SchemaVersion: 30,
		Version:     1,
		Editable:    true,
		HideControls: false,
		SharedCrosshair: true,
		GraphTooltip: 0,
		Time: GrafanaTimeRange{
			From: "now-1h",
			To:   "now",
		},
		Panels: []GrafanaPanel{
			gc.createRequestRatePanel(),
			gc.createResponseTimePanel(),
			gc.createErrorRatePanel(),
			gc.createRiskAssessmentPanel(),
			gc.createCompliancePanel(),
			gc.createSystemMetricsPanel(),
			gc.createTenantMetricsPanel(),
		},
		Annotations: GrafanaAnnotations{
			List: []GrafanaAnnotation{
				{
					BuiltIn:    1,
					Datasource: "-- Grafana --",
					Enable:     true,
					Hide:       true,
					IconColor:  "rgba(0, 211, 255, 1)",
					Name:       "Annotations & Alerts",
					Type:       "dashboard",
				},
			},
		},
		Templating: GrafanaTemplating{
			List: []GrafanaTemplate{
				{
					AllValue:   "",
					Current:    GrafanaCurrent{Selected: false, Text: "All", Value: "$__all"},
					Datasource: "Prometheus",
					Definition: "label_values(risk_assessment_http_requests_total, tenant_id)",
					Hide:       0,
					IncludeAll: true,
					Label:      "Tenant",
					Multi:      true,
					Name:       "tenant",
					Options:    []GrafanaOption{},
					Query:      "label_values(risk_assessment_http_requests_total, tenant_id)",
					Refresh:    1,
					Regex:      "",
					SkipURLSync: false,
					Sort:       0,
					Type:       "query",
				},
			},
		},
	}
	
	return gc.CreateDashboard(ctx, dashboard)
}

// createRequestRatePanel creates a request rate panel
func (gc *GrafanaClient) createRequestRatePanel() GrafanaPanel {
	return GrafanaPanel{
		ID:    1,
		Title: "Request Rate",
		Type:  "graph",
		GridPos: GrafanaGridPos{H: 8, W: 12, X: 0, Y: 0},
		Targets: []GrafanaTarget{
			{
				Expr:         "rate(risk_assessment_http_requests_total[5m])",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "{{method}} {{endpoint}}",
				RefID:        "A",
			},
		},
		YAxes: []GrafanaYAxis{
			{Label: "Requests/sec", Min: 0, Max: 0, Unit: "reqps", LogBase: 1},
			{Label: "", Min: 0, Max: 0, Unit: "short", LogBase: 1},
		},
		XAxis: GrafanaXAxis{Mode: "time", Name: "", Show: true, Values: []string{}},
		Legend: GrafanaLegend{
			Avg: false, Current: false, Max: false, Min: false, Show: true, Total: false, Values: false,
			AsTable: false, AsRightSide: false,
		},
		Thresholds: []GrafanaThreshold{},
		Options:    map[string]interface{}{},
		FieldConfig: GrafanaFieldConfig{
			Defaults: GrafanaFieldConfigDefaults{
				Color: GrafanaColor{Mode: "palette-classic"},
				Custom: map[string]interface{}{},
				Decimals: 2,
				DisplayName: "",
				Filterable: false,
				Links: []GrafanaLink{},
				Mappings: []interface{}{},
				Max: 0,
				Min: 0,
				NoValue: "",
				NullValueMode: "null",
				Path: "",
				Thresholds: GrafanaThresholds{
					Mode: "absolute",
					Steps: []GrafanaThreshold{
						{ColorMode: "critical", Fill: true, Line: true, Op: "gt", Value: 1000, YAxis: 0},
						{ColorMode: "warning", Fill: true, Line: true, Op: "gt", Value: 500, YAxis: 0},
					},
				},
				Type: "number",
				Unit: "reqps",
			},
			Overrides: []GrafanaFieldConfigOverride{},
		},
		Transparent: false,
		Datasource: "Prometheus",
		Description: "HTTP request rate over time",
		Links: []GrafanaLink{},
	}
}

// createResponseTimePanel creates a response time panel
func (gc *GrafanaClient) createResponseTimePanel() GrafanaPanel {
	return GrafanaPanel{
		ID:    2,
		Title: "Response Time",
		Type:  "graph",
		GridPos: GrafanaGridPos{H: 8, W: 12, X: 12, Y: 0},
		Targets: []GrafanaTarget{
			{
				Expr:         "histogram_quantile(0.50, rate(risk_assessment_http_request_duration_seconds_bucket[5m]))",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "P50",
				RefID:        "A",
			},
			{
				Expr:         "histogram_quantile(0.95, rate(risk_assessment_http_request_duration_seconds_bucket[5m]))",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "P95",
				RefID:        "B",
			},
			{
				Expr:         "histogram_quantile(0.99, rate(risk_assessment_http_request_duration_seconds_bucket[5m]))",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "P99",
				RefID:        "C",
			},
		},
		YAxes: []GrafanaYAxis{
			{Label: "Response Time", Min: 0, Max: 0, Unit: "s", LogBase: 1},
			{Label: "", Min: 0, Max: 0, Unit: "short", LogBase: 1},
		},
		XAxis: GrafanaXAxis{Mode: "time", Name: "", Show: true, Values: []string{}},
		Legend: GrafanaLegend{
			Avg: false, Current: false, Max: false, Min: false, Show: true, Total: false, Values: false,
			AsTable: false, AsRightSide: false,
		},
		Thresholds: []GrafanaThreshold{},
		Options:    map[string]interface{}{},
		FieldConfig: GrafanaFieldConfig{
			Defaults: GrafanaFieldConfigDefaults{
				Color: GrafanaColor{Mode: "palette-classic"},
				Custom: map[string]interface{}{},
				Decimals: 3,
				DisplayName: "",
				Filterable: false,
				Links: []GrafanaLink{},
				Mappings: []interface{}{},
				Max: 0,
				Min: 0,
				NoValue: "",
				NullValueMode: "null",
				Path: "",
				Thresholds: GrafanaThresholds{
					Mode: "absolute",
					Steps: []GrafanaThreshold{
						{ColorMode: "critical", Fill: true, Line: true, Op: "gt", Value: 2, YAxis: 0},
						{ColorMode: "warning", Fill: true, Line: true, Op: "gt", Value: 1, YAxis: 0},
					},
				},
				Type: "number",
				Unit: "s",
			},
			Overrides: []GrafanaFieldConfigOverride{},
		},
		Transparent: false,
		Datasource: "Prometheus",
		Description: "HTTP response time percentiles",
		Links: []GrafanaLink{},
	}
}

// createErrorRatePanel creates an error rate panel
func (gc *GrafanaClient) createErrorRatePanel() GrafanaPanel {
	return GrafanaPanel{
		ID:    3,
		Title: "Error Rate",
		Type:  "graph",
		GridPos: GrafanaGridPos{H: 8, W: 12, X: 0, Y: 8},
		Targets: []GrafanaTarget{
			{
				Expr:         "rate(risk_assessment_errors_total[5m])",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "{{error_type}}",
				RefID:        "A",
			},
		},
		YAxes: []GrafanaYAxis{
			{Label: "Errors/sec", Min: 0, Max: 0, Unit: "ops", LogBase: 1},
			{Label: "", Min: 0, Max: 0, Unit: "short", LogBase: 1},
		},
		XAxis: GrafanaXAxis{Mode: "time", Name: "", Show: true, Values: []string{}},
		Legend: GrafanaLegend{
			Avg: false, Current: false, Max: false, Min: false, Show: true, Total: false, Values: false,
			AsTable: false, AsRightSide: false,
		},
		Thresholds: []GrafanaThreshold{},
		Options:    map[string]interface{}{},
		FieldConfig: GrafanaFieldConfig{
			Defaults: GrafanaFieldConfigDefaults{
				Color: GrafanaColor{Mode: "palette-classic"},
				Custom: map[string]interface{}{},
				Decimals: 2,
				DisplayName: "",
				Filterable: false,
				Links: []GrafanaLink{},
				Mappings: []interface{}{},
				Max: 0,
				Min: 0,
				NoValue: "",
				NullValueMode: "null",
				Path: "",
				Thresholds: GrafanaThresholds{
					Mode: "absolute",
					Steps: []GrafanaThreshold{
						{ColorMode: "critical", Fill: true, Line: true, Op: "gt", Value: 10, YAxis: 0},
						{ColorMode: "warning", Fill: true, Line: true, Op: "gt", Value: 5, YAxis: 0},
					},
				},
				Type: "number",
				Unit: "ops",
			},
			Overrides: []GrafanaFieldConfigOverride{},
		},
		Transparent: false,
		Datasource: "Prometheus",
		Description: "Error rate over time",
		Links: []GrafanaLink{},
	}
}

// createRiskAssessmentPanel creates a risk assessment panel
func (gc *GrafanaClient) createRiskAssessmentPanel() GrafanaPanel {
	return GrafanaPanel{
		ID:    4,
		Title: "Risk Assessments",
		Type:  "graph",
		GridPos: GrafanaGridPos{H: 8, W: 12, X: 12, Y: 8},
		Targets: []GrafanaTarget{
			{
				Expr:         "rate(risk_assessment_total[5m])",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "{{assessment_type}}",
				RefID:        "A",
			},
		},
		YAxes: []GrafanaYAxis{
			{Label: "Assessments/sec", Min: 0, Max: 0, Unit: "ops", LogBase: 1},
			{Label: "", Min: 0, Max: 0, Unit: "short", LogBase: 1},
		},
		XAxis: GrafanaXAxis{Mode: "time", Name: "", Show: true, Values: []string{}},
		Legend: GrafanaLegend{
			Avg: false, Current: false, Max: false, Min: false, Show: true, Total: false, Values: false,
			AsTable: false, AsRightSide: false,
		},
		Thresholds: []GrafanaThreshold{},
		Options:    map[string]interface{}{},
		FieldConfig: GrafanaFieldConfig{
			Defaults: GrafanaFieldConfigDefaults{
				Color: GrafanaColor{Mode: "palette-classic"},
				Custom: map[string]interface{}{},
				Decimals: 2,
				DisplayName: "",
				Filterable: false,
				Links: []GrafanaLink{},
				Mappings: []interface{}{},
				Max: 0,
				Min: 0,
				NoValue: "",
				NullValueMode: "null",
				Path: "",
				Thresholds: GrafanaThresholds{
					Mode: "absolute",
					Steps: []GrafanaThreshold{},
				},
				Type: "number",
				Unit: "ops",
			},
			Overrides: []GrafanaFieldConfigOverride{},
		},
		Transparent: false,
		Datasource: "Prometheus",
		Description: "Risk assessment rate over time",
		Links: []GrafanaLink{},
	}
}

// createCompliancePanel creates a compliance panel
func (gc *GrafanaClient) createCompliancePanel() GrafanaPanel {
	return GrafanaPanel{
		ID:    5,
		Title: "Compliance Checks",
		Type:  "graph",
		GridPos: GrafanaGridPos{H: 8, W: 12, X: 0, Y: 16},
		Targets: []GrafanaTarget{
			{
				Expr:         "rate(compliance_checks_total[5m])",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "{{regulation}}",
				RefID:        "A",
			},
		},
		YAxes: []GrafanaYAxis{
			{Label: "Checks/sec", Min: 0, Max: 0, Unit: "ops", LogBase: 1},
			{Label: "", Min: 0, Max: 0, Unit: "short", LogBase: 1},
		},
		XAxis: GrafanaXAxis{Mode: "time", Name: "", Show: true, Values: []string{}},
		Legend: GrafanaLegend{
			Avg: false, Current: false, Max: false, Min: false, Show: true, Total: false, Values: false,
			AsTable: false, AsRightSide: false,
		},
		Thresholds: []GrafanaThreshold{},
		Options:    map[string]interface{}{},
		FieldConfig: GrafanaFieldConfig{
			Defaults: GrafanaFieldConfigDefaults{
				Color: GrafanaColor{Mode: "palette-classic"},
				Custom: map[string]interface{}{},
				Decimals: 2,
				DisplayName: "",
				Filterable: false,
				Links: []GrafanaLink{},
				Mappings: []interface{}{},
				Max: 0,
				Min: 0,
				NoValue: "",
				NullValueMode: "null",
				Path: "",
				Thresholds: GrafanaThresholds{
					Mode: "absolute",
					Steps: []GrafanaThreshold{},
				},
				Type: "number",
				Unit: "ops",
			},
			Overrides: []GrafanaFieldConfigOverride{},
		},
		Transparent: false,
		Datasource: "Prometheus",
		Description: "Compliance check rate over time",
		Links: []GrafanaLink{},
	}
}

// createSystemMetricsPanel creates a system metrics panel
func (gc *GrafanaClient) createSystemMetricsPanel() GrafanaPanel {
	return GrafanaPanel{
		ID:    6,
		Title: "System Metrics",
		Type:  "graph",
		GridPos: GrafanaGridPos{H: 8, W: 12, X: 12, Y: 16},
		Targets: []GrafanaTarget{
			{
				Expr:         "risk_assessment_active_connections",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "Active Connections",
				RefID:        "A",
			},
			{
				Expr:         "risk_assessment_database_connections",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "DB Connections ({{state}})",
				RefID:        "B",
			},
		},
		YAxes: []GrafanaYAxis{
			{Label: "Connections", Min: 0, Max: 0, Unit: "short", LogBase: 1},
			{Label: "", Min: 0, Max: 0, Unit: "short", LogBase: 1},
		},
		XAxis: GrafanaXAxis{Mode: "time", Name: "", Show: true, Values: []string{}},
		Legend: GrafanaLegend{
			Avg: false, Current: false, Max: false, Min: false, Show: true, Total: false, Values: false,
			AsTable: false, AsRightSide: false,
		},
		Thresholds: []GrafanaThreshold{},
		Options:    map[string]interface{}{},
		FieldConfig: GrafanaFieldConfig{
			Defaults: GrafanaFieldConfigDefaults{
				Color: GrafanaColor{Mode: "palette-classic"},
				Custom: map[string]interface{}{},
				Decimals: 0,
				DisplayName: "",
				Filterable: false,
				Links: []GrafanaLink{},
				Mappings: []interface{}{},
				Max: 0,
				Min: 0,
				NoValue: "",
				NullValueMode: "null",
				Path: "",
				Thresholds: GrafanaThresholds{
					Mode: "absolute",
					Steps: []GrafanaThreshold{},
				},
				Type: "number",
				Unit: "short",
			},
			Overrides: []GrafanaFieldConfigOverride{},
		},
		Transparent: false,
		Datasource: "Prometheus",
		Description: "System connection metrics",
		Links: []GrafanaLink{},
	}
}

// createTenantMetricsPanel creates a tenant metrics panel
func (gc *GrafanaClient) createTenantMetricsPanel() GrafanaPanel {
	return GrafanaPanel{
		ID:    7,
		Title: "Tenant Metrics",
		Type:  "graph",
		GridPos: GrafanaGridPos{H: 8, W: 24, X: 0, Y: 24},
		Targets: []GrafanaTarget{
			{
				Expr:         "rate(risk_assessment_tenant_requests_total[5m])",
				Format:       "time_series",
				Interval:     "",
				LegendFormat: "{{tenant_id}}",
				RefID:        "A",
			},
		},
		YAxes: []GrafanaYAxis{
			{Label: "Requests/sec", Min: 0, Max: 0, Unit: "reqps", LogBase: 1},
			{Label: "", Min: 0, Max: 0, Unit: "short", LogBase: 1},
		},
		XAxis: GrafanaXAxis{Mode: "time", Name: "", Show: true, Values: []string{}},
		Legend: GrafanaLegend{
			Avg: false, Current: false, Max: false, Min: false, Show: true, Total: false, Values: false,
			AsTable: false, AsRightSide: false,
		},
		Thresholds: []GrafanaThreshold{},
		Options:    map[string]interface{}{},
		FieldConfig: GrafanaFieldConfig{
			Defaults: GrafanaFieldConfigDefaults{
				Color: GrafanaColor{Mode: "palette-classic"},
				Custom: map[string]interface{}{},
				Decimals: 2,
				DisplayName: "",
				Filterable: false,
				Links: []GrafanaLink{},
				Mappings: []interface{}{},
				Max: 0,
				Min: 0,
				NoValue: "",
				NullValueMode: "null",
				Path: "",
				Thresholds: GrafanaThresholds{
					Mode: "absolute",
					Steps: []GrafanaThreshold{},
				},
				Type: "number",
				Unit: "reqps",
			},
			Overrides: []GrafanaFieldConfigOverride{},
		},
		Transparent: false,
		Datasource: "Prometheus",
		Description: "Request rate per tenant",
		Links: []GrafanaLink{},
	}
}
