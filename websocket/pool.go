package websocket

import (
	"errors"
	"log"
	"math/rand"
	"reflect"
	"time"
)

type SpeakerPool interface {
	AddClient(c Speaker) error
	RemoveClient(c Speaker) bool
	AddEventHandler(h SpeakerEventHandler)
	GetSpeakers() []Speaker
	GetSpeakerByRemoteHost(host string) Speaker
	PickConnectedSpeaker() Speaker
	BroadcastMessage(m Message)
	SendMessageTo(m Message, host string) error
}

type Pool struct {
	RWMutex
	name              string
	eventHandlers     []SpeakerEventHandler
	eventHandlersLock RWMutex
	speakers          []Speaker
}

func (s *Pool) CloneEventHandlers() (handlers []SpeakerEventHandler) {
	s.eventHandlersLock.RLock()
	handlers = append(handlers, s.eventHandlers...)
	s.eventHandlersLock.RUnlock()

	return
}

func (s *Pool) Start() {
	speakers := s.GetSpeakers()
	for _, speaker := range speakers {
		speaker.Start()
	}
}

func (s *Pool) AddClient(c Speaker) error {
	s.Lock()
	s.speakers = append(s.speakers, c)
	s.Unlock()

	// This is to call SpeakerPool.On{Message,Disconnected}
	c.AddEventHandler(s)

	return nil
}

func (s *Pool) RemoveClient(c Speaker) bool {
	s.Lock()
	defer s.Unlock()

	host := c.GetRemoteHost()
	for i, ic := range s.speakers {
		if ic.GetRemoteHost() == host {
			s.speakers = append(s.speakers[:i], s.speakers[i+1:]...)
			return true
		}
	}

	return false
}

func (s *Pool) GetStatus() map[string]ConnStatus {
	clients := make(map[string]ConnStatus)
	for _, client := range s.GetSpeakers() {
		clients[client.GetRemoteHost()] = client.GetStatus()
	}
	return clients
}

func (s *Pool) GetName() string {
	return s.name + " type : [" + (reflect.TypeOf(s).String()) + "]"
}

func (s *Pool) GetSpeakers() (speakers []Speaker) {
	s.RLock()
	speakers = append(speakers, s.speakers...)
	s.RUnlock()
	return
}

func (s *Pool) PickConnectedSpeaker() Speaker {
	s.RLock()
	defer s.RUnlock()

	length := len(s.speakers)
	if length == 0 {
		return nil
	}

	index := rand.Intn(length)
	for i := 0; i != length; i++ {
		if c := s.speakers[index]; c != nil && c.IsConnected() {
			return c
		}

		if index+1 >= length {
			index = 0
		} else {
			index++
		}
	}

	return nil
}

func (s *Pool) GetSpeakerByRemoteHost(host string) Speaker {
	s.RLock()
	defer s.RUnlock()

	for _, c := range s.speakers {
		if c.GetRemoteHost() == host {
			return c
		}
	}
	return nil
}

func (s *Pool) SendMessageTo(m Message, host string) error {
	c := s.GetSpeakerByRemoteHost(host)
	if c == nil {
		return errors.New("no result found")
	}

	return c.SendMessage(m)
}

func (s *Pool) BroadcastMessage(m Message) {
	s.RLock()
	defer s.RUnlock()

	for _, c := range s.speakers {
		if err := c.SendMessage(m); err != nil {
			log.Default().Printf("Unable to send message from pool %s to %s: %s", s.name, c.GetRemoteHost(), err)
		}
	}
}

func (s *Pool) AddEventHandler(h SpeakerEventHandler) {
	s.eventHandlersLock.Lock()
	s.eventHandlers = append(s.eventHandlers, h)
	s.eventHandlersLock.Unlock()
}

func (s *Pool) OnMessage(c Speaker, m Message) {
	for _, h := range s.eventHandlers {
		h.OnMessage(c, m)
	}
}

func (s *Pool) OnConnected(c Speaker) {
	for _, h := range s.eventHandlers {
		h.OnConnected(c)
		if !c.IsConnected() {
			break
		}
	}
}

func (s *Pool) OnDisconnected(c Speaker) {
	for _, h := range s.eventHandlers {
		h.OnDisconnected(c)
	}
}

type ClientPool struct {
	*Pool
}

type incomerPool struct {
	*Pool
}

type StructClientPool struct {
	*ClientPool
	*structSpeakerPoolEventDispatcher
}

func (s *incomerPool) AddClient(c Speaker) error {
	s.Lock()
	s.speakers = append(s.speakers, c)
	s.Unlock()

	// This is to call SpeakerPool.On{Message,Disconnected}
	c.AddEventHandler(s)

	return nil
}

func (s *StructClientPool) AddClient(c Speaker) error {
	if wc, ok := c.(*Client); ok {
		speaker := wc.UpgradeToStructSpeaker()
		s.ClientPool.AddClient(speaker)
	} else {
		return errors.New("wrong client type")
	}
	return nil
}

func (s *StructClientPool) AddStructMessageHandler(h SpeakerStructMessageHandler, namespaces []string) {
	for _, client := range s.ClientPool.GetSpeakers() {
		client.(*StructSpeaker).AddStructMessageHandler(h, namespaces)
	}
}

func (s *StructClientPool) Request(host string, request *StructMessage, timeout time.Duration) (*StructMessage, error) {
	c := s.ClientPool.GetSpeakerByRemoteHost(host)
	if c == nil {
		return nil, errors.New("no result found")
	}

	return c.(*StructSpeaker).Request(request, timeout)
}

func (s *ClientPool) ConnectAll() {
	s.RLock()
	// shuffle connections to avoid election of the same client as master
	indexes := rand.Perm(len(s.speakers))
	for _, i := range indexes {
		s.speakers[i].Start()
	}
	s.RUnlock()
}

func newPool(name string) *Pool {
	return &Pool{
		name: name,
	}
}

func newIncomerPool(name string) *incomerPool {
	return &incomerPool{
		Pool: newPool(name),
	}
}

type StructSpeakerPool interface {
	SpeakerPool
	SpeakerStructMessageDispatcher
	Request(host string, request *StructMessage, timeout time.Duration) (*StructMessage, error)
}

func NewClientPool(name string) *ClientPool {
	s := &ClientPool{
		Pool: newPool(name),
	}
	s.Start()

	return s
}

func NewStructClientPool(name string) *StructClientPool {
	pool := NewClientPool(name)
	return &StructClientPool{
		ClientPool: pool,
	}
}
