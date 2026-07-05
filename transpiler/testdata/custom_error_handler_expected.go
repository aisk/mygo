package main

import "fmt"

func process() error {
	r, err := SomeFunction()
	if err != nil {
		return fmt.Errorf("something failed: %v", err)
	}
	fmt.Println(r)
	return nil
}

func SomeFunction() (string, error) {
	return "hello", nil
}
