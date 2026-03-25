package api_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetItemsBySeller(t *testing.T) {
	t.Run("Получение объявлений существующего продавца", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		price := rand.Intn(maxInt) + 1
		likes := rand.Intn(maxInt) + 1
		views := rand.Intn(maxInt) + 1
		contacts := rand.Intn(maxInt) + 1
		uuids := make([]string, 2)
		for i := 0; i < 2; i++ {

			payload := map[string]interface{}{
				"sellerID":   sellerID,
				"name":       randomString(10, allChars),
				"price":      price,
				"statistics": map[string]int{"likes": likes, "viewCount": views, "contacts": contacts},
			}
			resp := postItem(t, payload)
			require.Equal(t, 200, resp.StatusCode)
			uuid := extractUUIDFromResponse(t, resp)
			uuids[i] = uuid
		}
		defer func() {
			for _, uuid := range uuids {
				deleteItem(t, uuid)
			}
		}()

		items, code := getItemsBySeller(t, sellerID)
		assert.Equal(t, 200, code)
		assert.Len(t, items, 2)
		found := 0
		for _, item := range items {
			for _, uuid := range uuids {
				if item.ID == uuid {
					found++
				}
			}
		}
		assert.Equal(t, 2, found)
	})

	t.Run("Получение объявлений существующего продавца, проверка идемпотентности ", func(t *testing.T) {
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

		items1, code1 := getItemsBySeller(t, sellerID)
		items2, code2 := getItemsBySeller(t, sellerID)
		assert.Equal(t, 200, code1)
		assert.Equal(t, 200, code2)
		assert.Equal(t, items1, items2)
	})

	t.Run("Получение объявлений продавца с sellerID=1", func(t *testing.T) {
		_, code := getItemsBySeller(t, 1)
		assert.Equal(t, 200, code)
	})

	t.Run("Получение объявлений продавца с sellerID=0", func(t *testing.T) {
		_, code := getItemsBySeller(t, 0)
		assert.Equal(t, 200, code)
	})

	t.Run("Получение объявлений продавца с sellerID=9223372036854775807", func(t *testing.T) {
		_, code := getItemsBySeller(t, 9223372036854775807)
		assert.Equal(t, 200, code)
	})

	t.Run("Получение объявлений продавца с валидным, но несуществующим sellerID", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		items, code := getItemsBySeller(t, sellerID)
		assert.Equal(t, 200, code)
		assert.Empty(t, items)
	})

	t.Run("Получение объявлений продавца,  проверка отсутствия поддержки других методов", func(t *testing.T) {
		sellerID := rand.Intn(999999-111111+1) + 111111
		methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions}
		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req, _ := http.NewRequest(method, fmt.Sprintf("%s/api/1/%d/item", baseURL, sellerID), nil)
				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
			})
		}
	})

	t.Run("Получение объявлений продавца, sellerId отсутствует", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/1//item")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Получение объявлений продавца, sellerId не в числовом формате", func(t *testing.T) {
		nonNumeric := []string{"abc", randomString(10, latinicLower)}
		for _, id := range nonNumeric {
			t.Run(id, func(t *testing.T) {
				url := fmt.Sprintf("%s/api/1/%s/item", baseURL, id)
				resp, err := http.Get(url)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		}
	})

	t.Run("Получение объявлений продавца с sellerID=9223372036854775808", func(t *testing.T) {
		url := baseURL + "/api/1/9223372036854775808/item"
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Получение объявлений продавца с sellerID=-9223372036854775809", func(t *testing.T) {
		url := baseURL + "/api/1/-9223372036854775809/item"
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Получение объявлений продавца с sellerID=-1", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/1/-1/item")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Получение объявлений продавца, проверка использования / в sellerId", func(t *testing.T) {
		url := baseURL + "/api/1/123/456/item"
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Получение объявлений продавца, проверка корректности маршрутизации", func(t *testing.T) {
		urls := []string{
			baseURL + "/api/1/statistics/item",
			baseURL + "/api/1/item/item",
		}
		for _, url := range urls {
			t.Run(url, func(t *testing.T) {
				resp, err := http.Get(url)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			})
		}
	})
}
