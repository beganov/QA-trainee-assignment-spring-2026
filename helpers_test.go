package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	maxInt        = 9223372036854775807
	baseURL       = "https://qa-internship.avito.com"
	latinicLower  = "abcdefghijklmnopqrstuvwxyz"
	cyrillicLower = "абвгдеёжзийклмнопрстуфхцчшщъыьэюя"
	uppers        = "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits        = "0123456789"
	specialChars  = ".,\\<>@#$%&*+=:;|[]{}~'!\"~`^()_-№"
	spaces        = " "
	emojis        = "😀😃😄😁😆😅🤣😂🙂🙃🫠😉😊😇🥰😍🤩😘😗😚😙😋😛😜🤪😝🤑"
	diacritics    = "áéíóúýčďěňřšťžäöüßçñÁÉÍÓÚÝČĎĚŇŘŠŤŽÄÖÜÇÑ"
	escapes       = "\n\r\t\\\""
	allChars      = latinicLower + uppers + cyrillicLower + digits + specialChars + spaces + emojis + diacritics + escapes
)

type Item struct {
	ID         string     `json:"id"`
	SellerID   int        `json:"sellerId"`
	Name       string     `json:"name"`
	Price      int        `json:"price"`
	Statistics Statistics `json:"statistics"`
	CreatedAt  string     `json:"createdAt"`
}

type Statistics struct {
	Likes     int `json:"likes"`
	ViewCount int `json:"viewCount"`
	Contacts  int `json:"contacts"`
}

type postResponse struct {
	Status string `json:"status"`
}

func postItem(t *testing.T, payload interface{}) *http.Response {
	body, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/api/1/item", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	return resp
}

func deleteItem(t *testing.T, id string) {
	req, _ := http.NewRequest("DELETE", baseURL+"/api/2/item/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
}

func randomString(n int, letters string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getStatistic(t *testing.T, uuid string) ([]Statistics, int) {
	resp, err := http.Get(baseURL + "/api/1/statistic/" + uuid)
	require.NoError(t, err)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var stats []Statistics
	err = json.NewDecoder(resp.Body).Decode(&stats)
	require.NoError(t, err)
	return stats, resp.StatusCode
}

func getItemsBySeller(t *testing.T, sellerID int) ([]Item, int) {
	url := fmt.Sprintf("%s/api/1/%d/item", baseURL, sellerID)
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var items []Item
	err = json.NewDecoder(resp.Body).Decode(&items)
	require.NoError(t, err)
	return items, resp.StatusCode
}

func extractUUIDFromResponse(t *testing.T, resp *http.Response) string {
	var pr postResponse
	err := json.NewDecoder(resp.Body).Decode(&pr)
	require.NoError(t, err)
	parts := strings.Split(pr.Status, " - ")
	require.Len(t, parts, 2, "не удалось извлечь UUID из ответа")
	return parts[1]
}

func getItemByID(t *testing.T, uuid string) Item {
	resp, err := http.Get(baseURL + "/api/1/item/" + uuid)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var items []Item
	err = json.NewDecoder(resp.Body).Decode(&items)
	require.NoError(t, err)
	require.Len(t, items, 1)
	return items[0]
}

func sumDurations(durs []time.Duration) time.Duration {
	var sum time.Duration
	for _, d := range durs {
		sum += d
	}
	return sum
}
