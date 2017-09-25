package portal

import (
	"time"
)

// Portal is the main access handle applications use to access the protocol
// system.  It is an abstraction of an application's "connection" to a
// messaging topology.  Applications can have more than one Socket open
// at a time.
type Portal interface {
	Connect(addr string) error
	Bind(addr string) error
	Close()

	Send(interface{})
	Recv() interface{}

	// SendMsg puts the message on the outbound send.  It works like Send,
	// but allows the caller to supply message headers.  AGAIN, the Socket
	// ASSUMES OWNERSHIP OF THE MESSAGE.
	SendMsg(*Message)

	// RecvMsg receives a complete message, including the message header,
	// which is useful for protocols in raw mode.
	RecvMsg() *Message
}

// Endpoint is used by the Protocol implementation to access the underlying
// channel transport
type Endpoint interface {
	// GetID() uint32
	Close()
	Done() <-chan struct{}
	SendMsg(*Message)
	RecvMsg() *Message
}

// Protocol implements messaging topography elements.  Each protocol will
// implement Protocol and each half of a protocol pair (e.g. REQ/REP) will
// implement Protocol.
type Protocol interface {
	// Init is called by the core to allow the protocol to perform
	// any initialization steps it needs.  It should save the handle
	// for future use, as well.
	Init(ProtocolPortal)

	// Shutdown is used to drain the send side.  It is only ever called
	// when the socket is being shutdown cleanly. Protocols should use
	// the linger time, and wait up to that time for sockets to drain.
	Shutdown(time.Time)

	// AddEndpoint is called when a new Endpoint is added to the socket.
	// Typically this is as a result of connect or accept completing.
	AddEndpoint(Endpoint)

	// RemoveEndpoint is called when an Endpoint is removed from the socket.
	// Typically this indicates a disconnected or closed connection.
	RemoveEndpoint(Endpoint)

	// ProtocolNumber returns a 16-bit value for the protocol number,
	// as assigned by the SP governing body. (IANA?)
	Number() uint16

	// Name returns our name.
	Name() string

	// PeerNumber() returns a 16-bit number for our peer protocol.
	PeerNumber() uint16

	// PeerName() returns the name of our peer protocol.
	PeerName() string

	// // GetOption is used to retrieve the current value of an option.
	// // If the protocol doesn't recognize the option, EBadOption should
	// // be returned.
	// GetOption(string) (interface{}, error)
	//
	// // SetOption is used to set an option.  EBadOption is returned if
	// // the option name is not recognized, EBadValue if the value is
	// // invalid.
	// SetOption(string, interface{}) error
}

// ProtocolPortal is the "handle" given to protocols to interface with the
// socket.  The Protocol implementation should not access any sockets or pipes
// except by using functions made available on the ProtocolSocket.  Note
// that all functions listed here are non-blocking.
type ProtocolPortal interface {
	// SendChannel represents the channel used to send messages.  The
	// application injects messages to it, and the protocol consumes
	// messages from it.  The channel may be closed when the core needs to
	// create a new channel, typically after an option is set that requires
	// the channel to be reconfigured.  (OptionWriteQLen) When the protocol
	// implementation notices this, it should call this function again to obtain
	// the value of the new channel.
	SendChannel() <-chan *Message

	// RecvChannel is the channel used to receive messages.  The protocol
	// should inject messages to it, and the application will consume them
	// later.
	RecvChannel() chan<- *Message

	// The protocol can wait on this channel to close.  When it is closed,
	// it indicates that the application has closed the upper read socket,
	// and the protocol should stop any further read operations on this
	// instance.
	CloseChannel() <-chan struct{}
}
