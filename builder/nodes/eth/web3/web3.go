package web3

import (
	"math/big"
	"strings"
)

var (
	wei  int64 = 1e18
	gwei int64 = 1e9
)

// HexToAddr 转化为40个长度的等长地址
func HexToAddr(hexAddr interface{}) string {
	switch addr := hexAddr.(type) {
	case string:
		addr = strings.ToLower(addr)
		addr = strings.TrimPrefix(addr, "0x")
		if len(addr) < 40 {
			addr = strPadPre(addr, "0", 40)
			return addr
		}
		addr = addr[len(addr)-40:]
		return "0x" + addr
	case []byte:
		return HexToAddr(string(addr))

	default:
		return ""
	}

}

// HexToBigint 16 进制转化为 big int
func HexToBigint(hex string) *big.Int {

	base := 16
	if strings.HasPrefix(strings.ToLower(hex), "0x") {
		base = 0
	}

	n, _ := big.NewInt(0).SetString(hex, base)

	return n

}

func Towei(number *big.Int) *big.Int {

	wei := big.NewInt(wei)

	return number.Mul(number, wei)
}

func Gweitowei(number *big.Int) *big.Int {

	wei := big.NewInt(gwei)

	return number.Mul(number, wei)
}

func WeitoEther(number *big.Int) *big.Int {

	wei := big.NewInt(wei)

	return number.Div(number, wei)
}

func BigToHex(number *big.Int, prefix bool) string {

	if prefix {
		return "0x" + number.Text(16)
	}
	return number.Text(16)
}

// CreateERC20Input erc20 transfer 事件
func CreateERC20Input(to string, value *big.Int) string {

	// 0xa9059cbb
	to = strings.TrimRight(strings.ToLower(to), "0x")

	v := BigToHex(value, false)

	input := "0xa9059cbb" + strPadPre(to, "0", 64) + strPadPre(v, "0", 64)
	return input
}

// 补齐
func strPadPre(str, pad string, length int) string {
	if len(str) > length {
		return str[:length]
	}

	for len(str) < length {
		str = pad + str
	}
	return str
}
