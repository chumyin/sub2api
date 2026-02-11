//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type accountTestOpenAIRepoStub struct {
	mockAccountRepoForGemini
	account *Account
}

func (r *accountTestOpenAIRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	if r.account != nil && r.account.ID == id {
		return r.account, nil
	}
	return nil, errors.New("account not found")
}

type accountTestOpenAIUpstreamStub struct {
	lastReq *http.Request
	body    []byte
	status  int
}

func (s *accountTestOpenAIUpstreamStub) Do(req *http.Request, proxyURL string, accountID int64, accountConcurrency int) (*http.Response, error) {
	return nil, errors.New("unexpected Do call")
}

func (s *accountTestOpenAIUpstreamStub) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, enableTLSFingerprint bool) (*http.Response, error) {
	s.lastReq = req
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		s.body = body
		req.Body = io.NopCloser(strings.NewReader(string(body)))
	}
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	stream := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"ok\"}\n\n" +
		"data: {\"type\":\"response.completed\"}\n\n"
	resp := &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(stream)),
	}
	return resp, nil
}

type accountTestOpenAITokenProviderStub struct {
	token string
	err   error
}

func (p *accountTestOpenAITokenProviderStub) GetAccessToken(ctx context.Context, account *Account) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	return p.token, nil
}

type accountTestTokenCacheStub struct {
	mu     sync.Mutex
	tokens map[string]string
}

func (s *accountTestTokenCacheStub) GetAccessToken(ctx context.Context, cacheKey string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokens == nil {
		return "", nil
	}
	return s.tokens[cacheKey], nil
}

func (s *accountTestTokenCacheStub) SetAccessToken(ctx context.Context, cacheKey string, token string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokens == nil {
		s.tokens = make(map[string]string)
	}
	s.tokens[cacheKey] = token
	return nil
}

func (s *accountTestTokenCacheStub) DeleteAccessToken(ctx context.Context, cacheKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tokens != nil {
		delete(s.tokens, cacheKey)
	}
	return nil
}

func (s *accountTestTokenCacheStub) AcquireRefreshLock(ctx context.Context, cacheKey string, ttl time.Duration) (bool, error) {
	return true, nil
}

func (s *accountTestTokenCacheStub) ReleaseRefreshLock(ctx context.Context, cacheKey string) error {
	return nil
}

func TestAccountTestService_OpenAIOAuthUsesTokenProviderAndGatewayHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	account := &Account{
		ID:       2001,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "stale-db-token",
			"chatgpt_account_id": "acct_123",
			"user_agent":         "custom-test-ua",
		},
	}

	repo := &accountTestOpenAIRepoStub{account: account}
	upstream := &accountTestOpenAIUpstreamStub{}
	tokenCache := &accountTestTokenCacheStub{tokens: map[string]string{}}
	cacheKey := OpenAITokenCacheKey(account)
	tokenCache.tokens[cacheKey] = "provider-token"
	provider := NewOpenAITokenProvider(repo, tokenCache, nil)
	svc := NewAccountTestService(repo, nil, provider, nil, upstream, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/2001/test", strings.NewReader(`{"model_id":"gpt-5"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	err := svc.TestAccountConnection(c, account.ID, "gpt-5")
	require.NoError(t, err)
	require.NotNil(t, upstream.lastReq)

	require.Equal(t, "chatgpt.com", upstream.lastReq.Host)
	require.Equal(t, "Bearer provider-token", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "responses=experimental", upstream.lastReq.Header.Get("OpenAI-Beta"))
	require.Equal(t, "opencode", upstream.lastReq.Header.Get("originator"))
	require.Equal(t, "text/event-stream", upstream.lastReq.Header.Get("accept"))
	require.Equal(t, "acct_123", upstream.lastReq.Header.Get("chatgpt-account-id"))
	require.Equal(t, "custom-test-ua", upstream.lastReq.Header.Get("user-agent"))
	require.NotEmpty(t, upstream.lastReq.Header.Get("conversation_id"))
	require.NotEmpty(t, upstream.lastReq.Header.Get("session_id"))

	var payload map[string]any
	require.NoError(t, json.Unmarshal(upstream.body, &payload))
	require.Equal(t, "gpt-5", payload["model"])
	require.Equal(t, true, payload["stream"])
	require.Equal(t, false, payload["store"])
	require.NotEmpty(t, payload["instructions"])
}
