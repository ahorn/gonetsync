package netsync

import (
	"testing"
	"os"
)

const (
	initId  = 0
	fixture = "fixture.txt"
)

// Data structure and methods under test
var fa *FileAcceptor

func TestInitPromisedId(t *testing.T) {
	if id := fa.PromisedId(); id != initId {
		t.Fatalf("TestInitPromisedId expected %q got %q", initId, id)
	}
}

func TestInitAcceptedId(t *testing.T) {
	if id := fa.AcceptedId(); id != initId {
		t.Fatalf("TestInitAcceptedId expected %q got %q", initId, id)
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
	request            ProposerMessage
	expectedOk         bool
	expectedPromisedId uint64
	expectedAcceptedId uint64
}

var (
	someValue      = []byte{0x07, 0x03}
	someOtherValue = []byte{0xA3, 0xB7}
	tests          = []Test{
		{NewPrepareMessage(1), true, 1, 0},
		{NewPrepareMessage(2), true, 2, 0},
		{NewPrepareMessage(1), false, 2, 0},
		{NewPrepareMessage(3), true, 3, 0},
		{NewProposeMessage(NewProposal(2, someValue)), false, 3, 0},
		{NewProposeMessage(NewProposal(3, someValue)), true, 3, 3},
		{NewPrepareMessage(2), false, 3, 3},
		{NewPrepareMessage(5), true, 5, 3},
		{NewPrepareMessage(7), true, 7, 3},
		{NewPrepareMessage(2), false, 7, 3},
		{NewProposeMessage(NewProposal(5, someValue)), false, 7, 3},
		{NewProposeMessage(NewProposal(7, someValue)), true, 7, 7},
		{NewPrepareMessage(7), true, 7, 7},
		{NewProposeMessage(NewProposal(7, someValue)), true, 7, 7},
		{NewPrepareMessage(8), true, 8, 7},
		{NewProposeMessage(NewProposal(8, someOtherValue)), true, 8, 8},
	}
)

func TestProcess(t *testing.T) {
	defer cleanup()

	err := fa.Start()
	if err != nil {
		t.Fatalf("TestProcess encountered unexpected error %q", err)
	}
	defer fa.Stop()

	for _, test := range tests {
		response, err := fa.Process(test.request)

		if err != nil {
			t.Fatalf("TestProcess encountered unexpected error %q", err)
		}

		if ok := response.IsOk(); ok != test.expectedOk {
			t.Fatalf("TestProcess expected response.IsOk() == %q", test.expectedOk)
		}

		if id := fa.PromisedId(); id != test.expectedPromisedId {
			t.Fatalf("TestProcess expected promised ID %d got %d", test.expectedPromisedId, id)
		}

		if id := fa.AcceptedId(); id != test.expectedAcceptedId {
			t.Fatalf("TestProcess expected accepted ID %d got %d", test.expectedAcceptedId, id)
		}
	}
}

func TestRestart(t *testing.T) {
	defer cleanup()

	err := fa.Start()
	if err != nil {
		t.Fatalf("TestRestart encountered unexpected error %q", err)
	}

	for _, test := range tests {
		response, err := fa.Process(test.request)

		if err != nil {
			t.Fatalf("TestRestart encountered unexpected error %q", err)
		}

		if ok := response.IsOk(); ok != test.expectedOk {
			t.Fatalf("TestRestart expected response.IsOk() == %q", test.expectedOk)
		}

		if id := fa.PromisedId(); id != test.expectedPromisedId {
			t.Fatalf("TestRestart expected promised ID %d got %d", test.expectedPromisedId, id)
		}

		if id := fa.AcceptedId(); id != test.expectedAcceptedId {
			t.Fatalf("TestRestart expected accepted ID %d got %d", test.expectedAcceptedId, id)
		}

		if err := fa.Stop(); err != nil {
			t.Fatalf("TestRestart encountered unexpected error %q", err)
		}

		fa = NewFileAcceptor(fixture)
		err = fa.Restart()

		if err != nil {
			t.Fatalf("TestRestart encountered unexpected error %q", err)
		}

		if id := fa.PromisedId(); id != test.expectedPromisedId {
			t.Fatalf("TestRestart expected promised ID %d got %d", test.expectedPromisedId, id)
		}

		if id := fa.AcceptedId(); id != test.expectedAcceptedId {
			t.Fatalf("TestRestart expected accepted ID %d got %d", test.expectedAcceptedId, id)
		}
	}
}

func cleanup() {
	defer setup()

	if fa != nil {
		fa.Stop()
	}
}

func setup() {
	os.Remove(fixture)
	fa = NewFileAcceptor(fixture)
}

func init() {
	setup()
}
