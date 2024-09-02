package main

/*
#cgo LDFLAGS: -L. -ldistance
#include "distance.h"
*/
import "C"

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
	"unsafe"
)

const (
	NUM_CLUSTERS = 10
	EPSILON      = 0.01
)

type ThreadSafeClusters struct {
	m        sync.Mutex
	Clusters map[[32]byte][][]float32
}

func NewClusters() *ThreadSafeClusters {
	return &ThreadSafeClusters{
		Clusters: make(map[[32]byte][][]float32),
	}
}

func (c *ThreadSafeClusters) appendAt(idx [32]byte, value []float32) {
	c.m.Lock()
	defer c.m.Unlock()
	c.Clusters[idx] = append(c.Clusters[idx], value)
}

func main() {
	file, err := os.Open("out20k.jsonl")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	vecData := [][]float32{}
	centroids := [][]float32{}

	clusters := NewClusters()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	t := time.Now()
	for scanner.Scan() {
		line := scanner.Bytes()
		var vec []float32
		err = json.Unmarshal(line, &vec)
		if err != nil {
			return
		}

		vecData = append(vecData, vec)
	}

	fmt.Println("read file time", time.Now().Sub(t))

	var guesses map[int]struct{}
	guesses = make(map[int]struct{})

	for len(guesses) < NUM_CLUSTERS {
		guess := rand.Intn(NUM_CLUSTERS + 1)
		if _, exists := guesses[guess]; !exists {
			guesses[guess] = struct{}{}
			centroids = append(centroids, vecData[guess])
		}
	}

	t = time.Now()
	var wg sync.WaitGroup
	for {
		for _, vec := range vecData {
			wg.Add(1)
			go func() {
				minDist := math.Inf(1)
				var bestCentroid [32]byte
				for _, centroid := range centroids {
					dist  := EuclideanDistance(vec, centroid)
					if dist < float32(minDist) {
						bestCentroid = HashFloat32Slice(centroid)
						minDist = float64(dist)
					}
				}

				wg.Done()
				clusters.appendAt(bestCentroid, vec)
			}()
		}
		wg.Wait()

		converged := true
		for i := range centroids {
			clust, exists := clusters.Clusters[HashFloat32Slice(centroids[i])]
			if !exists { // this sould never ever happen
				continue
			}

			newCentroid, _ := CalculateCentroid(clust)
			dist, _ := Distance(newCentroid, centroids[i])
			if dist > EPSILON {
				converged = false
			}
			centroids[i] = newCentroid
		}

		if converged {
			break
		}

		// we are doing another round so zero out clusters
		for key := range clusters.Clusters {
			delete(clusters.Clusters, key)
		}
	}
	//
	// for k, v := range clusters.Clusters {
	// 	fmt.Println("\t", k)
	// 	for _, i := range v {
	// 		fmt.Println(i)
	// 	}
	// }
	fmt.Println("cluster time", time.Now().Sub(t))
}

func CalculateCentroid(points [][]float32) ([]float32, error) {
	if len(points) == 0 {
		return nil, fmt.Errorf("the points slice must not be empty")
	}

	numPoints := len(points)
	numDimensions := len(points[0])

	centroid := make([]float32, numDimensions)

	for _, point := range points {
		if len(point) != numDimensions {
			return nil, fmt.Errorf("all points must have the same number of dimensions")
		}
		for i := 0; i < numDimensions; i++ {
			centroid[i] += point[i]
		}
	}

	for i := 0; i < numDimensions; i++ {
		centroid[i] /= float32(numPoints)
	}

	return centroid, nil
}

func Distance(p1, p2 []float32) (float32, error) {
	if len(p1) != len(p2) {
		return 0, fmt.Errorf("input slices must have the same length")
	}

	var sum float64
	for i := range p1 {
		difference := float64(p2[i] - p1[i])
		sum += difference * difference
	}
	return float32(math.Sqrt(sum)), nil
}

func HashFloat32Slice(data []float32) [32]byte {
	// Create a new SHA-256 hash instance
	hasher := sha256.New()

	// Convert each float32 value to bytes and write to the hasher
	for _, value := range data {
		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, math.Float32bits(value))
		hasher.Write(bytes)
	}

	// Compute the final hash
	return sha256.Sum256(hasher.Sum(nil))
}

func EuclideanDistance(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("slices must have the same length")
	}
	length := C.size_t(len(a))
	result := C.euclidean_distance_avx(
		(*C.float)(unsafe.Pointer(&a[0])),
		(*C.float)(unsafe.Pointer(&b[0])),
		length,
	)
	return float32(result)
}
