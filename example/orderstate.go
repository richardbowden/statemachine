package example

import (
	"context"
	"fmt"

	ss "github.com/richardbowden/statemachine"
)

// ==================== EXAMPLE 1: ORDER STATE MACHINE ====================

type OrderState string

const (
	OrderStatePending    OrderState = "Pending"
	OrderStateProcessing OrderState = "Processing"
	OrderStateShipped    OrderState = "Shipped"
	OrderStateDelivered  OrderState = "Delivered"
	OrderStateCancelled  OrderState = "Cancelled"
	OrderStateRefunded   OrderState = "Refunded"
)

func (s OrderState) String() string {
	return string(s)
}

type OrderEvent string

const (
	OrderEventConfirm OrderEvent = "Confirm"
	OrderEventShip    OrderEvent = "Ship"
	OrderEventDeliver OrderEvent = "Deliver"
	OrderEventCancel  OrderEvent = "Cancel"
	OrderEventRefund  OrderEvent = "Refund"
)

func (e OrderEvent) String() string {
	return string(e)
}

// Order represents an order in your system
type Order struct {
	ID     int64
	State  OrderState
	Total  float64
	UserID int64
}

func NewOrderStateMachine() *ss.StateMachine[OrderState, OrderEvent] {
	sm := ss.NewStateMachine[OrderState, OrderEvent]()

	sm.AddTransitions([]ss.Transition[OrderState, OrderEvent]{
		// Happy path
		{From: OrderStatePending, Event: OrderEventConfirm, To: OrderStateProcessing},
		{From: OrderStateProcessing, Event: OrderEventShip, To: OrderStateShipped},
		{From: OrderStateShipped, Event: OrderEventDeliver, To: OrderStateDelivered},

		// Cancellation path
		{From: OrderStatePending, Event: OrderEventCancel, To: OrderStateCancelled},
		{From: OrderStateProcessing, Event: OrderEventCancel, To: OrderStateCancelled},

		// Refund path
		{From: OrderStateDelivered, Event: OrderEventRefund, To: OrderStateRefunded},
		{From: OrderStateCancelled, Event: OrderEventRefund, To: OrderStateRefunded},
	})

	return sm
}

type OrderService struct {
	stateMachine *ss.StateMachine[OrderState, OrderEvent]
}

func NewOrderService() *OrderService {
	return &OrderService{
		stateMachine: NewOrderStateMachine(),
	}
}

func (os *OrderService) TransitionOrder(ctx context.Context, order *Order, event OrderEvent) error {
	newState, err := os.stateMachine.Transition(order.State, event)
	if err != nil {
		return err
	}

	// Update order state
	order.State = newState

	// Here you would save to database
	// return os.repo.UpdateOrderState(ctx, order.ID, newState)

	return nil
}

func ExampleOrderStateMachine() {
	os := NewOrderService()

	order := &Order{
		ID:     1,
		State:  OrderStatePending,
		Total:  99.99,
		UserID: 123,
	}

	// Confirm order
	err := os.TransitionOrder(context.Background(), order, OrderEventConfirm)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Order state: %s\n", order.State) // Processing

	// Ship order
	err = os.TransitionOrder(context.Background(), order, OrderEventShip)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Order state: %s\n", order.State) // Shipped

	// Try to cancel (should fail - can't cancel shipped orders)
	err = os.TransitionOrder(context.Background(), order, OrderEventCancel)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Deliver order
	err = os.TransitionOrder(context.Background(), order, OrderEventDeliver)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("Order state: %s\n", order.State) // Delivered
}

// ==================== RUNNING ALL EXAMPLES ====================

func RunAllExamples() {
	fmt.Println("=== Order State Machine ===")
	ExampleOrderStateMachine()

	fmt.Println("\n=== Document State Machine ===")
	ExampleDocumentStateMachine()

}
