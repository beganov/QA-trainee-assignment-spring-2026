package api_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestLoad_CreateItems(t *testing.T) {
	concurrency := 5
	iterations := 5

	var wg sync.WaitGroup
	results := make(chan time.Duration, concurrency*iterations)
	errors := make(chan error, concurrency*iterations)

	for i := 0; i < iterations; i++ {
		for j := 0; j < concurrency; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sellerID := rand.Intn(999999-111111+1) + 111111
				name := randomString(10, latinicLower)
				price := rand.Intn(1000) + 1
				stats := map[string]int{"likes": rand.Intn(maxInt) + 1, "viewCount": rand.Intn(maxInt) + 1, "contacts": rand.Intn(maxInt) + 1}

				payload := map[string]interface{}{
					"sellerID":   sellerID,
					"name":       name,
					"price":      price,
					"statistics": stats,
				}
				start := time.Now()
				resp := postItem(t, payload)
				elapsed := time.Since(start)
				if resp.StatusCode != http.StatusOK {
					errors <- fmt.Errorf("статус %d", resp.StatusCode)
				} else {
					results <- elapsed
					uuid := extractUUIDFromResponse(t, resp)
					defer deleteItem(t, uuid)
				}
			}()
		}
	}

	wg.Wait()
	close(results)
	close(errors)

	var durations []time.Duration
	for d := range results {
		durations = append(durations, d)
	}
	if len(durations) == 0 {
		t.Fatal("не удалось выполнить ни одного успешного запроса")
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p95 := durations[int(float64(len(durations))*0.95)]
	avg := sumDurations(durations) / time.Duration(len(durations))

	t.Logf("Успешных запросов: %d", len(durations))
	t.Logf("Среднее время ответа: %v", avg)
	t.Logf("p95 время ответа: %v", p95)

	var errs []error
	for e := range errors {
		errs = append(errs, e)
	}
	if len(errs) > 0 {
		t.Errorf("Обнаружено %d ошибок: %v", len(errs), errs)
	}
}
