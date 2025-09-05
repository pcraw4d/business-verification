package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LineageType represents the type of lineage
type LineageType string

const (
	LineageTypeDataFlow       LineageType = "data_flow"
	LineageTypeTransformation LineageType = "transformation"
	LineageTypeDependency     LineageType = "dependency"
	LineageTypeImpact         LineageType = "impact"
	LineageTypeSource         LineageType = "source"
	LineageTypeTarget         LineageType = "target"
	LineageTypeProcess        LineageType = "process"
	LineageTypeSystem         LineageType = "system"
)

// LineageStatus represents the lineage status
type LineageStatus string

const (
	LineageStatusActive     LineageStatus = "active"
	LineageStatusInactive   LineageStatus = "inactive"
	LineageStatusDeprecated LineageStatus = "deprecated"
	LineageStatusError      LineageStatus = "error"
)

// LineageDirection represents the lineage direction
type LineageDirection string

const (
	LineageDirectionUpstream      LineageDirection = "upstream"
	LineageDirectionDownstream    LineageDirection = "downstream"
	LineageDirectionBidirectional LineageDirection = "bidirectional"
)

// DataLineageRequest represents a data lineage request
type DataLineageRequest struct {
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Dataset         string                  `json:"dataset"`
	Type            LineageType             `json:"type"`
	Direction       LineageDirection        `json:"direction"`
	Depth           int                     `json:"depth"`
	Sources         []LineageSource         `json:"sources"`
	Targets         []LineageTarget         `json:"targets"`
	Processes       []LineageProcess        `json:"processes"`
	Transformations []LineageTransformation `json:"transformations"`
	Filters         LineageFilters          `json:"filters"`
	Options         LineageOptions          `json:"options"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// LineageSource represents a lineage source
type LineageSource struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Location   string                 `json:"location"`
	Format     string                 `json:"format"`
	Schema     map[string]interface{} `json:"schema"`
	Connection LineageConnection      `json:"connection"`
	Properties map[string]interface{} `json:"properties"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// LineageTarget represents a lineage target
type LineageTarget struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Location   string                 `json:"location"`
	Format     string                 `json:"format"`
	Schema     map[string]interface{} `json:"schema"`
	Connection LineageConnection      `json:"connection"`
	Properties map[string]interface{} `json:"properties"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// LineageTransformation represents a lineage transformation
type LineageTransformation struct {
	ID           string                      `json:"id"`
	Name         string                      `json:"name"`
	Type         string                      `json:"type"`
	Description  string                      `json:"description"`
	InputFields  []string                    `json:"input_fields"`
	OutputFields []string                    `json:"output_fields"`
	Logic        string                      `json:"logic"`
	Rules        []LineageTransformationRule `json:"rules"`
	Conditions   []TransformationCondition   `json:"conditions"`
	Metadata     map[string]interface{}      `json:"metadata"`
}

// LineageTransformationRule represents a lineage transformation rule
type LineageTransformationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Expression  string                 `json:"expression"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
}

// TransformationCondition represents a transformation condition
type TransformationCondition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Expression  string                 `json:"expression"`
	Parameters  map[string]interface{} `json:"parameters"`
	Operator    string                 `json:"operator"`
	Value       interface{}            `json:"value"`
}

// LineageConnection represents a lineage connection
type LineageConnection struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Protocol    string                 `json:"protocol"`
	Host        string                 `json:"host"`
	Port        int                    `json:"port"`
	Database    string                 `json:"database"`
	Schema      string                 `json:"schema"`
	Table       string                 `json:"table"`
	Credentials map[string]interface{} `json:"credentials"`
	Properties  map[string]interface{} `json:"properties"`
}

// LineageFilters represents lineage filters
type LineageFilters struct {
	Types     []string               `json:"types"`
	Statuses  []string               `json:"statuses"`
	DateRange DateRange              `json:"date_range"`
	Tags      []string               `json:"tags"`
	Owners    []string               `json:"owners"`
	Custom    map[string]interface{} `json:"custom"`
}

// DateRange represents a date range
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// LineageOptions represents lineage options
type LineageOptions struct {
	IncludeMetadata bool                   `json:"include_metadata"`
	IncludeSchema   bool                   `json:"include_schema"`
	IncludeStats    bool                   `json:"include_stats"`
	MaxDepth        int                    `json:"max_depth"`
	MaxNodes        int                    `json:"max_nodes"`
	Format          string                 `json:"format"`
	Direction       LineageDirection       `json:"direction"`
	Custom          map[string]interface{} `json:"custom"`
}

// DataLineageResponse represents a data lineage response
type DataLineageResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      LineageType            `json:"type"`
	Status    LineageStatus          `json:"status"`
	Dataset   string                 `json:"dataset"`
	Direction LineageDirection       `json:"direction"`
	Depth     int                    `json:"depth"`
	Nodes     []LineageNode          `json:"nodes"`
	Edges     []LineageEdge          `json:"edges"`
	Paths     []LineagePath          `json:"paths"`
	Impact    LineageImpact          `json:"impact"`
	Summary   LineageSummary         `json:"summary"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// LineageNode represents a lineage node
type LineageNode struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Category   string                 `json:"category"`
	Location   string                 `json:"location"`
	Status     string                 `json:"status"`
	Properties map[string]interface{} `json:"properties"`
	Schema     map[string]interface{} `json:"schema"`
	Stats      LineageStats           `json:"stats"`
	Metadata   map[string]interface{} `json:"metadata"`
	Position   Position               `json:"position"`
}

// LineageEdge represents a lineage edge
type LineageEdge struct {
	ID              string                 `json:"id"`
	Source          string                 `json:"source"`
	Target          string                 `json:"target"`
	Type            string                 `json:"type"`
	Direction       LineageDirection       `json:"direction"`
	Properties      map[string]interface{} `json:"properties"`
	Transformations []string               `json:"transformations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// LineagePath represents a lineage path
type LineagePath struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Nodes      []string               `json:"nodes"`
	Edges      []string               `json:"edges"`
	Length     int                    `json:"length"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// LineageImpact represents lineage impact analysis
type LineageImpact struct {
	AffectedNodes   []string               `json:"affected_nodes"`
	AffectedEdges   []string               `json:"affected_edges"`
	AffectedPaths   []string               `json:"affected_paths"`
	ImpactScore     float64                `json:"impact_score"`
	RiskLevel       string                 `json:"risk_level"`
	Recommendations []string               `json:"recommendations"`
	Analysis        map[string]interface{} `json:"analysis"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// LineageStats represents lineage statistics
type LineageStats struct {
	RowCount    int64                  `json:"row_count"`
	ColumnCount int                    `json:"column_count"`
	SizeBytes   int64                  `json:"size_bytes"`
	LastUpdated time.Time              `json:"last_updated"`
	RefreshRate string                 `json:"refresh_rate"`
	Quality     float64                `json:"quality"`
	Custom      map[string]interface{} `json:"custom"`
}

// Position represents a node position
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// LineageSummary represents a lineage summary
type LineageSummary struct {
	TotalNodes    int                    `json:"total_nodes"`
	TotalEdges    int                    `json:"total_edges"`
	TotalPaths    int                    `json:"total_paths"`
	NodeTypes     map[string]int         `json:"node_types"`
	EdgeTypes     map[string]int         `json:"edge_types"`
	PathTypes     map[string]int         `json:"path_types"`
	MaxDepth      int                    `json:"max_depth"`
	AvgPathLength float64                `json:"avg_path_length"`
	Complexity    string                 `json:"complexity"`
	Metrics       map[string]interface{} `json:"metrics"`
}

// LineageReport represents a lineage report
type LineageReport struct {
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	Type            string                  `json:"type"`
	Dataset         string                  `json:"dataset"`
	Period          string                  `json:"period"`
	Results         []DataLineageResponse   `json:"results"`
	Summary         LineageSummary          `json:"summary"`
	Trends          []LineageTrend          `json:"trends"`
	Recommendations []LineageRecommendation `json:"recommendations"`
	CreatedAt       time.Time               `json:"created_at"`
	Metadata        map[string]interface{}  `json:"metadata"`
}

// LineageTrend represents a lineage trend
type LineageTrend struct {
	Metric       string      `json:"metric"`
	Period       string      `json:"period"`
	Values       []float64   `json:"values"`
	Timestamps   []time.Time `json:"timestamps"`
	Direction    string      `json:"direction"`
	Change       float64     `json:"change"`
	Significance string      `json:"significance"`
}

// LineageRecommendation represents a lineage recommendation
type LineageRecommendation struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Impact      string   `json:"impact"`
	Effort      string   `json:"effort"`
	Actions     []string `json:"actions"`
	Benefits    []string `json:"benefits"`
	Risks       []string `json:"risks"`
	Timeline    string   `json:"timeline"`
}

// DataLineageHandler handles data lineage operations
type DataLineageHandler struct {
	logger   *zap.Logger
	lineages map[string]*DataLineageResponse
	jobs     map[string]*LineageJob
	reports  map[string]*LineageReport
	mutex    sync.RWMutex
}

// NewDataLineageHandler creates a new data lineage handler
func NewDataLineageHandler(logger *zap.Logger) *DataLineageHandler {
	return &DataLineageHandler{
		logger:   logger,
		lineages: make(map[string]*DataLineageResponse),
		jobs:     make(map[string]*LineageJob),
		reports:  make(map[string]*LineageReport),
	}
}

// CreateLineage handles POST /lineage
func (h *DataLineageHandler) CreateLineage(w http.ResponseWriter, r *http.Request) {
	var req DataLineageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validateLineageRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate unique ID
	id := fmt.Sprintf("lineage_%d", time.Now().UnixNano())

	// Create lineage response
	response := &DataLineageResponse{
		ID:        id,
		Name:      req.Name,
		Type:      req.Type,
		Status:    LineageStatusActive,
		Dataset:   req.Dataset,
		Direction: req.Direction,
		Depth:     req.Depth,
		Nodes:     h.generateLineageNodes(req),
		Edges:     h.generateLineageEdges(req),
		Paths:     h.generateLineagePaths(req),
		Impact:    h.generateLineageImpact(req),
		Summary:   h.generateLineageSummary(req),
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	h.mutex.Lock()
	h.lineages[id] = response
	h.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetLineage handles GET /lineage?id={id}
func (h *DataLineageHandler) GetLineage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Lineage ID is required", http.StatusBadRequest)
		return
	}

	h.mutex.RLock()
	lineage, exists := h.lineages[id]
	h.mutex.RUnlock()

	if !exists {
		http.Error(w, "Lineage not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lineage)
}

// ListLineages handles GET /lineage
func (h *DataLineageHandler) ListLineages(w http.ResponseWriter, r *http.Request) {
	h.mutex.RLock()
	lineages := make([]*DataLineageResponse, 0, len(h.lineages))
	for _, lineage := range h.lineages {
		lineages = append(lineages, lineage)
	}
	h.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lineages": lineages,
		"total":    len(lineages),
	})
}

// CreateLineageJob handles POST /lineage/jobs
func (h *DataLineageHandler) CreateLineageJob(w http.ResponseWriter, r *http.Request) {
	var req DataLineageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validateLineageRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate unique job ID
	jobID := fmt.Sprintf("lineage_job_%d", time.Now().UnixNano())

	// Create background job
	job := &LineageJob{
		ID:     jobID,
		Status: "pending",
	}

	h.mutex.Lock()
	h.jobs[jobID] = job
	h.mutex.Unlock()

	// Simulate background processing
	go h.processLineageJob(job, req)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// GetLineageJob handles GET /lineage/jobs?id={id}
func (h *DataLineageHandler) GetLineageJob(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	h.mutex.RLock()
	job, exists := h.jobs[id]
	h.mutex.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// ListLineageJobs handles GET /lineage/jobs
func (h *DataLineageHandler) ListLineageJobs(w http.ResponseWriter, r *http.Request) {
	h.mutex.RLock()
	jobs := make([]*LineageJob, 0, len(h.jobs))
	for _, job := range h.jobs {
		jobs = append(jobs, job)
	}
	h.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  jobs,
		"total": len(jobs),
	})
}

// validateLineageRequest validates the lineage request
func (h *DataLineageHandler) validateLineageRequest(req DataLineageRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Dataset == "" {
		return fmt.Errorf("dataset is required")
	}
	if req.Type == "" {
		return fmt.Errorf("type is required")
	}
	if req.Direction == "" {
		return fmt.Errorf("direction is required")
	}
	if req.Depth <= 0 {
		return fmt.Errorf("depth must be greater than 0")
	}

	return nil
}

// generateLineageNodes generates lineage nodes
func (h *DataLineageHandler) generateLineageNodes(req DataLineageRequest) []LineageNode {
	var nodes []LineageNode

	// Generate source nodes
	for i, source := range req.Sources {
		node := LineageNode{
			ID:       source.ID,
			Name:     source.Name,
			Type:     source.Type,
			Category: "source",
			Location: source.Location,
			Status:   "active",
			Properties: map[string]interface{}{
				"format":     source.Format,
				"connection": source.Connection,
			},
			Schema: source.Schema,
			Stats: LineageStats{
				RowCount:    1000000,
				ColumnCount: 50,
				SizeBytes:   1024000000,
				LastUpdated: time.Now(),
				RefreshRate: "daily",
				Quality:     0.95,
			},
			Metadata: source.Metadata,
			Position: Position{
				X: float64(i * 100),
				Y: 0,
			},
		}
		nodes = append(nodes, node)
	}

	// Generate process nodes
	for i, process := range req.Processes {
		node := LineageNode{
			ID:       process.ID,
			Name:     process.Name,
			Type:     process.Type,
			Category: "process",
			Location: "internal",
			Status:   process.Status,
			Properties: map[string]interface{}{
				"inputs":  []string{},
				"outputs": []string{},
				"logic":   "sample logic",
			},
			Schema: make(map[string]interface{}),
			Stats: LineageStats{
				RowCount:    500000,
				ColumnCount: 45,
				SizeBytes:   512000000,
				LastUpdated: time.Now(),
				RefreshRate: "hourly",
				Quality:     0.92,
			},
			Metadata: make(map[string]interface{}),
			Position: Position{
				X: float64(i * 150),
				Y: 100,
			},
		}
		nodes = append(nodes, node)
	}

	// Generate target nodes
	for i, target := range req.Targets {
		node := LineageNode{
			ID:       target.ID,
			Name:     target.Name,
			Type:     target.Type,
			Category: "target",
			Location: target.Location,
			Status:   "active",
			Properties: map[string]interface{}{
				"format":     target.Format,
				"connection": target.Connection,
			},
			Schema: target.Schema,
			Stats: LineageStats{
				RowCount:    800000,
				ColumnCount: 40,
				SizeBytes:   819200000,
				LastUpdated: time.Now(),
				RefreshRate: "real-time",
				Quality:     0.98,
			},
			Metadata: target.Metadata,
			Position: Position{
				X: float64(i * 100),
				Y: 200,
			},
		}
		nodes = append(nodes, node)
	}

	return nodes
}

// generateLineageEdges generates lineage edges
func (h *DataLineageHandler) generateLineageEdges(req DataLineageRequest) []LineageEdge {
	var edges []LineageEdge

	// Generate edges from sources to processes
	for _, source := range req.Sources {
		for _, process := range req.Processes {
			edge := LineageEdge{
				ID:        fmt.Sprintf("edge_%s_%s", source.ID, process.ID),
				Source:    source.ID,
				Target:    process.ID,
				Type:      "data_flow",
				Direction: LineageDirectionDownstream,
				Properties: map[string]interface{}{
					"flow_type": "extract",
					"frequency": "daily",
				},
				Transformations: []string{},
				Metadata:        make(map[string]interface{}),
			}
			edges = append(edges, edge)
		}
	}

	// Generate edges from processes to targets
	for _, process := range req.Processes {
		for _, target := range req.Targets {
			edge := LineageEdge{
				ID:        fmt.Sprintf("edge_%s_%s", process.ID, target.ID),
				Source:    process.ID,
				Target:    target.ID,
				Type:      "data_flow",
				Direction: LineageDirectionDownstream,
				Properties: map[string]interface{}{
					"flow_type": "load",
					"frequency": "hourly",
				},
				Transformations: []string{},
				Metadata:        make(map[string]interface{}),
			}
			edges = append(edges, edge)
		}
	}

	return edges
}

// generateLineagePaths generates lineage paths
func (h *DataLineageHandler) generateLineagePaths(req DataLineageRequest) []LineagePath {
	var paths []LineagePath

	// Generate paths from sources to targets
	for _, source := range req.Sources {
		for _, target := range req.Targets {
			path := LineagePath{
				ID:     fmt.Sprintf("path_%s_%s", source.ID, target.ID),
				Name:   fmt.Sprintf("Path from %s to %s", source.Name, target.Name),
				Nodes:  []string{source.ID, target.ID},
				Edges:  []string{fmt.Sprintf("edge_%s_%s", source.ID, target.ID)},
				Length: 2,
				Type:   "data_flow",
				Properties: map[string]interface{}{
					"path_type":  "direct",
					"complexity": "low",
				},
				Metadata: make(map[string]interface{}),
			}
			paths = append(paths, path)
		}
	}

	return paths
}

// generateLineageImpact generates lineage impact analysis
func (h *DataLineageHandler) generateLineageImpact(req DataLineageRequest) LineageImpact {
	impact := LineageImpact{
		AffectedNodes: []string{},
		AffectedEdges: []string{},
		AffectedPaths: []string{},
		ImpactScore:   0.75,
		RiskLevel:     "medium",
		Recommendations: []string{
			"Monitor data quality metrics",
			"Implement data validation checks",
			"Set up automated alerts",
		},
		Analysis: map[string]interface{}{
			"critical_paths": 2,
			"bottlenecks":    1,
			"dependencies":   5,
		},
		Metadata: make(map[string]interface{}),
	}

	// Add affected nodes
	for _, source := range req.Sources {
		impact.AffectedNodes = append(impact.AffectedNodes, source.ID)
	}
	for _, target := range req.Targets {
		impact.AffectedNodes = append(impact.AffectedNodes, target.ID)
	}

	return impact
}

// generateLineageSummary generates lineage summary
func (h *DataLineageHandler) generateLineageSummary(req DataLineageRequest) LineageSummary {
	summary := LineageSummary{
		TotalNodes:    len(req.Sources) + len(req.Processes) + len(req.Targets),
		TotalEdges:    len(req.Sources)*len(req.Processes) + len(req.Processes)*len(req.Targets),
		TotalPaths:    len(req.Sources) * len(req.Targets),
		NodeTypes:     make(map[string]int),
		EdgeTypes:     make(map[string]int),
		PathTypes:     make(map[string]int),
		MaxDepth:      req.Depth,
		AvgPathLength: 2.0,
		Complexity:    "medium",
		Metrics:       make(map[string]interface{}),
	}

	// Count node types
	summary.NodeTypes["source"] = len(req.Sources)
	summary.NodeTypes["process"] = len(req.Processes)
	summary.NodeTypes["target"] = len(req.Targets)

	// Count edge types
	summary.EdgeTypes["data_flow"] = summary.TotalEdges

	// Count path types
	summary.PathTypes["data_flow"] = summary.TotalPaths

	// Add metrics
	summary.Metrics["data_volume"] = "1.5TB"
	summary.Metrics["refresh_frequency"] = "hourly"
	summary.Metrics["data_quality"] = 0.95

	return summary
}

// processLineageJob processes a lineage job in the background
func (h *DataLineageHandler) processLineageJob(job *LineageJob, req DataLineageRequest) {
	// Simulate processing time
	time.Sleep(3 * time.Second)

	h.mutex.Lock()
	job.Status = "running"
	h.mutex.Unlock()

	time.Sleep(1 * time.Second)

	h.mutex.Lock()
	h.mutex.Unlock()

	time.Sleep(1 * time.Second)

	h.mutex.Lock()
	h.mutex.Unlock()

	time.Sleep(1 * time.Second)

	// Create result
	_ = &DataLineageResponse{
		ID:        job.ID,
		Name:      req.Name,
		Type:      req.Type,
		Status:    LineageStatusActive,
		Dataset:   req.Dataset,
		Direction: req.Direction,
		Depth:     req.Depth,
		Nodes:     h.generateLineageNodes(req),
		Edges:     h.generateLineageEdges(req),
		Paths:     h.generateLineagePaths(req),
		Impact:    h.generateLineageImpact(req),
		Summary:   h.generateLineageSummary(req),
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	completedAt := time.Now()

	h.mutex.Lock()
	job.Status = "completed"
	job.CompletedAt = &completedAt
	h.mutex.Unlock()
}

// String conversion functions for enums
func (lt LineageType) String() string {
	return string(lt)
}

func (ls LineageStatus) String() string {
	return string(ls)
}

func (ld LineageDirection) String() string {
	return string(ld)
}
