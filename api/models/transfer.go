package models

// var (
// 	UserList map[string]*User
// )

type Transfer struct {
	Chain     string
	Coin      string
	From      string `json:"from"`
	To        string
	Value     string
	TokenKey  string
	CreateID  string
	RequestID string `json:"requestId"`
}

type Balance struct {
	CommResp
	Code    int32
	Address string `json:"address"`
	Balance string
}
