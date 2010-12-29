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

func TestInitPromisedUuid(t *testing.T) {
	if uuid := fa.PromisedUuid(); uuid != initId {
		t.Fatalf("TestInitPromisedUuid expected %q got %q", initId, uuid)
	}
}

func TestInitAcceptedUuid(t *testing.T) {
	if uuid := fa.AcceptedUuid(); uuid != initId {
		t.Fatalf("TestInitAcceptedUuid expected %q got %q", initId, uuid)
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
	request            Message
	expectedOk         bool
	expectedPromisedUuid uint64
	expectedAcceptedUuid uint64
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

		if ok := response.ok; ok != test.expectedOk {
			t.Fatalf("TestProcess expected response.Ok == %q", test.expectedOk)
		}

		if uuid := fa.PromisedUuid(); uuid != test.expectedPromisedUuid {
			t.Fatalf("TestProcess expected promised ID %d got %d", test.expectedPromisedUuid, uuid)
		}

		if uuid := fa.AcceptedUuid(); uuid != test.expectedAcceptedUuid {
			t.Fatalf("TestProcess expected accepted ID %d got %d", test.expectedAcceptedUuid, uuid)
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

		if ok := response.ok; ok != test.expectedOk {
			t.Fatalf("TestRestart expected response.ok == %q", test.expectedOk)
		}

		if uuid := fa.PromisedUuid(); uuid != test.expectedPromisedUuid {
			t.Fatalf("TestRestart expected promised ID %d got %d", test.expectedPromisedUuid, uuid)
		}

		if uuid := fa.AcceptedUuid(); uuid != test.expectedAcceptedUuid {
			t.Fatalf("TestRestart expected accepted ID %d got %d", test.expectedAcceptedUuid, uuid)
		}

		if err := fa.Stop(); err != nil {
			t.Fatalf("TestRestart encountered unexpected error %q", err)
		}

		fa = NewFileAcceptor(fixture)
		err = fa.Restart()

		if err != nil {
			t.Fatalf("TestRestart encountered unexpected error %q", err)
		}

		if uuid := fa.PromisedUuid(); uuid != test.expectedPromisedUuid {
			t.Fatalf("TestRestart expected promised ID %d got %d", test.expectedPromisedUuid, uuid)
		}

		if uuid := fa.AcceptedUuid(); uuid != test.expectedAcceptedUuid {
			t.Fatalf("TestRestart expected accepted ID %d got %d", test.expectedAcceptedUuid, uuid)
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

func setup() {
	os.Remove(fixture)
	fa = NewFileAcceptor(fixture)
}

func init() {
	setup()
}
