package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
)

var (
	calcTime = flag.Duration("calcTime", 10*time.Second, "Calculate pi for this length of time")
	bucket   = flag.String("bucket", "", "Write Pi result to this GCS bucket")
)

func main() {
	flag.Parse()
	val := pi(*calcTime)
	strval := strconv.FormatFloat(val, 'f', -1, 64)
	log.Printf("Calculated Pi for %v: %s\n", *calcTime, strval)
	if *bucket != "" {
		writeToGcs(*bucket, strval)
	}
}

// approximate pi using the Leibniz formula for specified duration
func pi(calcTime time.Duration) float64 {
	f := 0.0
	k := 0.0
	for timeout := time.After(calcTime); ; {
		select {
		case <-timeout:
			return f
		default:
			f += 4 * math.Pow(-1, k) / (2*k + 1)
			k++
		}
	}
}

func writeToGcs(bucketName string, val string) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}

	bucket := client.Bucket(bucketName)
	filename := "pi-" + time.Now().UTC().Format(time.RFC3339)
	obj := bucket.Object(filename)

	w := obj.NewWriter(ctx)
	if _, err := fmt.Fprintf(w, "%s", val); err != nil {
		log.Fatalf("Failed to write GCS file: %v", err)
	}
	if err := w.Close(); err != nil {
		log.Fatalf("Failed to write GCS file: %v", err)
	}
	log.Printf("Wrote gs://%s/%s", bucketName, filename)
}
