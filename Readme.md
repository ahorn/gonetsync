# netsync.go

netsync.go achieves reliable synchronization among unreliable processes in a network.

## Motivation

Distributed algorithms consist of autonomous processes which participate in a network protocol. However, network communication is often asynchronous and protocol participants can fail. Both factors contribute to the complexity of distributed software architectures. netsync.go is an implementation of the [Paxos algorithm](http://en.wikipedia.org/wiki/Paxos_algorithm) to deal with some of these complexities by enabling unreliable, distributed processes to reach agreement on decisions. 

## Installation

1. Make sure you have got a working Go environment. See the [install instructions](http://golang.org/doc/install.html). netsync.go targets the 'release' tag. 
2. Since netsync.go installs itself as a Go package, the environment variable $GOROOT must be set. 
3. git clone git://github.com/ahorn/netsync.go.git
4. cd netsync.go && make install

## License

netsync.go is an open source project, distributed under the MIT license. 
