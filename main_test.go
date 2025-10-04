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
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},                         // 0 записей
		{1, 1},                         // 1 запись
		{2, 2},                         // 2 записи
		{100, len(cafeList["moscow"])}, // больше, чем есть в Москве (5)
		{3, len(cafeList["tula"])},     // больше, чем есть в Туле (3)
	}

	for _, test := range requests {
		url := fmt.Sprintf("/cafe?city=moscow&count=%d", test.count)

		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(body, ",")

		if body == "" {
			cafes = []string{}
		}

		assert.Equal(t, test.want, len(cafes))
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},  // Нет совпадений
		{"кофе", 2},    // "Мир кофе", "Кофе и завтраки"
		{"вилка", 1},   // "Ложка и вилка"
		{"слад", 1},    // "Сладкоежка"
		{"и", 3},       // 3 кафе содержат "и"
		{"студент", 1}, // "Сытый студент"
	}

	for _, test := range requests {
		url := fmt.Sprintf("/cafe?city=moscow&search=%s", test.search)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(body, ",")

		if body == "" {
			cafes = []string{}
		}

		assert.Equal(t, test.wantCount, len(cafes), fmt.Sprintf("Для поиска '%s' ожидалось %d кафе, найдено %d", test.search, test.wantCount, len(cafes)))

		for _, cafe := range cafes {

			assert.True(t, strings.Contains(strings.ToLower(cafe), strings.ToLower(test.search)), fmt.Sprintf("Кафе %q не содержит строку %q", cafe, test.search))
		}
	}
}
