package example

import (
	"fmt"

	ss "github.com/richardbowden/statemachine"
)

type DocumentState string

const (
	DocumentStateDraft     DocumentState = "Draft"
	DocumentStateSubmitted DocumentState = "Submitted"
	DocumentStateReviewing DocumentState = "Reviewing"
	DocumentStateApproved  DocumentState = "Approved"
	DocumentStateRejected  DocumentState = "Rejected"
	DocumentStatePublished DocumentState = "Published"
	DocumentStateArchived  DocumentState = "Archived"
)

func (s DocumentState) String() string {
	return string(s)
}

type DocumentEvent string

const (
	DocumentEventSubmit  DocumentEvent = "Submit"
	DocumentEventReview  DocumentEvent = "Review"
	DocumentEventApprove DocumentEvent = "Approve"
	DocumentEventReject  DocumentEvent = "Reject"
	DocumentEventPublish DocumentEvent = "Publish"
	DocumentEventArchive DocumentEvent = "Archive"
	DocumentEventRevise  DocumentEvent = "Revise"
)

func (e DocumentEvent) String() string {
	return string(e)
}

func NewDocumentStateMachine() *ss.StateMachine[DocumentState, DocumentEvent] {
	sm := ss.NewStateMachine[DocumentState, DocumentEvent]()

	sm.AddTransitions([]ss.Transition[DocumentState, DocumentEvent]{

		{From: DocumentStateDraft, Event: DocumentEventSubmit, To: DocumentStateSubmitted},
		{From: DocumentStateSubmitted, Event: DocumentEventReview, To: DocumentStateReviewing},

		{From: DocumentStateReviewing, Event: DocumentEventApprove, To: DocumentStateApproved},
		{From: DocumentStateApproved, Event: DocumentEventPublish, To: DocumentStatePublished},

		{From: DocumentStateReviewing, Event: DocumentEventReject, To: DocumentStateRejected},
		{From: DocumentStateRejected, Event: DocumentEventRevise, To: DocumentStateDraft},

		{From: DocumentStatePublished, Event: DocumentEventArchive, To: DocumentStateArchived},
	})

	return sm
}

type Document struct {
	ID       int64
	Title    string
	State    DocumentState
	AuthorID int64
}

type DocumentService struct {
	stateMachine *ss.StateMachine[DocumentState, DocumentEvent]
}

func NewDocumentService() *DocumentService {
	return &DocumentService{
		stateMachine: NewDocumentStateMachine(),
	}
}

func (ds *DocumentService) CanUserApprove(doc *Document) bool {
	return ds.stateMachine.CanTransition(doc.State, DocumentEventApprove)
}

func (ds *DocumentService) GetAvailableActions(doc *Document) []DocumentEvent {
	return ds.stateMachine.GetValidEvents(doc.State)
}

func ExampleDocumentStateMachine() {
	ds := NewDocumentService()

	doc := &Document{
		ID:       1,
		Title:    "Q4 Strategy",
		State:    DocumentStateDraft,
		AuthorID: 456,
	}

	actions := ds.GetAvailableActions(doc)
	fmt.Printf("Available actions: %v\n", actions) // [Submit]

	path := []DocumentEvent{
		DocumentEventSubmit,
		DocumentEventReview,
		DocumentEventApprove,
		DocumentEventPublish,
	}

	finalState, err := ds.stateMachine.ValidateTransitionPath(doc.State, path)
	if err != nil {
		fmt.Printf("Invalid workflow: %v\n", err)
	} else {
		fmt.Printf("Valid workflow! Final state: %s\n", finalState) // Published
	}

	isTerminal := ds.stateMachine.IsTerminalState(DocumentStateArchived)
	fmt.Printf("Is Archived terminal? %v\n", isTerminal) // true
}
