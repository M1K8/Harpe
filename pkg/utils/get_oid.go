package utils

import (
	"fmt"
	"strings"
)

func GetCode(ticker, contractType, day, month, year string, price float32) string {
	code := ""
	code += strings.ToUpper(ticker)
	code += year[2:]
	code += month
	code += day

	if len(contractType) > 1 {
		code += strings.ToUpper(contractType[:1])
	} else {
		code += strings.ToUpper(contractType)
	}
	price *= 1000

	code += fmt.Sprintf("%08d", int(price)) // problem?

	return code
}
