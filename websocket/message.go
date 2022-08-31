package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"
	"net/http"
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

type ProtobufObject interface {
	proto.Marshaler
	proto.Message
}

type structMessageState struct {
	value interface{}

	jsonCache     []byte
	protobufCache []byte
}

func (g *StructMessage) protobufBytes() ([]byte, error) {
	if len(g.XXX_state.protobufCache) > 0 {
		return g.XXX_state.protobufCache, nil
	}

	obj, ok := g.XXX_state.value.(ProtobufObject)
	if ok {
		g.Format = Format_Protobuf

		msgObj, err := obj.Marshal()
		if err != nil {
			return nil, err
		}
		g.Obj = msgObj
	} else {
		// fallback to json
		g.Format = Format_Json

		msgObj, err := json.Marshal(g.XXX_state.value)
		if err != nil {
			return nil, err
		}
		g.Obj = msgObj
	}

	msgBytes, err := g.Marshal()
	if err != nil {
		return nil, err
	}

	g.XXX_state.protobufCache = msgBytes

	return msgBytes, nil
}

func (g *StructMessage) jsonBytes() ([]byte, error) {
	if len(g.XXX_state.jsonCache) > 0 {
		return g.XXX_state.jsonCache, nil
	}

	m := struct {
		*StructMessage
		Obj interface{}
	}{
		StructMessage: g,
		Obj:           g.XXX_state.value,
	}

	msgBytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	g.XXX_state.jsonCache = msgBytes

	return msgBytes, nil
}

func (g *StructMessage) Bytes(protocol Protocol) ([]byte, error) {
	if protocol == ProtobufProtocol {
		return g.protobufBytes()
	}
	return g.jsonBytes()
}

func (g *StructMessage) unmarshalByProtocol(b []byte, protocol Protocol) error {
	if protocol == ProtobufProtocol {
		if err := g.Unmarshal(b); err != nil {
			return fmt.Errorf("Error while decoding Protobuf StructMessage %s", err)
		}
	} else {
		if err := g.UnmarshalJSON(b); err != nil {
			return fmt.Errorf("Error while decoding JSON StructMessage %s", err)
		}
	}

	return nil
}

func (g *StructMessage) UnmarshalJSON(b []byte) error {
	m := struct {
		Namespace string
		Type      string
		UUID      string
		Status    int64
		Obj       json.RawMessage
	}{}

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	g.Namespace = m.Namespace
	g.Type = m.Type
	g.UUID = m.UUID
	g.Status = m.Status
	g.Obj = []byte(m.Obj)

	return nil
}

func (g *StructMessage) Reply(v interface{}, kind string, status int) *StructMessage {
	msg := &StructMessage{
		Namespace: g.Namespace,
		Type:      kind,
		UUID:      g.UUID,
		Status:    int64(status),
	}
	msg.XXX_state.value = v

	return msg
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

func NewStructMessage(ns string, tp string, v interface{}, uuids ...string) *StructMessage {
	var u string
	if len(uuids) != 0 {
		u = uuids[0]
	} else {
		v4, _ := uuid.NewV4()
		u = v4.String()
	}

	msg := &StructMessage{
		Namespace: ns,
		Type:      tp,
		UUID:      u,
		Status:    int64(http.StatusOK),
	}
	msg.XXX_state.value = v

	return msg
}
