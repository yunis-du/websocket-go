package websocket

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/yunis-du/websocket-go/websocket/pb"
)

type Message interface {
	Bytes(protocol Protocol) ([]byte, error)
}

// Protocol used to transport messages
type Protocol string

const (
	// WildcardNamespace is the namespace used as wildcard. It is used by listeners to filter callbacks.
	WildcardNamespace = "*"
	// RawProtocol is used for raw messages
	RawProtocol Protocol = "raw"
	// ProtobufProtocol is used for protobuf encoded messages
	ProtobufProtocol Protocol = "protobuf"
	// JSONProtocol is used for JSON encoded messages
	JSONProtocol Protocol = "json"
)

type RawMessage []byte

func (m RawMessage) Bytes(protocol Protocol) ([]byte, error) {
	return m, nil
}

// StructMessage wraps pb.StructMessage and implements Message interface
type StructMessage struct {
	*pb.StructMessage
	Value any
}

// Bytes returns the message bytes in the specified protocol format
func (g *StructMessage) Bytes(protocol Protocol) ([]byte, error) {
	if protocol == ProtobufProtocol {
		return g.ProtobufBytes()
	}
	return g.JSONBytes()
}

// ProtobufBytes returns the protobuf encoded bytes of the message
func (m *StructMessage) ProtobufBytes() ([]byte, error) {
	if obj, ok := m.Value.(proto.Message); ok {
		m.Format = pb.Format_PROTOBUF

		msgObj, err := proto.Marshal(obj)
		if err != nil {
			return nil, err
		}
		m.Obj = msgObj
	} else {
		// fallback to json
		m.Format = pb.Format_JSON

		msgObj, err := json.Marshal(m.Value)
		if err != nil {
			return nil, err
		}
		m.Obj = msgObj
	}

	return proto.Marshal(m.StructMessage)
}

// JSONBytes returns the JSON encoded bytes of the message
func (m *StructMessage) JSONBytes() ([]byte, error) {
	type StructMessageAlias pb.StructMessage
	msg := struct {
		*StructMessageAlias
		Obj any `json:"obj"`
	}{
		StructMessageAlias: (*StructMessageAlias)(m.StructMessage),
		Obj:                m.Value,
	}

	return json.Marshal(msg)
}

// Unmarshal decodes the message based on the protocol
func (m *StructMessage) Unmarshal(data []byte, protocol Protocol) error {
	if protocol == ProtobufProtocol {
		if m.StructMessage == nil {
			m.StructMessage = &pb.StructMessage{}
		}
		return proto.Unmarshal(data, m.StructMessage)
	}

	// JSON
	msg := struct {
		Namespace string          `json:"namespace"`
		Type      string          `json:"type"`
		Uuid      string          `json:"uuid"`
		Status    int64           `json:"status"`
		Format    pb.Format       `json:"format"`
		Obj       json.RawMessage `json:"obj"`
	}{}

	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	if m.StructMessage == nil {
		m.StructMessage = &pb.StructMessage{}
	}
	m.Namespace = msg.Namespace
	m.Type = msg.Type
	m.Uuid = msg.Uuid
	m.Status = msg.Status
	m.Format = msg.Format
	m.Obj = []byte(msg.Obj)

	return nil
}

func (p *Protocol) parse(s string) error {
	switch s {
	case "", "json":
		*p = JSONProtocol
	case "protobuf":
		*p = ProtobufProtocol
	default:
		return errors.New("protocol not supported")
	}
	return nil
}

func (p *Protocol) String() string {
	return string(*p)
}

type SpeakerStructMessageDispatcher interface {
	AddStructMessageHandler(h SpeakerStructMessageHandler, namespaces []string)
}

type SpeakerStructMessageHandler interface {
	SpeakerEventHandler
	OnStructMessage(c Speaker, m *StructMessage)
}

func NewStructMessage(ns string, tp string, v any, uuids ...string) *StructMessage {
	var u string
	if len(uuids) != 0 {
		u = uuids[0]
	} else {
		v4, _ := uuid.NewV4()
		u = v4.String()
	}

	pbMsg := &pb.StructMessage{
		Namespace: ns,
		Type:      tp,
		Uuid:      u,
		Status:    int64(http.StatusOK),
	}

	return &StructMessage{StructMessage: pbMsg, Value: v}
}

// Reply creates a reply message with the same Namespace and UUID
func (m *StructMessage) Reply(v any, kind string, status int) *StructMessage {
	return NewStructMessage(m.Namespace, kind, v, m.Uuid)
}
