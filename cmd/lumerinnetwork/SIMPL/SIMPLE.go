package simple

import (
	"context"
	"errors"
	_ "fmt"
	"net"
	_ "time"

	_ "gitlab.com/TitanInd/lumerin/cmd/log"
	"gitlab.com/TitanInd/lumerin/cmd/lumerinnetwork/connectionmanager"
	"gitlab.com/TitanInd/lumerin/cmd/lumerinnetwork/lumerinconnection"
	"gitlab.com/TitanInd/lumerin/cmd/msgbus"
	"gitlab.com/TitanInd/lumerin/lumerinlib"

	//_ "gitlab.com/TitanInd/lumerin/cmd/config"
	_ "gitlab.com/TitanInd/lumerin/lumerinlib"
	contextlib "gitlab.com/TitanInd/lumerin/lumerinlib/context"
)

/*
The purpose of the simple layer is to accept any messages from the
protocol layer and pass messages to;
lower down the stack
to the message bus
It is also designed to return messages from the msg bus to the protocol
layer.
Refer to proxy router document
https://titanind.atlassian.net/wiki/spaces/PR/pages/5570561/Lumerin+Node
*/

//type SimpleStructProtocolFunc func(*SimpleStruct) chan *SimpleEvent

type ConnUniqueID uint
type URL string
type MsgType string
type ID string
type Data string
type EventHandler string
type SearchString string

// type NewProtocolFunc func(*SimpleStruct) chan *SimpleEvent
type NewProtocolInterface interface {
	NewProtocol(*SimpleStruct)
}

/*
The simple listen struct is used to establish a Listen port
(TCP, UDP, or TRUNK) and accept connections. The accepted
connections create a SimpleStruct{}, and are passed up to the
protocol layer where the connection is initialized with a new
context, which contains a protocol structure that allows for event handling.
*/
type SimpleListenStruct struct {
	ctx    context.Context
	cancel func()
	accept chan *SimpleStruct //channel to accept simple structs and process their message
	// lumerinlisten *lumerinconnection.LumerinListenStruct
	connectionListen *connectionmanager.ConnectionListenStruct
}

/*
The simple struct is used to point to a specific instance
of a connection manager and MsgBus. The structure ties these
to a protocol struct where events are directed to be handled.
*/
type SimpleStruct struct {
	ctx    context.Context
	cancel func() //it might make sense to use the WithCancel function instead
	//the event handler portion can be removed since the
	//EventHandler method in implemented on the SimpleStruct
	//eventHandler      interface{}                                             //this is the event handler function
	eventChan  chan *SimpleEvent  // Channle to get
	msgbusChan chan *msgbus.Event //
	//protocolChan      chan SimpleEvent                                        //channel for protocol to receive simple events
	maxMessageSize    uint                                                    //this value is not initially set so defaults to 0
	connectionMapping map[ConnUniqueID]*lumerinconnection.LumerinSocketStruct //mapping of uint to connections
	//connectionIndex   ConnUniqueID                                            //keeps track of connections in the mapping
	connectionStruct *connectionmanager.ConnectionStruct
}

/*
a struct that contains the data and the event type being passed into the SimpleStruct
*/
type SimpleEvent struct {
	EventType EventType
	Data      interface{}
}

/*
struct that tells the SimpleStruct which connection to provide
the encoded data to
*/
type SimpleConn struct {
	id   ConnUniqueID
	data []byte
}

type EventType string

const NoEvent EventType = "noevent"
const MsgBusEvent EventType = "msgbus"

//const MsgUpdateEvent EventType = "msgupdate"
//const MsgDeleteEvent EventType = "msgdelete"
//const MsgGetEvent EventType = "msgget"
//const MsgGetIndexEvent EventType = "msgindex"
//const MsgSearchEvent EventType = "msgsearch"
//const MsgSearchIndexEvent EventType = "msgsearchindex"
//const MsgPublishEvent EventType = "msgpublish"
//const MsgUnpublishEvent EventType = "msgunpublish"
//const MsgSubscribedEvent EventType = "msgsubscribe"
//const MsgUnsubscribedEvent EventType = "msgunsubscribe"
//const MsgRemovedEvent EventType = "msgremoved"
const ConnReadEvent EventType = "connread"
const ConnEOFEvent EventType = "conneof"
const ConnErrorEvent EventType = "connerror"
const ErrorEvent EventType = "error"
const MsgToProtocol EventType = "msgUp"

// this is a temporary function used to initialize a SimpleListenStruct
func dummyFunc() {}

// this is a dummy interface
type dummyInterface interface {
	dummy()
}

type dummyStruct struct {
}

func (d *dummyStruct) dummy() {
}

/*
create and return a struct with channels to listen to
call goroutine embedded in the struct
//assuming that the context being passed in will contain a ContextStruct in the value
*/
func New(ctx context.Context, listen net.Addr) (SimpleListenStruct, error) {
	//myContext may be used in the future
	//myContext := ctx.Value("ContextKey")

	ctx, cancel := context.WithCancel(ctx)

	c := ctx.Value(contextlib.ContextKey)
	if c == nil {
		contextlib.Logf(ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" called")
	}

	cs, ok := c.(*contextlib.ContextStruct)
	if !ok {
		contextlib.Logf(ctx, contextlib.LevelPanic, lumerinlib.FileLine()+" Context Structre not correct")
	}

	if cs.GetSrc() == nil {
		cs.Logf(contextlib.LevelPanic, "Context Src Addr not defined")
	}

	cls, e := connectionmanager.Listen(ctx)
	if e != nil {
		contextlib.Logf(ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" Lumerin Listen() returne error:%s", e)
	}

	myStruct := SimpleListenStruct{
		ctx:              ctx,
		cancel:           cancel,
		accept:           make(chan *SimpleStruct),
		connectionListen: cls,
	}
	// determine if a more robust error message is needed
	return myStruct, nil
}

//consider calling this as a gorouting from protocol layer, assuming
//protocll layer will have a layer to communicate with a chan over
func (s *SimpleListenStruct) Run() {

	cs := contextlib.GetContextStruct(s.ctx)
	if cs == nil {
		contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLine()+" Context Structre not correct")
	}

	if cs.GetProtocol() == nil {
		cs.Logf(contextlib.LevelPanic, "Context New Protocol Function not defined")
	}

	go s.goListenAccept()

}

//
//
//
func (s *SimpleListenStruct) goListenAccept() {

	cs := contextlib.GetContextStruct(s.ctx)
	if cs == nil {
		contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" Context Structre not correct")
	}

	if cs.GetProtocol() == nil {
		cs.Logf(contextlib.LevelPanic, lumerinlib.FileLineFunc()+" Context New Protocol Function not defined")
	}

	// This needs error checking....
	proto := cs.GetProtocol()
	cs.Logf(contextlib.LevelTrace, lumerinlib.FileLineFunc()+" GetProtocol func type:%t", proto)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Wait for a new connection here
		connectionStruct, e := s.connectionListen.Accept()
		if e != nil {
			contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLine()+" connectionListenStruct.Accept() returned error:%s", e)
		}

		//create a cancel function from the context in the SimpleListenStruct
		newctx, cancel := context.WithCancel(s.ctx)

		//creating a new simple struct to pass to the protocol layer
		newSimpleStruct := &SimpleStruct{
			ctx:               newctx,
			cancel:            cancel,
			eventChan:         make(chan *SimpleEvent),
			msgbusChan:        make(chan *msgbus.Event),
			maxMessageSize:    0,
			connectionMapping: map[ConnUniqueID]*lumerinconnection.LumerinSocketStruct{},
			connectionStruct:  connectionStruct,
		}

		// var np NewProtocolInterface = proto.(NewProtocolInterface)
		var np NewProtocolInterface
		np = proto.(NewProtocolInterface)
		np.NewProtocol(newSimpleStruct) // Call the supplied "new" protocol function here

		s.accept <- newSimpleStruct
	}
}

//
//
//
func (s *SimpleListenStruct) Accept() <-chan *SimpleStruct {
	return s.accept
}

// replacing the channel with a return statement containing the new simple struct
//func (s *SimpleListenStruct) NewSimpleStruct(ctx context.Context) {
//	go func() {
//		myStruct := &SimpleStruct{ //generate a new SimpleStruct
//			ctx:               ctx,
//			cancel:            dummyFunc,
//			eventHandler:      dummyStruct{},
//			eventChan:         make(chan *SimpleEvent),
//			connectionMapping: make(map[ConnUniqueID]*lumerinconnection.LumerinSocketStruct),
//		}
//		s.accept <- myStruct //push a SimpleStruct onto the SimpleListenStruct's accept channel
//	}()
//}

// Calls the listen context cancel function, which closes out the listener routine
func (s *SimpleListenStruct) Close() {
	s.cancel()
	//	_, cancel := context.WithCancel(s.ctx)
	//	cancel() //cancel is a function which terminates the associated goroutine
}

func (s *SimpleStruct) SetEventChan(eventchan chan *SimpleEvent) {
	s.eventChan = eventchan
}
func (s *SimpleStruct) GetEvent() <-chan *SimpleEvent {
	return s.eventChan
}

/*
Start a new go routine to handle the new connection context
after initialization by the protocol layer. There will be a
variable in the context that points to the protocol structure
containing all of the pertinent data for the state of the protocol
and event handler routines
All of the SimpleStruct functions that follow can be called
before and after Run() is called
It is assumed that Run() can only be called once
*/
/*
TODO pass context to SimpleListenStruct's designated connection layer
*/
func (s *SimpleStruct) Run() {

	// Just checking for good measure
	cs := contextlib.GetContextStruct(s.ctx)
	if cs == nil {
		contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" Context Structre not correct")
	}

	if s.maxMessageSize == 0 {
		s.maxMessageSize = 10 //setting the default max message size to 10 bytes
	}

	//
	// Using connection managers index as the UniqueID (for now?)
	//
	go func() {
		for {
			select {
			case <-s.Ctx().Done():
				contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" Closing down")
				return
			case comm := <-s.connectionStruct.ReadReady():
				ev := &SimpleEvent{
					EventType: ConnReadEvent,
					Data:      comm,
				}
				s.eventChan <- ev
			case msg := <-s.msgbusChan:
				ev := &SimpleEvent{
					EventType: MsgBusEvent,
					Data:      msg,
				}
				s.eventChan <- ev
			}
		}
	}()
}

/*
Calls the connection context cancel function which closes out the
currently established SRC connection and all of the associated DST connections
*/
func (s *SimpleStruct) Close() {
	_, cancel := context.WithCancel(s.ctx)
	cancel()
}

// Set IO buffer parameters
// this IO buffer parameters apply to the deque used to stage/proess messages
// for stage 1 this can be assumed to be unconfigurable and use defaults only
func (s *SimpleStruct) SetBuffer() {}

// Set message buffering to a certain delimiter, for example a newline character: ‘\n’
// for stage 1 this will assumed to be unconfigurable and only use '\n' as the
// new line
func (s *SimpleStruct) SetMessageDelimiterDefault() {}

// Set message buffering to be of a certain size
func (s *SimpleStruct) SetMessageSizeDefault(mSize uint) {
	s.maxMessageSize = mSize
}

// TODO not part of stage 1
// Set encryption parameters
func (s *SimpleStruct) SetEncryptionDefault() {}

// TODO not part of stage 1
// Set Compression parameters
func (s *SimpleStruct) SetCompressionDefault() {}

/*
Dial the a destination address (DST)
takes in a net.Addr object and feeds into the net.Dial function
the resulting Conn is then added to the SimpleStructs mapping and and associated
ConnUniqueID is returned from this function
*/
func (s *SimpleStruct) Dial(dst net.Addr) (int, error) {

	return s.connectionStruct.Dial(s.ctx, dst)

	// conn, err := lumerinconnection.Dial(s.ctx, dst) //creates a new net.Conn object
	//gets the current index value and asssigns to connection in mapping
	// var uID ConnUniqueID = s.connectionIndex
	// s.connectionMapping[uID] = conn
	// s.connectionIndex++ //increase the connectionIndex for the next time a conn is made
	//consider a mapping of connections and UID's
	// return uID, err
}

/*
function to retrieve the connection mapped to a unique id
*/
func (s *SimpleStruct) GetConnBasedOnConnUniqueID(x ConnUniqueID) *lumerinconnection.LumerinSocketStruct {
	return s.connectionMapping[x]
}

// Reconnect dropped connection
func (s *SimpleStruct) Redial(u ConnUniqueID) {}

// Used later to direct the default route
func (s *SimpleStruct) SetRoute(u int) error {
	return s.connectionStruct.SetRoute(u)
}

// Used later to direct the default route
func (s *SimpleStruct) GetRoute() {} //return of 1 to appease compiler

// Used later to direct the default route
func (s *SimpleStruct) GetLocalAddr(ConnUniqueID) {} //return of 1 to appease compiler

/*

network connection functions

*/

// Used later to direct the default route
func (s *SimpleStruct) GetRemoteAddr(ConnUniqueID) {} //return of 1 to appease compiler

func (s *SimpleStruct) SetDefaultReadHandler() {}

// Supply a handler function for incoming data reads for the connection ID
func (s *SimpleStruct) SetReadHandler() {}

// Writes buffer to the specified connection
func (s *SimpleStruct) Write(i int, msg []byte) (int, error) {
	return s.connectionStruct.IdxWrite(i, msg)
}

// Automatic duplication of writes to a MsgBus data channel
func (s *SimpleStruct) DupWrite() {}

// Flushes all IO Buffers
func (s *SimpleStruct) Flush() {}

// Reads low level connection status information
func (s *SimpleStruct) Status() {}

/*

msg bus functions

*/

func (s *SimpleStruct) Pub(MsgType, ID, Data) error            { return errors.New("") }
func (s *SimpleStruct) Unpub(MsgType, ID) error                { return errors.New("") }
func (s *SimpleStruct) Sub(MsgType, ID, EventHandler) error    { return errors.New("") }
func (s *SimpleStruct) Unsub(MsgType, ID, EventHandler) error  { return errors.New("") }
func (s *SimpleStruct) Get(MsgType, ID, EventHandler) error    { return errors.New("") }
func (s *SimpleStruct) Set(MsgType, ID, Data) error            { return errors.New("") }
func (s *SimpleStruct) SearchIP(MsgType, SearchString) error   { return errors.New("") }
func (s *SimpleStruct) SearchMac(MsgType, SearchString) error  { return errors.New("") }
func (s *SimpleStruct) SearchName(MsgType, SearchString) error { return errors.New("") }

func (s *SimpleStruct) Ctx() context.Context {
	return s.ctx
}

func (s *SimpleStruct) Cancel() {
	s.cancel()
}

/*


event handler related functionality


*/

// type EventType string
//
//var eventOne EventType = "eventOne"
//
////event handler function for the SimpleStruct which is viewable from the protocol layer
//func (s *SimpleStruct) EventHandler(e SimpleEvent) {
//	for {
//		switch e.EventType {
//		case NoEvent:
//			fallthrough
//		case MsgUpdateEvent:
//			fallthrough
//		case MsgDeleteEvent:
//			fallthrough
//		case MsgGetEvent:
//			fallthrough
//		case MsgGetIndexEvent:
//			fallthrough
//		case MsgSearchEvent:
//			fallthrough
//		case MsgSearchIndexEvent:
//			fallthrough
//		case MsgPublishEvent:
//			fallthrough
//		case MsgUnpublishEvent:
//			fallthrough
//		case MsgSubscribedEvent:
//			fallthrough
//		case MsgUnsubscribedEvent:
//			fallthrough
//		case MsgRemovedEvent:
//			fallthrough
//		case ConnReadEvent:
//			fallthrough
//		case ConnEOFEvent:
//			fallthrough
//		case ConnErrorEvent:
//			fallthrough
//		case ErrorEvent:
//			fallthrough
//		case MsgToProtocol:
//			fallthrough
//		default:
//			return
//		}
//	}
//}
//
