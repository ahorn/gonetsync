package netsync

// Encode messages broadcast by proposers.
type ProposerMessage interface {

	// Returns the proposal issued by the proposer
	GetProposal() *Proposal

	// Determines if message could be converted to a PrepareMessage.
	IsPrepareMessage() bool

	// Converts the proposer's message to a PrepareMessage.
	// Returns nil if the conversion fails.
	ToPrepareMessage() *PrepareMessage

	// Determines if message could be converted to an ProposeMessage.
	IsProposeMessage() bool

	// Converts the proposer's message to a ProposeMessage
	// Returns nil if the conversion fails.
	ToProposeMessage() *ProposeMessage
}

// Internal data structure for outgoing messages to acceptors
type proposerMessage struct {
	// Embed information about proposals
	*Proposal
}

type PrepareMessage struct {
	proposerMessage
}
type ProposeMessage struct {
	proposerMessage
}

func NewPrepareMessage(id uint64) *PrepareMessage {
	return &PrepareMessage{proposerMessage{NewProposal(id, nil)}}
}

func NewProposeMessage(proposal *Proposal) *ProposeMessage {
	return &ProposeMessage{proposerMessage{proposal}}
}

func (m *proposerMessage) GetProposal() *Proposal {
	return m.Proposal
}

func (m *proposerMessage) IsPrepareMessage() bool {
	return len(m.value) == 0
}

func (m *proposerMessage) IsProposeMessage() bool {
	return len(m.value) > 0
}

func (m *PrepareMessage) ToPrepareMessage() *PrepareMessage {
	return m
}

func (m *PrepareMessage) ToProposeMessage() *ProposeMessage {
	return nil
}

func (m *ProposeMessage) ToPrepareMessage() *PrepareMessage {
	return nil
}

func (m *ProposeMessage) ToProposeMessage() *ProposeMessage {
	return m
}

// Encode messages broadcast by acceptors.
type AcceptorMessage interface {

	// Returns the unique proposal number which triggered this message.
	ProposalId() uint64

	// Determines if the message is good news.
	IsOk() bool

	// Determines if message could be converted to a PromiseMessage.
	IsPromiseMessage() bool

	// Converts the acceptor's message to a PromiseMessage.
	// Returns nil if the conversion fails.
	ToPromiseMessage() *PromiseMessage

	// Determines if message could be converted to an AcceptMessage.
	IsAcceptMessage() bool

	// Converts the acceptor's message to an AcceptMessage.
	// Returns nil if the conversion fails.
	ToAcceptMessage() *AcceptMessage
}

// Internal data structure for outgoing messages to proposers
type acceptorMessage struct {
	// Unique proposal number which triggered this message.
	proposalId uint64

	// Flag to indicate if proposer's request has been successful
	ok bool
}

type PromiseMessage struct {
	acceptorMessage

	// Embed information about promise
	*Proposal "promiseInfo"
}

type AcceptMessage struct {
	acceptorMessage
}

func NewPromiseMessage(proposalId uint64, ok bool, proposal *Proposal) *PromiseMessage {
	return &PromiseMessage{acceptorMessage{proposalId, ok}, proposal}
}

func NewAcceptMessage(proposalId uint64, ok bool) *AcceptMessage {
	return &AcceptMessage{acceptorMessage{proposalId, ok}}
}

func (m *acceptorMessage) ProposalId() uint64 {
	return m.proposalId
}

func (m *acceptorMessage) IsOk() bool {
	return m.ok
}

func (m *PromiseMessage) IsPromiseMessage() bool {
	return true
}

func (m *PromiseMessage) ToPromiseMessage() *PromiseMessage {
	return m
}

func (m *PromiseMessage) IsAcceptMessage() bool {
	return false
}

func (m *PromiseMessage) ToAcceptMessage() *AcceptMessage {
	return nil
}

func (m *AcceptMessage) IsPromiseMessage() bool {
	return false
}

func (m *AcceptMessage) ToPromiseMessage() *PromiseMessage {
	return nil
}

func (m *AcceptMessage) IsAcceptMessage() bool {
	return true
}

func (m *AcceptMessage) ToAcceptMessage() *AcceptMessage {
	return m
}
