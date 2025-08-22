package architecture

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// EventType represents the type of event
type EventType string

const (
	// System events
	EventTypeModuleStarted   EventType = "module.started"
	EventTypeModuleStopped   EventType = "module.stopped"
	EventTypeModuleFailed    EventType = "module.failed"
	EventTypeModuleHealthy   EventType = "module.healthy"
	EventTypeModuleUnhealthy EventType = "module.unhealthy"

	// Communication events
	EventTypeMessageSent     EventType = "message.sent"
	EventTypeMessageReceived EventType = "message.received"
	EventTypeMessageFailed   EventType = "message.failed"

	// Business events
	EventTypeVerificationStarted     EventType = "verification.started"
	EventTypeVerificationCompleted   EventType = "verification.completed"
	EventTypeVerificationFailed      EventType = "verification.failed"
	EventTypeClassificationStarted   EventType = "classification.started"
	EventTypeClassificationCompleted EventType = "classification.completed"
	EventTypeClassificationFailed    EventType = "classification.failed"

	// Data events
	EventTypeDataExtracted EventType = "data.extracted"
	EventTypeDataProcessed EventType = "data.processed"
	EventTypeDataStored    EventType = "data.stored"
	EventTypeDataFailed    EventType = "data.failed"
)

// EventPriority represents the priority level of an event
type EventPriority int

const (
	EventPriorityLow      EventPriority = 1
	EventPriorityNormal   EventPriority = 2
	EventPriorityHigh     EventPriority = 3
	EventPriorityCritical EventPriority = 4
)

// Event represents an event in the system
type Event struct {
	ID            string                 `json:"id"`
	Type          EventType              `json:"type"`
	Source        string                 `json:"source"`
	Target        string                 `json:"target,omitempty"`
	Priority      EventPriority          `json:"priority"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
}

// Message represents a message between modules
type Message struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Source        string                 `json:"source"`
	Target        string                 `json:"target"`
	Priority      EventPriority          `json:"priority"`
	Timestamp     time.Time              `json:"timestamp"`
	Payload       map[string]interface{} `json:"payload"`
	Headers       map[string]string      `json:"headers"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
	TTL           time.Duration          `json:"ttl,omitempty"`
}

// EventHandler handles events
type EventHandler func(ctx context.Context, event Event) error

// MessageHandler handles messages
type MessageHandler func(ctx context.Context, message Message) error

// EventFilter filters events
type EventFilter func(event Event) bool

// MessageFilter filters messages
type MessageFilter func(message Message) bool

// EventBus manages event communication between modules
type EventBus struct {
	handlers    map[EventType][]EventHandler
	filters     map[EventType][]EventFilter
	subscribers map[string][]EventHandler
	events      chan Event
	mu          sync.RWMutex
	tracer      trace.Tracer
	ctx         context.Context
	cancel      context.CancelFunc
	config      EventBusConfig
}

// EventBusConfig holds configuration for the event bus
type EventBusConfig struct {
	BufferSize    int           `json:"buffer_size"`
	WorkerCount   int           `json:"worker_count"`
	EventTTL      time.Duration `json:"event_ttl"`
	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`
	EnableTracing bool          `json:"enable_tracing"`
	EnableMetrics bool          `json:"enable_metrics"`
	PersistEvents bool          `json:"persist_events"`
	EventStore    EventStore    `json:"-"`
}

// EventStore interface for persisting events
type EventStore interface {
	Store(ctx context.Context, event Event) error
	Retrieve(ctx context.Context, filter EventFilter) ([]Event, error)
	Delete(ctx context.Context, eventID string) error
}

// NewEventBus creates a new event bus
func NewEventBus(config EventBusConfig) *EventBus {
	ctx, cancel := context.WithCancel(context.Background())

	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.WorkerCount == 0 {
		config.WorkerCount = 10
	}
	if config.EventTTL == 0 {
		config.EventTTL = 24 * time.Hour
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}

	eb := &EventBus{
		handlers:    make(map[EventType][]EventHandler),
		filters:     make(map[EventType][]EventFilter),
		subscribers: make(map[string][]EventHandler),
		events:      make(chan Event, config.BufferSize),
		tracer:      otel.Tracer("event-bus"),
		ctx:         ctx,
		cancel:      cancel,
		config:      config,
	}

	// Start event processing workers
	for i := 0; i < config.WorkerCount; i++ {
		go eb.eventWorker()
	}

	return eb
}

// Subscribe registers an event handler for a specific event type
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// SubscribeToModule registers an event handler for events from a specific module
func (eb *EventBus) SubscribeToModule(moduleID string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[moduleID] = append(eb.subscribers[moduleID], handler)
}

// AddFilter adds a filter for a specific event type
func (eb *EventBus) AddFilter(eventType EventType, filter EventFilter) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.filters[eventType] = append(eb.filters[eventType], filter)
}

// Publish publishes an event to the event bus
func (eb *EventBus) Publish(ctx context.Context, event Event) error {
	_, span := eb.tracer.Start(ctx, "PublishEvent")
	defer span.End()

	// Set event metadata
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.TraceID == "" {
		event.TraceID = trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	}

	span.SetAttributes(
		attribute.String("event.id", event.ID),
		attribute.String("event.type", string(event.Type)),
		attribute.String("event.source", event.Source),
		attribute.String("event.target", event.Target),
		attribute.Int("event.priority", int(event.Priority)),
	)

	// Persist event if enabled
	if eb.config.PersistEvents && eb.config.EventStore != nil {
		if err := eb.config.EventStore.Store(ctx, event); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to store event: %w", err)
		}
	}

	// Send event to processing channel
	select {
	case eb.events <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event bus buffer full")
	}
}

// PublishAsync publishes an event asynchronously
func (eb *EventBus) PublishAsync(event Event) {
	go func() {
		ctx := context.Background()
		if err := eb.Publish(ctx, event); err != nil {
			// Log error but don't block
			// In a real implementation, you might want to use a logger here
		}
	}()
}

// eventWorker processes events from the event channel
func (eb *EventBus) eventWorker() {
	for {
		select {
		case event := <-eb.events:
			eb.processEvent(context.Background(), event)
		case <-eb.ctx.Done():
			return
		}
	}
}

// processEvent processes a single event
func (eb *EventBus) processEvent(ctx context.Context, event Event) {
	_, span := eb.tracer.Start(ctx, "ProcessEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.id", event.ID),
		attribute.String("event.type", string(event.Type)),
		attribute.String("event.source", event.Source),
	)

	// Apply filters
	if !eb.applyFilters(event) {
		span.AddEvent("Event filtered out")
		return
	}

	// Get handlers for this event type
	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.handlers[event.Type]))
	copy(handlers, eb.handlers[event.Type])

	// Get subscribers for the source module
	subscribers := make([]EventHandler, len(eb.subscribers[event.Source]))
	copy(subscribers, eb.subscribers[event.Source])
	eb.mu.RUnlock()

	// Execute handlers with retry logic
	allHandlers := append(handlers, subscribers...)
	for _, handler := range allHandlers {
		eb.executeHandlerWithRetry(ctx, event, handler)
	}
}

// applyFilters applies all filters for the event type
func (eb *EventBus) applyFilters(event Event) bool {
	eb.mu.RLock()
	filters := eb.filters[event.Type]
	eb.mu.RUnlock()

	for _, filter := range filters {
		if !filter(event) {
			return false
		}
	}
	return true
}

// executeHandlerWithRetry executes a handler with retry logic
func (eb *EventBus) executeHandlerWithRetry(ctx context.Context, event Event, handler EventHandler) {
	for attempt := 0; attempt <= eb.config.RetryAttempts; attempt++ {
		if err := handler(ctx, event); err == nil {
			return
		} else if attempt == eb.config.RetryAttempts {
			// Log final failure
			// In a real implementation, you might want to use a logger here
		} else {
			time.Sleep(eb.config.RetryDelay * time.Duration(attempt+1))
		}
	}
}

// Close shuts down the event bus
func (eb *EventBus) Close() error {
	eb.cancel()
	close(eb.events)
	return nil
}

// MessageBus manages message communication between modules
type MessageBus struct {
	handlers map[string][]MessageHandler
	filters  map[string][]MessageFilter
	routes   map[string]string
	messages chan Message
	mu       sync.RWMutex
	tracer   trace.Tracer
	ctx      context.Context
	cancel   context.CancelFunc
	config   MessageBusConfig
}

// MessageBusConfig holds configuration for the message bus
type MessageBusConfig struct {
	BufferSize      int           `json:"buffer_size"`
	WorkerCount     int           `json:"worker_count"`
	MessageTTL      time.Duration `json:"message_ttl"`
	RetryAttempts   int           `json:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay"`
	EnableTracing   bool          `json:"enable_tracing"`
	EnableMetrics   bool          `json:"enable_metrics"`
	PersistMessages bool          `json:"persist_messages"`
	MessageStore    MessageStore  `json:"-"`
}

// MessageStore interface for persisting messages
type MessageStore interface {
	Store(ctx context.Context, message Message) error
	Retrieve(ctx context.Context, filter MessageFilter) ([]Message, error)
	Delete(ctx context.Context, messageID string) error
}

// NewMessageBus creates a new message bus
func NewMessageBus(config MessageBusConfig) *MessageBus {
	ctx, cancel := context.WithCancel(context.Background())

	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.WorkerCount == 0 {
		config.WorkerCount = 10
	}
	if config.MessageTTL == 0 {
		config.MessageTTL = 1 * time.Hour
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}

	mb := &MessageBus{
		handlers: make(map[string][]MessageHandler),
		filters:  make(map[string][]MessageFilter),
		routes:   make(map[string]string),
		messages: make(chan Message, config.BufferSize),
		tracer:   otel.Tracer("message-bus"),
		ctx:      ctx,
		cancel:   cancel,
		config:   config,
	}

	// Start message processing workers
	for i := 0; i < config.WorkerCount; i++ {
		go mb.messageWorker()
	}

	return mb
}

// Subscribe registers a message handler for a specific message type
func (mb *MessageBus) Subscribe(messageType string, handler MessageHandler) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.handlers[messageType] = append(mb.handlers[messageType], handler)
}

// AddFilter adds a filter for a specific message type
func (mb *MessageBus) AddFilter(messageType string, filter MessageFilter) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.filters[messageType] = append(mb.filters[messageType], filter)
}

// AddRoute adds a routing rule
func (mb *MessageBus) AddRoute(source, target string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.routes[source] = target
}

// Send sends a message to the message bus
func (mb *MessageBus) Send(ctx context.Context, message Message) error {
	_, span := mb.tracer.Start(ctx, "SendMessage")
	defer span.End()

	// Set message metadata
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	if message.TraceID == "" {
		message.TraceID = trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	}

	span.SetAttributes(
		attribute.String("message.id", message.ID),
		attribute.String("message.type", message.Type),
		attribute.String("message.source", message.Source),
		attribute.String("message.target", message.Target),
		attribute.Int("message.priority", int(message.Priority)),
	)

	// Persist message if enabled
	if mb.config.PersistMessages && mb.config.MessageStore != nil {
		if err := mb.config.MessageStore.Store(ctx, message); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to store message: %w", err)
		}
	}

	// Send message to processing channel
	select {
	case mb.messages <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("message bus buffer full")
	}
}

// SendAsync sends a message asynchronously
func (mb *MessageBus) SendAsync(message Message) {
	go func() {
		ctx := context.Background()
		if err := mb.Send(ctx, message); err != nil {
			// Log error but don't block
			// In a real implementation, you might want to use a logger here
		}
	}()
}

// messageWorker processes messages from the message channel
func (mb *MessageBus) messageWorker() {
	for {
		select {
		case message := <-mb.messages:
			mb.processMessage(context.Background(), message)
		case <-mb.ctx.Done():
			return
		}
	}
}

// processMessage processes a single message
func (mb *MessageBus) processMessage(ctx context.Context, message Message) {
	_, span := mb.tracer.Start(ctx, "ProcessMessage")
	defer span.End()

	span.SetAttributes(
		attribute.String("message.id", message.ID),
		attribute.String("message.type", message.Type),
		attribute.String("message.source", message.Source),
		attribute.String("message.target", message.Target),
	)

	// Check message TTL
	if time.Since(message.Timestamp) > mb.config.MessageTTL {
		span.AddEvent("Message expired")
		return
	}

	// Apply filters
	if !mb.applyFilters(message) {
		span.AddEvent("Message filtered out")
		return
	}

	// Get handlers for this message type
	mb.mu.RLock()
	handlers := make([]MessageHandler, len(mb.handlers[message.Type]))
	copy(handlers, mb.handlers[message.Type])
	mb.mu.RUnlock()

	// Execute handlers with retry logic
	for _, handler := range handlers {
		mb.executeHandlerWithRetry(ctx, message, handler)
	}
}

// applyFilters applies all filters for the message type
func (mb *MessageBus) applyFilters(message Message) bool {
	mb.mu.RLock()
	filters := mb.filters[message.Type]
	mb.mu.RUnlock()

	for _, filter := range filters {
		if !filter(message) {
			return false
		}
	}
	return true
}

// executeHandlerWithRetry executes a handler with retry logic
func (mb *MessageBus) executeHandlerWithRetry(ctx context.Context, message Message, handler MessageHandler) {
	for attempt := 0; attempt <= mb.config.RetryAttempts; attempt++ {
		if err := handler(ctx, message); err == nil {
			return
		} else if attempt == mb.config.RetryAttempts {
			// Log final failure
			// In a real implementation, you might want to use a logger here
		} else {
			time.Sleep(mb.config.RetryDelay * time.Duration(attempt+1))
		}
	}
}

// Close shuts down the message bus
func (mb *MessageBus) Close() error {
	mb.cancel()
	close(mb.messages)
	return nil
}

// CommunicationManager manages both event and message communication
type CommunicationManager struct {
	eventBus   *EventBus
	messageBus *MessageBus
	tracer     trace.Tracer
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewCommunicationManager creates a new communication manager
func NewCommunicationManager(eventConfig EventBusConfig, messageConfig MessageBusConfig) *CommunicationManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &CommunicationManager{
		eventBus:   NewEventBus(eventConfig),
		messageBus: NewMessageBus(messageConfig),
		tracer:     otel.Tracer("communication-manager"),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// PublishEvent publishes an event
func (cm *CommunicationManager) PublishEvent(ctx context.Context, event Event) error {
	return cm.eventBus.Publish(ctx, event)
}

// SendMessage sends a message
func (cm *CommunicationManager) SendMessage(ctx context.Context, message Message) error {
	return cm.messageBus.Send(ctx, message)
}

// SubscribeToEvent subscribes to an event type
func (cm *CommunicationManager) SubscribeToEvent(eventType EventType, handler EventHandler) {
	cm.eventBus.Subscribe(eventType, handler)
}

// SubscribeToMessage subscribes to a message type
func (cm *CommunicationManager) SubscribeToMessage(messageType string, handler MessageHandler) {
	cm.messageBus.Subscribe(messageType, handler)
}

// Close shuts down the communication manager
func (cm *CommunicationManager) Close() error {
	if err := cm.eventBus.Close(); err != nil {
		return err
	}
	if err := cm.messageBus.Close(); err != nil {
		return err
	}
	cm.cancel()
	return nil
}

// Helper functions
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
