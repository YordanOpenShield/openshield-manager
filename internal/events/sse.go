package events

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// EventType represents the type of SSE event
type EventType string

const (
	// Agent events
	AgentConnected    EventType = "AGENT_CONNECTED"
	AgentDisconnected EventType = "AGENT_DISCONNECTED"
	AgentUpdated      EventType = "AGENT_UPDATED"

	// Task events
	TaskCreated   EventType = "TASK_CREATED"
	TaskStarted   EventType = "TASK_STARTED"
	TaskCompleted EventType = "TASK_COMPLETED"
	TaskFailed    EventType = "TASK_FAILED"

	// Query events
	QueryStarted   EventType = "QUERY_STARTED"
	QueryCompleted EventType = "QUERY_COMPLETED"
	QueryFailed    EventType = "QUERY_FAILED"

	// Bulk operation events
	BulkOpStarted   EventType = "BULK_OP_STARTED"
	BulkOpProgress  EventType = "BULK_OP_PROGRESS"
	BulkOpCompleted EventType = "BULK_OP_COMPLETED"
)

// Event represents an SSE event
type Event struct {
	Type      EventType   `json:"type"`
	ID        string      `json:"id,omitempty"`
	AgentID   string      `json:"agent_id,omitempty"`
	TaskID    string      `json:"task_id,omitempty"`
	QueryID   string      `json:"query_id,omitempty"`
	BulkOpID  string      `json:"bulk_op_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Changes   interface{} `json:"changes,omitempty"`
	Progress  interface{} `json:"progress,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a connected SSE client
type Client struct {
	ID     string
	Events chan Event
	Done   chan struct{}
}

// Hub manages SSE clients and broadcasts events
type Hub struct {
	clients    map[string]*Client
	clientsMux sync.RWMutex
	broadcast  chan Event
}

// Global hub instance
var globalHub = NewHub()

// NewHub creates a new SSE hub
func NewHub() *Hub {
	return &Hub{
		clients:   make(map[string]*Client),
		broadcast: make(chan Event, 100),
	}
}

// Start begins the event broadcasting loop
func (h *Hub) Start() {
	go func() {
		for event := range h.broadcast {
			h.clientsMux.RLock()
			for _, client := range h.clients {
				select {
				case client.Events <- event:
				case <-time.After(100 * time.Millisecond):
					// Client is slow, skip this event
				}
			}
			h.clientsMux.RUnlock()
		}
	}()
}

// Subscribe adds a new client to the hub
func (h *Hub) Subscribe(clientID string) *Client {
	client := &Client{
		ID:     clientID,
		Events: make(chan Event, 10),
		Done:   make(chan struct{}),
	}

	h.clientsMux.Lock()
	h.clients[clientID] = client
	h.clientsMux.Unlock()

	return client
}

// Unsubscribe removes a client from the hub
func (h *Hub) Unsubscribe(clientID string) {
	h.clientsMux.Lock()
	if client, exists := h.clients[clientID]; exists {
		close(client.Done)
		close(client.Events)
		delete(h.clients, clientID)
	}
	h.clientsMux.Unlock()
}

// Publish sends an event to all connected clients
func (h *Hub) Publish(event Event) {
	event.Timestamp = time.Now()
	select {
	case h.broadcast <- event:
	default:
		// Broadcast channel full, drop event
	}
}

// HandleSSE handles SSE connections
func HandleSSE(c *gin.Context) {
	clientID := c.GetString("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	client := globalHub.Subscribe(clientID)
	defer globalHub.Unsubscribe(clientID)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-client.Events:
			if !ok {
				return false
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: %s\n", event.Type)
			fmt.Fprintf(w, "data: %s\n\n", data)
			return true
		case <-client.Done:
			return false
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// Global publish functions for convenience

func PublishAgentConnected(agentID string, agent interface{}) {
	globalHub.Publish(Event{
		Type:    AgentConnected,
		AgentID: agentID,
		Data:    agent,
	})
}

func PublishAgentDisconnected(agentID string, changes interface{}) {
	globalHub.Publish(Event{
		Type:    AgentDisconnected,
		AgentID: agentID,
		Changes: changes,
	})
}

func PublishAgentUpdated(agentID string, changes interface{}) {
	globalHub.Publish(Event{
		Type:    AgentUpdated,
		AgentID: agentID,
		Changes: changes,
	})
}

func PublishTaskCreated(taskID, agentID, jobID string, data interface{}) {
	globalHub.Publish(Event{
		Type:   TaskCreated,
		ID:     taskID,
		AgentID: agentID,
		Data:   data,
	})
}

func PublishTaskCompleted(taskID, agentID string, result interface{}) {
	globalHub.Publish(Event{
		Type:    TaskCompleted,
		ID:      taskID,
		AgentID: agentID,
		Data:    result,
	})
}

func PublishTaskFailed(taskID, agentID string, err error) {
	globalHub.Publish(Event{
		Type:    TaskFailed,
		ID:      taskID,
		AgentID: agentID,
		Data:    map[string]string{"error": err.Error()},
	})
}

func PublishQueryStarted(queryID string, targets interface{}) {
	globalHub.Publish(Event{
		Type:   QueryStarted,
		QueryID: queryID,
		Data:   targets,
	})
}

func PublishQueryCompleted(queryID string, results interface{}) {
	globalHub.Publish(Event{
		Type:    QueryCompleted,
		QueryID: queryID,
		Data:    results,
	})
}

func PublishBulkOpStarted(bulkOpID string, data interface{}) {
	globalHub.Publish(Event{
		Type:   BulkOpStarted,
		BulkOpID: bulkOpID,
		Data:   data,
	})
}

func PublishBulkOpProgress(bulkOpID string, progress interface{}) {
	globalHub.Publish(Event{
		Type:     BulkOpProgress,
		BulkOpID: bulkOpID,
		Progress: progress,
	})
}

func PublishBulkOpCompleted(bulkOpID string, results interface{}) {
	globalHub.Publish(Event{
		Type:     BulkOpCompleted,
		BulkOpID: bulkOpID,
		Data:     results,
	})
}

// Init starts the global event hub
func Init() {
	globalHub.Start()
}
