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
		city  string
		count int
		want  int
	}{
		{"moscow", 0, 0},
		{"moscow", 1, 1},
		{"moscow", 2, 2},
		{"moscow", 5, 5},
		{"moscow", 100, len(cafeList["moscow"])},

		{"tula", 0, 0},
		{"tula", 1, 1},
		{"tula", 2, 2},
		{"tula", 3, 3},
		{"tula", 100, len(cafeList["tula"])},
	}

	for _, test := range requests {
		url := fmt.Sprintf("/cafe?city=%s&count=%d", test.city, test.count)

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

		for i := 0; i < len(cafes) && i < len(cafeList[test.city]); i++ {
			assert.Equal(t, cafeList[test.city][i], cafes[i])
		}
	}
}
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		city      string
		search    string
		wantCount int
		wantNames []string
	}{
		{"moscow", "фасоль", 0, []string{}},
		{"moscow", "кофе", 2, []string{"Мир кофе", "Кофе и завтраки"}},
		{"moscow", "вилка", 1, []string{"Ложка и вилка"}},
		{"moscow", "КОФЕ", 2, []string{"Мир кофе", "Кофе и завтраки"}},
		{"moscow", "завтра", 1, []string{"Кофе и завтраки"}},

		{"tula", "пир", 1, []string{"Пир и мир"}},
		{"tula", "красиво", 1, []string{"Красиво есть не запретишь"}},
		{"tula", "завтрак", 2, []string{"Красиво есть не запретишь", "Поздний завтрак"}},
		{"tula", "поздний", 1, []string{"Поздний завтрак"}},
	}

	for _, test := range requests {

		url := fmt.Sprintf("/cafe?city=%s&search=%s", test.city, test.search)
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)
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
			assert.True(t, strings.Contains(strings.ToLower(name), strings.ToLower(test.search)), "Название не содержит искомую строку")
		}
	}
}
