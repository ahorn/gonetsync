# netsync

netsync achieves reliable synchronization among unreliable processes in a network.

## Motivation

Distributed algorithms consist of autonomous processes which participate in a network protocol. However, network communication is often asynchronous and protocol participants can fail. Both factors contribute to the complexity of distributed software architectures. netsync is an implementation of the [Paxos algorithm](http://en.wikipedia.org/wiki/Paxos_algorithm) to deal with some of these complexities by enabling unreliable, distributed processes to reach agreement on decisions. 

## Installation

1. Make sure you have got a working Go environment. See the [install instructions](http://golang.org/doc/install.html). netsync targets the 'release' tag. 
2. Install the [Go implementation](http://code.google.com/p/goprotobuf/) of [Google's protocol buffers](http://code.google.com/p/protobuf/).
3. Since netsync installs itself as a Go package, the environment variable $GOROOT must be set. 
4. git clone git://github.com/ahorn/gonetsync.git
5. cd gonetsync && make install

## License

netsync is an open source project, distributed under the MIT license. 
