package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

// generateVectors generates a slice of vectors with specified length and number of rows
// func generateVectors(vectorLength, numRows, offset int) [][]float32 {
// 	rand.Seed(time.Now().UnixNano())
// 	vectors := make([][]float32, numRows)
// 	for i := 0; i < numRows; i++ {
// 		vector := make([]float32, vectorLength)
// 		for j := 0; j < vectorLength; j++ {
// 			vector[j] = rand.Float32() / float32(offset)
// 		}
// 		vectors[i] = vector
// 	}
// 	return vectors
// }
func generateVectors(vectorLength, numRows int, min, max float32) [][]float32 {
	rand.Seed(time.Now().UnixNano())
	vectors := make([][]float32, numRows)
	for i := 0; i < numRows; i++ {
		vector := make([]float32, vectorLength)
		for j := 0; j < vectorLength; j++ {
			vector[j] = min + rand.Float32()*(max-min)
		}
		vectors[i] = vector
	}
	return vectors
}
// printVectors prints the vectors to standard output in the desired format
func printVectors(vectors [][]float32) {
	for _, vector := range vectors {
		fmt.Printf("[")
		for i, val := range vector {
			if i > 0 {
				fmt.Printf(",")
			}
			fmt.Printf("%.7f", val)
		}
		fmt.Printf("]\n")
	}
}
func main() {
	vectorLength := flag.Int("vectorLength", 1024, "Length of each vector")
	numRows := flag.Int("numRows", 679123, "Number of rows (vectors)")
	min := flag.Float64("min", 0.0, "Minimum value for range")
	max := flag.Float64("max", 1.0, "Maximum value for range")

	flag.Parse()

	// Convert min and max to float32
	min32 := float32(*min)
	max32 := float32(*max)

	vectors := generateVectors(*vectorLength, *numRows, min32, max32)
	printVectors(vectors)
}
