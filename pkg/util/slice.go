package util

import (
	"math/rand"
	"time"
)

// Deduplicate removes duplicate values from any comparable slice
// This is a generic function that works with any comparable type
func Deduplicate[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}

	// Create a map to track unique values
	uniqueMap := make(map[T]struct{})
	result := make([]T, 0, len(slice))

	// Iterate through the slice and add unique values to the result
	for _, item := range slice {
		if _, exists := uniqueMap[item]; !exists {
			uniqueMap[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

func RandSelect[T any](slices []T) (t T) {
	if len(slices) == 0 {
		return t
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := rand.Intn(len(slices))
	return slices[idx]
}
