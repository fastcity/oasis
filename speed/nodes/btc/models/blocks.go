package models

import "math/big"

type Blocks struct {
	Height    string     `json:"height"`
	Hash      string     `json:"hash"`
	TimeStamp string     `json:"timeStamp"`
	Txid      string     `json:"txid"`
	Vin       []Txs      `json:"transactions"`
	Vout      []TxEvents `json:"txEvents"`
	// CreatedAt time.Time      `json:"height"`
	// UpdatedAt time.Time      `json:"height"`
}

type Txs struct {
	Height          string  `json:"height"`
	Txid            string  `json:"txID"`
	Chaincode       string  `json:"chaincode"`
	Method          string  `json:"method"`
	CreatedFlag     bool    `json:"createdFlag"`
	ChaincodeModule string  `json:"chaincodeModule"`
	Nonce           string  `json:"nonce"`
	Detail          *Detail `json:"detail"`
	TxHash          string  `json:"txHash"`
	// Details         *[]Detail `json:"detail"`
}

type TxEvents struct {
	// 			"Chaincode": "local",
	// 			"Name": "INVOKEERROR",
	// 			"Status": 1,
	// 			"Detail": "Local invoke error: handling method [MTOKEN.INIT] fail: Can not re-deploy existed data"
	Status    int64  `json:"status"`
	Txid      string `json:"txID"`
	Chaincode string `json:"chaincode"`
	Name      string `json:"name"`
	Detail    string `json:"detail"`
}

type Detail struct {
	Amount *big.Int `json:"amount"`
	From   string   `json:"from"`
	To     string   `json:"to"`
	Token  string   `json:"token"`

	//   Detail: {
	// 	"amount": "400000",
	// 	"to": "ARJtq6Q46oTnxDwvVqMgDtZeNxs7Ybt81A"
	//   },
}
