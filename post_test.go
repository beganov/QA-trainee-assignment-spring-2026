package api_test

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostItem(t *testing.T) {
	ints := make([]int, 240)
	sellerIds := make([]int, 240)
	names := make([]string, 240)
	for i := range ints {
		ints[i] = rand.Intn(maxInt) + 1
		sellerIds[i] = rand.Intn(999999-111111+1) + 111111
		names[i] = randomString(rand.Intn(25)+2, allChars)
	}
	names[0] = randomString(rand.Intn(25)+2, latinicLower)
	names[1] = randomString(rand.Intn(25)+2, cyrillicLower)
	names[2] = randomString(rand.Intn(25)+2, digits)
	names[3] = randomString(rand.Intn(25)+2, uppers)
	names[4] = randomString(rand.Intn(25)+2, latinicLower+spaces)
	names[5] = randomString(rand.Intn(25)+2, specialChars)
	names[6] = randomString(rand.Intn(25)+2, emojis)
	names[7] = randomString(rand.Intn(25)+2, spaces)
	names[8] = randomString(rand.Intn(25)+2, diacritics)
	names[9] = randomString(rand.Intn(25)+2, escapes+latinicLower)

	positiveCases := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			"Создание объявления с латиницей в названии и положительными числовыми полями", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[0],
				"name":       names[0],
				"price":      ints[0],
				"statistics": map[string]int{"likes": ints[1], "viewCount": ints[2], "contacts": ints[3]},
			},
		},
		{
			"Создание объявления с кириллицей в названии", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[1],
				"name":       names[1],
				"price":      ints[4],
				"statistics": map[string]int{"likes": ints[5], "viewCount": ints[6], "contacts": ints[7]},
			},
		},
		{
			"Создание объявления с числовой строкой в названии", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[2],
				"name":       names[2],
				"price":      ints[8],
				"statistics": map[string]int{"likes": ints[9], "viewCount": ints[10], "contacts": ints[11]},
			},
		},
		{
			"Создание объявления с названием в верхнем регистре", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[3],
				"name":       names[3],
				"price":      ints[12],
				"statistics": map[string]int{"likes": ints[13], "viewCount": ints[14], "contacts": ints[15]},
			},
		},
		{
			"Создание объявления с именем, содержащим пробелы", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[4],
				"name":       names[4],
				"price":      ints[16],
				"statistics": map[string]int{"likes": ints[17], "viewCount": ints[18], "contacts": ints[19]},
			},
		},
		{
			"Создание объявления с именем, содержащим спец.символы", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[5],
				"name":       names[5],
				"price":      ints[20],
				"statistics": map[string]int{"likes": ints[21], "viewCount": ints[22], "contacts": ints[23]},
			},
		},
		{
			"Создание объявления с именем, содержащим эмодзи", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[6],
				"name":       names[6],
				"price":      ints[24],
				"statistics": map[string]int{"likes": ints[25], "viewCount": ints[26], "contacts": ints[27]},
			},
		},
		{
			"Создание объявления с именем, содержащим только пробелы", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[7],
				"name":       names[7],
				"price":      ints[28],
				"statistics": map[string]int{"likes": ints[29], "viewCount": ints[30], "contacts": ints[31]},
			},
		},
		{
			"Создание объявления с именем содержащим диакретические символы", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[8],
				"name":       names[8],
				"price":      ints[32],
				"statistics": map[string]int{"likes": ints[33], "viewCount": ints[34], "contacts": ints[35]},
			},
		},
		{
			"Создание объявления с именем содержащим эскейп-последовательности", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[9],
				"name":       names[9],
				"price":      ints[36],
				"statistics": map[string]int{"likes": ints[37], "viewCount": ints[38], "contacts": ints[39]},
			},
		},
		{
			"Создание объявления с именем длиной в 1 символ", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[10],
				"name":       names[10][:1],
				"price":      ints[40],
				"statistics": map[string]int{"likes": ints[41], "viewCount": ints[42], "contacts": ints[43]},
			},
		},
		{
			"Создание объявления, в числовых полях нижняя граница положительных чисел", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   1,
				"name":       "positive lower bound",
				"price":      1,
				"statistics": map[string]int{"likes": 1, "viewCount": 1, "contacts": 1},
			},
		},
		{
			"Создание объявления, в числовых полях верхняя граница int", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   maxInt,
				"name":       "positive upper bound",
				"price":      maxInt,
				"statistics": map[string]int{"likes": maxInt, "viewCount": maxInt, "contacts": maxInt},
			},
		},
		{
			"Создание объявления, значение 0 в числовых полях", // падает в силу багов BR-4, BR-5, BR-6, BR-7, BR-8
			map[string]interface{}{
				"sellerID":   0,
				"name":       "zero",
				"price":      0,
				"statistics": map[string]int{"likes": 0, "viewCount": 0, "contacts": 0},
			},
		},
		{
			"Создание объявления, JSON с дополнительными полями", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":       sellerIds[11],
				"name":           names[11],
				"price":          ints[44],
				"sedddd":         sellerIds[12],
				"o hi mark":      names[12],
				"statistics":     map[string]interface{}{"likes": ints[45], "viewCount": ints[46], "contacts": ints[47]},
				"pe":             ints[48],
				"statisticsaaaa": map[string]interface{}{"likedds": ints[49], "viewfCount": ints[50], "constacts": ints[51]},
			},
		},
		{
			"Создание объявления, поля JSON дублируются в разных регистрах", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[13],
				"name":       names[13],
				"price":      ints[52],
				"sellerId":   sellerIds[14],
				"Name":       names[14],
				"statistics": map[string]interface{}{"likes": ints[53], "viewCount": ints[54], "contacts": ints[55]},
				"pricE":      ints[56],
				"Statistic":  map[string]interface{}{"likES": ints[57], "vieWCount": ints[58], "coNstacts": ints[59]},
			},
		},
		{
			"Создание объявления, поля JSON дублируются в разных регистрах c разными типами данных", // падает в силу бага BR-1
			map[string]interface{}{
				"sellerID":   sellerIds[15],
				"name":       names[15],
				"price":      ints[60],
				"sellerId":   names[16],
				"Name":       sellerIds[16],
				"statistics": map[string]interface{}{"likes": ints[61], "viewCount": ints[62], "contacts": ints[63]},
				"pricE":      names[16],
				"Statistic":  map[string]interface{}{"likES": names[16], "vieWCount": names[16], "coNtacts": names[16]},
			},
		},
	}
	negativeCases := []struct {
		name       string
		payload    map[string]interface{}
		wantStatus int
	}{

		{
			"Создание объявления, значение -1 в sellerID", // падает в силу бага BR-9
			map[string]interface{}{
				"sellerID":   -1,
				"name":       names[20],
				"price":      ints[100],
				"statistics": map[string]int{"likes": ints[101], "viewCount": ints[102], "contacts": ints[103]},
			},
			400,
		},
		{
			"Создание объявления, значение -1 в price", // падает в силу бага BR-10
			map[string]interface{}{
				"sellerID":   sellerIds[21],
				"name":       names[21],
				"price":      -1,
				"statistics": map[string]int{"likes": ints[105], "viewCount": ints[106], "contacts": ints[107]},
			},
			400,
		},
		{
			"Создание объявления, значение -1 в likes", // падает в силу бага BR-11
			map[string]interface{}{
				"sellerID":   sellerIds[22],
				"name":       names[22],
				"price":      ints[108],
				"statistics": map[string]int{"likes": -1, "viewCount": ints[110], "contacts": ints[111]},
			},
			400,
		},
		{
			"Создание объявления, значение -1 в viewCount", // падает в силу бага BR-12
			map[string]interface{}{
				"sellerID":   sellerIds[23],
				"name":       names[23],
				"price":      ints[112],
				"statistics": map[string]int{"likes": ints[113], "viewCount": -1, "contacts": ints[115]},
			},
			400,
		},
		{
			"Создание объявления, значение -1 в contacts", // падает в силу бага BR-13
			map[string]interface{}{
				"sellerID":   sellerIds[24],
				"name":       names[24],
				"price":      ints[116],
				"statistics": map[string]int{"likes": ints[117], "viewCount": -1, "contacts": ints[119]},
			},
			400,
		},
		{
			"Создание объявления, пустая строка в name",
			map[string]interface{}{
				"sellerID":   sellerIds[25],
				"name":       "",
				"price":      ints[120],
				"statistics": map[string]int{"likes": ints[121], "viewCount": ints[122], "contacts": ints[123]},
			},
			400,
		},
		{
			"Создание объявления, значение null в sellerID",
			map[string]interface{}{
				"sellerID":   nil,
				"name":       names[26],
				"price":      ints[124],
				"statistics": map[string]int{"likes": ints[125], "viewCount": ints[126], "contacts": ints[127]},
			},
			400,
		},
		{
			"Создание объявления, значение null в name",
			map[string]interface{}{
				"sellerID":   sellerIds[27],
				"name":       nil,
				"price":      ints[128],
				"statistics": map[string]int{"likes": ints[129], "viewCount": ints[130], "contacts": ints[131]},
			},
			400,
		},
		{
			"Создание объявления, значение null в price",
			map[string]interface{}{
				"sellerID":   sellerIds[28],
				"name":       names[28],
				"price":      nil,
				"statistics": map[string]int{"likes": ints[133], "viewCount": ints[134], "contacts": ints[135]},
			},
			400,
		},
		{
			"Создание объявления, значение null в statistic",
			map[string]interface{}{
				"sellerID":   sellerIds[29],
				"name":       names[29],
				"price":      ints[136],
				"statistics": nil,
			},
			400,
		},
		{
			"Создание объявления, значение null в likes",
			map[string]interface{}{
				"sellerID":   sellerIds[30],
				"name":       names[30],
				"price":      ints[140],
				"statistics": map[string]interface{}{"likes": nil, "viewCount": ints[142], "contacts": ints[143]},
			},
			400,
		},
		{
			"Создание объявления, значение null в viewCount",
			map[string]interface{}{
				"sellerID":   sellerIds[31],
				"name":       names[31],
				"price":      ints[144],
				"statistics": map[string]interface{}{"likes": ints[145], "viewCount": nil, "contacts": ints[147]},
			},
			400,
		},
		{
			"Создание объявления, значение null в contacts",
			map[string]interface{}{
				"sellerID":   sellerIds[32],
				"name":       names[32],
				"price":      ints[148],
				"statistics": map[string]interface{}{"likes": ints[149], "viewCount": ints[150], "contacts": nil},
			},
			400,
		},
		{
			"Создание объявления, sellerID отсутствует",
			map[string]interface{}{
				"name":       names[33],
				"price":      ints[152],
				"statistics": map[string]int{"likes": ints[153], "viewCount": ints[154], "contacts": ints[155]},
			},
			400,
		},
		{
			"Создание объявления, name отсутствует",
			map[string]interface{}{
				"sellerID":   sellerIds[34],
				"price":      ints[156],
				"statistics": map[string]int{"likes": ints[157], "viewCount": ints[158], "contacts": ints[159]},
			},
			400,
		},
		{
			"Создание объявления, price отсутствует",
			map[string]interface{}{
				"sellerID":   sellerIds[35],
				"name":       names[35],
				"statistics": map[string]int{"likes": ints[161], "viewCount": ints[162], "contacts": ints[163]},
			},
			400,
		},
		{
			"Создание объявления, statistic отсутствует",
			map[string]interface{}{
				"sellerID": sellerIds[36],
				"name":     names[36],
				"price":    ints[164],
			},
			400,
		},
		{
			"Создание объявления, likes отсутствует",
			map[string]interface{}{
				"sellerID":   sellerIds[37],
				"name":       names[37],
				"price":      ints[168],
				"statistics": map[string]interface{}{"viewCount": ints[170], "contacts": ints[171]},
			},
			400,
		},
		{
			"Создание объявления, viewCount отсутствует",
			map[string]interface{}{
				"sellerID":   sellerIds[38],
				"name":       names[38],
				"price":      ints[172],
				"statistics": map[string]interface{}{"likes": ints[173], "contacts": ints[175]},
			},
			400,
		},
		{
			"Создание объявления, contacts отсутствует",
			map[string]interface{}{
				"sellerID":   sellerIds[39],
				"name":       names[39],
				"price":      ints[176],
				"statistics": map[string]interface{}{"likes": ints[177], "viewCount": ints[178]},
			},
			400,
		},
		{
			"Создание объявления, поля JSON в разных регистрах", // падает в силу бага BR-3
			map[string]interface{}{
				"sEllerId":   sellerIds[40],
				"NaMe":       names[40],
				"pricE":      ints[180],
				"StatisticS": map[string]interface{}{"likES": ints[181], "vieWCount": ints[182], "coNtacts": ints[183]},
			},
			400,
		},
		{
			"Создание объявления, значение sellerID строка",
			map[string]interface{}{
				"sellerID":   names[41],
				"name":       names[41],
				"price":      ints[184],
				"statistics": map[string]int{"likes": ints[185], "viewCount": ints[186], "contacts": ints[187]},
			},
			400,
		},
		{
			"Создание объявления, значение name - число",
			map[string]interface{}{
				"sellerID":   sellerIds[42],
				"name":       sellerIds[42],
				"price":      ints[188],
				"statistics": map[string]int{"likes": ints[189], "viewCount": ints[190], "contacts": ints[191]},
			},
			400,
		},
		{
			"Создание объявления, значение price - строка",
			map[string]interface{}{
				"sellerID":   sellerIds[43],
				"name":       names[43],
				"price":      names[43],
				"statistics": map[string]int{"likes": ints[193], "viewCount": ints[194], "contacts": ints[195]},
			},
			400,
		},
		{
			"Создание объявления, значение likes  - строка",
			map[string]interface{}{
				"sellerID":   sellerIds[44],
				"name":       names[44],
				"price":      ints[196],
				"statistics": map[string]interface{}{"likes": names[44], "viewCount": ints[198], "contacts": ints[199]},
			},
			400,
		},
		{
			"Создание объявления, значение viewCount - строка",
			map[string]interface{}{
				"sellerID":   sellerIds[45],
				"name":       names[45],
				"price":      ints[200],
				"statistics": map[string]interface{}{"likes": ints[201], "viewCount": names[45], "contacts": ints[203]},
			},
			400,
		},
		{
			"Создание объявления, значение contacts - строка",
			map[string]interface{}{
				"sellerID":   sellerIds[46],
				"name":       names[46],
				"price":      ints[204],
				"statistics": map[string]interface{}{"likes": ints[205], "viewCount": ints[206], "contacts": names[46]},
			},
			400,
		},
		{
			"Создание объявления, булево значение sellerID",
			map[string]interface{}{
				"sellerID":   true,
				"name":       names[47],
				"price":      ints[208],
				"statistics": map[string]int{"likes": ints[209], "viewCount": ints[210], "contacts": ints[211]},
			},
			400,
		},
		{
			"Создание объявления, булево значение name",
			map[string]interface{}{
				"sellerID":   sellerIds[48],
				"name":       false,
				"price":      ints[212],
				"statistics": map[string]int{"likes": ints[213], "viewCount": ints[214], "contacts": ints[215]},
			},
			400,
		},
		{
			"Создание объявления, значение price - float",
			map[string]interface{}{
				"sellerID":   sellerIds[49],
				"name":       names[49],
				"price":      1.5,
				"statistics": map[string]int{"likes": ints[213], "viewCount": ints[214], "contacts": ints[215]},
			},
			400,
		},
		{
			"Создание объявления, значение likes  - float",
			map[string]interface{}{
				"sellerID":   sellerIds[50],
				"name":       names[50],
				"price":      ints[216],
				"statistics": map[string]interface{}{"likes": 1.5, "viewCount": ints[218], "contacts": ints[219]},
			},
			400,
		},
		{
			"Создание объявления, значение viewCount - массив",
			map[string]interface{}{
				"sellerID":   sellerIds[51],
				"name":       names[51],
				"price":      ints[220],
				"statistics": map[string]interface{}{"likes": ints[221], "viewCount": []int{ints[222]}, "contacts": ints[223]},
			},
			400,
		},
		{
			"Создание объявления, значение contacts - массив",
			map[string]interface{}{
				"sellerID":   sellerIds[52],
				"name":       names[52],
				"price":      ints[224],
				"statistics": map[string]interface{}{"likes": ints[225], "viewCount": ints[226], "contacts": []int{ints[227]}},
			},
			400,
		},
		{
			"Создание объявления, значение statistics - массив",
			map[string]interface{}{
				"sellerID":   sellerIds[53],
				"name":       names[53],
				"price":      ints[228],
				"statistics": []int{ints[229], ints[230], ints[231]},
			},
			400,
		},
	}

	t.Run("positive", func(t *testing.T) {
		for _, tc := range positiveCases {
			t.Run(tc.name, func(t *testing.T) {
				resp := postItem(t, tc.payload)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode, "статус не 200")
				var created Item
				err := json.NewDecoder(resp.Body).Decode(&created)
				require.NoError(t, err, "не удалось распарсить ответ")
				assert.NotEmpty(t, created.ID, "id пустой")
				assert.Equal(t, tc.payload["sellerID"], created.SellerID)
				assert.Equal(t, tc.payload["name"], created.Name)
				assert.Equal(t, tc.payload["price"], created.Price)
				if stats, ok := tc.payload["statistics"].(map[string]int); ok {
					assert.Equal(t, stats["likes"], created.Statistics.Likes)
					assert.Equal(t, stats["viewCount"], created.Statistics.ViewCount)
					assert.Equal(t, stats["contacts"], created.Statistics.Contacts)
				}
				deleteItem(t, created.ID)
			})
		}
	})
	IdempotenceJson := map[string]interface{}{
		"sellerID":   sellerIds[19],
		"name":       names[19],
		"price":      ints[96],
		"statistics": map[string]interface{}{"likes": ints[97], "viewCount": ints[98], "contacts": ints[99]},
	}
	t.Run("Idempotence", func(t *testing.T) {
		resp := postItem(t, IdempotenceJson)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode, "статус не соответствует ожидаемому")
		resp2 := postItem(t, IdempotenceJson)
		defer resp2.Body.Close()
		assert.Equal(t, 200, resp2.StatusCode, "статус не соответствует ожидаемому")
		var created, created2 Item
		err := json.NewDecoder(resp.Body).Decode(&created)
		require.NoError(t, err, "не удалось распарсить ответ")
		err2 := json.NewDecoder(resp2.Body).Decode(&created2)
		require.NoError(t, err2, "не удалось распарсить ответ")
		assert.NotEmpty(t, created.ID, "id пустой")
		assert.NotEmpty(t, created2.ID, "id пустой")
		assert.Equal(t, created.SellerID, created2.SellerID)
		assert.Equal(t, created.Name, created2.Name)
		assert.Equal(t, created.Price, created2.Price)
		assert.Equal(t, created.Statistics.Likes, created2.Statistics.Likes)
		assert.Equal(t, created.Statistics.ViewCount, created2.Statistics.ViewCount)
		assert.Equal(t, created.Statistics.Contacts, created2.Statistics.Contacts)
	})
	t.Run("negative", func(t *testing.T) {
		for _, tc := range negativeCases {
			t.Run(tc.name, func(t *testing.T) {
				resp := postItem(t, tc.payload)
				defer resp.Body.Close()
				assert.Equal(t, tc.wantStatus, resp.StatusCode, "статус не соответствует ожидаемому")
			})
		}
	})

	nonJSON := `{"sellerID":123456,"name":"test4775808,"statistics":{"likes":10,"viewCount":10,"contacts":10}}`
	t.Run("non valid json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/api/1/item", strings.NewReader(nonJSON))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	bigIntJSON := `{  "sellerID": 9223372036854775808,  "name": "nnШaмZ62уСзПгТЧцLр2P",  "price": 744665813,  "statistics": {    "likes": 746487284,"viewCount": 954282558,"contacts": 502532000}}`
	t.Run("big json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/api/1/item", strings.NewReader(bigIntJSON))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
	lowerIntJSON := `{  "sellerID": -9223372036854775809,  "name": "nnШaмZ62уСзПгТЧцLр2P",  "price": 744665813,  "statistics": {    "likes": 746487284,"viewCount": 954282558,"contacts": 502532000}}`
	t.Run("low json", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/api/1/item", strings.NewReader(lowerIntJSON))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
	anothermethodsJSON := `{  "sellerID": 1,  "name": "nnШaмZ62уСзПгТЧцLр2P",  "price": 744665813,  "statistics": {    "likes": 746487284,"viewCount": 954282558,"contacts": 502532000}}`
	t.Run("unsupported methods", func(t *testing.T) {
		methods := []struct {
			method string
			body   io.Reader
		}{
			{http.MethodGet, nil},
			{http.MethodPut, strings.NewReader(anothermethodsJSON)},
			{http.MethodPatch, strings.NewReader(anothermethodsJSON)},
			{http.MethodDelete, nil},
			{http.MethodHead, nil},
			{http.MethodOptions, nil},
		}
		for _, m := range methods {
			t.Run(m.method, func(t *testing.T) {
				req, err := http.NewRequest(m.method, baseURL+"/api/1/item", m.body)
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "метод %s должен возвращать 405", m.method)
			})
		}
	})
}
