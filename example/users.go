package example

import (
	"context"
	"fmt"
	"time"

	ss "github.com/richardbowden/statemachine"
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

type UserStateMachine = ss.StateMachine[UserState, UserEvent]

func NewUserStateMachine() *UserStateMachine {
	sm := ss.NewStateMachine[UserState, UserEvent]()

	// Define all valid transitions using the generic AddTransitions method
	sm.AddTransitions([]ss.Transition[UserState, UserEvent]{
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

type NewUser struct {
	FirstName      string `json:"first_name" validate:"required,minlen=2"`
	MiddleName     string
	Surname        string
	Username       string
	EMail          string
	State          UserState
	HashedPassword string
}

type User struct {
	ID             int64
	FirstName      string
	MiddleName     string
	Surname        string
	Username       string
	EMail          string
	State          UserState
	HashedPassword string
	Enabled        bool
	CreatedOn      time.Time
	UpdatedAt      time.Time
}

type EmailAddress struct {
	Id         int
	Email      string
	Verified   bool
	VerifiedOn time.Time
	UpdatedOn  time.Time
}

type EmailAddresses []EmailAddress

type UserRepository interface {
	Create(ctx context.Context, params NewUser) (User, error)
	DoesUserExist(ctx context.Context, email string, username string) (bool, bool, error)
	GetByID(ctx context.Context, id int64) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	UpdateState(ctx context.Context, userID int64, newState UserState) error
}

type NewUserRequest struct {
	Email      string `json:"email" validate:"required,email"`
	FirstName  string `json:"first_name" validate:"required,minlen=2"`
	MiddleName string `json:"middle_name,omitempty,omitzero"`
	Surname    string `json:"surname,omitempty,omitzero"`
	Username   string `json:"username"`
	Password1  string `json:"pwd1" validate:"required"`
	Password2  string `json:"pwd2" validate:"eqfieldsecure=Password1"`
}

type Authz struct{}

type UserService struct {
	repo         UserRepository
	ac           *Authz
	stateMachine *UserStateMachine
}

func NewUserService(repo UserRepository, authz *Authz) (*UserService, error) {
	us := &UserService{
		repo:         repo,
		ac:           authz,
		stateMachine: NewUserStateMachine(),
	}
	return us, nil
}

// ProcessEvent handles state transitions for users
func (us *UserService) ProcessEvent(ctx context.Context, userID int64, event UserEvent) error {
	user, err := us.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	newState, err := us.stateMachine.Transition(user.State, event)
	if err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	err = us.repo.UpdateState(ctx, userID, newState)
	if err != nil {
		return fmt.Errorf("failed to update user state: %w", err)
	}

	return nil
}

// CanProcessEvent checks if an event can be processed for a user
func (us *UserService) CanProcessEvent(ctx context.Context, userID int64, event UserEvent) (bool, error) {
	user, err := us.repo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	return us.stateMachine.CanTransition(user.State, event), nil
}

// GetValidEvents returns valid events for a user's current state
func (us *UserService) GetValidEvents(ctx context.Context, userID int64) ([]UserEvent, error) {
	user, err := us.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return us.stateMachine.GetValidEvents(user.State), nil
}

// VerifyEmail processes email verification event
func (us *UserService) VerifyEmail(ctx context.Context, userID int64) error {
	return us.ProcessEvent(ctx, userID, UserEventClickVerificationLink)
}

// CompleteProfile marks profile as complete
func (us *UserService) CompleteProfile(ctx context.Context, userID int64) error {
	return us.ProcessEvent(ctx, userID, UserEventCompleteProfile)
}

// RejectSignup marks signup as rejected/failed
func (us *UserService) RejectSignup(ctx context.Context, userID int64) error {
	return us.ProcessEvent(ctx, userID, UserEventSignupFailed)
}
