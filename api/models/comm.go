package models

type CommResp struct {
	Code int                    `json:"code"`
	Data map[string]interface{} `json:"data"` // 不能解析??
	// Data interface{}                 `json:"data"`
	Msg string `json:"msg"`
}

type RespIntserface interface {
	NewSuccess(map[interface{}]interface{}) *CommResp
	NewError(int, string) *CommResp
}

func (r *CommResp) NewSuccess(v map[string]interface{}) *CommResp {
	return &CommResp{
		Code: 0,
		Data: v,
	}
}

func (r *CommResp) NewError(code int, msg string) *CommResp {
	return &CommResp{
		Code: code,
		Msg:  msg,
	}
}
