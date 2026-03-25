package api_test

import (
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatisticsByID(t *testing.T) {
	t.Run("Получение статистики существующего объявления", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		price := rand.Intn(maxInt) + 1
		likes := rand.Intn(maxInt) + 1
		views := rand.Intn(maxInt) + 1
		contacts := rand.Intn(maxInt) + 1
		payload := map[string]interface{}{
			"sellerID":   sellerID,
			"name":       randomString(10, allChars),
			"price":      price,
			"statistics": map[string]int{"likes": likes, "viewCount": views, "contacts": contacts},
		}
		resp := postItem(t, payload)
		require.Equal(t, 200, resp.StatusCode)
		uuid := extractUUIDFromResponse(t, resp)
		defer deleteItem(t, uuid)

		stats, code := getStatistic(t, uuid)
		assert.Equal(t, 200, code)
		require.Len(t, stats, 1)
		assert.Equal(t, likes, stats[0].Likes)
		assert.Equal(t, views, stats[0].ViewCount)
		assert.Equal(t, contacts, stats[0].Contacts)
	})

	t.Run("Получение статистики существующего объявления, проверка идемпотентности ", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		price := rand.Intn(maxInt) + 1
		likes := rand.Intn(maxInt) + 1
		views := rand.Intn(maxInt) + 1
		contacts := rand.Intn(maxInt) + 1
		payload := map[string]interface{}{
			"sellerID":   sellerID,
			"name":       randomString(10, allChars),
			"price":      price,
			"statistics": map[string]int{"likes": likes, "viewCount": views, "contacts": contacts},
		}
		resp := postItem(t, payload)
		require.Equal(t, 200, resp.StatusCode)
		uuid := extractUUIDFromResponse(t, resp)
		defer deleteItem(t, uuid)

		stats1, code1 := getStatistic(t, uuid)
		stats2, code2 := getStatistic(t, uuid)
		assert.Equal(t, 200, code1)
		assert.Equal(t, 200, code2)
		assert.Equal(t, stats1, stats2)
	})

	t.Run("Получение статистики существующего объявления, UUID в верхнем регистре", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		price := rand.Intn(maxInt) + 1
		likes := rand.Intn(maxInt) + 1
		views := rand.Intn(maxInt) + 1
		contacts := rand.Intn(maxInt) + 1
		payload := map[string]interface{}{
			"sellerID":   sellerID,
			"name":       randomString(10, allChars),
			"price":      price,
			"statistics": map[string]int{"likes": likes, "viewCount": views, "contacts": contacts},
		}
		resp := postItem(t, payload)
		require.Equal(t, 200, resp.StatusCode)
		uuid := extractUUIDFromResponse(t, resp)
		defer deleteItem(t, uuid)

		upperUUID := strings.ToUpper(uuid)
		stats, code := getStatistic(t, upperUUID)
		assert.Equal(t, 200, code)
		assert.Len(t, stats, 1)
	})

	t.Run("Получение статистики объявления, id не в формате UUID", func(t *testing.T) {
		invalidID := randomString(36, allChars)
		resp, err := http.Get(baseURL + "/api/1/statistic/" + invalidID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Получение статистики объявления, id отстутствует", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/1/statistic/")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Contains(t, []int{404}, resp.StatusCode)
	})

	t.Run("Получение статистики несуществующего объявления, UUID валиден", func(t *testing.T) {
		_, code := getStatistic(t, "00000000-0000-0000-0000-000000000000")
		assert.Equal(t, http.StatusNotFound, code)
	})

	t.Run("Получение статистики объявления, проверка отсутствия поддержки других методов", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		payload := map[string]interface{}{
			"sellerID":   sellerID,
			"name":       randomString(10, allChars),
			"price":      rand.Intn(maxInt) + 1,
			"statistics": map[string]int{"likes": rand.Intn(maxInt) + 1, "viewCount": rand.Intn(maxInt) + 1, "contacts": rand.Intn(maxInt) + 1},
		}
		resp := postItem(t, payload)
		require.Equal(t, 200, resp.StatusCode)
		uuid := extractUUIDFromResponse(t, resp)
		defer deleteItem(t, uuid)

		methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions}
		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req, _ := http.NewRequest(method, baseURL+"/api/1/statistic/"+uuid, nil)
				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
			})
		}
	})

	t.Run("Получение статистики объявления, проверка использования / в id", func(t *testing.T) {
		idWithSlash := "550e8400/e29b-41d4-a716-446655440000"
		_, code := getStatistic(t, idWithSlash)
		assert.Equal(t, http.StatusNotFound, code)
	})
}
