package portal

type mockProto struct {
	mockProtoSig
	epAdded   chan Endpoint
	epRemoved chan Endpoint
}

func (m mockProto) Init(ProtocolPortal)        {}
func (m mockProto) AddEndpoint(ep Endpoint)    { m.epAdded <- ep }
func (m mockProto) RemoveEndpoint(ep Endpoint) { m.epRemoved <- ep }

type mockProtoSig struct {
	name, peerName     string
	number, peerNumber uint16
}

func (m mockProtoSig) Name() string       { return m.name }
func (m mockProtoSig) PeerName() string   { return m.peerName }
func (m mockProtoSig) Number() uint16     { return m.number }
func (m mockProtoSig) PeerNumber() uint16 { return m.peerNumber }

type mockEP struct {
	id  ID
	sc  chan *Message
	rc  chan *Message
	sig mockProtoSig
}

func (m mockEP) ID() ID                       { return m.id }
func (m mockEP) Close()                       {}
func (m mockEP) RecvChannel() chan<- *Message { return m.rc }
func (m mockEP) SendChannel() <-chan *Message { return m.sc }
func (m mockEP) Signature() ProtocolSignature { return m.sig }

type mockProtoExt struct {
	mockProto
	onSend, onRecv func(*Message) bool
}

func (m mockProtoExt) SendHook(msg *Message) bool { return m.onSend(msg) }
func (m mockProtoExt) RecvHook(msg *Message) bool { return m.onRecv(msg) }
