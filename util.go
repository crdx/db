package db

import (
	"fmt"
	"strconv"
)

// ToUint returns value as a uint, and an error if it could not be converted.
func ToUint[T ID](value T) (uint, error) {
	var n uint

	switch value := any(value).(type) {
	case string:
		if value, err := strconv.Atoi(value); err != nil {
			return 0, err
		} else {
			n = uint(value)
		}
	case int:
		n = uint(value)
	case uint:
		n = value
	default:
		return 0, fmt.Errorf("invalid type")
	}

	return n, nil
}
