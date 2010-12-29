include $(GOROOT)/src/Make.inc

TARG=netsync
GOFMT=gofmt

GOFILES=\
	netsync.go\
	messages.go\
	paxos.pb.go\

include $(GOROOT)/src/Make.pkg
include $(GOROOT)/src/pkg/goprotobuf.googlecode.com/hg/Make.protobuf

format:
	${GOFMT} -w -s netsync.go
	${GOFMT} -w -s messages.go
	${GOFMT} -w -s netsync_test.go
