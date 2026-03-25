package api_test

import (
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetItemByID(t *testing.T) {
	t.Run("Получение существующего объявления", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		name := randomString(10, latinicLower)
		price := rand.Intn(maxInt) + 1
		stats := map[string]int{"likes": rand.Intn(maxInt) + 1, "viewCount": rand.Intn(maxInt) + 1, "contacts": rand.Intn(maxInt) + 1}
		payload := map[string]interface{}{
			"sellerID":   sellerID,
			"name":       name,
			"price":      price,
			"statistics": stats,
		}
		resp := postItem(t, payload)
		require.Equal(t, 200, resp.StatusCode)
		uuid := extractUUIDFromResponse(t, resp)
		defer deleteItem(t, uuid)

		item := getItemByID(t, uuid)
		assert.Equal(t, sellerID, item.SellerID)
		assert.Equal(t, name, item.Name)
		assert.Equal(t, price, item.Price)
		assert.Equal(t, stats["likes"], item.Statistics.Likes)
		assert.Equal(t, stats["viewCount"], item.Statistics.ViewCount)
		assert.Equal(t, stats["contacts"], item.Statistics.Contacts)
		assert.NotEmpty(t, item.CreatedAt)
	})

	t.Run("Получение существующего объявления, проверка идемпотентности ", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		payload := map[string]interface{}{
			"sellerID":   sellerID,
			"name":       randomString(10, latinicLower),
			"price":      rand.Intn(10000),
			"statistics": map[string]int{"likes": rand.Intn(10000), "viewCount": rand.Intn(10000), "contacts": rand.Intn(10000)},
		}
		resp := postItem(t, payload)
		require.Equal(t, 200, resp.StatusCode)
		uuid := extractUUIDFromResponse(t, resp)
		defer deleteItem(t, uuid)

		item1 := getItemByID(t, uuid)
		item2 := getItemByID(t, uuid)
		assert.Equal(t, item1, item2)
	})

	t.Run("Получение существующего объявления, UUID  в верхнем регистре", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		payload := map[string]interface{}{
			"sellerID":   sellerID,
			"name":       randomString(10, latinicLower),
			"price":      rand.Intn(10000),
			"statistics": map[string]int{"likes": rand.Intn(maxInt) + 1, "viewCount": rand.Intn(maxInt) + 1, "contacts": rand.Intn(maxInt) + 1},
		}
		resp := postItem(t, payload)
		require.Equal(t, 200, resp.StatusCode)
		uuid := extractUUIDFromResponse(t, resp)
		defer deleteItem(t, uuid)

		upperUUID := strings.ToUpper(uuid)
		item := getItemByID(t, upperUUID)
		assert.Equal(t, sellerID, item.SellerID)
	})

	t.Run("Получение объявления, id не в формате UUID", func(t *testing.T) {
		invalidID := randomString(36, allChars)
		resp, err := http.Get(baseURL + "/api/1/item/" + invalidID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Получение объявления, id отстутствует", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/1/item/")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Contains(t, []int{404}, resp.StatusCode)
	})

	t.Run("Получение несуществующего объявления, UUID валиден", func(t *testing.T) {
		nonExistent := "00000000-0000-0000-0000-000000000000"
		resp, err := http.Get(baseURL + "/api/1/item/" + nonExistent)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Получение объявления, проверка отсутствия поддержки других методов", func(t *testing.T) {
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
				req, _ := http.NewRequest(method, baseURL+"/api/1/item/"+uuid, nil)
				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
			})
		}
	})

	t.Run("Получение объявления, проверка использования / в id", func(t *testing.T) {
		idWithSlash := "550e8400/e29b-41d4-a716-446655440000"
		resp, err := http.Get(baseURL + "/api/1/item/" + idWithSlash)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
