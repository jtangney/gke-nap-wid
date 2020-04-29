package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"cloud.google.com/go/storage"
)

var (
	workers = flag.Int("workerCount", 500000, "Number of go routines used to calculate pi")
	bucket  = flag.String("bucket", "", "Write Pi result to this GCS bucket")
)

func main() {
	flag.Parse()
	val := pi(*workers)
	log.Printf("Calculated Pi with %d goroutines: %f\n", *workers, val)
	if *bucket != "" {
		writeToGcs(*bucket, val)
	}
}

func writeToGcs(bucketName string, val float64) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}

	bucket := client.Bucket(bucketName)
	filename := time.Now().UTC().Format(time.RFC3339)
	obj := bucket.Object(filename)

	w := obj.NewWriter(ctx)
	if _, err := fmt.Fprintf(w, "%f", val); err != nil {
		log.Fatalf("Failed to write GCS file: %v", err)
	}
	if err := w.Close(); err != nil {
		log.Printf("Error closing writer: %v", err)
	}
	log.Printf("Wrote gs://%s/%s", bucketName, filename)
}

// pi launches n goroutines to compute an
// approximation of pi.
func pi(n int) float64 {
	ch := make(chan float64)
	for k := 0; k <= n; k++ {
		go term(ch, float64(k))
	}
	f := 0.0
	for k := 0; k <= n; k++ {
		f += <-ch
	}
	return f
}

func term(ch chan float64, k float64) {
	ch <- 4 * math.Pow(-1, k) / (2*k + 1)
}
