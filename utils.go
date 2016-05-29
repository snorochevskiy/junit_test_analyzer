package main

import (
	"strconv"
)

type ParsePanicErr struct {
	Message string
}

func (e ParsePanicErr) String() string {
	return e.Message
}

func ParseInt64(str string, errMsg string) int64 {
	val, parseErr := strconv.ParseInt(str, 10, 64)
	if parseErr != nil {
		panic(ParsePanicErr{Message: string(errMsg) + ":" + parseErr.Error()})
	}
	return val
}
