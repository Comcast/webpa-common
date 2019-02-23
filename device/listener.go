package device

import (
	"context"

	"github.com/Comcast/webpa-common/wrp"
)

// EventType is the type of device-related event
type EventType uint8

const (
	// Connect indicates a successful device connection.  After receipt of this event, the given
	// Device is able to receive requests.
	Connect EventType = iota

	// Disconnect indicates a device disconnection.  After receipt of this event, the given
	// Device can no longer receive requests.
	Disconnect

	// MessageSent indicates that a message was successfully dispatched to a device.
	MessageSent

	// MessageReceived indicates that a message has been successfully received and
	// dispatched to any goroutine waiting on it, as would be the case for a response.
	MessageReceived

	// MessageFailed indicates that a message could not be sent to a device, either because
	// of a communications error or due to the device disconnecting.  For each enqueued message
	// at the time of a device's disconnection, there will be (1) MessageFailed event.
	MessageFailed

	// TransactionComplete indicates that a response to a transaction has been received, and the
	// transaction completed successfully (at least as far as the routing infrastructure can tell).
	TransactionComplete

	// TransactionBroken indicates receipt of a message that had a transaction key for which there
	// was no waiting transaction
	TransactionBroken

	InvalidEventString string = "!!INVALID DEVICE EVENT TYPE!!"
)

var (
	// blankEvent is an Event in its initial state.  Useful for quick state reset.
	blankEvent Event
)

func (et EventType) String() string {
	switch et {
	case Connect:
		return "Connect"
	case Disconnect:
		return "Disconnect"
	case MessageSent:
		return "MessageSent"
	case MessageReceived:
		return "MessageReceived"
	case MessageFailed:
		return "MessageFailed"
	case TransactionComplete:
		return "TransactionComplete"
	case TransactionBroken:
		return "TransactionBroken"
	default:
		return InvalidEventString
	}
}

// Event represents a single occurrence of interest for device-related applications.
// Instances of Event should be considered immutable by application code.  Also, Event
// instances should not be stored across calls to a listener, as the infrastructure is
// free to reuse Event instances.
type Event struct {
	// Type describes the kind of this event.  This field is always set.
	Type EventType

	// Device refers to the device, possibly disconnected, for which this event is being set.
	// This field is always set.
	Device Interface

	// Message is the WRP message relevant to this event.
	//
	// Never assume that it is safe to use this Message outside the listener invocation.  Make
	// a copy if this Message is needed by other goroutines or if it needs to be part of a long-lived
	// data structure.
	Message wrp.Typed

	// Format is the encoding format of the Contents field
	Format wrp.Format

	// Contents is the encoded representation of the Message field.  It is always set if and only if
	// the Message field is set.
	//
	// Never assume that it is safe to use this byte slice outside the listener invocation.  Make
	// a copy if this byte slice is needed by other goroutines or if it needs to be part of a long-lived
	// data structure.
	Contents []byte

	// Ctx adds context to Event types to get the same functionally context package functionality as a http request would.
	//
	// For now ctx is used to carry money httptrackers is added to carry throughout the pumps.
	Context context.Context

	// Error is the error which occurred during an attempt to send a message.  This field is only populated
	// for MessageFailed events when there was an actual error.  For MessageFailed events that indicate a
	// device was disconnected with enqueued messages, this field will be nil.
	Error error
}

// WithContext is returns an Event modified with context.
func (e Event) WithContext(ctx context.Context) Event {
	if ctx == nil {
		panic("nil context")
	}

	e.Context = ctx
	return e
}

// Listener is an event sink.  Listeners should never modify events and should never
// store events for later use.  If data from an event is needed for another goroutine
// or for long-term storage, a copy should be made.
type Listener func(*Event)
