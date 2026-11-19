package test

import (
	"context"
	"net/http"
	"testing"

	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirect_ServedFromCache(t *testing.T) {
	clearLinks(t)

	link := entity.Link{Code: "cache01", OriginalURL: "https://go.dev"}
	require.NoError(t, db.Create(&link).Error)

	assert.Equal(t, http.StatusFound, getCode(t, "cache01").StatusCode)

	require.NoError(t, db.Exec("DELETE FROM links WHERE code = ?", "cache01").Error)

	assert.Equal(t, http.StatusFound, getCode(t, "cache01").StatusCode)
}

func TestDeleteLink_InvalidatesCache(t *testing.T) {
	clearLinks(t)
	owner := registerUser(t, "cache@example.com")

	resp := doJSON(t, http.MethodPost, "/api/links", `{"url":"https://go.dev"}`, owner)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	code := decode[model.LinkResponse](t, resp).Code

	assert.Equal(t, http.StatusFound, getCode(t, code).StatusCode)

	resp = doJSON(t, http.MethodDelete, "/api/links/"+code, "", owner)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	assert.Equal(t, http.StatusNotFound, getCode(t, code).StatusCode)
}

func TestClicksAreBatched(t *testing.T) {
	clearLinks(t)

	link := entity.Link{Code: "batch01", OriginalURL: "https://go.dev"}
	require.NoError(t, db.Create(&link).Error)

	for i := 0; i < 3; i++ {
		assert.Equal(t, http.StatusFound, getCode(t, "batch01").StatusCode)
	}

	var before entity.Link
	require.NoError(t, db.First(&before, link.ID).Error)
	assert.Equal(t, int64(0), before.Clicks)

	require.NoError(t, flusher.Flush(context.Background()))

	var after entity.Link
	require.NoError(t, db.First(&after, link.ID).Error)
	assert.Equal(t, int64(3), after.Clicks)

	require.NoError(t, flusher.Flush(context.Background()))
	require.NoError(t, db.First(&after, link.ID).Error)
	assert.Equal(t, int64(3), after.Clicks)
}

func TestNotFound_IsNegativeCached(t *testing.T) {
	clearLinks(t)

	assert.Equal(t, http.StatusNotFound, getCode(t, "ghost01").StatusCode)

	n, err := redisClient.Exists(context.Background(), "link:ghost01").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestRedirect_InvalidCode(t *testing.T) {
	clearLinks(t)

	assert.Equal(t, http.StatusNotFound, getCode(t, "favicon.ico").StatusCode)

	n, err := redisClient.Exists(context.Background(), "link:favicon.ico").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
}
