package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <number of clusters> <number of points in each cluster>")
		return
	}

	numClusters, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Error: Invalid number of clusters")
		return
	}

	numPointsInCluster, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Error: Invalid number of points in each cluster")
		return
	}

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < numClusters; i++ {
		clusterCenter := generateRandomPoint()
		for j := 0; j < numPointsInCluster; j++ {
			point := generatePointNear(clusterCenter)
			jsonPoint, err := json.Marshal(point)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}
			fmt.Println(string(jsonPoint))
		}
	}
}

func generateRandomPoint() [2]float64 {
	return [2]float64{
		rand.Float64() * 100, // Random x coordinate between 0 and 100
		rand.Float64() * 100, // Random y coordinate between 0 and 100
	}
}

func generatePointNear(center [2]float64) [2]float64 {
	const proximity = 5.0
	return [2]float64{
		center[0] + rand.Float64()*proximity - proximity/2, // Random x coordinate near center
		center[1] + rand.Float64()*proximity - proximity/2, // Random y coordinate near center
	}
}
