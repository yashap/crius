package crius

import "rsc.io/quote/v3"

// Hello returns a greeting
func Hello() string {
	return quote.HelloV3()
}

// Proverb returns a proverb
func Proverb() string {
	return quote.Concurrency()
}

// TODO: left off at Removing unused dependencies
