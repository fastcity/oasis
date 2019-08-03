package models

type Transaction struct {
	BlockHeight string `json:"height" bson:"blockHeight"`
	BlockTime   string `json:"blockTime" bson:"blockTime"`
	BlockHash   string `json:"blockHash" bson:"blockHash"`
	Txid        string `json:"txid" bson:"txid"`
	Method      string `json:"method" bson:"method"`
	TxHash      string `json:"txHash" bson:"txHash"`
	From        string `json:"from" bson:"from"`
	To          string `json:"to" bson:"to"`
	Value       string `json:"value" bson:"value"`
	TokenKey    string `json:"tokenKey" bson:"tokenKey"`
	OnChain     bool   `json:"onChain" bson:"onChain"`
	Log         string `json:"log" bson:"log"`
}

// 		"Height": "1",
// 		"Hash": "Local",
// 		"TimeStamp": "2019-02-14 19:57:12.1294157 +0800 CST m=+18.015939401",
// 		"Transactions": [
// 				{
// 				  "Height": "1",
// 				  "TxID": "78629A0F5F3F164F1583390EB8263C3C",
// 				  "Chaincode": "local",
// 				  "Method": "TOKEN.ASSIGN",
// 				  "CreatedFlag": false,
// 				  "ChaincodeModule": "AtomicEnergy_v1",
// 				  "Nonce": "B991CAF3783E7CFA43ABBF3A60D8D27314E3CB76",
// 				  "Detail": {
// 					"amount": "400000",
// 					"to": "ARJtq6Q46oTnxDwvVqMgDtZeNxs7Ybt81A"
// 				  },
// 				  "TxHash": "990B78AE548E3CB8B8D389A1371E7ECE8316A44878CDB4AD9DAD003329E47CD7"
// 				}
// 		]
// 		"TxEvents": [
// 		  {
// 			"TxID": "8866CB397916001E158368A7E2329318",
// 			"Chaincode": "local",
// 			"Name": "INVOKEERROR",
// 			"Status": 1,
// 			"Detail": "Local invoke error: handling method [MTOKEN.INIT] fail: Can not re-deploy existed data"
// 		  }
// 		]
