package test

import (
	"context"
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
	require.NoError(t, db.Exec("TRUNCATE TABLE link_tags, link_analytics_events, links, users RESTART IDENTITY").Error)
	require.NoError(t, redisClient.FlushDB(context.Background()).Err())
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

func visitorCookie(t *testing.T, resp *http.Response) *http.Cookie {
	t.Helper()
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "url_shortener_visitor" {
			return cookie
		}
	}
	t.Fatal("visitor cookie not set")
	return nil
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
	cookie := visitorCookie(t, resp)
	assert.True(t, cookie.HttpOnly)

	require.NoError(t, flusher.Flush(context.Background()))

	var updated entity.Link
	require.NoError(t, db.First(&updated, link.ID).Error)
	assert.Equal(t, int64(1), updated.Clicks)

	var event entity.LinkAnalyticsEvent
	require.NoError(t, db.Where("link_id = ?", link.ID).Take(&event).Error)
	assert.NotEmpty(t, event.VisitorHash)
	assert.NotEqual(t, cookie.Value, event.VisitorHash)
	assert.Equal(t, "unknown", event.DeviceType)
}

func TestRedirect_ReusesVisitorCookieAndCapturesMetadata(t *testing.T) {
	clearLinks(t)

	link := entity.Link{Code: "visit12", OriginalURL: "https://go.dev"}
	require.NoError(t, db.Create(&link).Error)

	firstResponse := getCode(t, link.Code)
	cookie := visitorCookie(t, firstResponse)

	req := httptest.NewRequest(http.MethodGet, "/"+link.Code, nil)
	req.AddCookie(cookie)
	req.Header.Set("Referer", "https://news.example/story?private=value")
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone) Mobile")
	secondResponse, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, secondResponse.StatusCode)

	var events []entity.LinkAnalyticsEvent
	require.NoError(t, db.Where("link_id = ?", link.ID).Order("occurred_at ASC").Find(&events).Error)
	require.Len(t, events, 2)
	assert.Equal(t, events[0].VisitorHash, events[1].VisitorHash)
	assert.Equal(t, "news.example", events[1].ReferrerHost)
	assert.Equal(t, "mobile", events[1].DeviceType)
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

func TestGetLink(t *testing.T) {
	clearLinks(t)

	link := entity.Link{Code: "abc1234", OriginalURL: "https://go.dev", Clicks: 3}
	require.NoError(t, db.Exec("TRUNCATE TABLE link_tags, link_analytics_events, links, users RESTART IDENTITY").Error)
	require.NoError(t, db.Create(&link).Error)

	req := httptest.NewRequest(http.MethodGet, "/api/links/abc1234", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body model.WebResponse[model.LinkResponse]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, int64(3), body.Data.Clicks)

	req = httptest.NewRequest(http.MethodGet, "/api/links/tidakada", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAnalyticsCleanupRemovesExpiredEvents(t *testing.T) {
	clearLinks(t)

	link := entity.Link{Code: "clean12", OriginalURL: "https://go.dev"}
	require.NoError(t, db.Create(&link).Error)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&[]entity.LinkAnalyticsEvent{
		{
			LinkID:      link.ID,
			VisitorHash: "old-visitor",
			DeviceType:  "desktop",
			OccurredAt:  now.Add(-91 * 24 * time.Hour),
		},
		{
			LinkID:      link.ID,
			VisitorHash: "recent-visitor",
			DeviceType:  "mobile",
			OccurredAt:  now.Add(-89 * 24 * time.Hour),
		},
	}).Error)

	analyticsCleaner.Cleanup(context.Background())

	var events []entity.LinkAnalyticsEvent
	require.NoError(t, db.Where("link_id = ?", link.ID).Find(&events).Error)
	require.Len(t, events, 1)
	assert.Equal(t, "recent-visitor", events[0].VisitorHash)
}
