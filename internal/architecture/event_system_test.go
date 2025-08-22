package architecture

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockEventStore implements EventStore for testing
type MockEventStore struct {
	events map[string]Event
	mu     sync.RWMutex
}

func NewMockEventStore() *MockEventStore {
	return &MockEventStore{
		events: make(map[string]Event),
	}
}

func (m *MockEventStore) Store(ctx context.Context, event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[event.ID] = event
	return nil
}

func (m *MockEventStore) Retrieve(ctx context.Context, filter EventFilter) ([]Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var events []Event
	for _, event := range m.events {
		if filter(event) {
			events = append(events, event)
		}
	}
	return events, nil
}

func (m *MockEventStore) Delete(ctx context.Context, eventID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.events, eventID)
	return nil
}

// MockMessageStore implements MessageStore for testing
type MockMessageStore struct {
	messages map[string]Message
	mu       sync.RWMutex
}

func NewMockMessageStore() *MockMessageStore {
	return &MockMessageStore{
		messages: make(map[string]Message),
	}
}

func (m *MockMessageStore) Store(ctx context.Context, message Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages[message.ID] = message
	return nil
}

func (m *MockMessageStore) Retrieve(ctx context.Context, filter MessageFilter) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var messages []Message
	for _, message := range m.messages {
		if filter(message) {
			messages = append(messages, message)
		}
	}
	return messages, nil
}

func (m *MockMessageStore) Delete(ctx context.Context, messageID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.messages, messageID)
	return nil
}

func TestNewEventBus(t *testing.T) {
	config := EventBusConfig{
		BufferSize:    100,
		WorkerCount:   5,
		EventTTL:      time.Hour,
		RetryAttempts: 3,
		RetryDelay:    time.Second,
		EnableTracing: true,
		EnableMetrics: true,
		PersistEvents: true,
		EventStore:    NewMockEventStore(),
	}

	eventBus := NewEventBus(config)

	assert.NotNil(t, eventBus)
	assert.NotNil(t, eventBus.handlers)
	assert.NotNil(t, eventBus.filters)
	assert.NotNil(t, eventBus.subscribers)
	assert.NotNil(t, eventBus.events)
	assert.Equal(t, config, eventBus.config)
}

func TestEventBusSubscribe(t *testing.T) {
	eventBus := NewEventBus(EventBusConfig{})

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	eventBus.Subscribe(EventTypeModuleStarted, handler)

	// Verify handler was registered
	assert.Len(t, eventBus.handlers[EventTypeModuleStarted], 1)
}

func TestEventBusSubscribeToModule(t *testing.T) {
	eventBus := NewEventBus(EventBusConfig{})

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	eventBus.SubscribeToModule("test_module", handler)

	// Verify handler was registered
	assert.Len(t, eventBus.subscribers["test_module"], 1)
}

func TestEventBusAddFilter(t *testing.T) {
	eventBus := NewEventBus(EventBusConfig{})

	filter := func(event Event) bool {
		return event.Priority >= EventPriorityHigh
	}

	eventBus.AddFilter(EventTypeModuleStarted, filter)

	// Verify filter was registered
	assert.Len(t, eventBus.filters[EventTypeModuleStarted], 1)
}

func TestEventBusPublish(t *testing.T) {
	eventBus := NewEventBus(EventBusConfig{})

	var receivedEvent Event
	handler := func(ctx context.Context, event Event) error {
		receivedEvent = event
		return nil
	}

	eventBus.Subscribe(EventTypeModuleStarted, handler)

	event := Event{
		Type:     EventTypeModuleStarted,
		Source:   "test_module",
		Priority: EventPriorityNormal,
		Data:     map[string]interface{}{"key": "value"},
	}

	ctx := context.Background()
	err := eventBus.Publish(ctx, event)

	// Wait for event processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, event.Type, receivedEvent.Type)
	assert.Equal(t, event.Source, receivedEvent.Source)
	assert.Equal(t, event.Data["key"], receivedEvent.Data["key"])
}

func TestEventBusPublishWithPersistence(t *testing.T) {
	eventStore := NewMockEventStore()
	config := EventBusConfig{
		PersistEvents: true,
		EventStore:    eventStore,
	}

	eventBus := NewEventBus(config)

	event := Event{
		Type:     EventTypeModuleStarted,
		Source:   "test_module",
		Priority: EventPriorityNormal,
		Data:     map[string]interface{}{"key": "value"},
	}

	ctx := context.Background()
	err := eventBus.Publish(ctx, event)

	// Wait for event processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)

	// Verify event was stored
	events, err := eventStore.Retrieve(ctx, func(e Event) bool {
		return e.Type == EventTypeModuleStarted
	})
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, event.Type, events[0].Type)
}

func TestEventBusFilter(t *testing.T) {
	eventBus := NewEventBus(EventBusConfig{})

	var receivedEvent Event
	handler := func(ctx context.Context, event Event) error {
		receivedEvent = event
		return nil
	}

	// Add filter that only allows high priority events
	filter := func(event Event) bool {
		return event.Priority >= EventPriorityHigh
	}

	eventBus.AddFilter(EventTypeModuleStarted, filter)
	eventBus.Subscribe(EventTypeModuleStarted, handler)

	// Publish low priority event
	lowPriorityEvent := Event{
		Type:     EventTypeModuleStarted,
		Source:   "test_module",
		Priority: EventPriorityLow,
	}

	ctx := context.Background()
	err := eventBus.Publish(ctx, lowPriorityEvent)

	// Wait for event processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Empty(t, receivedEvent.ID) // Event should be filtered out

	// Publish high priority event
	highPriorityEvent := Event{
		Type:     EventTypeModuleStarted,
		Source:   "test_module",
		Priority: EventPriorityHigh,
	}

	err = eventBus.Publish(ctx, highPriorityEvent)

	// Wait for event processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, highPriorityEvent.Type, receivedEvent.Type) // Event should be processed
}

func TestEventBusRetry(t *testing.T) {
	eventBus := NewEventBus(EventBusConfig{
		RetryAttempts: 2,
		RetryDelay:    10 * time.Millisecond,
	})

	attempts := 0
	handler := func(ctx context.Context, event Event) error {
		attempts++
		if attempts < 3 {
			return assert.AnError
		}
		return nil
	}

	eventBus.Subscribe(EventTypeModuleStarted, handler)

	event := Event{
		Type:     EventTypeModuleStarted,
		Source:   "test_module",
		Priority: EventPriorityNormal,
	}

	ctx := context.Background()
	err := eventBus.Publish(ctx, event)

	// Wait for event processing and retries
	time.Sleep(200 * time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts) // Should have retried twice
}

func TestEventBusClose(t *testing.T) {
	eventBus := NewEventBus(EventBusConfig{})

	err := eventBus.Close()
	assert.NoError(t, err)
}

func TestNewMessageBus(t *testing.T) {
	config := MessageBusConfig{
		BufferSize:      100,
		WorkerCount:     5,
		MessageTTL:      time.Hour,
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		EnableTracing:   true,
		EnableMetrics:   true,
		PersistMessages: true,
		MessageStore:    NewMockMessageStore(),
	}

	messageBus := NewMessageBus(config)

	assert.NotNil(t, messageBus)
	assert.NotNil(t, messageBus.handlers)
	assert.NotNil(t, messageBus.filters)
	assert.NotNil(t, messageBus.routes)
	assert.NotNil(t, messageBus.messages)
	assert.Equal(t, config, messageBus.config)
}

func TestMessageBusSubscribe(t *testing.T) {
	messageBus := NewMessageBus(MessageBusConfig{})

	handler := func(ctx context.Context, message Message) error {
		return nil
	}

	messageBus.Subscribe("test_message", handler)

	// Verify handler was registered
	assert.Len(t, messageBus.handlers["test_message"], 1)
}

func TestMessageBusAddFilter(t *testing.T) {
	messageBus := NewMessageBus(MessageBusConfig{})

	filter := func(message Message) bool {
		return message.Priority >= EventPriorityHigh
	}

	messageBus.AddFilter("test_message", filter)

	// Verify filter was registered
	assert.Len(t, messageBus.filters["test_message"], 1)
}

func TestMessageBusAddRoute(t *testing.T) {
	messageBus := NewMessageBus(MessageBusConfig{})

	messageBus.AddRoute("source_module", "target_module")

	// Verify route was registered
	assert.Equal(t, "target_module", messageBus.routes["source_module"])
}

func TestMessageBusSend(t *testing.T) {
	messageBus := NewMessageBus(MessageBusConfig{})

	var receivedMessage Message
	handler := func(ctx context.Context, message Message) error {
		receivedMessage = message
		return nil
	}

	messageBus.Subscribe("test_message", handler)

	message := Message{
		Type:     "test_message",
		Source:   "test_module",
		Target:   "target_module",
		Priority: EventPriorityNormal,
		Payload:  map[string]interface{}{"key": "value"},
	}

	ctx := context.Background()
	err := messageBus.Send(ctx, message)

	// Wait for message processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, message.Type, receivedMessage.Type)
	assert.Equal(t, message.Source, receivedMessage.Source)
	assert.Equal(t, message.Payload["key"], receivedMessage.Payload["key"])
}

func TestMessageBusSendWithPersistence(t *testing.T) {
	messageStore := NewMockMessageStore()
	config := MessageBusConfig{
		PersistMessages: true,
		MessageStore:    messageStore,
	}

	messageBus := NewMessageBus(config)

	message := Message{
		Type:     "test_message",
		Source:   "test_module",
		Target:   "target_module",
		Priority: EventPriorityNormal,
		Payload:  map[string]interface{}{"key": "value"},
	}

	ctx := context.Background()
	err := messageBus.Send(ctx, message)

	// Wait for message processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)

	// Verify message was stored
	messages, err := messageStore.Retrieve(ctx, func(m Message) bool {
		return m.Type == "test_message"
	})
	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, message.Type, messages[0].Type)
}

func TestMessageBusFilter(t *testing.T) {
	messageBus := NewMessageBus(MessageBusConfig{})

	var receivedMessage Message
	handler := func(ctx context.Context, message Message) error {
		receivedMessage = message
		return nil
	}

	// Add filter that only allows high priority messages
	filter := func(message Message) bool {
		return message.Priority >= EventPriorityHigh
	}

	messageBus.AddFilter("test_message", filter)
	messageBus.Subscribe("test_message", handler)

	// Send low priority message
	lowPriorityMessage := Message{
		Type:     "test_message",
		Source:   "test_module",
		Target:   "target_module",
		Priority: EventPriorityLow,
	}

	ctx := context.Background()
	err := messageBus.Send(ctx, lowPriorityMessage)

	// Wait for message processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Empty(t, receivedMessage.ID) // Message should be filtered out

	// Send high priority message
	highPriorityMessage := Message{
		Type:     "test_message",
		Source:   "test_module",
		Target:   "target_module",
		Priority: EventPriorityHigh,
	}

	err = messageBus.Send(ctx, highPriorityMessage)

	// Wait for message processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, highPriorityMessage.Type, receivedMessage.Type) // Message should be processed
}

func TestMessageBusTTL(t *testing.T) {
	config := MessageBusConfig{
		MessageTTL: 10 * time.Millisecond,
	}
	messageBus := NewMessageBus(config)

	var receivedMessage Message
	handler := func(ctx context.Context, message Message) error {
		receivedMessage = message
		return nil
	}

	messageBus.Subscribe("test_message", handler)

	message := Message{
		Type:      "test_message",
		Source:    "test_module",
		Target:    "target_module",
		Priority:  EventPriorityNormal,
		Timestamp: time.Now().Add(-20 * time.Millisecond), // Expired message
	}

	ctx := context.Background()
	err := messageBus.Send(ctx, message)

	// Wait for message processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Empty(t, receivedMessage.ID) // Message should be expired
}

func TestMessageBusRetry(t *testing.T) {
	messageBus := NewMessageBus(MessageBusConfig{
		RetryAttempts: 2,
		RetryDelay:    10 * time.Millisecond,
	})

	attempts := 0
	handler := func(ctx context.Context, message Message) error {
		attempts++
		if attempts < 3 {
			return assert.AnError
		}
		return nil
	}

	messageBus.Subscribe("test_message", handler)

	message := Message{
		Type:     "test_message",
		Source:   "test_module",
		Target:   "target_module",
		Priority: EventPriorityNormal,
	}

	ctx := context.Background()
	err := messageBus.Send(ctx, message)

	// Wait for message processing and retries
	time.Sleep(200 * time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts) // Should have retried twice
}

func TestMessageBusClose(t *testing.T) {
	messageBus := NewMessageBus(MessageBusConfig{})

	err := messageBus.Close()
	assert.NoError(t, err)
}

func TestNewCommunicationManager(t *testing.T) {
	eventConfig := EventBusConfig{
		BufferSize:    100,
		WorkerCount:   5,
		EnableTracing: true,
	}

	messageConfig := MessageBusConfig{
		BufferSize:    100,
		WorkerCount:   5,
		EnableTracing: true,
	}

	cm := NewCommunicationManager(eventConfig, messageConfig)

	assert.NotNil(t, cm)
	assert.NotNil(t, cm.eventBus)
	assert.NotNil(t, cm.messageBus)
	assert.NotNil(t, cm.tracer)
}

func TestCommunicationManagerPublishEvent(t *testing.T) {
	cm := NewCommunicationManager(EventBusConfig{}, MessageBusConfig{})

	var receivedEvent Event
	handler := func(ctx context.Context, event Event) error {
		receivedEvent = event
		return nil
	}

	cm.SubscribeToEvent(EventTypeModuleStarted, handler)

	event := Event{
		Type:     EventTypeModuleStarted,
		Source:   "test_module",
		Priority: EventPriorityNormal,
	}

	ctx := context.Background()
	err := cm.PublishEvent(ctx, event)

	// Wait for event processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, event.Type, receivedEvent.Type)
}

func TestCommunicationManagerSendMessage(t *testing.T) {
	cm := NewCommunicationManager(EventBusConfig{}, MessageBusConfig{})

	var receivedMessage Message
	handler := func(ctx context.Context, message Message) error {
		receivedMessage = message
		return nil
	}

	cm.SubscribeToMessage("test_message", handler)

	message := Message{
		Type:     "test_message",
		Source:   "test_module",
		Target:   "target_module",
		Priority: EventPriorityNormal,
	}

	ctx := context.Background()
	err := cm.SendMessage(ctx, message)

	// Wait for message processing
	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	assert.Equal(t, message.Type, receivedMessage.Type)
}

func TestCommunicationManagerClose(t *testing.T) {
	cm := NewCommunicationManager(EventBusConfig{}, MessageBusConfig{})

	err := cm.Close()
	assert.NoError(t, err)
}

func TestEventConcurrency(t *testing.T) {
	eventBus := NewEventBus(EventBusConfig{
		BufferSize:  1000,
		WorkerCount: 10,
	})

	var mu sync.Mutex
	receivedEvents := make([]Event, 0)

	handler := func(ctx context.Context, event Event) error {
		mu.Lock()
		defer mu.Unlock()
		receivedEvents = append(receivedEvents, event)
		return nil
	}

	eventBus.Subscribe(EventTypeModuleStarted, handler)

	// Publish multiple events concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := Event{
				Type:     EventTypeModuleStarted,
				Source:   "test_module",
				Priority: EventPriorityNormal,
				Data:     map[string]interface{}{"id": id},
			}

			ctx := context.Background()
			eventBus.Publish(ctx, event)
		}(i)
	}

	wg.Wait()

	// Wait for event processing
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	eventCount := len(receivedEvents)
	mu.Unlock()

	assert.Equal(t, 100, eventCount)
}

func TestMessageConcurrency(t *testing.T) {
	messageBus := NewMessageBus(MessageBusConfig{
		BufferSize:  1000,
		WorkerCount: 10,
	})

	var mu sync.Mutex
	receivedMessages := make([]Message, 0)

	handler := func(ctx context.Context, message Message) error {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, message)
		return nil
	}

	messageBus.Subscribe("test_message", handler)

	// Send multiple messages concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			message := Message{
				Type:     "test_message",
				Source:   "test_module",
				Target:   "target_module",
				Priority: EventPriorityNormal,
				Payload:  map[string]interface{}{"id": id},
			}

			ctx := context.Background()
			messageBus.Send(ctx, message)
		}(i)
	}

	wg.Wait()

	// Wait for message processing
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	messageCount := len(receivedMessages)
	mu.Unlock()

	assert.Equal(t, 100, messageCount)
}
