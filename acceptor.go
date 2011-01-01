package netsync

import (
	"os"
	"io"
	"encoding/binary"
)

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
	OnPrepare(uusn uint64) (*PromiseMessage, os.Error)

	// An acceptor accepts proposals with unique numbers greater than or
	// equal to PromisedUusn(). Before an acceptor broadcasts a successful
	// response, it must persist the newly accepted proposal number and
	// its value to stable storage which survives failures and restarts.
	OnPropose(uusn uint64, val []byte) (*AcceptMessage, os.Error)
}

// Abstract acceptor implementation which does not persist proposal information.
type acceptor struct {
	// Initially zero
	promisedUusn uint64

	// Initially nil;
	// accepted proposal number is strictly greater than zero iff 
	//     accepted proposal value is not nil
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

// Abstract OnPrepare(uint64) implementation which does not
// persist the promised proposal number to stable storage.
func (a *acceptor) OnPrepare(uusn uint64) (*PromiseMessage, os.Error) {
	ok := a.isNew(uusn)
	var info *proposal
	if ok {
		a.promisedUusn = uusn
		info = a.acceptedProposal
	} else {
		info = &proposal{uusn: a.promisedUusn}
	}

	return NewPromiseMessage(uusn, ok, info), nil
}

// Abstract OnPropose(uint64, []byte) implementation which
// does not persist the accepted proposal to stable storage.
func (a *acceptor) OnPropose(uusn uint64, val []byte) (*AcceptMessage, os.Error) {
	ok := a.isNew(uusn)
	if ok {
		a.acceptedProposal = &proposal{uusn, val}
	}
	return NewAcceptMessage(uusn, ok), nil
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

// Saves the accepted promised proposal number to a file if the request has been successful.
func (fa *FileAcceptor) OnPrepare(uusn uint64) (*PromiseMessage, os.Error) {
	promise, _ := fa.acceptor.OnPrepare(uusn)
	if *promise.Ok {
		err := fa.savePromisedUusn()
		if err != nil {
			return nil, err
		}
	}
	return promise, nil
}

// Saves the accepted proposal information to a file if the request has been successful.
func (fa *FileAcceptor) OnPropose(uusn uint64, val []byte) (*AcceptMessage, os.Error) {
	accept, _ := fa.acceptor.OnPropose(uusn, val)
	if *accept.Ok {
		err := fa.saveAcceptedProposal()
		if err != nil {
			return nil, err
		}
	}

	return accept, nil
}

// Restore the state of the acceptor before joining the protocol.
func (fa *FileAcceptor) Restart() os.Error {
	file, err := os.Open(fa.Name, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	defer func() { file.Close() }()

	dec := newAcceptorDecoder(file)
	fa.acceptor, err = dec.decode()
	if err != nil {
		return err
	}
	return fa.Start()
}

// Open file in which promised and accepted proposals should be saved.
func (fa *FileAcceptor) Start() (err os.Error) {
	fa.file, err = os.Open(fa.Name, os.O_WRONLY|os.O_CREATE, 0644)
	fa.encoder = &acceptorEncoder{fa.file}

	return
}

// Close the file in which promised and accepted proposals are saved.
func (fa *FileAcceptor) Stop() os.Error {
	if !fa.IsStarted() {
		return nil
	}

	defer func() { fa.file = nil }()
	return fa.file.Close()
}

// Determine if acceptor is enable to persist its state to a file.
func (fa *FileAcceptor) IsStarted() bool {
	return fa.file != nil
}

func (fa *FileAcceptor) savePromisedUusn() os.Error {
	fa.file.Seek(0, 0)
	return fa.encoder.encodePromisedUusn(fa.promisedUusn)
}

func (fa *FileAcceptor) saveAcceptedProposal() os.Error {
	fa.file.Seek(uusnByteCount, 0)
	return fa.encoder.encodeAcceptedProposal(fa.acceptedProposal)
}

// Byte encoding:
//	64 bits 	- promised proposal number
// 	64 bits 	- accepted proposal number (if any)
//	remaining bytes	- accepted value byte sequence (only if there is an accepted proposal number)
type acceptorEncoder struct {
	writer io.Writer
}

func (enc *acceptorEncoder) encodePromisedUusn(promisedUusn uint64) os.Error {
	return enc.write(promisedUusn)
}

func (enc *acceptorEncoder) encodeAcceptedProposal(acceptedProposal *proposal) os.Error {
	if err := enc.write(acceptedProposal.uusn); err != nil {
		return err
	}
	if err := enc.write(acceptedProposal.val); err != nil {
		return err
	}

	return nil
}

func (enc *acceptorEncoder) write(data interface{}) os.Error {
	return binary.Write(enc.writer, binary.LittleEndian, data)
}


const (
	// Number of bytes for promised or accepted proposal numbers
	uusnByteCount = 64 / 8

	// Total number of bytes needed for promised and accepted proposal numbers
	totalUusnByteCount = 2 * uusnByteCount
)

type acceptorDecoder struct {
	reader io.Reader

	// number of bytes which can be read from the reader
	size int64
}

func newAcceptorDecoder(file *os.File) *acceptorDecoder {
	stat, err := file.Stat()
	if err != nil {
		return nil
	}
	return &acceptorDecoder{reader: file, size: stat.Size}
}

// Instantiates a new acceptor and restores its state by
// decoding promised and accepted proposal information in the reader.
func (dec *acceptorDecoder) decode() (a acceptor, err os.Error) {
	if err = dec.read(&a.promisedUusn); err != nil {
		return
	}

	// if there is no accepted proposal number, then there is no accepted proposal value
	if dec.size < totalUusnByteCount {
		return
	}

	acceptedProposal := new(proposal)
	acceptedProposal.val = make([]byte, dec.size-totalUusnByteCount)
	if err = dec.read(&acceptedProposal.uusn); err != nil {
		return
	}
	if err = dec.read(acceptedProposal.val); err != nil {
		return
	}

	a.acceptedProposal = acceptedProposal

	return
}

func (dec *acceptorDecoder) read(data interface{}) os.Error {
	return binary.Read(dec.reader, binary.LittleEndian, data)
}
