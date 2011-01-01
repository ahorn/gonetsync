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

// Interface which dispatches a request to another module to build a response.
type Responder interface {

	// Builds a response based on a request.
	Respond(request Message) (response Message, err os.Error)
}

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
		prepare, err := request.toPrepareMessage()
		if err != nil {
			return nil, err
		}

		promise, err := proc.Acceptor.OnPrepare(*prepare.Uusn)
		if err != nil {
			return nil, err
		}

		return promise.Marshal()

	case Phase_PROPOSE:
		propose, err := request.toProposeMessage()
		if err != nil {
			return nil, err
		}

		accept, err := proc.Acceptor.OnPropose(*propose.Uusn, propose.Val)
		if err != nil {
			return nil, err
		}

		return accept.Marshal()

	}

	return nil, nil
}
