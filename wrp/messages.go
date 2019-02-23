package wrp

//go:generate codecgen -st "wrp" -o messages_codec.go messages.go

// Typed is implemented by any WRP type which is associated with a MessageType.  All
// message types implement this interface.
type Typed interface {
	// MessageType is the type of message represented by this Typed.
	MessageType() MessageType
}

// Routable describes an object which can be routed.  Implementations will most
// often also be WRP Message instances.  All Routable objects may be passed to
// Encoders and Decoders.
//
// Not all WRP messages are Routable.  Only messages that can be sent through
// routing software (e.g. talaria) implement this interface.
type Routable interface {
	Typed

	// To is the destination of this Routable instance.  It corresponds to the Destination field
	// in WRP messages defined in this package.
	To() string

	// From is the originator of this Routable instance.  It corresponds to the Source field
	// in WRP messages defined in this package.
	From() string

	// IsTransactionPart tests if this message represents part of a transaction.  For this to be true,
	// both (1) the msg_type field must be of a type that participates in transactions and (2) a transaction_uuid
	// must exist in the message (see TransactionKey).
	//
	// If this method returns true, TransactionKey will always return a non-empty string.
	IsTransactionPart() bool

	// TransactionKey corresponds to the transaction_uuid field.  If present, this field is used
	// to match up responses from devices.
	//
	// Not all Routables support transactions, e.g. SimpleEvent.  For those Routable messages that do
	// not possess a transaction_uuid field, this method returns an empty string.
	TransactionKey() string

	// Response produces a new Routable instance which is a response to this one.  The new Routable's
	// destination (From) is set to the original source (To), with the supplied newSource used as the response's source.
	// The requestDeliveryResponse parameter indicates the success or failure of this response.  The underlying
	// type of the returned Routable will be the same as this type, i.e. if this instance is a Message,
	// the returned Routable will also be a Message.
	//
	// If applicable, the response's payload is set to nil.  All other fields are copied as is into the response.
	Response(newSource string, requestDeliveryResponse int64) Routable
}

// Message is the union of all WRP fields, made optional (except for Type).  This type is
// useful for transcoding streams, since deserializing from non-msgpack formats like JSON
// has some undesirable side effects.
//
// IMPORTANT: Anytime a new WRP field is added to any message, or a new message with new fields,
// those new fields must be added to this struct for transcoding to work properly.  And of course:
// update the tests!
//
// For server code that sends specific messages, use one of the other WRP structs in this package.
//
// For server code that needs to read one format and emit another, use this struct as it allows
// client code to transcode without knowledge of the exact type of message.
type Message struct {
	Type                    MessageType         `wrp:"msg_type"`
	Source                  string              `wrp:"source,omitempty"`
	Destination             string              `wrp:"dest,omitempty"`
	TransactionUUID         string              `wrp:"transaction_uuid,omitempty"`
	ContentType             string              `wrp:"content_type,omitempty"`
	Accept                  string              `wrp:"accept,omitempty"`
	Status                  *int64              `wrp:"status,omitempty"`
	RequestDeliveryResponse *int64              `wrp:"rdr,omitempty"`
	Headers                 []string            `wrp:"headers,omitempty"`
	Metadata                map[string]string   `wrp:"metadata,omitempty"`
	Spans                   [][]string          `wrp:"spans,omitempty"`
	IncludeSpans            *bool               `wrp:"include_spans,omitempty"`
	Path                    string              `wrp:"path,omitempty"`
	Payload                 []byte              `wrp:"payload,omitempty"`
	ServiceName             string              `wrp:"service_name,omitempty"`
	URL                     string              `wrp:"url,omitempty"`
	PartnerIDs              []string            `wrp:"partner_ids,omitempty"`
	Traces                  []map[string]string `wrp:"traces,omitempty"`
}

func (msg *Message) MessageType() MessageType {
	return msg.Type
}

func (msg *Message) To() string {
	return msg.Destination
}

func (msg *Message) From() string {
	return msg.Source
}

func (msg *Message) HasMoney() bool {
	return msg.Traces != nil
}

func (msg *Message) IsTransactionPart() bool {
	return msg.Type.SupportsTransaction() && len(msg.TransactionUUID) > 0
}

func (msg *Message) TransactionKey() string {
	return msg.TransactionUUID
}

func (msg *Message) Response(newSource string, requestDeliveryResponse int64) Routable {
	response := *msg
	response.Destination = msg.Source
	response.Source = newSource
	response.RequestDeliveryResponse = &requestDeliveryResponse
	response.Payload = nil

	return &response
}

// SetStatus simplifies setting the optional Status field, which is a pointer type tagged with omitempty.
func (msg *Message) SetStatus(value int64) *Message {
	msg.Status = &value
	return msg
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
func (msg *Message) SetRequestDeliveryResponse(value int64) *Message {
	msg.RequestDeliveryResponse = &value
	return msg
}

// SetIncludeSpans simplifies setting the optional IncludeSpans field, which is a pointer type tagged with omitempty.
func (msg *Message) SetIncludeSpans(value bool) *Message {
	msg.IncludeSpans = &value
	return msg
}

// SimpleRequestResponse represents a WRP message of type SimpleRequestResponseMessageType.
//
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#simple-request-response-definition
type SimpleRequestResponse struct {
	// Type is exposed principally for encoding.  This field *must* be set to SimpleRequestResponseMessageType,
	// and is automatically set by the BeforeEncode method.
	Type                    MessageType       `wrp:"msg_type"`
	Source                  string            `wrp:"source"`
	Destination             string            `wrp:"dest"`
	ContentType             string            `wrp:"content_type,omitempty"`
	Accept                  string            `wrp:"accept,omitempty"`
	TransactionUUID         string            `wrp:"transaction_uuid,omitempty"`
	Status                  *int64            `wrp:"status,omitempty"`
	RequestDeliveryResponse *int64            `wrp:"rdr,omitempty"`
	Headers                 []string          `wrp:"headers,omitempty"`
	Metadata                map[string]string `wrp:"metadata,omitempty"`
	Spans                   [][]string        `wrp:"spans,omitempty"`
	IncludeSpans            *bool             `wrp:"include_spans,omitempty"`
	Payload                 []byte            `wrp:"payload,omitempty"`
	PartnerIDs              []string          `wrp:"partner_ids,omitempty"`
}

// SetStatus simplifies setting the optional Status field, which is a pointer type tagged with omitempty.
func (msg *SimpleRequestResponse) SetStatus(value int64) *SimpleRequestResponse {
	msg.Status = &value
	return msg
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
func (msg *SimpleRequestResponse) SetRequestDeliveryResponse(value int64) *SimpleRequestResponse {
	msg.RequestDeliveryResponse = &value
	return msg
}

// SetIncludeSpans simplifies setting the optional IncludeSpans field, which is a pointer type tagged with omitempty.
func (msg *SimpleRequestResponse) SetIncludeSpans(value bool) *SimpleRequestResponse {
	msg.IncludeSpans = &value
	return msg
}

func (msg *SimpleRequestResponse) BeforeEncode() error {
	msg.Type = SimpleRequestResponseMessageType
	return nil
}

func (msg *SimpleRequestResponse) MessageType() MessageType {
	return msg.Type
}

func (msg *SimpleRequestResponse) To() string {
	return msg.Destination
}

func (msg *SimpleRequestResponse) From() string {
	return msg.Source
}

func (msg *SimpleRequestResponse) IsTransactionPart() bool {
	return len(msg.TransactionUUID) > 0
}

func (msg *SimpleRequestResponse) TransactionKey() string {
	return msg.TransactionUUID
}

func (msg *SimpleRequestResponse) Response(newSource string, requestDeliveryResponse int64) Routable {
	response := *msg
	response.Destination = msg.Source
	response.Source = newSource
	response.RequestDeliveryResponse = &requestDeliveryResponse
	response.Payload = nil

	return &response
}

// SimpleEvent represents a WRP message of type SimpleEventMessageType.
//
// This type implements Routable, and as such has a Response method.  However, in actual practice
// failure responses are not sent for messages of this type.  Response is merely supplied in order to satisfy
// the Routable interface.
//
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#simple-event-definition
type SimpleEvent struct {
	// Type is exposed principally for encoding.  This field *must* be set to SimpleEventMessageType,
	// and is automatically set by the BeforeEncode method.
	Type        MessageType       `wrp:"msg_type"`
	Source      string            `wrp:"source"`
	Destination string            `wrp:"dest"`
	ContentType string            `wrp:"content_type,omitempty"`
	Headers     []string          `wrp:"headers,omitempty"`
	Metadata    map[string]string `wrp:"metadata,omitempty"`
	Payload     []byte            `wrp:"payload,omitempty"`
	PartnerIDs  []string          `wrp:"partner_ids,omitempty"`
}

func (msg *SimpleEvent) BeforeEncode() error {
	msg.Type = SimpleEventMessageType
	return nil
}

func (msg *SimpleEvent) MessageType() MessageType {
	return msg.Type
}

func (msg *SimpleEvent) To() string {
	return msg.Destination
}

func (msg *SimpleEvent) From() string {
	return msg.Source
}

// IsTransactionPart for SimpleEvent types always returns false
func (msg *SimpleEvent) IsTransactionPart() bool {
	return false
}

func (msg *SimpleEvent) TransactionKey() string {
	return ""
}

func (msg *SimpleEvent) Response(newSource string, requestDeliveryResponse int64) Routable {
	response := *msg
	response.Destination = msg.Source
	response.Source = newSource
	response.Payload = nil

	return &response
}

// CRUD represents a WRP message of one of the CRUD message types.  This type does not implement BeforeEncode,
// and so does not automatically set the Type field.  Client code must set the Type code appropriately.
//
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#crud-message-definition
type CRUD struct {
	Type                    MessageType       `wrp:"msg_type"`
	Source                  string            `wrp:"source"`
	Destination             string            `wrp:"dest"`
	TransactionUUID         string            `wrp:"transaction_uuid,omitempty"`
	ContentType             string            `wrp:"content_type,omitempty"`
	Headers                 []string          `wrp:"headers,omitempty"`
	Metadata                map[string]string `wrp:"metadata,omitempty"`
	Spans                   [][]string        `wrp:"spans,omitempty"`
	IncludeSpans            *bool             `wrp:"include_spans,omitempty"`
	Status                  *int64            `wrp:"status,omitempty"`
	RequestDeliveryResponse *int64            `wrp:"rdr,omitempty"`
	Path                    string            `wrp:"path"`
	Payload                 []byte            `wrp:"payload,omitempty"`
	PartnerIDs              []string          `wrp:"partner_ids,omitempty"`
}

// SetStatus simplifies setting the optional Status field, which is a pointer type tagged with omitempty.
func (msg *CRUD) SetStatus(value int64) *CRUD {
	msg.Status = &value
	return msg
}

// SetRequestDeliveryResponse simplifies setting the optional RequestDeliveryResponse field, which is a pointer type tagged with omitempty.
func (msg *CRUD) SetRequestDeliveryResponse(value int64) *CRUD {
	msg.RequestDeliveryResponse = &value
	return msg
}

// SetIncludeSpans simplifies setting the optional IncludeSpans field, which is a pointer type tagged with omitempty.
func (msg *CRUD) SetIncludeSpans(value bool) *CRUD {
	msg.IncludeSpans = &value
	return msg
}

func (msg *CRUD) MessageType() MessageType {
	return msg.Type
}

func (msg *CRUD) To() string {
	return msg.Destination
}

func (msg *CRUD) From() string {
	return msg.Source
}

func (msg *CRUD) IsTransactionPart() bool {
	return len(msg.TransactionUUID) > 0
}

func (msg *CRUD) TransactionKey() string {
	return msg.TransactionUUID
}

func (msg *CRUD) Response(newSource string, requestDeliveryResponse int64) Routable {
	response := *msg
	response.Destination = msg.Source
	response.Source = newSource
	response.RequestDeliveryResponse = &requestDeliveryResponse

	return &response
}

// ServiceRegistration represents a WRP message of type ServiceRegistrationMessageType.
//
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#on-device-service-registration-message-definition
type ServiceRegistration struct {
	// Type is exposed principally for encoding.  This field *must* be set to ServiceRegistrationMessageType,
	// and is automatically set by the BeforeEncode method.
	Type        MessageType `wrp:"msg_type"`
	ServiceName string      `wrp:"service_name"`
	URL         string      `wrp:"url"`
}

func (msg *ServiceRegistration) BeforeEncode() error {
	msg.Type = ServiceRegistrationMessageType
	return nil
}

// ServiceAlive represents a WRP message of type ServiceAliveMessageType.
//
// https://github.com/Comcast/wrp-c/wiki/Web-Routing-Protocol#on-device-service-alive-message-definition
type ServiceAlive struct {
	// Type is exposed principally for encoding.  This field *must* be set to ServiceAliveMessageType,
	// and is automatically set by the BeforeEncode method.
	Type MessageType `wrp:"msg_type"`
}

func (msg *ServiceAlive) BeforeEncode() error {
	msg.Type = ServiceAliveMessageType
	return nil
}
