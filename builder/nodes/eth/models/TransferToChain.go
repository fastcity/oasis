package models

import (
	"math/big"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransferToChain struct {
	ID          primitive.ObjectID `bson:"_id"`
	RequestId   string             `bson:"-"`
	Chain       string             `bson:"-"`
	Coin        string             `bson:"-"`
	TokenKey    string             `bson:"-"`
	From        string             `bson:"from"`
	To          string             `bson:"to"`
	Value       string             `bson:"value"`
	Txid        string             `bson:"txid"`
	Fee         *big.Int           `bson:"fee"`
	BlockHeight int64              `bson:"blockheight"`
	BlockTime   string             `bson:"blocktime"`
	CreatedAt   int64              `bson:"createdAt"`
	UpdatedAt   int64              `bson:"updatedAt"`
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
