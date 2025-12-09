package websocket

import (
	"errors"
	"net/http"

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
}

// Bytes returns the message bytes in the specified protocol format
func (g *StructMessage) Bytes(protocol Protocol) ([]byte, error) {
	if protocol == ProtobufProtocol {
		return g.ProtobufBytes()
	}
	return g.JSONBytes()
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
		UUID:      u,
		Status:    int64(http.StatusOK),
	}
	pbMsg.SetValue(v)

	return &StructMessage{StructMessage: pbMsg}
}
