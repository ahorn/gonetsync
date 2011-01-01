// This package uses the Paxos algorithm to reach consensus in a network of unreliable processors.
package netsync

import "os"

// Since several proposers can simultaneously propose distinct values, no
// proposal might gain the majority of acceptors. Therefore, it could require
// multiple proposals before one value is chosen. For this reason, acceptors
// must accept more than one proposal. We keep track of each different proposal
// with a unique number. The protocol between Proposer and Acceptor processes
// guarantees that all chosen proposals have the same value. 
type proposal struct {
	uusn uint64
	val  []byte
}

var (
	ErrUnsupportedMessage = os.NewError("Message phase is unsupported")
	ErrCorruptedMessage = os.NewError("Message has been jumbled")
)

// Interface which dispatches a request to another module to build a response.
type Responder interface {

	// Builds a response based on a request.
	// Returns a nil response if no reply should be sent.
	Respond(request Message) (response Message, err os.Error)
}

// Structure to delegate messages to the appropriate modules
type Proc struct {
	// Embed interface to accept proposals
	Acceptor
}

func NewProc(fa *FileAcceptor) *Proc {
	return &Proc{Acceptor: fa}
}

// Dispatches proposer requests to acceptor implementation.
// Returns a nil response if incoming message does not conform to the Paxos protocol.
func (proc *Proc) Respond(request Message) (response Message, err os.Error) {
	switch request.Phase() {
	case Phase_PREPARE:
		request, err := request.toPrepareMessage()
		if err != nil {
			return nil, err
		}

		response, err := proc.Acceptor.OnPrepare(request)
		if err != nil {
			return nil, err
		}

		return response.Marshal()

	case Phase_PROPOSE:
		request, err := request.toProposeMessage()
		if err != nil {
			return nil, err
		}

		response, err := proc.Acceptor.OnPropose(request)
		if err != nil {
			return nil, err
		}

		return response.Marshal()

	}

	return nil, ErrUnsupportedMessage
}
