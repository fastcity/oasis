package models

type Transaction struct {
	BlockHeight int64                    `json:"height" bson:"blockHeight"`
	BlockTime   int                      `json:"blockTime" bson:"blockTime"`
	BlockHash   string                   `json:"blockHash" bson:"blockHash"`
	Txid        string                   `json:"txid" bson:"txid"`
	Version     int                      `json:"version" bson:"version"`
	Weight      int                      `json:"weight" bson:"weight"`
	Vsize       int                      `json:"vsize" bson:"vsize"`
	Size        int                      `json:"size" bson:"size"`
	Vins        []map[string]interface{} `json:"vin" bson:"vin"`
	Vouts       []map[string]interface{} `json:"vout" bson:"vout"`
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
// 			{
// 				"coinbase": "03f20b090402d96c5d424a2f48756f42692ffabe6d6d0461ff74508944818defe5cf2fe7213033dec489a74a0e96e2d9a228cff1b87708000000940e71b60a4bc8665958555200000000",
// 				"sequence": 4294967295
// 			}
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
