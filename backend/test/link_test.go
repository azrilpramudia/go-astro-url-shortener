package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearLinks(t *testing.T) {
	t.Helper()
	require.NoError(t, db.Exec("TRUNCATE TABLE links RESTART IDENTITY").Error)
}

func postLink(t *testing.T, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func getCode(t *testing.T, code string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func TestCreateLink_Success(t *testing.T) {
	clearLinks(t)

	resp := postLink(t, `{"url":"https://go.dev"}`)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var body model.WebResponse[model.LinkResponse]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Len(t, body.Data.Code, 7)
	assert.Equal(t, "https://go.dev", body.Data.OriginalURL)
	assert.True(t, strings.HasSuffix(body.Data.ShortURL, "/"+body.Data.Code))

	var link entity.Link
	require.NoError(t, db.Where("code = ?", body.Data.Code).Take(&link).Error)
	assert.Equal(t, "https://go.dev", link.OriginalURL)
	assert.Equal(t, int64(0), link.Clicks)
}

func TestCreateLink_Invalid(t *testing.T) {
	clearLinks(t)

	cases := map[string]string{
		"bukan url":            `{"url":"halo"}`,
		"skema berbahaya":      `{"url":"javascript:alert(1)"}`,
		"url kosong":           `{"url":""}`,
		"expires di masa lalu": `{"url":"https://go.dev","expires_at":"2020-01-01T00:00:00Z"}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			resp := postLink(t, body)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestRedirect_Success(t *testing.T) {
	clearLinks(t)

	link := entity.Link{Code: "abc1234", OriginalURL: "https://go.dev"}
	require.NoError(t, db.Create(&link).Error)

	resp := getCode(t, "abc1234")
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "https://go.dev", resp.Header.Get("Location"))

	var updated entity.Link
	require.NoError(t, db.First(&updated, link.ID).Error)
	assert.Equal(t, int64(1), updated.Clicks)
}

func TestRedirect_NotFound(t *testing.T) {
	clearLinks(t)

	resp := getCode(t, "tidakada")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRedirect_Expired(t *testing.T) {
	clearLinks(t)

	expired := time.Now().Add(-time.Hour)
	link := entity.Link{Code: "old1234", OriginalURL: "https://go.dev", ExpiresAt: &expired}
	require.NoError(t, db.Create(&link).Error)

	resp := getCode(t, "old1234")
	assert.Equal(t, http.StatusGone, resp.StatusCode)

	var updated entity.Link
	require.NoError(t, db.First(&updated, link.ID).Error)
	assert.Equal(t, int64(0), updated.Clicks)
}
