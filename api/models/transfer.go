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
	CreateID  string `bson:"createId"`
	RequestID string `json:"requestId" bson:"requestId"`
}

type Balance struct {
	CommResp
	Code    int32
	Address string `json:"address"`
	Balance string
}
