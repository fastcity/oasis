package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type TransferFromChain struct {
	// BlockHeight string `json:"height"`
	// BlockTime   string `json:"blockTime"`
	// BlockHash   string `json:"blockHash"`
	// Txid        string `json:"txid"`
	// Method      string `json:"method"`
	// TxHash      string `json:"txHash"`
	// From        string `json:"from"`
	// To          string `json:"to"`
	// Value       string `json:"value"`
	// TokenKey    string `json:"token"`
	// OnChain     bool   `json:"onChain"`
	// Log         string `json:"log"`
	ID          primitive.ObjectID `bson:"_id"`
	RequestId   string             `bson:"-"`
	Chain       string             `bson:"-"`
	Coin        string             `bson:"-"`
	TokenKey    string             `bson:"tokenkey"`
	From        string             `bson:"from"`
	To          string             `bson:"to"`
	Value       string             `bson:"value"`
	Txid        string             `bson:"txid"`
	BlockHeight string             `bson:"blockheight"`
	BlockTime   string             `bson:"blocktime"`
	CreatedAt   int64              `bson:"createdAt"`
	UpdatedAt   int64              `bson:"updatedAt"`
	// createdAt: { type: Number, default: Date.now },
	// updatedAt: { type: Number, default: Date.now },
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