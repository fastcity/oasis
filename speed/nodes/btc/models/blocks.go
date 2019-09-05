package models

type Blocks struct {
	// "bits": "171a213e",
	// "blockHash": "000000000000000000081b95177389fd5d53a07fb607ce41fd407839d29f5737",
	// "chainwork": "00000000000000000000000000000000000000000831937c653abaacc7809301",
	// "confirmations": 1,
	// "difficulty": 10771996663680.4,
	// "hash": "000000000000000000081b95177389fd5d53a07fb607ce41fd407839d29f5737",
	// "height": 592882,
	// "mediantime": 1567413325,
	// "merkleroot": "eebafdbd9deaa3ef060457473592c9b6400e4bbc0a24c3ef9ad305ac17a5bdba",
	// "nTx": 2164,
	// "nonce": 1623728037,
	// "previousblockhash": "0000000000000000001370df341c10b65fa568ae6ab82307ecee2196f4533642",
	// "size": 1199203,
	// "strippedsize": 931271,
	// "time": 1567414541,

	Height       int64  `json:"height"`
	Hash         string `json:"hash"`
	PreviousHash string `json:"previousblockhash"`
	Time         int    `json:"time"`
	Txid         string `json:"txid"`
	Tx           []TXs  `json:"tx"`
	// CreatedAt time.Time      `json:"height"`
	// UpdatedAt time.Time      `json:"height"`
}

type TXs struct {
	// "hash": "ef73d945a69948211bfff7e57f0f557171697622acfe4564b0d513ed16dc068d",
	// 		"hex": "020000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff4a03f20b090402d96c5d424a2f48756f42692ffabe6d6d0461ff74508944818defe5cf2fe7213033dec489a74a0e96e2d9a228cff1b87708000000940e71b60a4bc8665958555200000000ffffffff03d2a2ae4b000000001976a914e582933875bedfdc448473c00b474f8f053a467588ac0000000000000000266a24aa21a9ed6d93e6c8347eec37600d9831de8eb8c5ae47294aad62e1777d92bb8e64f4803d0000000000000000266a24b9e11b6d322bb2948b633959c79c657206a9df6c4c8c4009191c768829702394866b54790120000000000000000000000000000000000000000000000000000000000000000000000000",
	// 		"locktime": 0,
	// 		"size": 289,
	// 		"txid": "854a5bf4e7d3f39f4b76976aba1d312f3d43020e751f4c6c71018fb4e47060e5",
	// 		"version": 2,
	Hash     string                   `json:"hash"`
	Hex      string                   `json:"hex"`
	Locktime string                   `json:"locktime"`
	Size     int                      `json:"size"`
	Txid     string                   `json:"txid"`
	Version  int                      `json:"version"`
	Vsize    int                      `json:"vsize" bson:"vsize"`
	Weight   int                      `json:"weight" bson:"weight"`
	Vin      []map[string]interface{} `json:"vin"`
	Vout     []map[string]interface{} `json:"vout"`
}

type Vins struct {
	Txid     string `json:"txid"`
	Vout     int    `json:"vout"`
	Coinbase string `json:"coinbase"`
	Sequence string `json:"sequence"`
}
type Vouts struct {
	ScriptPubKey interface{} `json:"ScriptPubKey"`
	Value        float32     `json:"from"`
	Vout         int         `json:"n"`
	Addresses    []string    `json:"addresses"`
}

// "bits": "171a213e",
// "blockHash": "000000000000000000081b95177389fd5d53a07fb607ce41fd407839d29f5737",
// "chainwork": "00000000000000000000000000000000000000000831937c653abaacc7809301",
// "confirmations": 1,
// "difficulty": 10771996663680.4,
// "hash": "000000000000000000081b95177389fd5d53a07fb607ce41fd407839d29f5737",
// "height": 592882,
// "mediantime": 1567413325,
// "merkleroot": "eebafdbd9deaa3ef060457473592c9b6400e4bbc0a24c3ef9ad305ac17a5bdba",
// "nTx": 2164,
// "nonce": 1623728037,
// "previousblockhash": "0000000000000000001370df341c10b65fa568ae6ab82307ecee2196f4533642",
// "size": 1199203,
// "strippedsize": 931271,
// "time": 1567414541,
// "tx": [
// 	{
// 		"hash": "ef73d945a69948211bfff7e57f0f557171697622acfe4564b0d513ed16dc068d",
// 		"hex": "020000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff4a03f20b090402d96c5d424a2f48756f42692ffabe6d6d0461ff74508944818defe5cf2fe7213033dec489a74a0e96e2d9a228cff1b87708000000940e71b60a4bc8665958555200000000ffffffff03d2a2ae4b000000001976a914e582933875bedfdc448473c00b474f8f053a467588ac0000000000000000266a24aa21a9ed6d93e6c8347eec37600d9831de8eb8c5ae47294aad62e1777d92bb8e64f4803d0000000000000000266a24b9e11b6d322bb2948b633959c79c657206a9df6c4c8c4009191c768829702394866b54790120000000000000000000000000000000000000000000000000000000000000000000000000",
// 		"locktime": 0,
// 		"size": 289,
// 		"txid": "854a5bf4e7d3f39f4b76976aba1d312f3d43020e751f4c6c71018fb4e47060e5",
// 		"version": 2,
// 		"vin": [
// 		 {
// 	"scriptSig": {
// 		"asm": "",
// 		"hex": ""
// 	},
// 	"sequence": 4294967295,
// 	"txid": "cb15d3e2b7480f59a45d040568cb70475b212f85952e422adfe33580b5829cf8",
// 	"txinwitness": [
// 		"",
// 		"30440220199522a812d48377d8cdbc30c7a15129078c5101e07d4a79473edc025b6743d7022079cb1fb1f214e4d9e0ae5d8aff18cfadf1c536ed7fddaeb31744b39561836ff701",
// 		"304402204039cbdc9ae81fd0897c5dcfdb8046aed30dccfef773573dec98de11d10c0a1b02202968a43b47ae6ee83c3a249b6d498c7232ce910b6b2102aee021f0b161959b6c01",
// 		"52210375e00eb72e29da82b89367947f29ef34afb75e8654f6ea368e0acdfd92976b7c2103a1b26313f430c4b15bb1fdce663207659d8cac749a0e53d70eff01874496feff2103c96d495bfdd5ba4145e3e046fee45e84a8a48ad05bd8dbb395c011a32cf9f88053ae"
// 	],
// 	"vout": 1
// }
// 		],
// 		"vout": [
// 			{
// 				"n": 0,
// 				"scriptPubKey": {
// 					"addresses": [
// 						"1MvYASoHjqynMaMnP7SBmenyEWiLsTqoU6"
// 					],
// 					"asm": "OP_DUP OP_HASH160 e582933875bedfdc448473c00b474f8f053a4675 OP_EQUALVERIFY OP_CHECKSIG",
// 					"hex": "76a914e582933875bedfdc448473c00b474f8f053a467588ac",
// 					"reqSigs": 1,
// 					"type": "pubkeyhash"
// 				},
// 				"value": 12.69736146
// 			},
// 			{
// 				"n": 1,
// 				"scriptPubKey": {
// 					"asm": "OP_RETURN aa21a9ed6d93e6c8347eec37600d9831de8eb8c5ae47294aad62e1777d92bb8e64f4803d",
// 					"hex": "6a24aa21a9ed6d93e6c8347eec37600d9831de8eb8c5ae47294aad62e1777d92bb8e64f4803d",
// 					"type": "nulldata"
// 				},
// 				"value": 0
// 			},
// 			{
// 				"n": 2,
// 				"scriptPubKey": {
// 					"asm": "OP_RETURN b9e11b6d322bb2948b633959c79c657206a9df6c4c8c4009191c768829702394866b5479",
// 					"hex": "6a24b9e11b6d322bb2948b633959c79c657206a9df6c4c8c4009191c768829702394866b5479",
// 					"type": "nulldata"
// 				},
// 				"value": 0
// 			}
// 		],
// 		"vsize": 262,
// 		"weight": 1048
// 	},]
