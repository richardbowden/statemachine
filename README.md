# State Machine

A type-safe, generic state machine library for Go.

## What is a State Machine?

A state machine models how something transitions between different states based on events. For example, a user signup might flow: `Initial → EmailPending → EmailVerified → Complete`.

State machines enforce **valid transitions**. You can't jump from `Initial` to `Complete` - you must go through the proper steps.

## Why Use a State Machine?

**Without:**
```go
user.State = "Complete"  // Oops, skipped email verification
```

**With:**
```go
newState, err := sm.Transition(user.State, UserEventComplete)
// Error: invalid transition
```

State machines prevent bugs by:
- Enforcing valid state transitions
- Making workflows explicit and documented in code
- Catching logic errors at runtime (or test time)
- Providing a clear audit trail of how states change

## When to Use Them

Use state machines when:
- Order processing (pending → processing → shipped → delivered)
- User onboarding flows
- Document approval workflows
- Job/task lifecycle management
- Any process with distinct stages and rules about progression

Don't use them for:
- Simple boolean flags (`active`/`inactive`)
- States with no transition rules
- Overly complex workflows (consider a workflow engine instead)

## Installation

```bash
go get github.com/richardbowden/statemachine
```

## Usage

### 1. Define States and Events

```go
type OrderState string

const (
    OrderStatePending    OrderState = "Pending"
    OrderStateProcessing OrderState = "Processing"
    OrderStateShipped    OrderState = "Shipped"
)

func (s OrderState) String() string {
    return string(s)
}

type OrderEvent string

const (
    OrderEventConfirm OrderEvent = "Confirm"
    OrderEventShip    OrderEvent = "Ship"
)

func (e OrderEvent) String() string {
    return string(e)
}
```

### 2. Create Your State Machine

```go
func NewOrderStateMachine() *StateMachine[OrderState, OrderEvent] {
    sm := NewStateMachine[OrderState, OrderEvent]()
    
    sm.AddTransitions([]Transition[OrderState, OrderEvent]{
        {From: OrderStatePending, Event: OrderEventConfirm, To: OrderStateProcessing},
        {From: OrderStateProcessing, Event:  OrderEventShip, To: OrderStateShipped},
    })
    
    return sm
}
```

### 3. Use It

```go
sm := NewOrderStateMachine()

// Execute a transition
newState, err := sm.Transition(OrderStatePending, OrderEventConfirm)
if err != nil {
    // Invalid transition
}
// newState = OrderStateProcessing

// Check before transitioning
if sm.CanTransition(order.State, OrderEventShip) {
    // Safe to ship
}

// Get available actions
events := sm.GetValidEvents(order.State)
// Returns all valid events from current state

// Validate a workflow path
finalState, err := sm.ValidateTransitionPath(
    OrderStatePending,
    []OrderEvent{OrderEventConfirm, OrderEventShip},
)
// finalState = OrderStateShipped
```

## API

| Method | Description |
|--------|-------------|
| `Transition(from, event)` | Execute a state transition, returns new state or error |
| `CanTransition(from, event)` | Check if transition is valid without executing |
| `GetValidEvents(from)` | Get all valid events for a state |
| `ValidateTransitionPath(start, events)` | Validate a sequence of transitions |
| `IsTerminalState(state)` | Check if state has no outgoing transitions |
| `GetAllStates()` | Get all registered states |
| `GetTransitions(from)` | Get all transitions from a state |

## Integration Example

```go
type OrderService struct {
    repo         OrderRepository
    stateMachine *StateMachine[OrderState, OrderEvent]
}

func NewOrderService(repo OrderRepository) *OrderService {
    return &OrderService{
        repo:         repo,
        stateMachine: NewOrderStateMachine(),
    }
}

func (s *OrderService) ShipOrder(ctx context.Context, orderID int64) error {
    order, err := s.repo.GetByID(ctx, orderID)
    if err != nil {
        return err
    }

    newState, err := s.stateMachine.Transition(order.State, OrderEventShip)
    if err != nil {
        return fmt.Errorf("cannot ship order: %w", err)
    }

    return s.repo.UpdateState(ctx, orderID, newState)
}
```

## Database Storage

Store state as a string column:

```sql
CREATE TABLE orders (
    id    BIGSERIAL PRIMARY KEY,
    state VARCHAR(50) NOT NULL DEFAULT 'Pending',
    -- other fields...
);

CREATE INDEX idx_orders_state ON orders(state);
```

## Testing

```go
func TestOrderFlow(t *testing.T) {
    sm := NewOrderStateMachine()
    
    // Test valid transition
    state, err := sm.Transition(OrderStatePending, OrderEventConfirm)
    assert.NoError(t, err)
    assert.Equal(t, OrderStateProcessing, state)
    
    // Test invalid transition
    _, err = sm.Transition(OrderStatePending, OrderEventShip)
    assert.Error(t, err)
}
```

## License

MIT
