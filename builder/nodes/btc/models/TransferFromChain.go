package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type TransferFromChain struct {
	ID          primitive.ObjectID       `bson:"_id"`
	RequestId   string                   `bson:"requestId"`
	Chain       string                   `bson:"chain"`
	Coin        string                   `bson:"coin"`
	TokenKey    string                   `bson:"tokenKey"`
	From        interface{}              `bson:"from"`
	To          interface{}              `bson:"to"`
	Value       string                   `bson:"value"`
	Txid        string                   `bson:"txid"`
	BlockHeight int64                    `bson:"blockHeight"`
	BlockTime   int                      `bson:"blockTime"`
	Vins        []map[string]interface{} `json:"vin" bson:"vin"`
	Vouts       []map[string]interface{} `json:"vout" bson:"vout"`
	CreatedAt   primitive.DateTime       `bson:"createdAt"`
	UpdatedAt   string                   `bson:"updatedAt"`
}
