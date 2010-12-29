package netsync

import (
	"os"
	"goprotobuf.googlecode.com/hg/proto"
)

type Message []byte
type Response struct {
	Message
	ok bool
}

func (m Message) Phase() Phase {
	// TODO: document the purpose of the first byte (index 0)
	raw, _ := proto.DecodeVarint(m[1:])
	return Phase(raw)
}

func (m Message) toPrepareMessage() (pb *PrepareMessage, err os.Error) {
	pb = &PrepareMessage{}
	err = proto.Unmarshal(m, pb)
	return
}

func NewPrepareMessage(uusn uint64) *PrepareMessage {
	return &PrepareMessage{Phase: NewPhase(Phase_PREPARE), Uusn: &uusn}
}

func (m Message) ToProposeMessage() (pb *ProposeMessage, err os.Error) {
	pb = &ProposeMessage{}
	err = proto.Unmarshal(m, pb)
	return
}

func NewProposeMessage(p *proposal) *ProposeMessage {
	return &ProposeMessage{Phase: NewPhase(Phase_PROPOSE), Uusn: &p.uusn, Val: p.val}
}

func (m Message) ToPromiseMessage() (pb *PromiseMessage, err os.Error) {
	pb = &PromiseMessage{}
	err = proto.Unmarshal(m, pb)
	return
}

func NewPromiseMessage(uusn uint64, ok bool, p *proposal) *PromiseMessage {
	if p == nil {
		return &PromiseMessage{Phase: NewPhase(Phase_PROMISE), ReUusn: &uusn, Ok: &ok}
	}
	return &PromiseMessage{Phase: NewPhase(Phase_PROMISE), ReUusn: &uusn, Ok: &ok, Uusn: &p.uusn, Val: p.val}
}

func (m *PromiseMessage) toResponse() (*Response, os.Error) {
	data, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &Response{Message: data, ok: *m.Ok}, nil
}

func (m Message) ToAcceptMessage() (pb *AcceptMessage, err os.Error) {
	pb = &AcceptMessage{}
	err = proto.Unmarshal(m, pb)
	return
}

func NewAcceptMessage(uusn uint64, ok bool) *AcceptMessage {
	return &AcceptMessage{Phase: NewPhase(Phase_ACCEPT), ReUusn: &uusn, Ok: &ok}
}

func (m *AcceptMessage) toResponse() (*Response, os.Error) {
	data, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &Response{Message: data, ok: *m.Ok}, nil
}
