package models

import (
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Blocks struct {
	// difficulty:0x2
	// extraData:0xd783010600846765746887676f312e372e33856c696e757800000000000000004e10f96536e45ceca7e34cc1bdda71db3f3bb029eb69afd28b57eb0202c0ec0859d383a99f63503c4df9ab6c1dc63bf6b9db77be952f47d86d2d7b208e77397301
	// gasLimit:0x47e7c4
	// gasUsed:0x0
	// hash:0x9eb9db9c3ec72918c7db73ae44e520139e95319c421ed6f9fc11fa8dd0cddc56
	// logsBloom:0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 miner:0x0000000000000000000000000000000000000000 mixHash:0x0000000000000000000000000000000000000000000000000000000000000000
	// nonce:0x0000000000000000
	// number:0x3
	// parentHash:0x9b095b36c15eaf13044373aef8ee0bd3a382a5abb92e402afa44b8249c3a90e9
	// receiptsRoot:0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421
	// sha3Uncles:0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347
	// size:0x25e
	// stateRoot:0x53580584816f617295ea26c0e17641e0120cab2f0a8ffb53a866fd53aa8e8c2d
	// timestamp:0x58ee45f9
	// totalDifficulty:0x7
	// transactions:[]
	// transactionsRoot:0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421
	// uncles:[]

	Height          int64            `json:"number"  bson:"height"`
	Hash            string           `json:"hash"  bson:"hash"`
	PreviousHash    string           `json:"parentHash"  bson:"parentHash"`
	GasLimit        int64            `json:"gasLimit"  bson:"gasUsed"`
	GasUsed         int64            `json:"gasUsed"  bson:"gasUsed"`
	Timestamp       int64            `json:"timestamp"  bson:"timestamp"`
	TotalDifficulty int64            `json:"totalDifficulty"  bson:"totalDifficulty"`
	Difficulty      int64            `json:"difficulty"  bson:"difficulty"`
	Nonce           string           `json:"nonce"  bson:"nonce"`
	Sha3Uncles      string           `json:"sha3Uncles"  bson:"sha3Uncles"`
	Transactions    []TransactionHex `json:"transactions"  bson:"transactions"`

	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt"`
	UpdatedAt primitive.DateTime `json:"updatedAt" bson:"updatedAt"`
}

type BlocksHex struct {
	Height          string           `json:"number"  bson:"height"`
	Hash            string           `json:"hash"  bson:"hash"`
	PreviousHash    string           `json:"parentHash"  bson:"parentHash"`
	GasLimit        string           `json:"gasLimit"  bson:"gasUsed"`
	GasUsed         string           `json:"gasUsed"  bson:"gasUsed"`
	Timestamp       string           `json:"timestamp"  bson:"timestamp"`
	TotalDifficulty string           `json:"totalDifficulty"  bson:"totalDifficulty"`
	Difficulty      string           `json:"difficulty"  bson:"difficulty"`
	Nonce           string           `json:"nonce"  bson:"nonce"`
	Sha3Uncles      string           `json:"sha3Uncles"  bson:"sha3Uncles"`
	Transactions    []TransactionHex `json:"transactions"  bson:"transactions"`

	// CreatedAt time.Time      `json:"height"`
	// UpdatedAt time.Time      `json:"height"`
}

func (bh *BlocksHex) HexToRaw() *Blocks {

	// Height          string   `json:"number"  bson:"height"`
	// Hash            string   `json:"hash"  bson:"hash"`
	// PreviousHash    string   `json:"parentHash"  bson:"parentHash"`
	// GasLimit        string   `json:"gasLimit"  bson:"gasUsed"`
	// GasUsed         string   `json:"gasUsed"  bson:"gasUsed"`
	// Timestamp       string   `json:"timestamp"  bson:"timestamp"`
	// TotalDifficulty string   `json:"totalDifficulty"  bson:"totalDifficulty"`
	// Difficulty      string   `json:"difficulty"  bson:"difficulty"`
	// Nonce           string   `json:"nonce"  bson:"nonce"`
	// Transactions    []string `json:"transactions"  bson:"transactions"`
	b := &Blocks{}

	h, err := strconv.ParseInt(bh.Height, 0, 32) ////0x4d1a35 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	b.Height = h

	gl, err := strconv.ParseInt(bh.GasLimit, 0, 32) ////0x4d1a35 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	b.GasLimit = gl

	gu, err := strconv.ParseInt(bh.GasUsed, 0, 32) ////0x4d1a35 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	b.GasUsed = gu

	time, err := strconv.ParseInt(bh.Timestamp, 0, 32) ////0x4d1a35 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	b.Timestamp = time

	td, err := strconv.ParseInt(bh.TotalDifficulty, 0, 32) ////0x4d1a35 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	b.TotalDifficulty = td

	d, err := strconv.ParseInt(bh.Difficulty, 0, 32) ////0x4d1a35 写 0 后 他自己判断去除前面的0x
	if err != nil {

	}

	b.Difficulty = d

	b.Hash = bh.Hash
	b.PreviousHash = bh.PreviousHash
	b.Nonce = bh.Nonce
	b.Transactions = bh.Transactions
	b.Sha3Uncles = bh.Sha3Uncles

	return b
}
