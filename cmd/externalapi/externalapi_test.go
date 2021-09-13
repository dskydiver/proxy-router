package externalapi

import (
	"testing"
	"time"

	"gitlab.com/TitanInd/lumerin/cmd/externalapi/msgdata"
	"gitlab.com/TitanInd/lumerin/cmd/msgbus"
)

func TestMsgBusDataAddedToApiRepos(t *testing.T) {
	ech := make(msgbus.EventChan)
	ps := msgbus.New(1)

	dest := msgbus.Dest{
		ID:   "DestID01",
		IP:   "127.0.0.1",
		Port: 80,
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
		//NewContracts:           make(map[contract.ID]bool),
		//ReadyContracts:         make(map[ContractID]bool),
		//ActiveContracts:        make(map[ContractID]bool),
	}
	contract := msgbus.Contract{
		ID:						"ContractID01",
		State: 					msgbus.ContActiveState,
		Buyer: 					"Buyer ID01",
		Dest:					"DestID01",
		CommitedHashRate: 		9000,		
		TargetHashRate:   		10000,
		CurrentHashRate:		8000,
		Tolerance:				10,
		Penalty:				100,
		Priority:				1,
		StartDate:				time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:				time.Date(2021, 10, 31, 0, 0, 0, 0, time.UTC),
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

	configRepo, connectionRepo, contractRepo, destRepo, minerRepo, sellerRepo := InitializeJSONRepos()

	go func(ech msgbus.EventChan) {
		for e := range ech {
			switch e.Msg {
			case msgbus.ConfigMsg:
				configRepo.AddConfigInfoFromMsgBus(config)
				if configRepo.ConfigInfoJSONs[0].ID != "ConfigID01" {
					t.Errorf("Failed to add Config to Repo")
				} 
			case msgbus.DestMsg:
				destRepo.AddDestFromMsgBus(dest)
				if destRepo.DestJSONs[0].ID != "DestID01" {
					t.Errorf("Failed to add Dest to Repo")
				} 
			case msgbus.SellerMsg:
				sellerRepo.AddSellerFromMsgBus(seller)
				if sellerRepo.SellerJSONs[0].ID != "SellerID01" {
					t.Errorf("Failed to add Seller to Repo")
				} 
			case msgbus.ContractMsg:
				contractRepo.AddContractFromMsgBus(contract)
				if contractRepo.ContractJSONs[0].ID != "ContractID01" {
					t.Errorf("Failed to add Contract to Repo")
				} 
			case msgbus.MinerMsg:
				minerRepo.AddMinerFromMsgBus(miner)
				if minerRepo.MinerJSONs[0].ID != "MinerID01" {
					t.Errorf("Failed to add Miner to Repo")
				} 
			case msgbus.ConnectionMsg:
				connectionRepo.AddConnectionFromMsgBus(connection)
				if connectionRepo.ConnectionJSONs[0].ID != "ConnectionID01" {
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
	destJSON := msgdata.DestJSON {
		ID:   "DestID01",
		IP:   "127.0.0.1",
		Port: 80,
	}
	configJSON := msgdata.ConfigInfoJSON{
		ID:          "ConfigID01",
		DefaultDest: "DestID01",
		Seller:      "SellerID01",
	}
	connectionJSON := msgdata.ConnectionJSON{
		ID:        				"ConnectionID01",
		Miner:    				"MinerID01",
		Dest:      				"DestID01",
		State:     				"0",
		TotalHash: 				10000,
		StartDate: 				"80000",
	}
	contractJSON := msgdata.ContractJSON{
		ID:						"ContractID01",
		State: 					"0",
		Buyer: 					"Buyer ID01",
		Dest:					"DestID01",
		CommitedHashRate: 		9000,		
		TargetHashRate:   		10000,
		CurrentHashRate:		8000,
		Tolerance:				10,
		Penalty:				100,
		Priority:				1,
		StartDate:				"80000",
		EndDate:				"90000",
	}

	minerJSON := msgdata.MinerJSON{
		ID:						"MinerID01",
		State: 					"0",
		Seller:   				"SellerID01",
		Dest:					"DestID01",	
		InitialMeasuredHashRate: 10000,
		CurrentHashRate:         9000,
	}
	sellerJSON := msgdata.SellerJSON{
		ID:                     "SellerID01",
		DefaultDest:            "DestID01",
		TotalAvailableHashRate: 1000,
		UnusedHashRate:         100,
		//NewContracts:           make(map[contract.ID]bool),
		//ReadyContracts:         make(map[ContractID]bool),
		//ActiveContracts:        make(map[ContractID]bool),
	}
	
	ech := make(msgbus.EventChan)
	ps := msgbus.New(1)

	configRepo, connectionRepo, contractRepo, destRepo, minerRepo, sellerRepo := InitializeJSONRepos()

	configRepo.AddConfigInfo(configJSON)
	connectionRepo.AddConnection(connectionJSON)
	contractRepo.AddContract(contractJSON)
	destRepo.AddDest(destJSON)
	minerRepo.AddMiner(minerJSON)
	sellerRepo.AddSeller(sellerJSON)

	var ConfigMSG msgbus.ConfigInfo
	var ConnectionMSG msgbus.Connection
	var ContractMSG msgbus.Contract
	var DestMSG msgbus.Dest
	var MinerMSG msgbus.Miner
	var SellerMSG msgbus.Seller

	configMSG := msgdata.ConvertConfigInfoJSONtoConfigInfoMSG(configRepo.ConfigInfoJSONs[0], ConfigMSG)
	connectionMSG := msgdata.ConvertConnectionJSONtoConnectionMSG(connectionRepo.ConnectionJSONs[0], ConnectionMSG)
	contractMSG := msgdata.ConvertContractJSONtoContractMSG(contractRepo.ContractJSONs[0], ContractMSG)
	destMSG := msgdata.ConvertDestJSONtoDestMSG(destRepo.DestJSONs[0], DestMSG)
	minerMSG := msgdata.ConvertMinerJSONtoMinerMSG(minerRepo.MinerJSONs[0], MinerMSG)
	sellerMSG := msgdata.ConvertSellerJSONtoSellerMSG(sellerRepo.SellerJSONs[0], SellerMSG)
	
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