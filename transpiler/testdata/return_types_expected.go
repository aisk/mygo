package main

import (
	"bytes"
	"os"
)

type intAlias int

func example() (int, bool, string, intAlias, *bytes.Buffer, bytes.Buffer, byte, uintptr, float32, error) {
	_, err := os.Open("hello.mygo")
	if err != nil {
		return 0, false, "", *new(intAlias), nil, *new(bytes.Buffer), 0, 0, 0, err
	}
	return 0, false, "", *new(intAlias), nil, *new(bytes.Buffer), 0, 0, 0, err
}

func hello() error {
	_, err := os.Open("hello.mygo")
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if err := hello(); err != nil {
		panic(err)
	}
}
