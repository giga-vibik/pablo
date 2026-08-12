// Package zernio — клиент публикующего API zernio (getlate.dev).
//
// Модель: Profile объединяет подключённые Account'ы; Post таргетится в один или
// несколько аккаунтов.
//
// Профиль клиент находит сам: Pablo однопользовательский, профиль там ровно
// один, поэтому держать его id в конфиге незачем — см. resolveProfileID.
//
// Base: https://zernio.com/api/v1
// Auth: Authorization: Bearer sk_<...>
// ID возвращаются в Mongo-стиле как "_id".
//
// Медиа передаётся ПУБЛИЧНЫМ URL — zernio сам скачивает файл. Поэтому видео
// сначала должно уехать в S3 с публичным доступом.
//
// Подключение аккаунтов (OAuth) делается через GetConnectURL: zernio хостит
// страницу авторизации и редиректит обратно.
package zernio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pablo/backend/internal/config"
)

const (
	defaultBaseURL = "https://zernio.com/api/v1"
	defaultTimeout = 45 * time.Second

	// Имя профиля, который создаётся, если под ключом нет ни одного.
	defaultProfileName = "Pablo"
)

// ErrNotConfigured — не задан API-ключ.
var ErrNotConfigured = errors.New("zernio: API key not configured")

// ErrNoAccount — под профилем нет подключённого аккаунта для площадки.
var ErrNoAccount = errors.New("zernio: no connected account for platform")

// ErrDuplicateContent — zernio отверг пост с HTTP 409: тот же контент уже
// запланирован/публикуется или был опубликован в этот аккаунт за последние 24ч.
// При ретрае это значит, что ПРЕДЫДУЩАЯ попытка прошла, поэтому вызывающий
// трактует ошибку как идемпотентный успех.
var ErrDuplicateContent = errors.New("zernio: duplicate content (409) — already scheduled/publishing/posted")

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string

	// Профиль резолвится при первом обращении и кэшируется. Не sync.Once:
	// сеть может лечь, и тогда повтор должен пройти заново.
	profileMu sync.Mutex
	profileID string
}

func New(cfg config.Zernio, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		httpClient: httpClient,
		apiKey:     strings.TrimSpace(cfg.APIKey),
		baseURL:    baseURL,
	}
}

func (c *Client) IsConfigured() bool { return c.apiKey != "" }

// ─────────────────────────────────────────────────────────────
// Profiles
// ─────────────────────────────────────────────────────────────

// Profile — профиль zernio, группирующий подключённые аккаунты.
type Profile struct {
	ID    string `json:"_id"`
	AltID string `json:"id"`
	Name  string `json:"name"`
}

func (p Profile) identifier() string { return firstNonEmpty(p.ID, p.AltID) }

func (c *Client) ListProfiles(ctx context.Context) ([]Profile, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	var parsed struct {
		Profiles []Profile `json:"profiles"`
	}
	if err := c.do(ctx, http.MethodGet, "/profiles", nil, &parsed); err != nil {
		return nil, err
	}

	return parsed.Profiles, nil
}

// resolveProfileID отдаёт id профиля, под которым работает Pablo. Профиль
// ровно один, поэтому берём первый существующий, а если их нет — заводим сами.
// Результат кэшируется: он не меняется в течение жизни процесса.
func (c *Client) resolveProfileID(ctx context.Context) (string, error) {
	c.profileMu.Lock()
	defer c.profileMu.Unlock()

	if c.profileID != "" {
		return c.profileID, nil
	}

	profiles, err := c.ListProfiles(ctx)
	if err != nil {
		return "", err
	}

	for _, p := range profiles {
		if id := p.identifier(); id != "" {
			c.profileID = id
			log.Printf("zernio: using profile %q (%s)", p.Name, id)

			return id, nil
		}
	}

	id, err := c.CreateProfile(ctx, defaultProfileName, "Публикации Pablo")
	if err != nil {
		return "", err
	}

	c.profileID = id
	log.Printf("zernio: created profile %q (%s)", defaultProfileName, id)

	return id, nil
}

// CreateProfile создаёт профиль и возвращает его id.
func (c *Client) CreateProfile(ctx context.Context, name, description string) (string, error) {
	if !c.IsConfigured() {
		return "", ErrNotConfigured
	}

	body, _ := json.Marshal(map[string]any{
		"name":        name,
		"description": description,
	})

	var parsed struct {
		ID      string `json:"_id"`
		AltID   string `json:"id"`
		Profile struct {
			ID    string `json:"_id"`
			AltID string `json:"id"`
		} `json:"profile"`
	}
	if err := c.do(ctx, http.MethodPost, "/profiles", body, &parsed); err != nil {
		return "", err
	}

	id := firstNonEmpty(parsed.ID, parsed.AltID, parsed.Profile.ID, parsed.Profile.AltID)
	if id == "" {
		return "", errors.New("zernio.CreateProfile: response had no profile id")
	}

	return id, nil
}

// ─────────────────────────────────────────────────────────────
// Accounts
// ─────────────────────────────────────────────────────────────

// Account — подключённый аккаунт соцсети под профилем.
type Account struct {
	ID       string `json:"_id"`
	Platform string `json:"platform"`
	Username string `json:"username"`
	Status   string `json:"status"` // connected | disconnected (если приходит)
}

func (c *Client) ListAccounts(ctx context.Context) ([]Account, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	profileID, err := c.resolveProfileID(ctx)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Accounts []Account `json:"accounts"`
	}
	if err = c.do(ctx, http.MethodGet, "/accounts?profileId="+url.QueryEscape(profileID), nil, &parsed); err != nil {
		return nil, err
	}

	return parsed.Accounts, nil
}

// ResolveAccountID находит id подключённого аккаунта для площадки. Предпочитает
// аккаунт со статусом connected; если статус не сообщается — берёт первый
// подходящий. ErrNoAccount, если не подключён ни один.
func (c *Client) ResolveAccountID(ctx context.Context, platform string) (string, error) {
	accounts, err := c.ListAccounts(ctx)
	if err != nil {
		return "", err
	}

	platform = strings.ToLower(strings.TrimSpace(platform))

	var fallback string
	for _, a := range accounts {
		if strings.ToLower(a.Platform) != platform {
			continue
		}
		if a.Status == "" || strings.EqualFold(a.Status, "connected") {
			return a.ID, nil
		}
		if fallback == "" {
			fallback = a.ID
		}
	}

	if fallback != "" {
		return fallback, nil
	}

	return "", fmt.Errorf("%w: %s", ErrNoAccount, platform)
}

// ─────────────────────────────────────────────────────────────
// Connect (OAuth)
// ─────────────────────────────────────────────────────────────

// GetConnectURL запускает hosted-OAuth и возвращает authUrl, который открывает
// пользователь. После авторизации zernio редиректит на redirectURL с
// ?connected={platform}&profileId=…&accountId=…&username=…
func (c *Client) GetConnectURL(ctx context.Context, platform, redirectURL string) (string, error) {
	if !c.IsConfigured() {
		return "", ErrNotConfigured
	}

	platform = strings.ToLower(strings.TrimSpace(platform))
	if platform == "" {
		return "", errors.New("zernio.GetConnectURL: empty platform")
	}

	profileID, err := c.resolveProfileID(ctx)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Set("profileId", profileID)
	if strings.TrimSpace(redirectURL) != "" {
		q.Set("redirect_url", redirectURL)
	}

	var parsed struct {
		AuthURL    string `json:"authUrl"`
		URL        string `json:"url"`
		ConnectURL string `json:"connectUrl"`
	}
	if err = c.do(ctx, http.MethodGet, "/connect/"+url.PathEscape(platform)+"?"+q.Encode(), nil, &parsed); err != nil {
		return "", err
	}

	authURL := firstNonEmpty(parsed.AuthURL, parsed.URL, parsed.ConnectURL)
	if authURL == "" {
		return "", errors.New("zernio.GetConnectURL: response had no authUrl")
	}

	return authURL, nil
}

// ─────────────────────────────────────────────────────────────
// Posts
// ─────────────────────────────────────────────────────────────

// MediaItem — одно прикреплённое медиа (публичный URL, zernio скачает сам).
type MediaItem struct {
	Type string `json:"type"` // "image" | "video"
	URL  string `json:"url"`
}

// PostTarget — один аккаунт для публикации с опциями площадки.
type PostTarget struct {
	Platform             string         `json:"platform"`
	AccountID            string         `json:"accountId"`
	PlatformSpecificData map[string]any `json:"platformSpecificData,omitempty"`
}

type PostRequest struct {
	Content    string       `json:"content,omitempty"`
	MediaItems []MediaItem  `json:"mediaItems,omitempty"`
	Platforms  []PostTarget `json:"platforms"`
	PublishNow bool         `json:"publishNow,omitempty"`
	Timezone   string       `json:"timezone,omitempty"`
}

// PostPlatformResult — исход по конкретному аккаунту.
type PostPlatformResult struct {
	Platform        string `json:"platform"`
	Status          string `json:"status"` // published | failed | scheduled | publishing | ...
	PlatformPostURL string `json:"platformPostUrl"`
	Error           string `json:"error"`
}

type PostResult struct {
	ID        string               `json:"_id"`
	Status    string               `json:"status"` // published | partial | failed | scheduled
	Platforms []PostPlatformResult `json:"platforms"`
	RawBody   []byte               `json:"-"`
}

func (c *Client) Post(ctx context.Context, req PostRequest) (PostResult, error) {
	if !c.IsConfigured() {
		return PostResult{}, ErrNotConfigured
	}
	if len(req.Platforms) == 0 {
		return PostResult{}, errors.New("zernio.Post: no target platforms")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return PostResult{}, fmt.Errorf("zernio.Post: marshal: %w", err)
	}

	start := time.Now()
	log.Printf("zernio.Post start platforms=%d media=%d publish_now=%t", len(req.Platforms), len(req.MediaItems), req.PublishNow)

	raw, err := c.doRaw(ctx, http.MethodPost, "/posts", body)
	if err != nil {
		log.Printf("zernio.Post error duration=%s: %s", time.Since(start).Truncate(time.Millisecond), err.Error())
		return PostResult{}, err
	}

	// Ответ приходит либо как {post:{...}}, либо самим объектом поста.
	var wrap struct {
		Post *PostResult `json:"post"`
	}
	if jerr := json.Unmarshal(raw, &wrap); jerr == nil && wrap.Post != nil {
		wrap.Post.RawBody = raw
		log.Printf("zernio.Post finished id=%s status=%s platforms=%d duration=%s",
			wrap.Post.ID, wrap.Post.Status, len(wrap.Post.Platforms), time.Since(start).Truncate(time.Millisecond))
		return *wrap.Post, nil
	}

	var flat PostResult
	if jerr := json.Unmarshal(raw, &flat); jerr != nil {
		return PostResult{RawBody: raw}, fmt.Errorf("zernio.Post: decode: %w", jerr)
	}
	flat.RawBody = raw

	log.Printf("zernio.Post finished id=%s status=%s platforms=%d duration=%s",
		flat.ID, flat.Status, len(flat.Platforms), time.Since(start).Truncate(time.Millisecond))

	return flat, nil
}

// IsFailureStatus — только ЯВНЫЙ провал считается ошибкой. Публикация в
// Instagram асинхронная: успешный вызов возвращает publishing/processing, а не
// published. Всё остальное (published, publishing, processing, pending,
// scheduled) трактуем как принято.
func IsFailureStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "failed", "error", "rejected", "cancelled", "canceled":
		return true
	}
	return false
}

// IsPublishedStatus — площадка подтвердила публикацию.
func IsPublishedStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "published", "posted", "success", "completed":
		return true
	}
	return false
}

// ─────────────────────────────────────────────────────────────
// HTTP helpers
// ─────────────────────────────────────────────────────────────

func (c *Client) do(ctx context.Context, method, path string, body []byte, out any) error {
	raw, err := c.doRaw(ctx, method, path, body)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("zernio: decode %s %s: %w", method, path, err)
	}
	return nil
}

func (c *Client) doRaw(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("zernio: build %s %s: %w", method, path, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zernio: http %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("zernio: rate limited (429) on %s %s, retry-after=%s",
				method, path, resp.Header.Get("Retry-After"))
		}
		if resp.StatusCode == http.StatusConflict {
			return nil, fmt.Errorf("%w (%s %s): %s",
				ErrDuplicateContent, method, path, truncate(string(respBody), 300))
		}
		return nil, fmt.Errorf("zernio: %s %s HTTP %d: %s",
			method, path, resp.StatusCode, truncate(string(respBody), 500))
	}

	return respBody, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
