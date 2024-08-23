package main

import (
	"fmt"
	"math/rand"
	"time"
)

// generateVectors generates a slice of vectors with specified length and number of rows
func generateVectors(vectorLength, numRows int) [][]int {
	rand.Seed(time.Now().UnixNano())
	vectors := make([][]int, numRows)
	for i := 0; i < numRows; i++ {
		vector := make([]int, vectorLength)
		for j := 0; j < vectorLength; j++ {
			vector[j] = rand.Intn(100000) // Generate a random integer between 0 and 99
		}
		vectors[i] = vector
	}
	return vectors
}

// printVectors prints the vectors to standard output in the desired format
func printVectors(vectors [][]int) {
	for _, vector := range vectors {
		fmt.Printf("[")
		for i, val := range vector {
			if i > 0 {
				fmt.Printf(",")
			}
			fmt.Printf("%d", val)
		}
		fmt.Printf("]\n")
	}
}

func main() {
	// Change these variables to generate different vector lengths and number of rows
	vectorLength := 512
	numRows := 1000000

	vectors := generateVectors(vectorLength, numRows)
	printVectors(vectors)
}
