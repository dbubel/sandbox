package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

// generateVectors generates a slice of vectors with specified length and number of rows
func generateVectors(vectorLength, numRows, offset int) [][]int {
	rand.Seed(time.Now().UnixNano())
	vectors := make([][]int, numRows)
	for i := 0; i < numRows; i++ {
		vector := make([]int, vectorLength)
		for j := 0; j < vectorLength; j++ {
			vector[j] = rand.Intn(10) + offset
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
			fmt.Printf("%f.2", val)
		}
		fmt.Printf("]\n")
	}
}

func main() {
	vectorLength := flag.Int("vectorLength", 1024, "Length of each vector")
	numRows := flag.Int("numRows", 679123, "Number of rows (vectors)")
	offset := flag.Int("offset", 0, "Offset for random numbers")

	flag.Parse()

	vectors := generateVectors(*vectorLength, *numRows, *offset)
	printVectors(vectors)
}
