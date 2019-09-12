package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type TransferFromChain struct {
	ID          primitive.ObjectID `bson:"_id"`
	RequestId   string             `bson:"requestId"`
	Chain       string             `bson:"chain"`
	Coin        string             `bson:"coin"`
	BlockHeight int64              `json:"height" bson:"blockHeight"`
	BlockTime   primitive.DateTime `json:"blockTime" bson:"blockTime"`
	BlockHash   string             `json:"blockHash" bson:"blockHash"`
	Nonce       string             `json:"nonce"  bson:"nonce"`
	Txid        string             `json:"hash" bson:"txid"`
	From        string             `json:"from" bson:"from"`
	To          string             `json:"to" bson:"to"`
	Value       string             `json:"value" bson:"value"`
	Gas         string             `json:"gas"  bson:"gas"`
	GasPrice    string             `json:"gasPrice"  bson:"gasPrice"`
	GasUsed     string             `json:"gasUsed"  bson:"gasUsed"`
	Fee         string             `json:"fee"  bson:"fee"`
	TokenKey    string             `bson:"tokenKey"`
	Status      string             `json:"status" bson:"status"`
	CreatedAt   primitive.DateTime `bson:"createdAt"`
	UpdatedAt   string             `bson:"updatedAt"`
}
