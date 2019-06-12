package main

import (
	"century/oasis/nodes/vct/util"
	"fmt"
	"math/big"
	"net/http"
)

func main() {
	http.HandleFunc("/CreateTransactionDataHandler", createTransactionDataHandler)
	http.HandleFunc("/getBlockHeight", getBlockHeight)
	// http.HandleFunc("/login/", loginHandler)
	// http.HandleFunc("/ajax/", ajaxHandler)
	// http.HandleFunc("/", NotFoundHandler)
	err := http.ListenAndServe("0.0.0.0:7799", nil)
	if err != nil {
		fmt.Println("http listen failed.")
	}
}

func createTransactionDataHandler(w http.ResponseWriter, r *http.Request) {
	u := util.Util{BaseURL: "127.0.0.1:7080"}
	from := r.PostFormValue("from")
	to := r.PostFormValue("to")
	// value := r.PostFormValue("value")
	tokenKey := r.PostFormValue("tokenKey")

	amount, ok := big.NewInt(0).SetString(r.PostFormValue("value"), 0)

	if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
		// s.NormalErrorF(rw, 0, "Invalid amount")
		fmt.Println("Invalid amount")
		return
	}
	u.CreateTransactionData(from, to, tokenKey, amount)

}

func getBlockHeight(w http.ResponseWriter, r *http.Request) {
	u := util.Util{BaseURL: "http://127.0.0.1:7080/api/v1"}

	h, err := u.GetBlockHeight()
	if err != nil {
		fmt.Println("--------------err", err)
	}
	fmt.Fprintf(w, "h=%s", h)
}
