package stratumv1

//
// ToDo
// Move GetRandomIDString to be getRandomIDString containing it in this module
//
//
//
//

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gitlab.com/TitanInd/lumerin/lumerinlib"
)

type stratumMethods string
type stratumStates string
type stratumErrors string

const (
	SErrNull           stratumErrors = "null"
	SErrUnknown        stratumErrors = "-1"
	SErrSrvNotFound    stratumErrors = "-2"
	SErrMethodNotFound stratumErrors = "-3"
	SErrFeeReq         stratumErrors = "-10"
	SErrSigReq         stratumErrors = "-20"
	SErrSigUnavail     stratumErrors = "-21"
	SErrUnkSigTyp      stratumErrors = "-22"
	SErrBadSig         stratumErrors = "-23"
)

const (
	StratumNew        stratumStates = "StratumNew"
	StratumSubscribed stratumStates = "StratumSubscribed"
	StratumAuthorized stratumStates = "StratumAuthorized"
	StratumMsgError   stratumStates = "StratumMsgError"
)

const (
	CLIENT_MINING_AUTHORIZE          stratumMethods = "mining.authorize"
	CLIENT_MINING_CAPABILITIES       stratumMethods = "mining.capabilities"
	CLIENT_MINING_EXTRANONCE         stratumMethods = "mining.extranonce.subscribe"
	CLIENT_MINING_SUBMIT             stratumMethods = "mining.submit"
	CLIENT_MINING_SUBSCRIBE          stratumMethods = "mining.subscribe"
	CLIENT_MINING_SUGGEST_DIFFICULTY stratumMethods = "mining.suggest_difficulty"
	CLIENT_MINING_SUGGEST_TARGET     stratumMethods = "mining.suggest_target"
	MINING_CONFIGURE                 stratumMethods = "mining.configure"
	MINING_SET_TARGET                stratumMethods = "mining.set_target"
	SERVER_GET_VERSION               stratumMethods = "client.get_version"
	SERVER_RECONNECT                 stratumMethods = "client.reconnect"
	SERVER_SHOW_MESSAGE              stratumMethods = "client.show_message"
	SERVER_MINING_NOTIFY             stratumMethods = "mining.notify"
	SERVER_MINING_PING               stratumMethods = "mining.ping"
	SERVER_MINING_SET_DIFFICULTY     stratumMethods = "mining.set_difficulty"
	SERVER_MINING_SET_EXTRANONCE     stratumMethods = "mining.set_extranonce"
	SERVER_MINING_SET_GOAL           stratumMethods = "mining.set_goal"
)

//
// Used for recieving incoming stratum JSON  messages
//
type StratumMsgStruct struct {
	ID      interface{} `json:"id,omitempty"`
	Jsonrpc interface{} `json:"jsonrpc,omitempty"`
	Method  interface{} `json:"method,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Reject  interface{} `json:"reject-reason,omitempty"`
}

//
// Used to build outgoing JSON message
//
type stratumRequest struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// notice ID is always null
type stratumNotice struct {
	ID     *string     `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type noticeMiningSetDifficulty struct {
	ID     *string `json:"id"`
	Method string  `json:"method"`
	Params []int   `json:"params"`
}

type noticeMiningNotifyParams struct {
	jobID          string
	prevBlockHash  string
	gen1           string
	gen2           string
	merkelBranches []string
	blockVersion   string
	nBits          string
	nTime          string
	CleanJob       bool
}
type noticeMiningNotify struct {
	ID      *string                  `json:"id"`
	Jsonrpc *string                  `json:"jsonrpc"`
	Method  string                   `json:"method"`
	Params  noticeMiningNotifyParams `json:"params"`
}

type stratumResponse struct {
	ID     int         `json:"id"`
	Error  *string     `json:"error"`
	Result interface{} `json:"result"`
	Reject interface{} `json:"reject-reason,omitempty"`
}

//
// unmarshalMsg() take []byte and translate it into a StratumMsgStruct
//
func unmarshalMsg(b []byte) (ret interface{}, err error) {

	msg := StratumMsgStruct{}
	j := map[string]interface{}{}

	err = json.Unmarshal(b, &j)

	if err == nil {
		for key, value := range j {
			switch key {
			case "id":
				switch vtype := value.(type) {
				case float32:
					msg.ID = int(value.(float32))
				case float64:
					msg.ID = int(value.(float64))
				case int:
					msg.ID = value.(int)
				case string:
					msg.ID = -1
				case nil:
					msg.ID = nil
				default:
					panic(fmt.Sprintf("Value Type: %t", vtype))
				}

			case "method":
				switch vtype := value.(type) {
				case string:
					msg.Method = value.(string)
				case nil:
					msg.Method = ""
				default:
					panic(fmt.Sprintf("Value Type: %t", vtype))
				}

			case "error":
				switch vtype := value.(type) {
				case string:
					msg.Error = value.(string)
				case nil:
					msg.Error = nil
				case interface{}:
					msg.Error = fmt.Sprintf(" Error: %f, %s", vtype.([]interface{})[0], vtype.([]interface{})[1])
				default:
					panic(fmt.Sprintf("Value Type: %t", vtype))
				}

			case "reject-reason":
				msg.Reject = value

			case "params":
				msg.Params = value
				//msg.Params = make([]interface{}, 0)
				//for _, v := range value.([]interface{}) {
				//	msg.Params = append(msg.Params, v)
				//}

			case "result":
				msg.Result = value

			case "jsonrpc":
				msg.Jsonrpc = value

			default:
				panic(fmt.Sprintf("Key Value: %s", key))
			}
		}

		// Is this a Response Msg?
		if msg.Result != nil {
			r := stratumResponse{}
			r.ID = msg.ID.(int)
			r.Result = msg.Result
			if msg.Error == nil {
				r.Error = nil
			} else {
				e := msg.Error.(string)
				r.Error = &e
			}
			if msg.Reject == nil {
				r.Reject = nil
			} else {
				r.Reject = msg.Reject
			}
			ret = &r

			// Is this a Notice?
		} else if msg.ID == nil {
			ret = &stratumNotice{
				ID:     nil,
				Method: msg.Method.(string),
				Params: msg.Params,
			}

		} else {
			// Must be a Request
			r := &stratumRequest{
				ID:     msg.ID.(int),
				Method: msg.Method.(string),
				Params: make([]interface{}, 0),
			}

			switch msg.Method.(string) {
			case string(SERVER_MINING_NOTIFY):
				r.Params = msg.Params.([]interface{})

				//for _, v := range msg.Params.([]interface{}) {
				//	switch v.(type) {
				//	case string:
				//		r.Params = append(r.Params, v.(string))
				//	case []interface{}:
				//		z := make([]string, 1)
				//		for _, u := range v.([]string) {
				//			z = append(z, u)
				//		}
				//		r.Params = append(r.Params, z)
				//	default:
				//		panic(fmt.Sprintf(lumerinlib.FileLineFunc()+" Error bad type:%T\n", v))
				//	}
				//}

			case string(SERVER_MINING_SET_DIFFICULTY):
				for _, v := range msg.Params.([]interface{}) {
					switch v.(type) {
					case string:
						r.Params = append(r.Params, v.(string))
					case float32:
						r.Params = append(r.Params, fmt.Sprintf("%f", v.(float32)))
					case float64:
						r.Params = append(r.Params, fmt.Sprintf("%f", v.(float64)))
					default:
						panic(fmt.Sprintf(lumerinlib.FileLineFunc()+" Error bad type:%T\n", v))
					}
				}

			default:

				for _, v := range msg.Params.([]interface{}) {
					r.Params = append(r.Params, v.(interface{}))
				}
			}

			ret = r
		}

	} else {
		fmt.Printf(lumerinlib.FileLineFunc()+"Error unmarshaling msg:%s\n", err)
	}

	return ret, err
}

//
//
//
func getStratumMsg(msg []byte) (ret interface{}, err error) {

	ret, err = unmarshalMsg(msg)

	return ret, err
}

// ---------------------------------------------------------------------------------------
//                   *StratumRequest
// ---------------------------------------------------------------------------------------

//------------------------------------------------------
//
// {"id":2,"method":"mining.authorize","params":["testrig.worker1",""]}
//------------------------------------------------------
func (r *stratumRequest) getID() (id int, err error) {

	switch r.Method {
	case string(CLIENT_MINING_AUTHORIZE):
	case string(CLIENT_MINING_CAPABILITIES):
	case string(CLIENT_MINING_EXTRANONCE):
	case string(CLIENT_MINING_SUBMIT):
	case string(CLIENT_MINING_SUBSCRIBE):
	case string(CLIENT_MINING_SUGGEST_DIFFICULTY):
	case string(CLIENT_MINING_SUGGEST_TARGET):
	default:
		return 0, fmt.Errorf(lumerinlib.FileLineFunc()+" wrong method, got: %s", r.Method)
	}

	id = r.ID

	return id, err
}

//------------------------------------------------------
//
// {"id":2,"method":"mining.authorize","params":["testrig.worker1",""]}
//------------------------------------------------------
func (r *stratumRequest) getAuthName() (name string, err error) {

	if r.Method != string(CLIENT_MINING_AUTHORIZE) {
		return "", fmt.Errorf(lumerinlib.FileLineFunc()+" wrong method, expetecting mining.authorize, got: %s", r.Method)
	}

	// fmt.Printf(" type:%T", r.Params)

	// name = r.Params[0]
	name = r.Params[0].(string)

	return name, err
}

//------------------------------------------------------
//
// {"id":0,"jsonrpc":"2.0","method":"mining.set_difficulty","params":[65535]}
//------------------------------------------------------
func (r *stratumRequest) getSetDifficulty() (difficulty float64, err error) {

	difficulty = 0.0

	if r.Method != string(SERVER_MINING_SET_DIFFICULTY) {
		err = fmt.Errorf(lumerinlib.FileLineFunc()+" wrong method, expetecting mining.set_difficulty, got: %s", r.Method)
	} else {

		switch t := r.Params[0].(type) {
		case string:
			if s, err := strconv.ParseFloat(r.Params[0].(string), 64); err == nil {
				difficulty = s
			}
		case float32:
			difficulty = r.Params[0].(float64)
		case float64:
			difficulty = r.Params[0].(float64)
		default:
			err = fmt.Errorf(lumerinlib.FileLineFunc()+" Error bad type:%T\n", t)
		}
	}

	return difficulty, err
}

//------------------------------------------------------
//
// {
//	"id": null,
//	"method": "mining.notify",
//	"params": [
//		"bf",  -- JOB ID
//		"4d16b6f85af6e2198f44ae2a6de67f78487ae5611b77c6c0440b921e00000000", -- HEX-ENCODED PREV BLOCK HASH
// 		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008", -- HEX-ENCODED PREFIX
//		"072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000", -- HEX-ENCODED SUFFIX
//		[], -- MERKEL ROOT
//		"00000002", -- HEX-ENCODED BLOCK VERSION
//		"1c2ac4af", -- HEX-ENCODED NETWORK DIFFICULTY REQUIRED
//		"504e86b9", -- HEX-ENCODED CURRENT TIME FOR THE BLOCK
//		false --
//	]
// }
//------------------------------------------------------
func (r *stratumRequest) getMiningNotifyJobID() (jobid string, err error) {

	if r.Method != string(SERVER_MINING_NOTIFY) {
		err = fmt.Errorf(lumerinlib.FileLineFunc()+" wrong method, expetecting mining.notify, got: %s", r.Method)
	} else {
		jobid = r.Params[0].(string)
	}

	return jobid, err

}

//------------------------------------------------------
//
//------------------------------------------------------
func (r *stratumRequest) createRequestMsg() (msg []byte, err error) {

	err = nil
	// fmt.Printf("Create Stratum Request: %v\n", r)

	msg, err = json.Marshal(r)
	if err != nil {
		fmt.Printf(lumerinlib.FileLineFunc()+"Error Marshaling Request Err:%s\n", err)
		return nil, err
	}

	return msg, err
}

//------------------------------------------------------
//
//------------------------------------------------------
func (r *stratumRequest) createNoticeMiningNotifyStruct() (nsd noticeMiningNotify, err error) {

	fmt.Printf(lumerinlib.Funcname()+": %v\n", r)

	err = nil

	// nsd.ID = r.ID
	nsd.Method = r.Method

	params := r.Params

	if len(params) != 9 {
		panic(fmt.Sprintf(lumerinlib.FileLineFunc() + " Not enough parameter for Notify\n"))
	}

	for i, v := range params {
		switch i {
		case 0:
			nsd.Params.jobID = v.(string)
		case 1:
			nsd.Params.prevBlockHash = v.(string)
		case 2:
			nsd.Params.gen1 = v.(string)
			fallthrough
		case 3:
			nsd.Params.gen2 = v.(string)

		case 4:
			nsd.Params.merkelBranches = make([]string, 0)
			if len(v.([]interface{})) > 0 {
				for _, w := range v.([]interface{}) {
					nsd.Params.merkelBranches = append(nsd.Params.merkelBranches, w.(string))
				}
			}

		case 5:
			nsd.Params.blockVersion = v.(string)
		case 6:
			nsd.Params.nBits = v.(string)
		case 7:
			nsd.Params.nTime = v.(string)
		case 8:
			nsd.Params.CleanJob = (v == "true")
		}
	}

	return nsd, err
}

// ---------------------------------------------------------------------------------------
//                   *StratumNotice
// ---------------------------------------------------------------------------------------

//------------------------------------------------------
//
// -->> {"id":0,"jsonrpc":"2.0","method":"mining.set_difficulty","params":[65535]}
//------------------------------------------------------
func (n *stratumNotice) getSetDifficulty() (difficulty float64, err error) {

	difficulty = 0.0

	if n.Method != string(SERVER_MINING_SET_DIFFICULTY) {
		err = fmt.Errorf(lumerinlib.FileLineFunc()+" wrong method, expetecting mining.set_difficulty, got: %s", n.Method)
	} else {

		switch t := n.Params.(type) {
		case string:
			if s, err := strconv.ParseFloat(n.Params.(string), 64); err == nil {
				difficulty = s
			}
		case int:
			difficulty = n.Params.(float64)
		case float32:
			difficulty = n.Params.(float64)
		case float64:
			difficulty = n.Params.(float64)
			// This is what is used.
		case interface{}:
			v := n.Params
			arr := v.([]interface{})
			difficulty = arr[0].(float64)
		default:
			err = fmt.Errorf(lumerinlib.FileLineFunc()+" Error bad type:%T\n", t)
		}
	}

	return difficulty, err
}

//------------------------------------------------------
//
// {
//	"id": null,
//	"method": "mining.notify",
//	"params": [
//		"bf",  -- JOB ID
//		"4d16b6f85af6e2198f44ae2a6de67f78487ae5611b77c6c0440b921e00000000", -- HEX-ENCODED PREV BLOCK HASH
// 		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff20020862062f503253482f04b8864e5008", -- HEX-ENCODED PREFIX
//		"072f736c7573682f000000000100f2052a010000001976a914d23fcdf86f7e756a64a7a9688ef9903327048ed988ac00000000", -- HEX-ENCODED SUFFIX
//		[], -- MERKEL ROOT
//		"00000002", -- HEX-ENCODED BLOCK VERSION
//		"1c2ac4af", -- HEX-ENCODED NETWORK DIFFICULTY REQUIRED
//		"504e86b9", -- HEX-ENCODED CURRENT TIME FOR THE BLOCK
//		false --
//	]
// }
//------------------------------------------------------
func (n *stratumNotice) getMiningNotifyJobID() (jobid string, err error) {

	if n.Method != string(SERVER_MINING_NOTIFY) {
		err = fmt.Errorf(lumerinlib.FileLineFunc()+" wrong method, expetecting mining.notify, got: %s", n.Method)
	} else {

		switch t := n.Params.(type) {
		case string:
			jobid = "string"
		case interface{}:
			v := n.Params
			arr := v.([]interface{})
			jobid = arr[0].(string)
		default:
			err = fmt.Errorf(lumerinlib.FileLineFunc()+" Error bad type:%T\n", t)
		}
	}

	return jobid, err

}

//------------------------------------------------------
//
//------------------------------------------------------
func (n *stratumNotice) createNoticeMsg() (msg []byte, err error) {

	err = nil
	// fmt.Printf("Create Stratum Notice: %v\n", n)

	switch n.Method {
	case string(SERVER_MINING_SET_DIFFICULTY):
		msg, err = n.createNoticeSetDifficultyMsg()
	case string(SERVER_MINING_NOTIFY):
		msg, err = n.createNoticeMiningNotify()
	default:
		msg, err = json.Marshal(n)
	}

	if err != nil {
		fmt.Printf(lumerinlib.FileLineFunc()+"Error Marshaling Request Err:%s\n", err)
		return nil, err
	}

	return msg, err
}

//------------------------------------------------------
//
//------------------------------------------------------
func (n *stratumNotice) createNoticeSetDifficultyMsg() (msg []byte, err error) {

	// fmt.Printf(lumerinlib.Funcname()+": %v\n", n)

	err = nil

	var nsd noticeMiningSetDifficulty
	nsd.ID = n.ID
	nsd.Method = n.Method
	nsd.Params = make([]int, 0)

	for _, v := range n.Params.([]interface{}) {
		i := int(v.(float64))
		if err != nil {
			panic("")
		}
		nsd.Params = append(nsd.Params, i)
	}

	msg, err = json.Marshal(nsd)
	if err != nil {
		fmt.Printf(lumerinlib.FileLineFunc()+"Error Marshaling Request Err:%s\n", err)
		return nil, err
	}

	return msg, err
}

//------------------------------------------------------
//
// {
//   "params": [
// #0(string)   "613a0f04000001bc",
// #1(string)   "36847fbbe629819b9c0e23ddb4a80e68339e1b420002630c0000000000000000",
// #2(string)   "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff360394ad0a0004f14e3a61046ff4f4050c",
// #3(string)   "0a636b706f6f6c122f6d696e6564206279204c756d6572696e2fffffffff031cda8225000000001976a91422ddd9233f44ac2e9f183ec755adf134c12cdbf188ac0000000000000000266a24aa21a9ed852376e1fca95e42b3b1c080c3dc14b0db71c1b683511b0037b091c6f28acab96a413000000000001976a91422ddd9233f44ac2e9f183ec755adf134c12cdbf188ac00000000",
// #4([]string) [
// (string)      "773418c442067fdd5c3caf10653537041db14d13249cab724d9d892d8427a66a",
// (string)      "4126854f7bd3dc91bf666f53c35930685ee245239242ced1254f43e7b51b97e2",
// (string)      "d89213f7501f4f6123c5d24403801b7d978957e9ecbee82869fefb295025caff",
// (string)      "b4817f2f1e86914186c5acf715db97f753b84b9cc2cbd3a977e021df09ccf46d",
// (string)      "51c91bbfb65e328063dbfe020913a5e92c2973796f7cd84c74806e33eaf48116",
// (string)      "6e006d18ed55017612adf0e334b94d52e16b06f11adb14058a91caee161a304f",
// (string)      "633c5a641b57c0fc0fc9ed669d04686634f17ff34b6d509cc9a50c58e7cd9771",
// (string)      "e90773f4f44dc4a6a13e60956cad1612549e5c23a8f4ba42e760eb8661177464",
// (string)      "de5fc02be1faa3dbbb59e9799ea1fae886ab25e6b154413d2e2d35204fedbaf2",
// (string)      "79b109bdf26dd068446afa66c62f7d5ba30b179fcf032bb299f5a2591e0e3fce",
// (string)      "fe16f0630558f6564ec212ed700b1d5469b0a9d1cd39f4b7ce344d3d01d650b7",
// (string)      "03802c6be8643a09f8f74254ebf6f3704cfc622ab55f94687299fc32ca4a31da"
//              ],
// #5(string)   "20000000",
// #6(string)   "170f48e4",
// #7(string)   "613a4ee8",
// #8(bool)     true
//   ],
//   "id": null,
//   "method": "mining.notify"
// }
//
//
//------------------------------------------------------
func (n *stratumNotice) createNoticeMiningNotify() (msg []byte, err error) {

	err = nil

	var nsd noticeMiningNotify
	nsd.ID = n.ID
	nsd.Method = n.Method

	if len(n.Params.([]interface{})) != 9 {
		panic("")
	}

	// nsd.Params = make(noticeMiningNotifyParams)

	// nsd.Params = append(nsd.Params, a)
	for i, v := range n.Params.([]interface{}) {
		switch i {
		case 0:
			nsd.Params.jobID = v.(string)
		case 1:
			nsd.Params.prevBlockHash = v.(string)
		case 2:
			nsd.Params.gen1 = v.(string)
			fallthrough
		case 3:
			nsd.Params.gen2 = v.(string)

		case 4:
			nsd.Params.merkelBranches = make([]string, 0)
			if len(v.([]interface{})) > 0 {
				for _, w := range v.([]interface{}) {
					nsd.Params.merkelBranches = append(nsd.Params.merkelBranches, w.(string))
				}
			}

		case 5:
			nsd.Params.blockVersion = v.(string)
		case 6:
			nsd.Params.nBits = v.(string)
		case 7:
			nsd.Params.nTime = v.(string)
		case 8:
			nsd.Params.CleanJob = (v == "true")
		}
	}

	msg, err = json.Marshal(nsd)
	if err != nil {
		fmt.Printf(lumerinlib.FileLineFunc()+"Error Marshaling Request Err:%s\n", err)
		return nil, err
	}

	return msg, err
}

//------------------------------------------------------
//
//------------------------------------------------------
func (n *stratumNotice) createNoticeMiningNotifyStruct() (nsd noticeMiningNotify, err error) {

	fmt.Printf(lumerinlib.Funcname()+": %v\n", n)

	err = nil

	nsd.Method = n.Method

	params := n.Params.([]interface{})

	if len(params) != 9 {
		panic(fmt.Sprintf(lumerinlib.FileLineFunc() + " Not enough parameter for Notify\n"))
	}

	for i, v := range params {
		switch i {
		case 0:
			nsd.Params.jobID = v.(string)
		case 1:
			nsd.Params.prevBlockHash = v.(string)
		case 2:
			nsd.Params.gen1 = v.(string)
			fallthrough
		case 3:
			nsd.Params.gen2 = v.(string)

		case 4:
			nsd.Params.merkelBranches = make([]string, 0)
			if len(v.([]interface{})) > 0 {
				for _, w := range v.([]interface{}) {
					nsd.Params.merkelBranches = append(nsd.Params.merkelBranches, w.(string))
				}
			}

		case 5:
			nsd.Params.blockVersion = v.(string)
		case 6:
			nsd.Params.nBits = v.(string)
		case 7:
			nsd.Params.nTime = v.(string)
		case 8:
			nsd.Params.CleanJob = (v == "true")
		}
	}

	return nsd, err
}

// ---------------------------------------------------------------------------------------
//                   *StratumResponse
// ---------------------------------------------------------------------------------------

//------------------------------------------------------
//
//------------------------------------------------------
func (r *stratumResponse) getAuthResult() (ret bool, err error) {

	_, ok := r.Result.(bool)
	if !ok {
		err = fmt.Errorf(lumerinlib.FileLineFunc()+" result is wrong type:%T", r.Result)
	} else {
		ret = r.Result.(bool)
	}

	return ret, err
}

//------------------------------------------------------
//
//------------------------------------------------------
func (r *stratumResponse) createResponseMsg() (msg []byte, err error) {

	err = nil
	// fmt.Printf("Create Stratum Response: %v\n", r)

	msg, err = json.Marshal(r)
	if err != nil {
		fmt.Printf(lumerinlib.FileLineFunc()+"Error Marshaling Response Err:%s\n", err)
		return nil, err
	}

	return msg, err
}
