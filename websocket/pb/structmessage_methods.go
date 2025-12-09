package pb

import (
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// ProtobufObject interface for objects that can be marshaled to protobuf
type ProtobufObject interface {
	proto.Marshaler
	proto.Message
}

// SetValue sets the value in the state
func (m *StructMessage) SetValue(v any) {
	m.XXX_state.Value = v
	m.UUID = m.Uuid // Sync UUID alias
}

// GetValue gets the value from the state
func (m *StructMessage) GetValue() any {
	return m.XXX_state.Value
}

// ProtobufBytes returns the protobuf encoded bytes of the message
func (m *StructMessage) ProtobufBytes() ([]byte, error) {
	if len(m.XXX_state.ProtobufCache) > 0 {
		return m.XXX_state.ProtobufCache, nil
	}

	obj, ok := m.XXX_state.Value.(ProtobufObject)
	if ok {
		m.Format = Format_PROTOBUF

		msgObj, err := obj.Marshal()
		if err != nil {
			return nil, err
		}
		m.Obj = msgObj
	} else {
		// fallback to json
		m.Format = Format_JSON

		msgObj, err := json.Marshal(m.XXX_state.Value)
		if err != nil {
			return nil, err
		}
		m.Obj = msgObj
	}

	msgBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	m.XXX_state.ProtobufCache = msgBytes

	return msgBytes, nil
}

// JSONBytes returns the JSON encoded bytes of the message
func (m *StructMessage) JSONBytes() ([]byte, error) {
	if len(m.XXX_state.JsonCache) > 0 {
		return m.XXX_state.JsonCache, nil
	}

	type StructMessageAlias StructMessage
	msg := struct {
		*StructMessageAlias
		Obj any `json:"obj"`
	}{
		StructMessageAlias: (*StructMessageAlias)(m),
		Obj:                m.XXX_state.Value,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	m.XXX_state.JsonCache = msgBytes

	return msgBytes, nil
}

// UnmarshalByProtocol unmarshals the message based on the protocol
func (m *StructMessage) UnmarshalByProtocol(b []byte, protocol string) error {
	if protocol == "protobuf" {
		if err := proto.Unmarshal(b, m); err != nil {
			return fmt.Errorf("error while decoding Protobuf StructMessage: %w", err)
		}
	} else {
		if err := m.UnmarshalJSON(b); err != nil {
			return fmt.Errorf("error while decoding JSON StructMessage: %w", err)
		}
	}

	return nil
}

// UnmarshalJSON custom unmarshaller for JSON
func (m *StructMessage) UnmarshalJSON(b []byte) error {
	msg := struct {
		Namespace string
		Type      string
		UUID      string
		Status    int64
		Obj       json.RawMessage
	}{}

	if err := json.Unmarshal(b, &msg); err != nil {
		return err
	}

	m.Namespace = msg.Namespace
	m.Type = msg.Type
	m.Uuid = msg.UUID
	m.UUID = msg.UUID
	m.Status = msg.Status
	m.Obj = []byte(msg.Obj)

	return nil
}

// Reply creates a reply message
func (m *StructMessage) Reply(v any, kind string, status int) *StructMessage {
	msg := &StructMessage{
		Namespace: m.Namespace,
		Type:      kind,
		Uuid:      m.Uuid,
		UUID:      m.UUID,
		Status:    int64(status),
	}
	msg.SetValue(v)

	return msg
}

// BytesWithProtocol returns the message bytes in the specified protocol format
func (m *StructMessage) BytesWithProtocol(protocol string) ([]byte, error) {
	if protocol == "protobuf" {
		return m.ProtobufBytes()
	}
	return m.JSONBytes()
}
