package netsync

import "os"

// Interface which dispatches a request to another module to build a response.
type Responder interface {

	// Builds a response based on a request.
	Respond(request Message) (response Message, err os.Error)

}

type Proc struct {
	fa	*FileAcceptor
}

func NewProc(fa *FileAcceptor) *Proc {
	return &Proc{ fa }
}

// Saves the acceptor state to a file if the request has been successful.
// Returns a nil response if incoming message does not conform to the Paxos protocol.
func (proc *Proc) Respond(request Message) (response Message, err os.Error) {
	switch request.Phase() {
	case Phase_PREPARE:
		prepare, err := request.toPrepareMessage()
		if err != nil {
			return nil, err
		}

		promise := proc.fa.OnPrepare(*prepare.Uusn)
		if *promise.Ok {
			// TODO: Optimize to save only changes in state
			err = proc.fa.saveState()
			if err != nil {
				return nil, err
			}

		}

		return promise.Marshal()

	case Phase_PROPOSE:
		propose, err := request.toProposeMessage()
		if err != nil {
			return nil, err
		}

		accept := proc.fa.OnPropose(*propose.Uusn, propose.Val)
		if *accept.Ok {
			// TODO: Optimize to save only changes in state
			err = proc.fa.saveState()
			if err != nil {
				return nil, err
			}
		}

		return accept.Marshal()
	}

	return nil, nil
}
