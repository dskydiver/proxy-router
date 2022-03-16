package protocol

import (
	"context"
	"fmt"
	"net"

	simple "gitlab.com/TitanInd/lumerin/cmd/lumerinnetwork/SIMPL"
	"gitlab.com/TitanInd/lumerin/lumerinlib"
	contextlib "gitlab.com/TitanInd/lumerin/lumerinlib/context"
)

//
// Top layer protocol template functions that a new protocol will use to access the SIMPLe layer
//

type ProtocolListenStruct struct {
	ctx          context.Context
	cancel       func()
	simplelisten *simple.SimpleListenStruct
}

type ProtocolStruct struct {
	ctx       context.Context
	cancel    func()
	simple    *simple.SimpleStruct
	eventchan chan *simple.SimpleEvent
	srcconn   *ProtocolConnectionStruct
	dstconn   *ProtocolDstStruct
	msgbus    *ProtocolMsgBusStruct
}

//
// NewListen() Create a new ProtocolListenStruct
// Opens the default destination
//
func NewListen(ctx context.Context) (pls *ProtocolListenStruct, e error) {
	var ok bool

	contextlib.Logf(ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")

	//
	// Basic error checking, make sure that the ContextStruct is
	// filled out correctly
	//
	c := ctx.Value(contextlib.ContextKey)
	if c == nil {
		contextlib.Logf(ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" called")
	}

	cs, ok := c.(*contextlib.ContextStruct)
	if !ok {
		contextlib.Logf(ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" Context Structre not correct")
	}
	if cs.GetProtocol() == nil {
		cs.Logf(contextlib.LevelPanic, "Context Protocol not defined")
	}
	if cs.GetMsgBus() == nil {
		cs.Logf(contextlib.LevelPanic, "Context MsgBus not defined")
	}
	if cs.GetSrc() == nil {
		cs.Logf(contextlib.LevelPanic, "Context Src Addr not defined")
	}
	if cs.GetDst() == nil {
		cs.Logf(contextlib.LevelPanic, "Context Dst Addr not defined")
	}

	ctx, cancel := context.WithCancel(ctx)

	listenaddr := contextlib.GetSrc(ctx)
	sls, err := simple.New(ctx, listenaddr)
	sls.Run()

	if err != nil {
		lumerinlib.PanicHere(fmt.Sprintf("Error:%s", err))
	}

	pls = &ProtocolListenStruct{
		ctx:          ctx,
		cancel:       cancel,
		simplelisten: &sls,
	}

	return pls, e
}

//
// Ctx() returns the current context of the ProtocolListenStruct
//
func (p *ProtocolListenStruct) Ctx() context.Context {
	return p.ctx
}

//
// Cancel() calls the simple layer Cancel function on the SimpleListenStruct
//
func (pls *ProtocolListenStruct) Cancel() {

	contextlib.Logf(pls.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")

	pls.cancel()
}

//
// Run() calls the simple layer Run function on the SimpleListenStruct
//
func (pls *ProtocolListenStruct) Run() {

	contextlib.Logf(pls.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")

	go pls.goAccept()
}

//
// goAccept()
//
func (pls *ProtocolListenStruct) goAccept() {

	contextlib.Logf(pls.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")

	go func() {
		for {
			select {
			case <-pls.ctx.Done():
				return
			case newSimpleStruct := <-pls.simplelisten.Accept():
				contextlib.Logf(pls.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" simplelisten.Accept() recieved")

				ps, e := NewProtocol(newSimpleStruct)
				if e != nil {
					contextlib.Logf(pls.ctx, contextlib.LevelPanic, lumerinlib.FileLineFunc()+" NewProtocol() error:%s", e)
				}
				ps.Run()
			}
		}
	}()
}

// --------------------------------------------
// ProtocolStruct functions
//

//
// NewProtocol() takes a simple struct and creates a ProtocolStruct, pulls the Src and Dst from the context
// and initiates a connection to the defualt Dst address
// This function is called from the layer above to initalize the common protocol functions, and enable
// access to the standard functions provided by this layer.
//
func NewProtocol(s *simple.SimpleStruct) (ps *ProtocolStruct, e error) {

	contextlib.Logf(s.Ctx(), contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")

	cs := contextlib.GetContextStruct(s.Ctx())
	src := cs.GetSrc()
	dst := cs.GetDst()

	ctx, cancel := context.WithCancel(s.Ctx())
	eventchan := make(chan *simple.SimpleEvent)

	s.SetEventChan(eventchan)

	ps = &ProtocolStruct{
		ctx:       ctx,
		cancel:    cancel,
		simple:    s,
		eventchan: eventchan,
		srcconn: &ProtocolConnectionStruct{
			Addr: src,
			Id:   0,
		},
		dstconn: &ProtocolDstStruct{
			conn: make(map[int]*ProtocolConnectionStruct),
		},
		msgbus: &ProtocolMsgBusStruct{},
	}

	// Open the default connection
	// What happens if the Default Destination does not take the connection?
	//
	index, e := ps.OpenConn(dst)
	// First connection should be zero
	if e != nil {
		contextlib.Logf(s.Ctx(), contextlib.LevelError, lumerinlib.FileLineFunc()+" openConn(dst) returned error:%s", e)
		return nil, e
	}
	if index != 0 {
		contextlib.Logf(s.Ctx(), contextlib.LevelError, lumerinlib.FileLineFunc()+" wrong index returned")
		return nil, fmt.Errorf("OpenConn returnd index == %d", index)
	}

	s.Run()

	e = ps.SetDefaultRoute(index)
	if e != nil {
		contextlib.Logf(s.Ctx(), contextlib.LevelPanic, lumerinlib.FileLineFunc()+" SetDefaultRoute(index) returned error:%s", e)

	}

	return ps, e
}

//
// Ctx() returns the context of the ProtocolStruct
//
func (ps *ProtocolStruct) Ctx() context.Context {
	if ps == nil {
		panic(lumerinlib.FileLineFunc() + " ProtocolStruct nil ProtocolStruct pointer")
	}
	return ps.ctx
}

//
// Cancel() calls the simple layer Cancel function on the SimpleListenStruct
//
func (ps *ProtocolStruct) Cancel() {
	contextlib.Logf(ps.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")
	ps.cancel()
}

//
// Run() calls the simple layer Run function on the SimpleListenStruct
//
func (ps *ProtocolStruct) Run() {
	contextlib.Logf(ps.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")
	ps.simple.Run()
}

//
// Run() calls the simple layer Run function on the SimpleListenStruct
//
func (ps *ProtocolStruct) GetSimpleEvent() <-chan *simple.SimpleEvent {
	contextlib.Logf(ps.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")
	return ps.simple.GetEvent()
}

//
//
//
//func (ps *ProtocolStruct) Event() chan *simple.SimpleEvent {
//
//	if ps == nil {
//		panic(lumerinlib.FileLineFunc() + " empty pointer")
//	}
//
//	contextlib.Logf(ps.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")
//
//	return ps.eventchan
//}

//
// OpenConn()
// opens a new connection to the desitnation and returns the index of it
//
func (ps *ProtocolStruct) OpenConn(dst net.Addr) (index int, e error) {

	if ps == nil {
		panic(lumerinlib.FileLineFunc() + "ProtocolStruct is nil")
	}
	if ps.simple == nil {
		contextlib.Logf(ps.ctx, contextlib.LevelError, lumerinlib.FileLineFunc()+" simple struct is nil")
		return -2, fmt.Errorf(lumerinlib.FileLineFunc() + " ProtoclStruct.SimpleStruct is nil")
	}
	if ps.simple.ConnectionStruct == nil {
		contextlib.Logf(ps.ctx, contextlib.LevelError, lumerinlib.FileLineFunc()+" simple struct CoonectionStruct is nil")
		return -2, fmt.Errorf(lumerinlib.FileLineFunc() + " ProtoclStruct.SimpleStruct.ConnectionStruct is nil")
	}

	contextlib.Logf(ps.ctx, contextlib.LevelTrace, lumerinlib.FileLineFunc()+" called")

	//
	// Call simple layer to dial up a connection to the dst
	//
	id, e := ps.simple.Dial(dst)
	if e != nil {
		return -1, e
	}

	//
	// Add new connection to the ProtocolDstStruct
	//
	return ps.dstconn.addConn(dst, id)
}

//
// SetDefaultRoute()
// Set the SIMPL layer default route
//
func (ps *ProtocolStruct) SetDefaultRoute(index int) (e error) {

	id, e := ps.dstconn.getConnID(index)
	if e != nil {
		return e
	}

	// Set the default route to the first route
	// e := ps.simple.SetRoute(id)
	// if e != nil {}

	ps.simple.SetRoute(id)

	return nil
}

//
// GetDefaultRoute()
// get the  SIMPL layer default route
//
func (ps *ProtocolStruct) GetDefaultRoute() (index int, e error) {

	// Set the default route to the first route
	ps.simple.GetRoute()
	var id int = 0
	//id, e := ps.simple.GetRoute()
	//if e != nil {
	//	return -1, e
	//}

	index, e = ps.dstconn.getConnIndex(id)

	return index, e

}

//
// Write()
//
func (ps *ProtocolStruct) Write(msg []byte) (count int, e error) {
	count = 0
	// l := len(msg)

	index, e := ps.GetDefaultRoute()
	if e != nil {
		return -1, e
	}

	id, e := ps.dstconn.getConnID(index)
	if e != nil {
		return -1, e
	}

	ps.simple.Write(id, msg)

	// count, e = ps.simple.Write(id, msg)
	// if e != nil {
	// 	return count, e
	// }
	// if l != count {
	// 	return count, fmt.Errorf(lumerinlib.FileLineFunc() + " full msg length was not written")
	// }

	return count, e
}

//
// WriteSrc()
//
func (ps *ProtocolStruct) WriteSrc(msg []byte) (count int, e error) {
	count = 0

	return count, e
}

//
// WriteDst()
//
func (ps *ProtocolStruct) WriteDst(index int, msg []byte) (count int, e error) {
	count = 0

	return count, e
}

//
// Pub() publishes data, and stores the request ID to match the Completion Event
//
func (ps *ProtocolStruct) Pub(msgtype simple.MsgType, id simple.ID, data simple.Data) (rID int, e error) {

	// rID, e = ps.simple.Pub(msgtype, id, data)
	e = ps.simple.Pub(msgtype, id, data)

	return 0, e
}

//
//
//
func (ps *ProtocolStruct) Unpub(msgtype simple.MsgType, id simple.ID) (rID int, e error) {

	return 0, nil
}

//
//
//
func (ps *ProtocolStruct) Sub(msgtype simple.MsgType, id simple.ID, eh func()) (rID int, e error) {

	return 0, nil
}

//
//
//
func (ps *ProtocolStruct) Unsub(msgtype simple.MsgType, id simple.ID, eh func()) (rID int, e error) {

	return 0, nil
}

//
//
//
func (ps *ProtocolStruct) Get(msgtype simple.MsgType, id simple.ID, eh func()) (rID int, e error) {

	return 0, nil
}

//
//
//
func (ps *ProtocolStruct) Set(msgtype simple.MsgType, id simple.ID, data interface{}) (rID int, e error) {

	return 0, nil
}

//
//
//
func (ps *ProtocolStruct) SearchIP(msgtype simple.MsgType, search string) (rID int, e error) {

	return 0, nil
}

//
//
//
func (ps *ProtocolStruct) SearchMac(msgtype simple.MsgType, search string) (rID int, e error) {

	return 0, nil
}

//
//
//
func (ps *ProtocolStruct) SearchName(msgtype simple.MsgType, search string) (rID int, e error) {

	return 0, nil
}
