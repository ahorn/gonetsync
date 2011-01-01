package netsync

import (
	"testing"
	"os"
	"goprotobuf.googlecode.com/hg/proto"
)

const (
	initId  = 0
	fixture = "fixture.txt"
)

// Data structure and methods under test
var fa *FileAcceptor

func TestInitPromisedUusn(t *testing.T) {
	if uusn := fa.PromisedUusn(); uusn != initId {
		t.Fatalf("TestInitPromisedUusn expected %q got %q", initId, uusn)
	}
}

func TestInitAcceptedUusn(t *testing.T) {
	if uusn := fa.AcceptedUusn(); uusn != initId {
		t.Fatalf("TestInitAcceptedUusn expected %q got %q", initId, uusn)
	}
}

func TestIsStopped(t *testing.T) {
	if ok := fa.IsStarted(); ok {
		t.Fatalf("TestIsStopped expected stopped acceptor")
	}
}

func TestIsStarted(t *testing.T) {
	defer cleanup()

	err := fa.Start()
	if err != nil {
		t.Fatalf("TestIsStarted encountered unexpected error %q", err)
	}
	defer fa.Stop()

	if ok := fa.IsStarted(); !ok {
		t.Fatalf("TestIsStarted expected started acceptor")
	}
}

func TestStart(t *testing.T) {
	defer cleanup()

	if ok := fa.IsStarted(); ok {
		t.Fatalf("TestStart expected stopped acceptor")
	}

	err := fa.Start()
	if err != nil {
		t.Fatalf("TestStart encountered unexpected error %q", err)
	}
	defer fa.Stop()

	file, err := os.Open(fixture, os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("TestStart encountered unexpected error %q", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("TestStart encountered unexpected error %q", err)
	}

	if size := stat.Size; size != 0 {
		t.Fatalf("TestStart expected file size %d got %d", 0, size)
	}
}

func TestStop(t *testing.T) {
	if err := fa.Stop(); err != nil {
		// acceptor has never been started so it should stop
		t.Fatalf("TestStop encountered unexpected error %q", err)
	}
}

type Test struct {
	request              Message
	expectedOk           bool
	expectedPromisedUusn uint64
	expectedAcceptedUusn uint64
}

var (
	someValue      = []byte{0x07, 0x03}
	someOtherValue = []byte{0xA3, 0xB7}
	tests          = []Test{
		{toMessage(NewPrepareMessage(1)), true, 1, 0},
		{toMessage(NewPrepareMessage(2)), true, 2, 0},
		{toMessage(NewPrepareMessage(1)), false, 2, 0},
		{toMessage(NewPrepareMessage(3)), true, 3, 0},
		{toMessage(NewProposeMessage(&proposal{2, someValue})), false, 3, 0},
		{toMessage(NewProposeMessage(&proposal{3, someValue})), true, 3, 3},
		{toMessage(NewPrepareMessage(2)), false, 3, 3},
		{toMessage(NewPrepareMessage(5)), true, 5, 3},
		{toMessage(NewPrepareMessage(7)), true, 7, 3},
		{toMessage(NewPrepareMessage(2)), false, 7, 3},
		{toMessage(NewProposeMessage(&proposal{5, someValue})), false, 7, 3},
		{toMessage(NewProposeMessage(&proposal{7, someValue})), true, 7, 7},
		{toMessage(NewPrepareMessage(7)), true, 7, 7},
		{toMessage(NewProposeMessage(&proposal{7, someValue})), true, 7, 7},
		{toMessage(NewPrepareMessage(8)), true, 8, 7},
		{toMessage(NewProposeMessage(&proposal{8, someOtherValue})), true, 8, 8},
	}
)

func TestResponder(t *testing.T) {
	defer cleanup()

	err := fa.Start()
	if err != nil {
		t.Fatalf("TestResponder encountered unexpected error %q", err)
	}
	defer fa.Stop()

	proc := NewProc(fa)
	for _, test := range tests {
		response, err := proc.Respond(test.request)

		if err != nil {
			t.Fatalf("TestResponder encountered unexpected error %q", err)
		}

		if ok := isOk(response); ok != test.expectedOk {
			t.Fatalf("TestResponder expected isOk(response) == %q", test.expectedOk)
		}

		if uusn := fa.PromisedUusn(); uusn != test.expectedPromisedUusn {
			t.Fatalf("TestResponder expected promised ID %d got %d", test.expectedPromisedUusn, uusn)
		}

		if uusn := fa.AcceptedUusn(); uusn != test.expectedAcceptedUusn {
			t.Fatalf("TestResponder expected accepted ID %d got %d", test.expectedAcceptedUusn, uusn)
		}
	}
}

func TestRestart(t *testing.T) {
	defer cleanup()

	err := fa.Start()
	if err != nil {
		t.Fatalf("TestRestart encountered unexpected error %q", err)
	}

	proc := NewProc(fa)
	for _, test := range tests {
		response, err := proc.Respond(test.request)

		if err != nil {
			t.Fatalf("TestRestart encountered unexpected error %q", err)
		}

		if ok := isOk(response); ok != test.expectedOk {
			t.Fatalf("TestRestart expected isOk(response) == %q", test.expectedOk)
		}

		if uusn := fa.PromisedUusn(); uusn != test.expectedPromisedUusn {
			t.Fatalf("TestRestart expected promised ID %d got %d", test.expectedPromisedUusn, uusn)
		}

		if uusn := fa.AcceptedUusn(); uusn != test.expectedAcceptedUusn {
			t.Fatalf("TestRestart expected accepted ID %d got %d", test.expectedAcceptedUusn, uusn)
		}

		if err := fa.Stop(); err != nil {
			t.Fatalf("TestRestart encountered unexpected error %q", err)
		}

		err = fa.Restart()

		if err != nil {
			t.Fatalf("TestRestart encountered unexpected error %q", err)
		}

		if uusn := fa.PromisedUusn(); uusn != test.expectedPromisedUusn {
			t.Fatalf("TestRestart expected promised ID %d got %d", test.expectedPromisedUusn, uusn)
		}

		if uusn := fa.AcceptedUusn(); uusn != test.expectedAcceptedUusn {
			t.Fatalf("TestRestart expected accepted ID %d got %d", test.expectedAcceptedUusn, uusn)
		}
	}
}

func cleanup() {
	defer setup()

	if fa != nil {
		fa.Stop()
	}
}

func toMessage(pb interface{}) Message {
	data, _ := proto.Marshal(pb)
	return data
}

// Internal helper function to determine if a request has been successful.
func isOk(m Message) bool {
	switch m.Phase() {
	case Phase_PROMISE:
		promise, _ := m.toPromiseMessage()
		return *promise.Ok
	case Phase_ACCEPT:
		accept, _ := m.toAcceptMessage()
		return *accept.Ok
	}
	return false
}

func setup() {
	os.Remove(fixture)
	fa = NewFileAcceptor(fixture)
}

func init() {
	setup()
}
