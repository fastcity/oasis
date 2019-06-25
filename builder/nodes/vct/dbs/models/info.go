package models

import "time"

type Info struct {
	Height       int64
	Hash         string
	Time         string
	TxCount      int64 // 截至当前块的总交易数量
	Tps          int64 // tps
	AccountCount int64 // 账户数量

	CreatedAt time.Time
	UpdatedAt time.Time
}
