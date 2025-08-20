package event

import (
	"reflect"
	"sync"

	"github.com/gr4vediggr/stellarlight/pkg/messages"
)

type EventHandler func(event interface{})

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[reflect.Type][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[reflect.Type][]EventHandler),
	}
}

func (eb *EventBus) Subscribe(eventType interface{}, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	typ := reflect.TypeOf(eventType)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	eb.subscribers[typ] = append(eb.subscribers[typ], handler)
}

func (eb *EventBus) SubscribeToMultiple(eventTypes []interface{}, handler EventHandler) {
	for _, eventType := range eventTypes {
		eb.Subscribe(eventType, handler)
	}
}

func (eb *EventBus) Publish(event interface{}) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	typ := reflect.TypeOf(event)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if handlers, ok := eb.subscribers[typ]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

func (eb *EventBus) PublishGameMessage(msg *messages.GameMessage) {
	eb.Publish(msg)

	switch content := msg.Content.(type) {
	case *messages.GameMessage_GameState:
		eb.Publish(content.GameState)
	case *messages.GameMessage_GameEvent:
		eb.Publish(content.GameEvent)
	case *messages.GameMessage_TurnUpdate:
		eb.Publish(content.TurnUpdate)
	default:
		// Handle unknown message type
	}
}
