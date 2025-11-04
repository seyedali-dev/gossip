// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package main. main demonstrates e-commerce order processing with event-driven batch processing.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	gossip "github.com/seyedali-dev/gossip/event"
)

// E-commerce event types.
const (
	OrderCreated   gossip.EventType = "order.created"
	OrderPaid      gossip.EventType = "order.paid"
	OrderShipped   gossip.EventType = "order.shipped"
	OrderDelivered gossip.EventType = "order.delivered"
	OrderCancelled gossip.EventType = "order.cancelled"
)

// -------------------------------------------- Event Data Structs --------------------------------------------

// OrderData contains order information.
type OrderData struct {
	OrderID    string
	CustomerID string
	Amount     float64
	Items      []string
}

// -------------------------------------------- Handlers --------------------------------------------

// inventoryHandler updates inventory when orders are created.
func inventoryHandler(ctx context.Context, event *gossip.Event) error {
	data := event.Data.(*OrderData)
	log.Printf("[Inventory] Reserving items for order %s: %v", data.OrderID, data.Items)
	return nil
}

// paymentHandler processes payments.
func paymentHandler(ctx context.Context, event *gossip.Event) error {
	data := event.Data.(*OrderData)
	log.Printf("[Payment] Processing payment of $%.2f for order %s", data.Amount, data.OrderID)
	return nil
}

// shippingHandler handles shipping logistics.
func shippingHandler(ctx context.Context, event *gossip.Event) error {
	data := event.Data.(*OrderData)
	log.Printf("[Shipping] Creating shipping label for order %s", data.OrderID)
	return nil
}

// batchEmailHandler processes multiple orders in a batch for efficiency.
func batchEmailHandler(ctx context.Context, events []*gossip.Event) error {
	log.Printf("[Email Batch] Processing %d order notifications", len(events))

	for _, event := range events {
		data := event.Data.(*OrderData)
		log.Printf("[Email] Sending confirmation for order %s to customer %s", data.OrderID, data.CustomerID)
	}

	return nil
}

// analyticsHandler tracks order metrics with filtering.
func analyticsHandler(ctx context.Context, event *gossip.Event) error {
	data := event.Data.(*OrderData)
	log.Printf("[Analytics] Recording order value: $%.2f", data.Amount)
	return nil
}

// -------------------------------------------- Service Logic --------------------------------------------

// OrderService handles order processing.
type OrderService struct {
	bus            *gossip.EventBus
	batchProcessor *gossip.BatchProcessor
}

// NewOrderService creates a new order service.
func NewOrderService(bus *gossip.EventBus, bp *gossip.BatchProcessor) *OrderService {
	return &OrderService{
		bus:            bus,
		batchProcessor: bp,
	}
}

// CreateOrder creates a new order and publishes an event.
func (s *OrderService) CreateOrder(customerID string, items []string, amount float64) error {
	orderID := fmt.Sprintf("order_%d", time.Now().Unix())

	log.Printf("[Service] Creating order %s for customer %s", orderID, customerID)

	eventData := &OrderData{
		OrderID:    orderID,
		CustomerID: customerID,
		Amount:     amount,
		Items:      items,
	}

	event := gossip.NewEvent(OrderCreated, eventData).
		WithMetadata("source", "web").
		WithMetadata("priority", "high")

	s.bus.Publish(event)

	return nil
}

// -------------------------------------------- Main --------------------------------------------

func main() {
	// Initialize event bus
	config := &gossip.Config{
		Workers:    15,
		BufferSize: 2000,
	}
	bus := gossip.NewEventBus(config)
	defer bus.Shutdown()

	// Create batch processor for email notifications
	batchConfig := gossip.BatchConfig{
		BatchSize:   10,
		FlushPeriod: 5 * time.Second,
	}
	batchProcessor := gossip.NewBatchProcessor(OrderCreated, batchConfig, batchEmailHandler)
	defer batchProcessor.Shutdown()

	// Register handlers with various patterns

	// Standard handlers
	bus.Subscribe(OrderCreated, gossip.Chain(
		gossip.WithRetry(3, 100*time.Millisecond),
		gossip.WithTimeout(5*time.Second),
	)(inventoryHandler))

	bus.Subscribe(OrderPaid, paymentHandler)
	bus.Subscribe(OrderShipped, shippingHandler)

	// Batch handler for emails
	bus.Subscribe(OrderCreated, batchProcessor.AsEventHandler())

	// Filtered handler - only track high-value orders
	highValueFilter := func(event *gossip.Event) bool {
		data := event.Data.(*OrderData)
		return data.Amount > 100.0
	}
	bus.Subscribe(OrderCreated, gossip.NewFilteredHandler(highValueFilter, analyticsHandler))

	// Create service
	orderService := NewOrderService(bus, batchProcessor)

	// Simulate order activities
	log.Println("=== Starting E-commerce Demo ===")

	orderService.CreateOrder("customer_1", []string{"laptop", "mouse"}, 1299.99)
	time.Sleep(100 * time.Millisecond)

	orderService.CreateOrder("customer_2", []string{"keyboard"}, 79.99)
	time.Sleep(100 * time.Millisecond)

	orderService.CreateOrder("customer_3", []string{"monitor"}, 449.99)
	time.Sleep(100 * time.Millisecond)

	// Simulate multiple orders to trigger batch processing
	for i := 0; i < 5; i++ {
		orderService.CreateOrder(fmt.Sprintf("customer_%d", i+4), []string{"item"}, 50.0)
	}

	log.Println("=== Demo Complete - Waiting for batch flush ===")
	time.Sleep(6 * time.Second)
}
