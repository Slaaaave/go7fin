package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	city := "moscow"
	countCity := len(cafeList[city])

	requests := []struct {
		count  int
		expect int
	}{
		{count: 0, expect: 0},
		{count: 1, expect: 1},
		{count: 2, expect: 2},
		{count: 100, expect: countCity},
	}

	handler := http.HandlerFunc(mainHandle)
	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count)
		req := httptest.NewRequest("GET", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())

		var actualCount int
		if body == "" {
			actualCount = 0
		} else {
			cafes := strings.Split(body, ",")
			actualCount = len(cafes)
		}

		assert.Equal(t, v.expect, actualCount, "Для count=%d ожидалось %d кафе, получили %d", v.count, v.expect, actualCount)
	}
}

func TestCafeSearch(t *testing.T) {
	city := "moscow"

	requests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 0},
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}
	handler := http.HandlerFunc(mainHandle)
	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=%s&search=%s", city, v.search)
		req := httptest.NewRequest("GET", url, nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)
		body := strings.TrimSpace(response.Body.String())

		if v.wantCount == 0 {
			assert.Empty(t, body, "В случае поиска '%s' ожидался нулевой ответ", v.search)
			continue // Переходим к следующему кейсу в цикле
		}

		cafes := strings.Split(body, ",")

		assert.Len(t, cafes, v.wantCount, "В случае поиска '%s' ожидалось %d кафе", v.search, v.wantCount)

		for _, name := range cafes {
			assert.Containsf(
				t,
				strings.ToLower(name),
				strings.ToLower(v.search),
				"В названии кафе '%s' не содержится строка '%s'",
				name,
				v.search,
			)
		}
	}
}
