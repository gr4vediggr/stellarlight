package events

import (
	"reflect"

	"github.com/google/uuid"
)

type EventBus struct {
	subscribers map[string][]func(GameEvent)
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(GameEvent)),
	}
}

func (eb *EventBus) Subscribe(eventType string, handler func(GameEvent)) func() {
	eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
	return func() {
		for i, h := range eb.subscribers[eventType] {
			// Compare function pointers using reflect
			if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
				eb.subscribers[eventType] = append(eb.subscribers[eventType][:i], eb.subscribers[eventType][i+1:]...)
				break
			}
		}
	}
}

func (eb *EventBus) Publish(event GameEvent) {
	if handlers, ok := eb.subscribers[event.GetType()]; ok {
		for _, handler := range handlers {
			handler(event)
		}
	}
}

// GameEvent interface for all game events
type GameEvent interface {
	GetSessionID() uuid.UUID
	GetType() string
	GetTimestamp() int64
}
