package statemachine

import (
	"testing"
)

type UserState string

const (
	UserStateInitial                  UserState = "Initial"
	UserStateEmailPendingVerification UserState = "EmailPendingVerification"
	UserStateEmailVerified            UserState = "EmailVerified"
	UserStateSignUpComplete           UserState = "SignUpComplete"
	UserStateRejected                 UserState = "SignupRejected"
)

// String implements the State interface
func (s UserState) String() string {
	return string(s)
}

type UserEvent string

const (
	UserEventSubmitSignUp          UserEvent = "SubmitSignup"
	UserEventClickVerificationLink UserEvent = "ClickVerificationLink"
	UserEventSignupFailed          UserEvent = "SignUpFailed"
	UserEventCompleteProfile       UserEvent = "CompleteProfile"
)

// String implements the Event interface
func (e UserEvent) String() string {
	return string(e)
}

type UserStateMachine = StateMachine[UserState, UserEvent]

func NewUserStateMachine() *UserStateMachine {
	sm := NewStateMachine[UserState, UserEvent]()

	// Define all valid transitions using the generic AddTransitions method
	sm.AddTransitions([]Transition[UserState, UserEvent]{
		// From Initial state
		{UserStateInitial, UserEventSubmitSignUp, UserStateEmailPendingVerification},
		{UserStateInitial, UserEventSignupFailed, UserStateRejected},

		// From EmailPendingVerification state
		{UserStateEmailPendingVerification, UserEventClickVerificationLink, UserStateEmailVerified},
		{UserStateEmailPendingVerification, UserEventSignupFailed, UserStateRejected},

		// From EmailVerified state
		{UserStateEmailVerified, UserEventCompleteProfile, UserStateSignUpComplete},
		{UserStateEmailVerified, UserEventSignupFailed, UserStateRejected},
	})

	return sm
}

// ==================== TESTS FOR USER STATE MACHINE ====================

func TestGenericUserStateMachine_ValidTransitions(t *testing.T) {
	sm := NewUserStateMachine()

	tests := []struct {
		name      string
		from      UserState
		event     UserEvent
		wantState UserState
		wantErr   bool
	}{
		{
			name:      "Initial to EmailPending via SubmitSignUp",
			from:      UserStateInitial,
			event:     UserEventSubmitSignUp,
			wantState: UserStateEmailPendingVerification,
			wantErr:   false,
		},
		{
			name:      "EmailPending to EmailVerified via ClickLink",
			from:      UserStateEmailPendingVerification,
			event:     UserEventClickVerificationLink,
			wantState: UserStateEmailVerified,
			wantErr:   false,
		},
		{
			name:      "EmailVerified to Complete via CompleteProfile",
			from:      UserStateEmailVerified,
			event:     UserEventCompleteProfile,
			wantState: UserStateSignUpComplete,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotState, err := sm.Transition(tt.from, tt.event)

			if (err != nil) != tt.wantErr {
				t.Errorf("Transition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotState != tt.wantState {
				t.Errorf("Transition() gotState = %v, want %v", gotState, tt.wantState)
			}
		})
	}
}

func TestGenericUserStateMachine_InvalidTransitions(t *testing.T) {
	sm := NewUserStateMachine()

	tests := []struct {
		name  string
		from  UserState
		event UserEvent
	}{
		{
			name:  "Cannot skip to Complete from Initial",
			from:  UserStateInitial,
			event: UserEventCompleteProfile,
		},
		{
			name:  "Cannot verify from Complete state",
			from:  UserStateSignUpComplete,
			event: UserEventClickVerificationLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sm.Transition(tt.from, tt.event)

			if err == nil {
				t.Errorf("Transition() expected error for invalid transition")
			}
		})
	}
}

// ==================== GENERIC FUNCTIONALITY TESTS ====================

func TestGenericStateMachine_GetAllStates(t *testing.T) {
	sm := NewUserStateMachine()

	states := sm.GetAllStates()

	// Should have 5 states
	expectedCount := 5
	if len(states) != expectedCount {
		t.Errorf("GetAllStates() returned %d states, want %d", len(states), expectedCount)
	}

	// Check that all expected states are present
	stateMap := make(map[UserState]bool)
	for _, s := range states {
		stateMap[s] = true
	}

	expectedStates := []UserState{
		UserStateInitial,
		UserStateEmailPendingVerification,
		UserStateEmailVerified,
		UserStateSignUpComplete,
		UserStateRejected,
	}

	for _, expected := range expectedStates {
		if !stateMap[expected] {
			t.Errorf("GetAllStates() missing expected state %v", expected)
		}
	}
}

func TestGenericStateMachine_GetTransitions(t *testing.T) {
	sm := NewUserStateMachine()

	// Test getting transitions from Initial state
	transitions := sm.GetTransitions(UserStateInitial)

	if len(transitions) != 2 {
		t.Errorf("GetTransitions(Initial) returned %d transitions, want 2", len(transitions))
	}

	// Check specific transitions
	if transitions[UserEventSubmitSignUp] != UserStateEmailPendingVerification {
		t.Errorf("Expected SubmitSignUp to lead to EmailPendingVerification")
	}

	if transitions[UserEventSignupFailed] != UserStateRejected {
		t.Errorf("Expected SignupFailed to lead to Rejected")
	}
}
