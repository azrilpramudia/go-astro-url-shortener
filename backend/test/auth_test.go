package test

import (
	"encoding/json"
	"fmt"
	"io"
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

func doJSON(t *testing.T, method, path, body string, cookie *http.Cookie) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	var body model.WebResponse[T]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body.Data
}

func tokenCookie(t *testing.T, resp *http.Response) *http.Cookie {
	t.Helper()
	for _, c := range resp.Cookies() {
		if c.Name == "token" {
			return c
		}
	}
	t.Fatal("token cookie not set")
	return nil
}

func registerUser(t *testing.T, email string) *http.Cookie {
	t.Helper()
	body := fmt.Sprintf(`{"name":"Test User","email":%q,"password":"password123"}`, email)
	resp := doJSON(t, http.MethodPost, "/api/auth/register", body, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	return tokenCookie(t, resp)
}

func TestRegister_Success(t *testing.T) {
	clearLinks(t)

	body := `{"name":"Test User","email":"Test@Example.com","password":"password123"}`
	resp := doJSON(t, http.MethodPost, "/api/auth/register", body, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	cookie := tokenCookie(t, resp)
	assert.True(t, cookie.HttpOnly)
	assert.NotEmpty(t, cookie.Value)

	user := decode[model.UserResponse](t, resp)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	clearLinks(t)
	registerUser(t, "dup@example.com")

	body := `{"name":"Other","email":"DUP@example.com","password":"password123"}`
	resp := doJSON(t, http.MethodPost, "/api/auth/register", body, nil)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestRegister_Invalid(t *testing.T) {
	clearLinks(t)

	cases := map[string]string{
		"password pendek":    `{"name":"Test","email":"a@example.com","password":"short"}`,
		"email salah":        `{"name":"Test","email":"bukan-email","password":"password123"}`,
		"nama kosong":        `{"name":"","email":"a@example.com","password":"password123"}`,
		"password > 72 byte": fmt.Sprintf(`{"name":"Test","email":"a@example.com","password":%q}`, strings.Repeat("a", 73)),
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			resp := doJSON(t, http.MethodPost, "/api/auth/register", body, nil)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestLogin(t *testing.T) {
	clearLinks(t)
	registerUser(t, "login@example.com")

	resp := doJSON(t, http.MethodPost, "/api/auth/login", `{"email":"login@example.com","password":"password123"}`, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, tokenCookie(t, resp).Value)

	wrongPass := doJSON(t, http.MethodPost, "/api/auth/login", `{"email":"login@example.com","password":"salahsalah"}`, nil)
	unknown := doJSON(t, http.MethodPost, "/api/auth/login", `{"email":"tidakada@example.com","password":"password123"}`, nil)
	assert.Equal(t, http.StatusUnauthorized, wrongPass.StatusCode)
	assert.Equal(t, http.StatusUnauthorized, unknown.StatusCode)

	b1, _ := io.ReadAll(wrongPass.Body)
	b2, _ := io.ReadAll(unknown.Body)
	assert.JSONEq(t, string(b1), string(b2))
}

func TestMe(t *testing.T) {
	clearLinks(t)
	cookie := registerUser(t, "me@example.com")

	resp := doJSON(t, http.MethodGet, "/api/auth/me", "", cookie)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "me@example.com", decode[model.UserResponse](t, resp).Email)

	resp = doJSON(t, http.MethodGet, "/api/auth/me", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp = doJSON(t, http.MethodGet, "/api/auth/me", "", &http.Cookie{Name: "token", Value: "bukan.token.valid"})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogout(t *testing.T) {
	clearLinks(t)

	resp := doJSON(t, http.MethodPost, "/api/auth/logout", "", nil)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Empty(t, tokenCookie(t, resp).Value)
}

func TestLinkOwnership(t *testing.T) {
	clearLinks(t)
	alice := registerUser(t, "alice@example.com")
	bob := registerUser(t, "bob@example.com")

	resp := doJSON(t, http.MethodPost, "/api/links", `{"url":"https://go.dev"}`, alice)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	aliceCode := decode[model.LinkResponse](t, resp).Code
	link := entity.Link{}
	require.NoError(t, db.Where("code = ?", aliceCode).Take(&link).Error)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&[]entity.LinkAnalyticsEvent{
		{
			LinkID:       link.ID,
			VisitorHash:  "hashed-visitor",
			ReferrerHost: "news.example",
			DeviceType:   "mobile",
			CountryCode:  "ID",
			OccurredAt:   now,
		},
		{
			LinkID:       link.ID,
			VisitorHash:  "hashed-visitor",
			ReferrerHost: "news.example",
			DeviceType:   "mobile",
			CountryCode:  "ID",
			OccurredAt:   now.Add(time.Second),
		},
	}).Error)

	resp = doJSON(t, http.MethodGet, "/api/me/links/"+aliceCode+"/analytics?days=30", "", alice)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	analytics := decode[model.LinkAnalyticsResponse](t, resp)
	assert.Equal(t, int64(2), analytics.TotalClicks)
	assert.Equal(t, int64(1), analytics.UniqueVisitors)
	require.Len(t, analytics.Referrers, 1)
	assert.Equal(t, "news.example", analytics.Referrers[0].Name)
	assert.Equal(t, "mobile", analytics.Devices[0].Name)
	assert.Equal(t, "ID", analytics.Countries[0].Name)

	resp = doJSON(t, http.MethodGet, "/api/me/links/"+aliceCode+"/analytics", "", bob)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp = doJSON(t, http.MethodGet, "/api/me/links/"+aliceCode+"/analytics", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp = doJSON(t, http.MethodPost, "/api/links", `{"url":"https://astro.build"}`, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp = doJSON(t, http.MethodGet, "/api/me/links", "", alice)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	list := decode[model.ListLinksResponse](t, resp)
	assert.Equal(t, int64(1), list.Total)
	require.Len(t, list.Items, 1)
	assert.Equal(t, aliceCode, list.Items[0].Code)

	resp = doJSON(t, http.MethodGet, "/api/me/links", "", bob)
	assert.Equal(t, int64(0), decode[model.ListLinksResponse](t, resp).Total)

	resp = doJSON(t, http.MethodGet, "/api/me/links", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp = doJSON(t, http.MethodDelete, "/api/links/"+aliceCode, "", bob)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp = doJSON(t, http.MethodDelete, "/api/links/"+aliceCode, "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp = doJSON(t, http.MethodDelete, "/api/links/"+aliceCode, "", alice)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp = doJSON(t, http.MethodGet, "/api/me/links", "", alice)
	assert.Equal(t, int64(0), decode[model.ListLinksResponse](t, resp).Total)
}

func TestCustomAliasEditAndLinkFilters(t *testing.T) {
	clearLinks(t)
	alice := registerUser(t, "alias@example.com")
	bob := registerUser(t, "alias-other@example.com")

	resp := doJSON(
		t,
		http.MethodPost,
		"/api/links",
		`{"url":"https://old.example/path","custom_code":"my-blog","tags":["Product","campaign","product"]}`,
		alice,
	)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	created := decode[model.LinkResponse](t, resp)
	assert.Equal(t, "my-blog", created.Code)
	assert.Equal(t, []string{"product", "campaign"}, created.Tags)

	resp = doJSON(
		t,
		http.MethodPost,
		"/api/links",
		`{"url":"https://other.example","custom_code":"my-blog"}`,
		bob,
	)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	resp = doJSON(
		t,
		http.MethodPost,
		"/api/links",
		`{"url":"https://other.example","custom_code":"bad alias"}`,
		bob,
	)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp = doJSON(
		t,
		http.MethodPost,
		"/api/links",
		`{"url":"https://other.example","tags":["campaign"]}`,
		nil,
	)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	redirect := getCode(t, created.Code)
	require.Equal(t, http.StatusFound, redirect.StatusCode)
	assert.Equal(t, "https://old.example/path", redirect.Header.Get("Location"))

	expiresAt := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	updateBody := fmt.Sprintf(
		`{"url":"https://new.example/path","expires_at":%q,"tags":["campaign","launch"]}`,
		expiresAt,
	)
	resp = doJSON(t, http.MethodPut, "/api/links/"+created.Code, updateBody, alice)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	updated := decode[model.LinkResponse](t, resp)
	assert.Equal(t, "https://new.example/path", updated.OriginalURL)
	assert.Equal(t, []string{"campaign", "launch"}, updated.Tags)

	redirect = getCode(t, created.Code)
	require.Equal(t, http.StatusFound, redirect.StatusCode)
	assert.Equal(t, "https://new.example/path", redirect.Header.Get("Location"))

	resp = doJSON(
		t,
		http.MethodGet,
		"/api/me/links?q=new.example&status=active&tag=campaign",
		"",
		alice,
	)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	list := decode[model.ListLinksResponse](t, resp)
	require.Len(t, list.Items, 1)
	assert.Equal(t, created.Code, list.Items[0].Code)
	assert.Equal(t, []string{"campaign", "launch"}, list.Items[0].Tags)

	resp = doJSON(t, http.MethodGet, "/api/me/links?q=new.example", "", bob)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int64(0), decode[model.ListLinksResponse](t, resp).Total)

	resp = doJSON(t, http.MethodPut, "/api/links/"+created.Code, updateBody, bob)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
