package simple

import (
	"context"
	"errors"
	"fmt"
	"net"

	"gitlab.com/TitanInd/lumerin/cmd/lumerinnetwork/connectionmanager"
	"gitlab.com/TitanInd/lumerin/cmd/lumerinnetwork/lumerinconnection"
	"gitlab.com/TitanInd/lumerin/cmd/msgbus"
	"gitlab.com/TitanInd/lumerin/lumerinlib"
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

type ConnUniqueID int
type URL string
type MsgType string
type ID string
type Data string
type EventHandler string
type SearchString string

/*
The simple listen struct is used to establish a Listen port
(TCP, UDP, or TRUNK) and accept connections. The accepted
connections create a SimpleStruct{}, and are passed up to the
protocol layer where the connection is initialized with a new
context, which contains a protocol structure that allows for event handling.
*/
type SimpleListenStruct struct {
	ctx              context.Context
	cancel           func()
	accept           chan *SimpleStruct //channel to accept simple structs and process their message
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
	openChan   chan *SimpleConnOpenEvent
	//protocolChan      chan SimpleEvent                                        //channel for protocol to receive simple events
	// maxMessageSize    uint                                                    //this value is not initially set so defaults to 0
	connectionMapping map[ConnUniqueID]*lumerinconnection.LumerinSocketStruct //mapping of uint to connections
	//connectionIndex   ConnUniqueID                                            //keeps track of connections in the mapping
	ConnectionStruct *connectionmanager.ConnectionStruct
}

/*
a struct that contains the data and the event type being passed into the SimpleStruct
*/
type SimpleEvent struct {
	EventType     EventType
	ConnReadEvent *SimpleConnReadEvent
	ConnOpenEvent *SimpleConnOpenEvent
	MsgBusEvent   *SimpleMsgBusEvent
}

type SimpleConnReadEvent struct {
	uniqueID ConnUniqueID
	data     []byte
	count    int
	err      error
}

type SimpleConnOpenEvent struct {
	requestID int
	uniqueID  ConnUniqueID
	err       error
}

func (s *SimpleConnReadEvent) UniqueID() ConnUniqueID { return s.uniqueID }
func (s *SimpleConnReadEvent) Data() []byte           { return s.data }
func (s *SimpleConnReadEvent) Count() int             { return s.count }
func (s *SimpleConnReadEvent) Err() error             { return s.err }

func (s *SimpleConnOpenEvent) UniqueID() ConnUniqueID { return s.uniqueID }
func (s *SimpleConnOpenEvent) RequestID() int         { return s.requestID }
func (s *SimpleConnOpenEvent) Err() error             { return s.err }

type SimpleMsgBusEvent struct {
	EventType msgbus.EventType
	Msg       msgbus.MsgType
	ID        msgbus.IDString
	RequestID int
	Data      interface{}
	Err       error
}

/*
struct that tells the SimpleStruct which connection to provide
the encoded data to
*/

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
const ConnOpenEvent EventType = "connopen"
const ConnReadEvent EventType = "connread"
const ConnEOFEvent EventType = "conneof"
const ConnErrorEvent EventType = "connerror"
const ErrorEvent EventType = "error"
const MsgToProtocol EventType = "msgUp"

/*
create and return a struct with channels to listen to
call goroutine embedded in the struct
//assuming that the context being passed in will contain a ContextStruct in the value
*/
func NewListen(ctx context.Context, listen net.Addr) (SimpleListenStruct, error) {
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

	cls, e := connectionmanager.NewListen(ctx)
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

	contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")

	s.connectionListen.Run()
	go s.goListenAccept()

}

//
//
//
func (s *SimpleListenStruct) goListenAccept() {

	contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" enter")

	cs := contextlib.GetContextStruct(s.ctx)
	if cs == nil {
		contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" Context Structre not correct")
	}

	// if cs.GetProtocol() == nil {
	// 	contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" Context New Protocol Function not defined")
	// }

	// This needs error checking....
	// proto := cs.GetProtocol()
	// if proto == nil {
	// 	contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" GetProtocol() returned nil")
	// }

	connectionStructChan := s.connectionListen.Accept()

FORLOOP:
	for {
		select {
		case <-s.ctx.Done():
			contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" context canceled")
			break FORLOOP
		case connectionStruct := <-connectionStructChan:

			if connectionStruct == nil {
				contextlib.Logf(s.ctx, contextlib.LevelError, lumerinlib.FileLineFunc()+" Connection Listen Accept returned nil")
				s.cancel()
				break
			}

			//create a cancel function from the context in the SimpleListenStruct
			newctx, cancel := context.WithCancel(s.ctx)
			eventchan := make(chan *SimpleEvent)

			//creating a new simple struct to pass to the protocol layer
			newSimpleStruct := &SimpleStruct{
				ctx:               newctx,
				cancel:            cancel,
				eventChan:         eventchan,
				msgbusChan:        make(chan *msgbus.Event),
				openChan:          make(chan *SimpleConnOpenEvent),
				connectionMapping: map[ConnUniqueID]*lumerinconnection.LumerinSocketStruct{},
				ConnectionStruct:  connectionStruct,
			}

			s.accept <- newSimpleStruct
		}
	}

	contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" exit")
}

//
//
//
func (s *SimpleListenStruct) GetAccept() <-chan *SimpleStruct {
	return s.accept
}

// Calls the listen context cancel function, which closes out the listener routine
func (s *SimpleListenStruct) Close() {
	s.cancel()
}

func (s *SimpleStruct) GetEventChan() <-chan *SimpleEvent {

	if s.eventChan == nil {
		contextlib.Logf(s.Ctx(), contextlib.LevelPanic, lumerinlib.FileLineFunc()+" EventChan is nil SimpleStruct:%v", s)
	}

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
	contextlib.Logf(s.Ctx(), contextlib.LevelInfo, lumerinlib.FileLineFunc()+" enter")

	if s == nil {
		panic(lumerinlib.FileLineFunc() + " SimpleStruct is nil")
	}
	if s.ConnectionStruct == nil {
		panic(lumerinlib.FileLineFunc() + " SimpleStruct.ConnectionStruct is nil")
	}

	// Just checking for good measure
	cs := contextlib.GetContextStruct(s.ctx)
	if cs == nil {
		contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" Context Structre not correct")
	}

	go s.goEvent()
}

func (s *SimpleStruct) goEvent() {
	contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" enter ")

	//
	// Using connection managers index as the UniqueID (for now?)
	//

	readchan := s.ConnectionStruct.GetReadChan()
	openchan := s.openChan
	msgbuschan := s.msgbusChan

FORLOOP:
	for {
		select {
		case <-s.Ctx().Done():
			contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" Closing down")
			break FORLOOP
		case comm := <-readchan:
			contextlib.Logf(s.ctx, contextlib.LevelInfo, lumerinlib.FileLineFunc()+" ReadChan Event ")

			if comm == nil {
				contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" readchan returned nil")
				break FORLOOP
			}
			//
			// Grossly inefficient... will fix later...
			scre := &SimpleConnReadEvent{}
			scre.uniqueID = ConnUniqueID(comm.Index())
			scre.data = comm.Data()
			scre.count = comm.Count()
			scre.err = comm.Err()

			ev := &SimpleEvent{
				EventType:     ConnReadEvent,
				ConnReadEvent: scre,
				MsgBusEvent:   nil,
				ConnOpenEvent: nil,
			}
			s.eventChan <- ev

		case open := <-openchan:
			contextlib.Logf(s.ctx, contextlib.LevelInfo, lumerinlib.FileLineFunc()+" OpenChan Event ")

			ev := &SimpleEvent{
				EventType:     ConnOpenEvent,
				ConnOpenEvent: open,
				ConnReadEvent: nil,
				MsgBusEvent:   nil,
			}

			s.eventChan <- ev
		case msg := <-msgbuschan:
			contextlib.Logf(s.ctx, contextlib.LevelInfo, lumerinlib.FileLineFunc()+" MsgBusChan Event ")

			if msg == nil {
				contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" msgbuschan returned nil")
				break FORLOOP
			}
			smbe := &SimpleMsgBusEvent{}
			smbe.Data = msg.Data
			smbe.Err = msg.Err
			smbe.EventType = msg.EventType
			smbe.ID = msg.ID
			smbe.RequestID = msg.RequestID

			ev := &SimpleEvent{
				EventType:     MsgBusEvent,
				MsgBusEvent:   smbe,
				ConnReadEvent: nil,
				ConnOpenEvent: nil,
			}
			s.eventChan <- ev
		}
	}

	contextlib.Logf(s.ctx, contextlib.LevelWarn, lumerinlib.FileLineFunc()+" exit")
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
// func (s *SimpleStruct) SetBuffer() {}

// Set message buffering to a certain delimiter, for example a newline character: ‘\n’
// for stage 1 this will assumed to be unconfigurable and only use '\n' as the
// new line
//func (s *SimpleStruct) SetMessageDelimiterDefault() {}

// Set message buffering to be of a certain size
//func (s *SimpleStruct) SetMessageSizeDefault(mSize uint) {
//	s.maxMessageSize = mSize
//}

// TODO not part of stage 1
// Set encryption parameters
//func (s *SimpleStruct) SetEncryptionDefault() {}

// TODO not part of stage 1
// Set Compression parameters
// func (s *SimpleStruct) SetCompressionDefault() {}

/*
Dial the a destination address (DST)
takes in a net.Addr object and feeds into the net.Dial function
the resulting Conn is then added to the SimpleStructs mapping and and associated
ConnUniqueID is returned from this function

id is the calling ID info, the uid is returned from the connectionmanager layer, and is used to index the connection
*/
func (s *SimpleStruct) AsyncDial(id int, dst net.Addr) error {

	contextlib.Logf(s.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")

	if s == nil {
		return fmt.Errorf(lumerinlib.FileLineFunc() + " SimpleStruct == nil ")
	}
	if s.ConnectionStruct == nil {
		return fmt.Errorf(lumerinlib.FileLineFunc() + " SimpleStruct.ConnectionStruct == nil ")
	}

	go func() {

		uid, e := s.ConnectionStruct.Dial(dst)

		open := &SimpleConnOpenEvent{
			requestID: id,
			uniqueID:  ConnUniqueID(uid),
			err:       e,
		}

		s.openChan <- open
	}()

	// conn, err := lumerinconnection.Dial(s.ctx, dst) //creates a new net.Conn object
	//gets the current index value and asssigns to connection in mapping
	// var uID ConnUniqueID = s.connectionIndex
	// s.connectionMapping[uID] = conn
	// s.connectionIndex++ //increase the connectionIndex for the next time a conn is made
	//consider a mapping of connections and UID's
	// return uID, err
	return nil
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
func (s *SimpleStruct) SetRoute(u ConnUniqueID) error {
	if u < 0 {
		contextlib.Logf(s.ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" index out of range:%d", u)
	}
	return s.ConnectionStruct.SetRoute(int(u))
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
func (s *SimpleStruct) Write(i ConnUniqueID, msg []byte) (count int, e error) {
	if i < 0 {
		count, e = s.ConnectionStruct.SrcWrite(msg)
		// Need to so some error checking here, if src is closed, then the whole thing needs to be shutdown.
		if e != nil {
			contextlib.Logf(s.ctx, contextlib.LevelError, lumerinlib.FileLineFunc()+" ConnectionStruct.SrcRead() error:%s", e)
		}

	} else {
		count, e = s.ConnectionStruct.IdxWrite(int(i), msg)
	}

	return count, e
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
