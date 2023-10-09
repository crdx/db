package db_test

import (
	"fmt"
	"testing"

	"crdx.org/db"
	"github.com/stretchr/testify/assert"
)

func TestToUint(t *testing.T) {
	testCases := []struct {
		input    any
		expected uint
		ok       bool
	}{
		{"123", 123, true},
		{"abc", 0, false},
		{123, 123, true},
		{uint(123), 123, true},
		{nil, 0, false},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%v", testCase.input), func(t *testing.T) {
			switch input := testCase.input.(type) {
			case string:
			case uint:
			case int:
				actual, err := db.ToUint(input)
				if !testCase.ok {
					assert.Error(t, err)
				} else {
					assert.Equal(t, testCase.expected, actual)
				}
			}
		})
	}
}
