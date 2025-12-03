package statemachine

import "fmt"

// State is a constraint for types that can be used as states
type State interface {
	comparable
	String() string
}

// Event is a constraint for types that can be used as events
type Event interface {
	comparable
	String() string
}

// StateMachine is a generic state machine that works with any State and Event types
type StateMachine[S State, E Event] struct {
	transitions map[S]map[E]S
}

// NewStateMachine creates a new generic state machine
func NewStateMachine[S State, E Event]() *StateMachine[S, E] {
	return &StateMachine[S, E]{
		transitions: make(map[S]map[E]S),
	}
}

// AddTransition adds a valid transition to the state machine
func (sm *StateMachine[S, E]) AddTransition(from S, event E, to S) {
	if sm.transitions[from] == nil {
		sm.transitions[from] = make(map[E]S)
	}
	sm.transitions[from][event] = to
}

// AddTransitions adds multiple transitions at once
func (sm *StateMachine[S, E]) AddTransitions(transitions []Transition[S, E]) {
	for _, t := range transitions {
		sm.AddTransition(t.From, t.Event, t.To)
	}
}

// Transition represents a single transition rule
type Transition[S State, E Event] struct {
	From  S
	Event E
	To    S
}

// CanTransition checks if a transition is valid
func (sm *StateMachine[S, E]) CanTransition(from S, event E) bool {
	if transitions, exists := sm.transitions[from]; exists {
		_, allowed := transitions[event]
		return allowed
	}
	return false
}

// Transition attempts to transition from current state via event
// Returns the new state or an error if transition is invalid
func (sm *StateMachine[S, E]) Transition(from S, event E) (S, error) {
	if transitions, exists := sm.transitions[from]; exists {
		if newState, allowed := transitions[event]; allowed {
			return newState, nil
		}
	}
	var zero S
	return zero, fmt.Errorf("invalid transition: cannot process event '%s' from state '%s'", event.String(), from.String())
}

// GetValidEvents returns all valid events for a given state
func (sm *StateMachine[S, E]) GetValidEvents(from S) []E {
	events := []E{}
	if transitions, exists := sm.transitions[from]; exists {
		for event := range transitions {
			events = append(events, event)
		}
	}
	return events
}

// GetNextState returns the state that would result from an event, without validation
func (sm *StateMachine[S, E]) GetNextState(from S, event E) (S, bool) {
	if transitions, exists := sm.transitions[from]; exists {
		if newState, allowed := transitions[event]; allowed {
			return newState, true
		}
	}
	var zero S
	return zero, false
}

// IsTerminalState checks if a state is terminal (no outgoing transitions)
func (sm *StateMachine[S, E]) IsTerminalState(state S) bool {
	transitions, exists := sm.transitions[state]
	return !exists || len(transitions) == 0
}

// ValidateTransitionPath checks if a sequence of events is valid from a starting state
func (sm *StateMachine[S, E]) ValidateTransitionPath(start S, events []E) (S, error) {
	currentState := start
	for i, event := range events {
		newState, err := sm.Transition(currentState, event)
		if err != nil {
			return currentState, fmt.Errorf("invalid path at step %d: %w", i+1, err)
		}
		currentState = newState
	}
	return currentState, nil
}

// GetAllStates returns all states that have been registered in the state machine
func (sm *StateMachine[S, E]) GetAllStates() []S {
	states := make([]S, 0, len(sm.transitions))
	seen := make(map[S]bool)

	for from := range sm.transitions {
		if !seen[from] {
			states = append(states, from)
			seen[from] = true
		}
		for _, to := range sm.transitions[from] {
			if !seen[to] {
				states = append(states, to)
				seen[to] = true
			}
		}
	}

	return states
}

// GetTransitions returns all transitions from a given state
func (sm *StateMachine[S, E]) GetTransitions(from S) map[E]S {
	if transitions, exists := sm.transitions[from]; exists {
		// Return a copy to prevent external modification
		result := make(map[E]S, len(transitions))
		for k, v := range transitions {
			result[k] = v
		}
		return result
	}
	return make(map[E]S)
}
