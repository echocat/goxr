package main

import "fmt"

type target interface {
	fmt.Stringer
	name() string
	createUriFor(t test) string
}
