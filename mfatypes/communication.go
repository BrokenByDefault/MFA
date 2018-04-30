package mfatypes

type DefaultRequest struct {
	Action string `json:"action"`
	App    string `json:"app"`
}

// CLIENT: RECEIVER
type SubscribeRequest struct {
	PublicKey []byte `json:"pk"`
	DefaultRequest
}

// CLIENT: SENDER
type EnvelopeRequest struct {
	PublicKey []byte `json:"pk"`
	Envelope  []byte `json:"envelope"`
	DefaultRequest
}

// SERVER
type Subscription struct {
	PublicKey []byte `json:"pk,omitempty"`
	Error     string `json:"error,omitempty"`
}

type CallbackResponse struct {
	Envelope []byte `json:"envelope"`
	Error    string `json:"error,omitempty"`
}
