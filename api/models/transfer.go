package models

// var (
// 	UserList map[string]*User
// )

type Transfer struct {
	From     string
	To       string
	Value    string
	TokenKey string
}

type Balance struct {
	CommResp
	Code    int32
	Address string
	Balance string
}
