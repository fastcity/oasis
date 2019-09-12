package models

import (
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {

	// "blockHash": "0x6ba2f988f1e6354710be622fc0f1155ed689e41e4ad04a36af0c1fbc2d020256",
	// "blockNumber": "0x4d30d2",
	// "contractAddress": "0xf4be8e28450fff2f938d6f8ee2af485a83a91a2e",
	// "cumulativeGasUsed": "0x1eff08",
	// "from": "0x558150acf40d522eaca490390e89fcb83e95c314",
	// "gasUsed": "0x1eff08",
	// "logs": [
	//     {
	//         "address": "0xf4be8e28450fff2f938d6f8ee2af485a83a91a2e",
	//         "blockHash": "0x6ba2f988f1e6354710be622fc0f1155ed689e41e4ad04a36af0c1fbc2d020256",
	//         "blockNumber": "0x4d30d2",
	//         "data": "0x",
	//         "logIndex": "0x0",
	//         "removed": false,
	//         "topics": [
	//             "0x6ae172837ea30b801fbfcdd4108aa1d5bf8ff775444fd70256b44e6bf3dfc3f6",
	//             "0x000000000000000000000000558150acf40d522eaca490390e89fcb83e95c314"
	//         ],
	//         "transactionHash": "0xd1092f37b28f82c77808d4de1eec14436d2f7592869b1be627b0f399b6eb7fe0",
	//         "transactionIndex": "0x0"
	//     },
	//     {
	//         "address": "0xf4be8e28450fff2f938d6f8ee2af485a83a91a2e",
	//         "blockHash": "0x6ba2f988f1e6354710be622fc0f1155ed689e41e4ad04a36af0c1fbc2d020256",
	//         "blockNumber": "0x4d30d2",
	//         "data": "0x00000000000000000000000000000000000000000052b7d2dcc80cd2e4000000",
	//         "logIndex": "0x1",
	//         "removed": false,
	//         "topics": [
	//             "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
	//             "0x0000000000000000000000000000000000000000000000000000000000000000",
	//             "0x000000000000000000000000558150acf40d522eaca490390e89fcb83e95c314"
	//         ],
	//         "transactionHash": "0xd1092f37b28f82c77808d4de1eec14436d2f7592869b1be627b0f399b6eb7fe0",
	//         "transactionIndex": "0x0"
	//     }
	// ],
	// "logsBloom": "0x00800200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000008040000000000000000040000000000000000000000000000020000000000000000000c00000000008000000000000010000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000020000000000000000000000000000000000100000000000000000000000000000000",
	// "status": "0x1",
	// "to": null,
	// "transactionHash": "0xd1092f37b28f82c77808d4de1eec14436d2f7592869b1be627b0f399b6eb7fe0",
	// "transactionIndex": "0x0"

	// "blockHash": "0x6ba2f988f1e6354710be622fc0f1155ed689e41e4ad04a36af0c1fbc2d020256",
	// "blockNumber": "0x4d30d2",
	// "from": "0x7b7840a7643359d759d9510ade7c9211cc1c727b",
	// "gas": "0xdc36",
	// "gasPrice": "0x37e11d600",
	// "hash": "0x912b5d9f7305dfbb40e855274bf36186696ed8c5a261ae8aa96a5ae3daa69500",
	// "input": "0xa9059cbb0000000000000000000000005d69231952d8d8d7e23e53612e83292993e7da4d000000000000000000000000000000000000000000000000000000003b9aca00",
	// "nonce": "0x64",
	// "r": "0xd2c98bf1dd69762cdd9847f7d56608964d1cbd0c2e105d7b026613b52f1ac786",
	// "s": "0x7f3a9f26d4163ace2aaaf7a364414d1573e883b049f8b5a9c823c7d08d85d757",
	// "to": "0x25b581fb3ee9feb8b6ccc50d07fadfd71c7ed529",
	// "transactionIndex": "0x1",
	// "v": "0x1c",
	// "value": "0x0"

	BlockHeight      int64                    `json:"blockNumber" bson:"blockHeight"`
	BlockTime        primitive.DateTime       `json:"blockTime" bson:"blockTime"`
	BlockHash        string                   `json:"blockHash" bson:"blockHash"`
	Txid             string                   `json:"hash" bson:"txid"`
	From             string                   `json:"from" bson:"from"`
	To               string                   `json:"to" bson:"to"`
	Value            int64                    `json:"value" bson:"value"`
	TokenKey         string                   `json:"contractAddress" bason:"tokenKey"`
	Gas              string                   `json:"gas"  bson:"gas"`
	GasPrice         string                   `json:"gasPrice"  bson:"gasPrice"`
	GasUsed          int64                    `json:"gasUsed"  bson:"gasUsed"`
	Nonce            int64                    `json:"nonce"  bson:"nonce"`
	Input            string                   `json:"input"  bson:"input"`
	TransactionIndex string                   `json:"transactionIndex"  bson:"transactionIndex"`
	Logs             []map[string]interface{} `json:"logs" bson:"logs"`
	Status           string                   `json:"status" bson:"status"`
}

type TransactionHex struct {
	BlockHeight      string        `json:"blockNumber" bson:"blockHeight"`
	BlockTime        string        `json:"blockTime" bson:"blockTime"`
	BlockHash        string        `json:"blockHash" bson:"blockHash"`
	Txid             string        `json:"hash" bson:"txid"`
	From             string        `json:"from" bson:"from"`
	To               string        `json:"to" bson:"to"`
	Value            string        `json:"value" bson:"value"`
	TokenKey         string        `json:"contractAddress" bason:"tokenKey"`
	Gas              string        `json:"gas"  bson:"gas"`
	GasPrice         string        `json:"gasPrice"  bson:"gasPrice"`
	GasUsed          string        `json:"gasUsed"  bson:"gasUsed"`
	Nonce            string        `json:"nonce"  bson:"nonce"`
	Input            string        `json:"input"  bson:"input"`
	TransactionIndex string        `json:"transactionIndex"  bson:"transactionIndex"`
	Logs             []interface{} `json:"logs" bson:"logs"`
	Status           string        `json:"status" bson:"status"`
}

func (th *TransactionHex) HexToRaw() Transaction {
	tx := Transaction{}

	h, err := strconv.ParseInt(th.BlockHeight, 0, 32) // 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	tx.BlockHeight = h

	gu, err := strconv.ParseInt(th.GasUsed, 0, 32) // 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	tx.GasUsed = gu

	nonce, err := strconv.ParseInt(th.Nonce, 0, 32) // 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	tx.Nonce = nonce

	time, err := strconv.ParseInt(th.BlockTime, 0, 32) // 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	tx.BlockTime = primitive.DateTime(time * 1000)
	tx.BlockHash = th.BlockHash
	tx.Txid = th.Txid
	tx.From = th.From
	tx.To = th.To
	tx.TokenKey = th.TokenKey
	tx.TransactionIndex = th.TransactionIndex[2:]
	tx.Input = th.Input
	tx.Logs = th.Logs
	tx.Status = th.Status[2:]

	return tx
}
