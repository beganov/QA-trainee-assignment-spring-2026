package api_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_ItemAfterDelete_Allure(t *testing.T) {
	origT := t

	runner.Run(t, "Проверка, что объявление недоступно после удаления", func(t provider.T) {
		t.Epic("E2E")
		t.Feature("Удаление объявления")
		t.Tags("e2e", "delete")
		t.Title("После удаления GET /item/{id} возвращает 404")
		t.Description("Создаётся объявление, удаляется, затем проверяется, что по его ID возвращается 404 Not Found.")

		var (
			sellerID int
			name     string
			price    int
			likes    int
			views    int
			contacts int
			uuid     string
		)

		t.WithNewStep("Создание объявления", func(sCtx provider.StepCtx) {
			sellerID = rand.Intn(999999-111111+1) + 111111
			name = randomString(10, latinicLower)
			price = rand.Intn(maxInt) + 1
			likes = rand.Intn(maxInt) + 1
			views = rand.Intn(maxInt) + 1
			contacts = rand.Intn(maxInt) + 1

			payload := map[string]interface{}{
				"sellerID": sellerID,
				"name":     name,
				"price":    price,
				"statistics": map[string]int{
					"likes":     likes,
					"viewCount": views,
					"contacts":  contacts,
				},
			}

			// Прикрепляем вложение
			if reqBody, err := json.MarshalIndent(payload, "", "  "); err == nil {
				sCtx.WithNewAttachment("Запрос POST", "application/json", reqBody)
			}

			resp := postItem(origT, payload)
			require.Equal(origT, http.StatusOK, resp.StatusCode, "POST должен вернуть 200")
			uuid = extractUUIDFromResponse(origT, resp)
			sCtx.Logf("Создано объявление с ID: %s", uuid)
		})

		t.WithNewStep("Удаление объявления", func(sCtx provider.StepCtx) {
			deleteItem(origT, uuid)
			sCtx.Logf("Объявление %s удалено", uuid)
		})

		t.WithNewStep("Проверка GET запроса после удаления", func(sCtx provider.StepCtx) {
			getResp, err := http.Get(baseURL + "/api/1/item/" + uuid)
			require.NoError(origT, err)
			defer getResp.Body.Close()
			sCtx.Assert().Equal(http.StatusNotFound, getResp.StatusCode, "GET должен вернуть 404")
		})

		defer deleteItem(origT, uuid)
	})
}
func TestE2E_DuplicateFieldsDifferentCase(t *testing.T) {
	sellerIDCorrect := rand.Intn(999999-111111+1) + 111111
	sellerIDWrong := sellerIDCorrect + 1
	nameCorrect := randomString(10, latinicLower)
	nameWrong := randomString(10, latinicLower)
	priceCorrect := rand.Intn(1000) + 1
	priceWrong := priceCorrect + 1
	likesCorrect := rand.Intn(100) + 1
	likesWrong := likesCorrect + 1
	viewCountCorrect := rand.Intn(100) + 1
	viewCountWrong := viewCountCorrect + 1
	contactsCorrect := rand.Intn(100) + 1
	contactsWrong := contactsCorrect + 1
	jsonStr := fmt.Sprintf(`{
        "sellerID": %d,
        "SellerID": %d,
        "name": "%s",
        "Name": "%s",
        "price": %d,
        "Price": %d,
        "statistics": {
            "likes": %d,
            "Likes": %d,
            "viewCount": %d,
            "ViewCount": %d,
            "contacts": %d,
            "Contacts": %d
        }
    }`, sellerIDCorrect, sellerIDWrong,
		nameCorrect, nameWrong,
		priceCorrect, priceWrong,
		likesCorrect, likesWrong,
		viewCountCorrect, viewCountWrong,
		contactsCorrect, contactsWrong)

	req, _ := http.NewRequest("POST", baseURL+"/api/1/item", strings.NewReader(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	uuid := extractUUIDFromResponse(t, resp)
	defer deleteItem(t, uuid)

	getResp, err := http.Get(baseURL + "/api/1/item/" + uuid)
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)
	var items []Item
	err = json.NewDecoder(getResp.Body).Decode(&items)
	require.NoError(t, err)
	require.Len(t, items, 1)
	item := items[0]

	assert.Equal(t, sellerIDCorrect, item.SellerID, "должен использоваться sellerID, а не SellerID")
	assert.Equal(t, nameCorrect, item.Name, "должен использоваться name, а не Name")
	assert.Equal(t, priceCorrect, item.Price, "должен использоваться price, а не Price")
	assert.Equal(t, likesCorrect, item.Statistics.Likes, "должен использоваться likes, а не Likes")
	assert.Equal(t, viewCountCorrect, item.Statistics.ViewCount, "должен использоваться viewCount, а не ViewCount")
	assert.Equal(t, contactsCorrect, item.Statistics.Contacts, "должен использоваться contacts, а не Contacts")
}

func TestE2E_CreateDuplicate(t *testing.T) {
	sellerID := rand.Intn(999999-111111+1) + 111111
	name := randomString(10, latinicLower)
	price := rand.Intn(1000) + 1
	likes := rand.Intn(100) + 1
	views := rand.Intn(100) + 1
	contacts := rand.Intn(100) + 1
	payload := map[string]interface{}{
		"sellerID": sellerID,
		"name":     name,
		"price":    price,
		"statistics": map[string]int{
			"likes":     likes,
			"viewCount": views,
			"contacts":  contacts,
		},
	}

	resp1 := postItem(t, payload)
	require.Equal(t, http.StatusOK, resp1.StatusCode)
	uuid1 := extractUUIDFromResponse(t, resp1)
	defer deleteItem(t, uuid1)

	resp2 := postItem(t, payload)
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	uuid2 := extractUUIDFromResponse(t, resp2)
	defer deleteItem(t, uuid2)

	assert.NotEqual(t, uuid1, uuid2, "идентификаторы объявлений должны быть разными")

	item1 := getItemByID(t, uuid1)
	assert.Equal(t, sellerID, item1.SellerID)
	assert.Equal(t, name, item1.Name)
	assert.Equal(t, price, item1.Price)
	assert.Equal(t, likes, item1.Statistics.Likes)
	assert.Equal(t, views, item1.Statistics.ViewCount)
	assert.Equal(t, contacts, item1.Statistics.Contacts)

	item2 := getItemByID(t, uuid2)
	assert.Equal(t, sellerID, item2.SellerID)
	assert.Equal(t, name, item2.Name)
	assert.Equal(t, price, item2.Price)
	assert.Equal(t, likes, item2.Statistics.Likes)
	assert.Equal(t, views, item2.Statistics.ViewCount)
	assert.Equal(t, contacts, item2.Statistics.Contacts)
}
