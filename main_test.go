package main

import (
	"testing"
)

func TestConvertToMB(t *testing.T) {
	result := convertToMB(2000)
	var expected uint64 = 2
	if result != expected {
		t.Errorf("convertToMB(2000) = %d; wanted %d", result, expected)
	}
}
