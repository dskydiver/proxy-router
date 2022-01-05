package externalapi

import (
	"encoding/json"
	"testing"
	"time"

	"gitlab.com/TitanInd/lumerin/cmd/configurationmanager"
	"gitlab.com/TitanInd/lumerin/cmd/externalapi/msgdata"
	"gitlab.com/TitanInd/lumerin/cmd/msgbus"
)

func TestMsgBusDataAddedToApiRepos(t *testing.T) {
	ech := make(msgbus.EventChan)
	ps := msgbus.New(1)

	dest := msgbus.Dest{
		ID:   		"DestID01",
		NetUrl: 	"stratum+tcp://127.0.0.1:3334/",	
	}
	config := msgbus.ConfigInfo{
		ID:          "ConfigID01",
		DefaultDest: "DestID01",
		Seller:      "SellerID01",
	}
	seller := msgbus.Seller{
		ID:                     "SellerID01",
		DefaultDest:            "DestID01",
		TotalAvailableHashRate: 0,
		UnusedHashRate:         0,
	}
	seller.Contracts = map[msgbus.ContractID]msgbus.ContractState{
		"0x50937C047DB93CB5C87F65B6EFFEA47D03DF0F7D": msgbus.ContRunningState,
        "0xFB610E4C269DA110C97B92F5F34EAA50E5F3D500": msgbus.ContAvailableState,
        "0x397729E80F77BA09D930FE24E8D1FC74372E86D3": msgbus.ContAvailableState,
	}  
	contract := msgbus.Contract{
		ID:				"ContractID01",
		State: 			msgbus.ContRunningState,
		Buyer: 			"Buyer ID01",
		Price: 			100,
		Limit: 			100,
		Speed: 			100,
		Length: 		100,
		StartingBlockTimestamp: 100,
	}
	
	miner := msgbus.Miner{
		ID:						"MinerID01",
		State: 					msgbus.OnlineState,
		Seller:   				"SellerID01",
		Dest:					"DestID01",	
		InitialMeasuredHashRate: 10000,
		CurrentHashRate:         9000,

	}
	connection := msgbus.Connection{
		ID:        				"ConnectionID01",
		Miner:    				"MinerID01",
		Dest:      				"DestID01",
		State:     				msgbus.ConnAuthState,
		TotalHash: 				10000,
		StartDate: 				time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC),
	}

	var api APIRepos
	api.InitializeJSONRepos()

	go func(ech msgbus.EventChan) {
		for e := range ech {
			switch e.Msg {
			case msgbus.ConfigMsg:
				api.Config.AddConfigInfoFromMsgBus(config)
				if api.Config.ConfigInfoJSONs[0].ID != "ConfigID01" {
					t.Errorf("Failed to add Config to Repo")
				} 
			case msgbus.DestMsg:
				api.Dest.AddDestFromMsgBus(dest)
				if api.Dest.DestJSONs[0].ID != "DestID01" {
					t.Errorf("Failed to add Dest to Repo")
				} 
			case msgbus.SellerMsg:
				api.Seller.AddSellerFromMsgBus(seller)
				if api.Seller.SellerJSONs[0].ID != "SellerID01" {
					t.Errorf("Failed to add Seller to Repo")
				} 
			case msgbus.ContractMsg:
				api.Contract.AddContractFromMsgBus(contract)
				if api.Contract.ContractJSONs[0].ID != "ContractID01" {
					t.Errorf("Failed to add Contract to Repo")
				} 
			case msgbus.MinerMsg:
				api.Miner.AddMinerFromMsgBus(miner)
				if api.Miner.MinerJSONs[0].ID != "MinerID01" {
					t.Errorf("Failed to add Miner to Repo")
				} 
			case msgbus.ConnectionMsg:
				api.Connection.AddConnectionFromMsgBus(connection)
				if api.Connection.ConnectionJSONs[0].ID != "ConnectionID01" {
					t.Errorf("Failed to add Connection to Repo")
				} 
			default:
			
			} 
		}
	}(ech)

	ps.Pub(msgbus.ConfigMsg, "configMsg01", msgbus.ConfigInfo{})
	ps.Pub(msgbus.DestMsg, "destMsg01", msgbus.Dest{})
	ps.Pub(msgbus.SellerMsg, "sellerMsg01", msgbus.Seller{})
	ps.Pub(msgbus.ContractMsg, "contractMsg01", msgbus.Contract{})
	ps.Pub(msgbus.MinerMsg, "minerMsg01", msgbus.Miner{})
	ps.Pub(msgbus.ConnectionMsg, "connectionMsg01", msgbus.Connection{})

	ps.Sub(msgbus.ConfigMsg, "configMsg01", ech)
	ps.Sub(msgbus.DestMsg, "destMsg01", ech)
	ps.Sub(msgbus.SellerMsg, "sellerMsg01", ech)
	ps.Sub(msgbus.ContractMsg, "contractMsg01", ech)
	ps.Sub(msgbus.MinerMsg, "minerMsg01", ech)
	ps.Sub(msgbus.ConnectionMsg, "connectionMsg01", ech)

	ps.Set(msgbus.ConfigMsg, "configMsg01", config)
	ps.Set(msgbus.DestMsg, "destMsg01", dest)
	ps.Set(msgbus.SellerMsg, "sellerMsg01", seller)
	ps.Set(msgbus.ContractMsg, "contractMsg01", contract)
	ps.Set(msgbus.MinerMsg, "minerMsg01", miner)
	ps.Set(msgbus.ConnectionMsg, "connectionMsg01", connection)
}

func TestMockPOSTAddedToMsgBus(t *testing.T) {	
	// Mock POST Requests by declaring new JSON structures and adding them to api repos
	eaConfig,err := configurationmanager.LoadConfiguration("../configurationmanager/testconfig.json", "externalAPI")
	if err != nil {
		t.Errorf("LoadConfiguration returned error")
	}

	dest := eaConfig["dest"].(map[string]interface{})
	destMarshaled,_ := json.Marshal(dest)
	destJSON := msgdata.DestJSON {}
	json.Unmarshal(destMarshaled, &destJSON)

	config := eaConfig["config"].(map[string]interface{})
	configMarshaled,_ := json.Marshal(config)
	configJSON := msgdata.ConfigInfoJSON {}
	json.Unmarshal(configMarshaled, &configJSON)

	connection := eaConfig["connection"].(map[string]interface{})
	connectionMarshaled,_ := json.Marshal(connection)
	connectionJSON := msgdata.ConnectionJSON {}
	json.Unmarshal(connectionMarshaled, &connectionJSON)

	contract := eaConfig["contract"].(map[string]interface{})
	contractMarshaled,_ := json.Marshal(contract)
	contractJSON := msgdata.ContractJSON {}
	json.Unmarshal(contractMarshaled, &contractJSON)

	miner := eaConfig["miner"].(map[string]interface{})
	minerMarshaled,_ := json.Marshal(miner)
	minerJSON := msgdata.MinerJSON {}
	json.Unmarshal(minerMarshaled, &minerJSON)

	seller := eaConfig["seller"].(map[string]interface{})
	sellerMarshaled,_ := json.Marshal(seller)
	sellerJSON := msgdata.SellerJSON {}
	json.Unmarshal(sellerMarshaled, &sellerJSON)
	
	ech := make(msgbus.EventChan)
	ps := msgbus.New(1)

	var api APIRepos
	api.InitializeJSONRepos()

	api.Config.AddConfigInfo(configJSON)
	api.Connection.AddConnection(connectionJSON)
	api.Contract.AddContract(contractJSON)
	api.Dest.AddDest(destJSON)
	api.Miner.AddMiner(minerJSON)
	api.Seller.AddSeller(sellerJSON)

	var ConfigMSG msgbus.ConfigInfo
	var ConnectionMSG msgbus.Connection
	var ContractMSG msgbus.Contract
	var DestMSG msgbus.Dest
	var MinerMSG msgbus.Miner
	var SellerMSG msgbus.Seller

	configMSG := msgdata.ConvertConfigInfoJSONtoConfigInfoMSG(api.Config.ConfigInfoJSONs[0], ConfigMSG)
	connectionMSG := msgdata.ConvertConnectionJSONtoConnectionMSG(api.Connection.ConnectionJSONs[0], ConnectionMSG)
	contractMSG := msgdata.ConvertContractJSONtoContractMSG(api.Contract.ContractJSONs[0], ContractMSG)
	destMSG := msgdata.ConvertDestJSONtoDestMSG(api.Dest.DestJSONs[0], DestMSG)
	minerMSG := msgdata.ConvertMinerJSONtoMinerMSG(api.Miner.MinerJSONs[0], MinerMSG)
	sellerMSG := msgdata.ConvertSellerJSONtoSellerMSG(api.Seller.SellerJSONs[0], SellerMSG)
	
	go func(ech msgbus.EventChan) {
		for e := range ech {
			if e.EventType == msgbus.GetEvent {
				switch e.Msg {
				case msgbus.ConfigMsg:
					if e.Data == nil {
						t.Errorf("Failed to add Config to message bus")
					} 
				case msgbus.DestMsg:
					if e.Data == nil {
						t.Errorf("Failed to add Dest to message bus")
					} 
				case msgbus.SellerMsg:
					if e.Data == nil {
						t.Errorf("Failed to add Seller to message bus")
					} 
				case msgbus.ContractMsg:
					if e.Data == nil {
						t.Errorf("Failed to add Contract to message bus")
					} 
				case msgbus.MinerMsg:
					if e.Data == nil {
						t.Errorf("Failed to add Miner to message bus")
					} 
				case msgbus.ConnectionMsg:
					if e.Data == nil {
						t.Errorf("Failed to add Connection to message bus")
					} 
				default:
				
				} 
			}
		}
	}(ech)

	ps.Pub(msgbus.ConfigMsg, msgbus.IDString(configMSG.ID), msgbus.ConfigInfo{})
	ps.Pub(msgbus.DestMsg, msgbus.IDString(destMSG.ID), msgbus.Dest{})
	ps.Pub(msgbus.SellerMsg, msgbus.IDString(sellerMSG.ID), msgbus.Seller{})
	ps.Pub(msgbus.ContractMsg, msgbus.IDString(contractMSG.ID), msgbus.Contract{})
	ps.Pub(msgbus.MinerMsg, msgbus.IDString(minerMSG.ID), msgbus.Miner{})
	ps.Pub(msgbus.ConnectionMsg, msgbus.IDString(connectionMSG.ID), msgbus.Connection{})

	ps.Sub(msgbus.ConfigMsg, msgbus.IDString(configMSG.ID), ech)
	ps.Sub(msgbus.DestMsg, msgbus.IDString(destMSG.ID), ech)
	ps.Sub(msgbus.SellerMsg, msgbus.IDString(sellerMSG.ID), ech)
	ps.Sub(msgbus.ContractMsg, msgbus.IDString(contractMSG.ID), ech)
	ps.Sub(msgbus.MinerMsg, msgbus.IDString(minerMSG.ID), ech)
	ps.Sub(msgbus.ConnectionMsg, msgbus.IDString(connectionMSG.ID), ech)

	ps.Set(msgbus.ConfigMsg, msgbus.IDString(configMSG.ID), configMSG)
	ps.Set(msgbus.DestMsg, msgbus.IDString(destMSG.ID), destMSG)
	ps.Set(msgbus.SellerMsg, msgbus.IDString(sellerMSG.ID), sellerMSG)
	ps.Set(msgbus.ContractMsg, msgbus.IDString(contractMSG.ID), contractMSG)
	ps.Set(msgbus.MinerMsg, msgbus.IDString(minerMSG.ID), minerMSG)
	ps.Set(msgbus.ConnectionMsg, msgbus.IDString(connectionMSG.ID), connectionMSG)

	ps.Get(msgbus.ConfigMsg, msgbus.IDString(configMSG.ID), ech)
	ps.Get(msgbus.DestMsg, msgbus.IDString(destMSG.ID), ech)
	ps.Get(msgbus.SellerMsg, msgbus.IDString(sellerMSG.ID), ech)
	ps.Get(msgbus.ContractMsg, msgbus.IDString(contractMSG.ID), ech)
	ps.Get(msgbus.MinerMsg, msgbus.IDString(minerMSG.ID), ech)
	ps.Get(msgbus.ConnectionMsg, msgbus.IDString(connectionMSG.ID), ech)
}