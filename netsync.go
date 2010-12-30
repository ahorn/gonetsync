// This package uses the Paxos algorithm to reach consensus in a network of unreliable processors.
package netsync

import (
	"os"
	"io"
	"encoding/binary"
)

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

// Protocol participant of the distributed consensus algorithm.
// A majority of acceptors must accept a proposal for it to be chosen.
// Synchronization must be enforced by callers.
// Therefore, acceptor implementations need not be thread-safe.
type Acceptor interface {

	// Returns the most recently promised proposal number.
	// Note that promised proposal numbers are always increasing.
	// Moreover, if PromisedUusn() is strictly less than AcceptedUusn(),
	// the acceptor is part of a minority of acceptors which accepted
	// a proposal without having received the preceding prepare message.
	// If no promise has been made, then the returned integer is zero.
	PromisedUusn() uint64

	// Returns the most recently accepted proposal number.
	// Note that accepted proposal numbers are always increasing.
	// If no proposal has been accepted, the returned integer is zero.
	AcceptedUusn() uint64

	// An acceptor updates PromisedUusn() to higher-numbered proposals.
	// Henceforth, acceptors promise to reject lower-numbered proposals.
	// Before an acceptor replies with such a promise, it must persist the
	// promised proposal number to stable storage which survives failures.
	OnPrepare(uusn uint64) *PromiseMessage

	// An acceptor accepts proposals with unique numbers greater than or
	// equal to PromisedUusn(). Before an acceptor broadcasts a successful
	// response, it must persist the newly accepted proposal number and
	// its value to stable storage which survives failures and restarts.
	OnPropose(uusn uint64, val []byte) *AcceptMessage
}

// Abstract acceptor implementation which does not persist proposal information.
type acceptor struct {
	// Initially zero
	promisedUusn uint64

	// Initially nil
	acceptedProposal *proposal
}

func (a *acceptor) PromisedUusn() uint64 {
	return a.promisedUusn
}

func (a *acceptor) AcceptedUusn() uint64 {
	if a.acceptedProposal == nil {
		return 0
	}

	return a.acceptedProposal.uusn
}

// Determines if the acceptor can proceed with the proposal.
func (a *acceptor) isNew(uusn uint64) bool {
	return a.promisedUusn <= uusn
}

func (a *acceptor) OnPrepare(uusn uint64) (response *PromiseMessage) {
	ok := a.isNew(uusn)
	var info *proposal
	if ok {
		a.promisedUusn = uusn
		info = a.acceptedProposal
	} else {
		info = &proposal{uusn: a.promisedUusn}
	}

	return NewPromiseMessage(uusn, ok, info)
}

func (a *acceptor) OnPropose(uusn uint64, val []byte) (response *AcceptMessage) {
	ok := a.isNew(uusn)
	if ok {
		a.acceptedProposal = &proposal{uusn, val}
	}
	return NewAcceptMessage(uusn, ok)
}

type acceptorEncoder struct {
	writer io.Writer
}

// Byte encoding:
//	64 bits 	- promised proposal number
// 	64 bits 	- accepted proposal number (if any)
//	remaining bytes	- accepted value byte sequence (if any)
func (enc *acceptorEncoder) encode(a *acceptor) os.Error {
	if err := enc.write(a.promisedUusn); err != nil {
		return err
	}
	if a.acceptedProposal == nil {
		return nil
	}

	if err := enc.write(a.acceptedProposal.uusn); err != nil {
		return err
	}
	if err := enc.write(a.acceptedProposal.val); err != nil {
		return err
	}

	return nil
}

func (enc *acceptorEncoder) write(data interface{}) os.Error {
	return binary.Write(enc.writer, binary.LittleEndian, data)
}

// Number of bytes for the promised and accepted uint64 proposal numbers
const headerSize = 2 * 64 / 8

type acceptorDecoder struct {
	reader io.Reader
	size   int64
}

func newAcceptorDecoder(file *os.File) *acceptorDecoder {
	stat, err := file.Stat()
	if err != nil {
		return nil
	}
	return &acceptorDecoder{file, stat.Size}
}

func (dec *acceptorDecoder) decode() (a acceptor, err os.Error) {
	if err = dec.read(&a.promisedUusn); err != nil {
		return
	}

	if dec.size < headerSize {
		return
	}

	acceptedProposal := new(proposal)
	acceptedProposal.val = make([]byte, dec.size-headerSize)
	if err = dec.read(&acceptedProposal.uusn); err != nil {
		return
	}
	if err = dec.read(&acceptedProposal.val); err != nil {
		return
	}

	a.acceptedProposal = acceptedProposal

	return
}

func (dec *acceptorDecoder) read(data interface{}) os.Error {
	return binary.Read(dec.reader, binary.LittleEndian, data)
}

// An acceptor which persists promised and accepted proposal to a file.
type FileAcceptor struct {
	Name string

	// Embed abstract acceptor
	acceptor

	// After Start() not nil until Stop() has been called
	file *os.File

	// After Start() not nil
	encoder *acceptorEncoder

	// After Restart() not nil
	decoder *acceptorDecoder
}

// Initialize an acceptor which persists accepted proposals in a named file.
func NewFileAcceptor(name string) *FileAcceptor {
	return &FileAcceptor{Name: name}
}

// Restore the state of the acceptor before joining the protocol.
func (f *FileAcceptor) Restart() os.Error {
	file, err := os.Open(f.Name, os.O_RDONLY, 0644)
	if err != nil {
		return err
	} else {
		defer func() { file.Close() }()
	}

	dec := newAcceptorDecoder(file)
	f.acceptor, err = dec.decode()
	if err != nil {
		return err
	}
	return f.Start()
}

// Open file in which promised and accepted proposals should be saved.
func (f *FileAcceptor) Start() (err os.Error) {
	f.file, err = os.Open(f.Name, os.O_WRONLY|os.O_CREATE, 0644)
	f.encoder = &acceptorEncoder{f.file}

	return
}

// Close the file in which promised and accepted proposals are saved.
func (f *FileAcceptor) Stop() os.Error {
	if f.file == nil {
		return nil
	}

	defer func() { f.file = nil }()
	return f.file.Close()
}

// Determine if acceptor is enable to persist its state to a file.
func (f *FileAcceptor) IsStarted() bool {
	return f.file != nil
}

func (f *FileAcceptor) saveState() os.Error {
	f.file.Seek(0, 0)
	return f.encoder.encode(&f.acceptor)
}
