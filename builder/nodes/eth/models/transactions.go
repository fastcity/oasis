package models

type Transaction struct {

	// hash: { type: String, required: true, unique: true }, // "0x9fc76417374aa880d4449a1f7f31ec597f00b1f6f3dd2d66f4c9c6c445836d8b",
	// nonce: { type: Number, required: true }, // 2,
	// blockHash: { type: String, required: true }, // "0xef95f2f1ed3ca60b048b4bf67cde2195961e0bba6f70bcbea9a2c4e133e34b46",
	// blockNumber: { type: Number, required: true }, // 3,
	// blockTime: { type: Number, required: false },
	// transactionIndex: { type: Number, required: true }, // 0,
	// from: { type: String, required: true }, // "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b",
	// to: { type: String, required: false }, // "0x6295ee1b4f6dd65047762f924ecd367c17eabf8f",
	// value: { type: String, required: false }, // '123450000000000000',
	// gas: { type: Number, required: false }, // 314159,
	// gasPrice: { type: Number, required: false }, // '2000000000000',
	// input: { type: String, required: false }, // "0x57cb2fc4",
	// gasUsed: { type: Number, required: false },
	// size: { type: Number, required: false },
	// cumulativeGasUsed: { type: Number, required: false },
	// contractAddress: { type: String, required: false },
	// status: { type: String, required: false },
	// logs: [],
	// createdAt: { type: Number, default: Date.now },

	BlockHeight int64  `json:"height" bson:"blockHeight"`
	BlockTime   int64  `json:"blockTime" bson:"blockTime"`
	BlockHash   string `json:"blockHash" bson:"blockHash"`
	Nonce       string `json:"nonce"  bson:"nonce"`
	Txid        string `json:"hash" bson:"txid"`
	From        string `json:"from" bson:"from"`
	To          string `json:"to" bson:"to"`
	Value       string `json:"value" bson:"value"`
	Gas         string `json:"gas"  bson:"gas"`
	GasPrice    string `json:"gasPrice"  bson:"gasPrice"`
	GasUsed     string `json:"gasUsed"  bson:"gasUsed"`
	Status      string `json:"status" bson:"status"`
}

type TransactionHex struct {
	BlockHeight  int64    `json:"number"`
	BlockTime    int      `json:"blockTime"`
	BlockHash    string   `json:"hash"`
	ParentHash   string   `json:"parentHash"`
	Timestamp    string   `json:"timestamp"`
	Nonce        string   `json:"nonce"`
	Transactions []string `json:"transactions"`
}
