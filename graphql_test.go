package main

import (
	"os"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {

	date := time.Now().Format(time.RFC3339)
	_, err := getCloudflareCacheMetrics(buildGraphQLQuery(date, "d88b6d7f404e420305cd6c9a73c60576"), os.Getenv("APIEMAIL"), os.Getenv("APIKEY"))
	if err != nil {
		t.Errorf("Error: %v", err)
	} else {
		t.Logf("Test succeeded with %v and %v", os.Getenv("apiEmail"), os.Getenv("apiKey"))
	}
}
