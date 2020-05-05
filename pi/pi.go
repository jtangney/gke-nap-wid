package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"k8s.io/klog"
)

var (
	workers  = flag.Int("workers", 500000, "Number of go routines used to calculate pi")
	calcTime = flag.Duration("calcTime", 10*time.Second, "Calculate pi for this length of time (seconds)")
	bucket   = flag.String("bucket", "", "Write Pi result to this GCS bucket")
)

func main() {
	flag.Parse()
	val := pi2(*calcTime)
	// val := pi(*workers)
	strval := strconv.FormatFloat(val, 'f', -1, 64)
	// log.Printf("Calculated Pi with %d goroutines: %s\n", *workers, strval)
	klog.Infof("Calculated Pi for %v: %s\n", *calcTime, strval)
	if *bucket != "" {
		writeToGcs(*bucket, strval)
	}
}

func writeToGcs(bucketName string, val string) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		klog.Exitf("Failed to create GCS client: %v", err)
	}

	bucket := client.Bucket(bucketName)
	filename := time.Now().UTC().Format(time.RFC3339)
	obj := bucket.Object(filename)

	w := obj.NewWriter(ctx)
	if _, err := fmt.Fprintf(w, "%s", val); err != nil {
		klog.Exitf("Failed to write GCS file: %v", err)
	}
	if err := w.Close(); err != nil {
		klog.Exitf("Failed to write GCS file: %v", err)
	}
	klog.Infof("Wrote gs://%s/%s", bucketName, filename)
}

// approximate pi using the Leibniz formula for specified duration
func pi2(calcTime time.Duration) float64 {
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
