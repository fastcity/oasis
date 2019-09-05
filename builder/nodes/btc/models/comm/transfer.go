package models

type Transaction struct {
	BlockHeight string `bson:"blockHeight"`
	BlockTime   string `bson:"blockTime"`
	BlockHash   string `bson:"blockHash"`
	Txid        string `bson:"txid"`
	Method      string `bson:"method"`
	TxHash      string `bson:"txHash"`
	From        string `bson:"from"`
	To          string `bson:"to"`
	Value       string `bson:"value"`
	TokenKey    string `bson:"token"`
	OnChain     bool   `bson:"onChain"`
	Log         string `bson:"log"`

	// 	_account: { type: mongoose.Schema.Types.ObjectId, ref: 'Account', required: false },
	// chain: { type: String, default: 'ETH' }, // 链
	// coin: { type: String },
	// requestBody: {},
	// txData: {},
	// txRaw: { type: String, required: false }, // 签名后的事务
	// kafkaMsg: {},
	// kafkaMsgResult: {},
	// // 提交上链请求后
	// txid: { type: String, required: false }, // 上链txid
	// blockHeight: { type: Number, default: 0, required: false }, // 上链后所载区块

	// code: { type: Number, default: 0 }, // 状态码
	// status: { type: String, default: 'ready' }, // 状态码
	// logs: [],

	// createdAt: { type: Number, default: Date.now },
	// updatedAt: { type: Number, default: Date.now },
}
