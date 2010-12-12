include $(GOROOT)/src/Make.inc

TARG=netsync
GOFMT=gofmt

GOFILES=\
	netsync.go\
	comm.go\

include $(GOROOT)/src/Make.pkg

format:
	${GOFMT} -w -s netsync.go
	${GOFMT} -w -s netsync_test.go
	${GOFMT} -w -s comm.go
