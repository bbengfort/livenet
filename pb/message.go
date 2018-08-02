package pb

import "time"

// Wrap a message that has already been serialized into an Envelop for dispatch
func Wrap(sender string, mtype MessageType, message []byte) *Envelope {
	return &Envelope{
		Sender:    sender,
		Type:      mtype,
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Message:   message,
	}
}

// ParseTimestamp returns the parsed time struct from the envelope.
func (e *Envelope) ParseTimestamp() (time.Time, error) {
	return time.Parse(time.RFC3339Nano, e.GetTimestamp())
}
