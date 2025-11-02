// Package event. filter provides event filtering capabilities for conditional handler execution.
package event

import "context"

// Filter determines if an event should be processed by a handler.
type Filter func(*Event) bool

// FilteredHandler wraps a handler with a filter condition.
type FilteredHandler struct {
	filter  Filter
	handler EventHandler
}

// -------------------------------------------- Public Functions --------------------------------------------

// NewFilteredHandler creates a handler that only executes when the filter returns true.
func NewFilteredHandler(filter Filter, handler EventHandler) EventHandler {
	return func(ctx context.Context, event *Event) error {
		if filter(event) {
			return handler(ctx, event)
		}
		return nil
	}
}

// FilterByMetadata creates a filter that checks for specific metadata key-value pairs.
func FilterByMetadata(key string, value interface{}) Filter {
	return func(event *Event) bool {
		if v, exists := event.Metadata[key]; exists {
			return v == value
		}
		return false
	}
}

// FilterByMetadataExists creates a filter that checks if metadata key exists.
func FilterByMetadataExists(key string) Filter {
	return func(event *Event) bool {
		_, exists := event.Metadata[key]
		return exists
	}
}

// And combines multiple filters with AND logic.
func And(filters ...Filter) Filter {
	return func(event *Event) bool {
		for _, f := range filters {
			if !f(event) {
				return false
			}
		}
		return true
	}
}

// Or combines multiple filters with OR logic.
func Or(filters ...Filter) Filter {
	return func(event *Event) bool {
		for _, f := range filters {
			if f(event) {
				return true
			}
		}
		return false
	}
}

// Not negates a filter.
func Not(filter Filter) Filter {
	return func(event *Event) bool {
		return !filter(event)
	}
}
