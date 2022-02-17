package simple

import (
	"context" //this can probably be removed once gitlab packages can be imported
	"fmt"
	"net"
	_ "time"
	//the below packages need to have their gitlab branches sorted out prior to being
	//imported via go mod tidy
	//_ "gitlab.com/TitanInd/lumerin/cmd/lumerinnetwork/lumerinconnection"
	//double check that these imports were formatted correctly
	//_ "gitlab.com/TitanInd/lumerin/cmd/config"
	//_ "gitlab.com/TitanInd/lumerin/cmd/msgbus"
	//_ "gitlab.com/TitanInd/lumerin/lumerinlib"
)

/*
The purpose of the simple layer is to accept any messages from the
protocol layer and pass messages to;
lower down the stack
to the message bus
It is also designed to return messages from the msg bus to the protocol
layer.
Refer to proxy router document section 7.7.2.2 (update to reflect location in confluence)
*/

type messageAction string
type actionResponse string

//
//
//
//
//
//
//the following structs are temporary
//
//
//
//
//

//
// Listen Struct for new SRC connections coming in
//
type ConnectionListenStruct struct {
	listen *LumerinListenStruct
	ctx    context.Context
	port   int
	ip     net.IPAddr
	//  cancel func()
}

//
// Struct for existing SRC connections and the associated outgoing DST connections
type ConnectionStruct struct {
	src    *LumerinSocketStruct
	dst    []*LumerinSocketStruct
	defidx int
	ctx    context.Context
	cancel func()
}

//
// This will contain a regular socket or virtual socket structure
//
type LumerinListenStruct struct {
	listener interface{}
}

type LumerinSocketStruct struct {
	socket interface{}
}

//
//
//
//
//
// end temporary structs
//
//
//
//
//

/*
MsgDeque is a last in first out datastructue which can accept
messages of any struct type and in constantly processed
*/
type SIMPLE struct {
	ProtocolChan   chan ProtocolMessage
	MSGChan        chan MSGBusMessage
	ConnectionChan chan ConnectionMessage
	done           chan string
	MsgDeque       []workerStruct
}

type workerStruct struct {
	msg    []byte
	action uint
}

//struct to handle/accept messages from the layer 1 channel
type ProtocolMessage struct {
	WorkerName      string
	MessageContents []byte
	MessageActions  []uint
}

// struct to handle/accept messages from the message bus
type MSGBusMessage struct {
	WorkerName      string
	MessageContents []byte
	MessageActions  []uint
}

// struct to handle messages from further down in the stack
type ConnectionMessage struct {
	WorkerName      string
	MessageContents []byte
	MessageActions  []uint
}

//define available actions
const (
	//constants to define requested incoming messages
	HashSubmit    messageAction = "HashSubmit"
	HashrateCount messageAction = "HashrateCount"
)

//define available return messages
const (
	HashValid     actionResponse = "HashValid"
	HashInvalid   actionResponse = "HashInvalid"
	HashrateValue actionResponse = "HashrateValue"
)

//this is a public function, being provided to convert a worker struct into
//a connection struct
/*
func WorkerStructToConnectionStruct(w workerStruct) (ConnectionStruct) {
}
*/

//this takes a ProtocolMessage struct and moves the data into a workerStruct
func (pm *ProtocolMessage) Actions(x uint) workerStruct {
	return workerStruct{
		msg:    pm.MessageContents,
		action: x,
	}
}

//this takes a ProtocolMessage struct and moves the data into a workerStruct
func (pm *ProtocolMessage) Message() []byte {
		return pm.MessageContents
}

//this takes a MSGBusMessage struct and moves the data into a workerStruct
func (mm *MSGBusMessage) Actions(x uint) workerStruct {
	return workerStruct{
		msg:    mm.MessageContents,
		action: x,
	}
}

//this takes a ProtocolMessage struct and moves the data into a workerStruct
func (pm *MSGBusMessage) Message() []byte {
		return pm.MessageContents
}

//this takes a ConnectionMessage struct and moves the data into a workerStruct
func (lm *ConnectionMessage) Actions(x uint) workerStruct {
	return workerStruct{
		msg:    lm.MessageContents,
		action: x,
	}
}

//this takes a ProtocolMessage struct and moves the data into a workerStruct
func (pm *ConnectionMessage) Message() []byte {
		return pm.MessageContents
}

// takes the byte array destined for the protocol layer and unmarshals it into a ProtocolMessage struct
// then it pushes the ProtocolMessage onto the ProtocolChan
func (s *SIMPLE) msgToProtocol(b []byte) {
	//create an in-memory temporary struct to pass to the ProtocolChan
	tempStruct := ProtocolMessage{
		WorkerName:      "",       //this field will probably be removed
		MessageContents: b,        //msg content to be passed back from msg
		MessageActions:  []uint{}, //empty array to keep compiler happy
	}
	//pass the struct to the protocol chan
	s.ProtocolChan <- tempStruct
}

// takes the byte array destined for the protocol layer and unmarshals it into a MSGBusMessage struct
// then it pushes the MSGBusMessage onto the MSGChan
func (s *SIMPLE) msgToMSGBus(b []byte) {
	//create an in-memory temporary struct to pass to the MSGChan
	tempStruct := MSGBusMessage{
		WorkerName:      "",       //this field will probably be removed
		MessageContents: b,        //msg content to be passed back from msg
		MessageActions:  []uint{}, //empty array to keep compiler happy
	}
	//pass the struct to the protocol chan
	s.MSGChan <- tempStruct
}

// takes the byte array destined for the protocol layer and unmarshals it into a ConnectionMessage struct
// then it pushes the ConnectionMessage onto the ConnectionChan
func (s *SIMPLE) msgToConnection(b []byte) {
	//create an in-memory temporary struct to pass to the ConnectionChan
	tempStruct := ConnectionMessage{
		WorkerName:      "",       //this field will probably be removed
		MessageContents: b,        //msg content to be passed back from msg
		MessageActions:  []uint{}, //empty array to keep compiler happy
	}
	//pass the struct to the protocol chan
	s.ConnectionChan <- tempStruct
}

type StandardMessager interface {
	Actions() []string
	Message() []byte
}

//function to constantly monitor MsgDeque and process the last item on it
func (s *SIMPLE) ActivateSIMPLELayer() {
	for {
		if len(s.MsgDeque) > 0 {
			//msg is the last element in the msg deque and is processed
			//newDeque is to rewrite the MsgDeque in lieu of another popping method
			msg := s.MsgDeque[0]
			newDeque := s.MsgDeque[1:]
			s.processIncomingMessage(msg)
			s.MsgDeque = newDeque

		}
	}
}

//listens for messages coming in through the various channels
//oldest item will always be index 0
func (s *SIMPLE) ListenForIncomingMessages() {
	for {
		select {
		case pc := <-s.ProtocolChan:
			for _, x := range pc.MessageActions {
				s.MsgDeque = append(s.MsgDeque, pc.Actions(x))
			}
		case mc := <-s.MSGChan:
			for _, x := range mc.MessageActions {
				s.MsgDeque = append(s.MsgDeque, mc.Actions(x))
			}
		case lc := <-s.ConnectionChan:
			for _, x := range lc.MessageActions {
				s.MsgDeque = append(s.MsgDeque, lc.Actions(x))
			}
		case <-s.done:
			return
		}
	}
}

/*
this function is where the majority of the work for the SIMPLE layer will be done
Each message coming in will have a [task] field which tells the SIMPLE layer
how to process the message. The idea here is that anybody can create a [task] and
associated function and add to the processing request.
Rules to follow
1. this is a function, so for every input there's only 1 output
2. do not break the interface of the output
3. design functions in a maintainable manner
*/
func (s *SIMPLE) processIncomingMessage(m workerStruct) {
	switch m.action {
	case 0: //route message to protocol channel
		s.msgToProtocol(m.msg)
	case 1: //route message to msgbus channel
		s.msgToMSGBus(m.msg)
	case 2: //route message to connection channel
		s.msgToConnection(m.msg)
	default:
		fmt.Println("lord bogdanoff demands elon tank the price of dogecoin")
	}
}

/*
create and return a struct with channels to listen to
call goroutine embedded in the struct
*/
func New() SIMPLE {
	var deque []workerStruct
	return SIMPLE{
		ProtocolChan:   make(chan ProtocolMessage),
		MSGChan:        make(chan MSGBusMessage),
		ConnectionChan: make(chan ConnectionMessage),
		done:           make(chan string),
		MsgDeque:       deque,
	}
}

func (s *SIMPLE) Close() {
	close(s.done)
	close(s.ProtocolChan)
	close(s.MSGChan)
	close(s.ConnectionChan)
}

//create a listener for the msg bus
