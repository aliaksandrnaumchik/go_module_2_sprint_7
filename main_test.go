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

const testRequestTypeGet = "GET"

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
		req := httptest.NewRequest(testRequestTypeGet, v.request, nil)
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
		req := httptest.NewRequest(testRequestTypeGet, v, nil)

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
		{0, 0},
		{1, 1},
		{2, 2},
		{5, 5},
		{10, 5},
		{100, 5},
	}

	for _, test := range requests {
		if test.count < 0 {
			t.Skip("Пропускаем отрицательные значения count")
			continue
		}

		url := fmt.Sprintf("/cafe?city=moscow&count=%d", test.count)
		response := httptest.NewRecorder()
		req := httptest.NewRequest(testRequestTypeGet, url, nil)
		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(body, ",")

		if body == "" {
			cafes = []string{}
		}

		assert.Equal(t, test.want, len(cafes))

		for i := 0; i < len(cafes) && i < len(cafeList["moscow"]); i++ {
			assert.Equal(t, cafeList["moscow"][i], cafes[i])
		}
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
		wantNames []string
	}{
		{"", 5, cafeList["moscow"]},
		{"фасоль", 0, []string{}},
		{"кофе", 2, []string{"Мир кофе", "Кофе и завтраки"}},
		{"вилка", 1, []string{"Ложка и вилка"}},
		{"КОФЕ", 2, []string{"Мир кофе", "Кофе и завтраки"}},
		{"завтра", 1, []string{"Кофе и завтраки"}},
		{"и", 4, []string{"Мир кофе", "Кофе и завтраки", "Сытый студент", "Ложка и вилка"}},
		{"слад", 1, []string{"Сладкоежка"}},
		{"студент", 1, []string{"Сытый студент"}},
	}

	for _, test := range requests {
		url := fmt.Sprintf("/cafe?city=moscow&search=%s", test.search)
		response := httptest.NewRecorder()
		req := httptest.NewRequest(testRequestTypeGet, url, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(body, ",")

		if body == "" {
			cafes = []string{}
			assert.Equal(t, test.wantCount, len(cafes))
			return
		}

		assert.Equal(t, test.wantCount, len(cafes))

		for _, name := range cafes {
			assert.Contains(t, test.wantNames, name)
			assert.True(t, strings.Contains(strings.ToLower(name), strings.ToLower(test.search)))
		}
	}
}
