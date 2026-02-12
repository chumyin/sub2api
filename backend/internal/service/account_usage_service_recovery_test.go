//go:build unit

package service

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/oauth"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type accountUsageFetcherStub struct {
	fetchCalls int
	responses  []*ClaudeUsageResponse
	errors     []error
}

func (s *accountUsageFetcherStub) next() (*ClaudeUsageResponse, error) {
	idx := s.fetchCalls
	s.fetchCalls++

	if idx < len(s.errors) && s.errors[idx] != nil {
		return nil, s.errors[idx]
	}
	if idx < len(s.responses) && s.responses[idx] != nil {
		return s.responses[idx], nil
	}

	return nil, fmt.Errorf("unexpected usage fetch call: %d", idx+1)
}

func (s *accountUsageFetcherStub) FetchUsage(ctx context.Context, accessToken, proxyURL string) (*ClaudeUsageResponse, error) {
	return s.next()
}

func (s *accountUsageFetcherStub) FetchUsageWithOptions(ctx context.Context, opts *ClaudeUsageFetchOptions) (*ClaudeUsageResponse, error) {
	return s.next()
}

type accountUsageLogRepoStub struct {
	UsageLogRepository
	windowStatsCalls int
}

func (s *accountUsageLogRepoStub) GetAccountWindowStats(ctx context.Context, accountID int64, startTime time.Time) (*usagestats.AccountStats, error) {
	s.windowStatsCalls++
	return &usagestats.AccountStats{}, nil
}

type accountUsageAccountRepoStub struct {
	mockAccountRepoForGemini
	clearErrorCalls int
}

func (s *accountUsageAccountRepoStub) ClearError(ctx context.Context, id int64) error {
	s.clearErrorCalls++
	return nil
}

type usageTokenInvalidatorStub struct {
	calls int
}

func (s *usageTokenInvalidatorStub) InvalidateToken(ctx context.Context, account *Account) error {
	s.calls++
	return nil
}

type claudeOAuthClientRefreshStub struct {
	ClaudeOAuthClient
	refreshCalls int
	refreshErr   error
}

func (s *claudeOAuthClientRefreshStub) RefreshToken(ctx context.Context, refreshToken, proxyURL string) (*oauth.TokenResponse, error) {
	s.refreshCalls++
	if s.refreshErr != nil {
		return nil, s.refreshErr
	}
	return &oauth.TokenResponse{
		AccessToken:  "new-access-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "new-refresh-token",
		Scope:        oauth.ScopeOAuth,
	}, nil
}

func buildOAuthUsageAccount(id int64) *Account {
	return &Account{
		ID:       id,
		Platform: PlatformAnthropic,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":  "old-access-token",
			"refresh_token": "old-refresh-token",
		},
	}
}

func TestAccountUsageService_GetUsage_OAuthAuthErrorAutoRecoverySuccess(t *testing.T) {
	account := buildOAuthUsageAccount(1001)
	accountRepo := &accountUsageAccountRepoStub{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	usageRepo := &accountUsageLogRepoStub{}

	resetAt := time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339)
	fetcher := &accountUsageFetcherStub{
		errors: []error{
			infraerrors.New(http.StatusUnauthorized, "UPSTREAM_UNAUTHORIZED", "upstream unauthorized"),
			nil,
		},
		responses: []*ClaudeUsageResponse{
			nil,
			{
				FiveHour: struct {
					Utilization float64 `json:"utilization"`
					ResetsAt    string  `json:"resets_at"`
				}{
					Utilization: 15,
					ResetsAt:    resetAt,
				},
			},
		},
	}
	cache := NewUsageCache()
	service := NewAccountUsageService(accountRepo, usageRepo, fetcher, nil, nil, cache, nil)

	oauthClient := &claudeOAuthClientRefreshStub{}
	oauthService := NewOAuthService(nil, oauthClient)
	service.SetOAuthRecoveryServices(oauthService, nil, nil, nil)

	rateLimitService := NewRateLimitService(accountRepo, nil, &config.Config{}, nil, nil)
	service.SetRateLimitService(rateLimitService)

	invalidator := &usageTokenInvalidatorStub{}
	service.SetTokenCacheInvalidator(invalidator)

	usage, err := service.GetUsage(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotNil(t, usage)
	require.NotNil(t, usage.FiveHour)
	require.Equal(t, 2, fetcher.fetchCalls, "should retry once after auth recovery")
	require.Equal(t, 1, oauthClient.refreshCalls, "should trigger token refresh once")
	require.Equal(t, 1, accountRepo.clearErrorCalls, "should clear account error after recovery")
	require.Equal(t, 1, invalidator.calls, "should invalidate token cache after recovery")
	require.Equal(t, 1, usageRepo.windowStatsCalls, "should continue loading window stats")
	require.Equal(t, "new-access-token", account.GetCredential("access_token"), "auto recovery should persist new access token")
	require.Equal(t, "new-refresh-token", account.GetCredential("refresh_token"), "auto recovery should persist rotated refresh token")
}

func TestAccountUsageService_GetUsage_OAuthAuthErrorAutoRecoveryFailed(t *testing.T) {
	account := buildOAuthUsageAccount(1002)
	accountRepo := &accountUsageAccountRepoStub{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	usageRepo := &accountUsageLogRepoStub{}
	fetcher := &accountUsageFetcherStub{
		errors: []error{
			infraerrors.New(http.StatusUnauthorized, "UPSTREAM_UNAUTHORIZED", "upstream unauthorized"),
		},
	}
	service := NewAccountUsageService(accountRepo, usageRepo, fetcher, nil, nil, NewUsageCache(), nil)

	oauthClient := &claudeOAuthClientRefreshStub{refreshErr: fmt.Errorf("refresh failed")}
	oauthService := NewOAuthService(nil, oauthClient)
	service.SetOAuthRecoveryServices(oauthService, nil, nil, nil)
	service.SetRateLimitService(NewRateLimitService(accountRepo, nil, &config.Config{}, nil, nil))
	invalidator := &usageTokenInvalidatorStub{}
	service.SetTokenCacheInvalidator(invalidator)

	usage, err := service.GetUsage(context.Background(), account.ID)
	require.Nil(t, usage)
	require.Error(t, err)
	require.Equal(t, http.StatusBadGateway, infraerrors.Code(err))
	appErr := infraerrors.FromError(err)
	require.Equal(t, "ACCOUNT_USAGE_AUTH_FAILED", appErr.Reason)
	require.Equal(t, "refresh_token_or_reset_status", appErr.Metadata["suggested_action"])
	require.Equal(t, 1, fetcher.fetchCalls)
	require.Equal(t, 1, oauthClient.refreshCalls)
	require.Equal(t, 0, invalidator.calls)
}

func TestAccountUsageService_GetUsage_OAuthNonAuthErrorNoAutoRecovery(t *testing.T) {
	account := buildOAuthUsageAccount(1003)
	accountRepo := &accountUsageAccountRepoStub{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	usageRepo := &accountUsageLogRepoStub{}
	fetcher := &accountUsageFetcherStub{
		errors: []error{
			infraerrors.New(http.StatusInternalServerError, "UPSTREAM_FAILURE", "upstream timeout"),
		},
	}
	service := NewAccountUsageService(accountRepo, usageRepo, fetcher, nil, nil, NewUsageCache(), nil)

	oauthClient := &claudeOAuthClientRefreshStub{}
	oauthService := NewOAuthService(nil, oauthClient)
	service.SetOAuthRecoveryServices(oauthService, nil, nil, nil)

	usage, err := service.GetUsage(context.Background(), account.ID)
	require.Nil(t, usage)
	require.Error(t, err)
	require.Equal(t, http.StatusBadGateway, infraerrors.Code(err))
	appErr := infraerrors.FromError(err)
	require.Equal(t, "ACCOUNT_USAGE_FETCH_FAILED", appErr.Reason)
	require.Equal(t, "not_applicable", appErr.Metadata["auto_recovery"])
	require.Equal(t, 1, fetcher.fetchCalls)
	require.Equal(t, 0, oauthClient.refreshCalls, "non-auth error should not trigger refresh")
}

func TestAccountUsageService_GetUsage_OAuthPermanentAuthErrorSkipAutoRecovery(t *testing.T) {
	account := buildOAuthUsageAccount(1004)
	accountRepo := &accountUsageAccountRepoStub{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	usageRepo := &accountUsageLogRepoStub{}
	fetcher := &accountUsageFetcherStub{
		errors: []error{
			infraerrors.New(http.StatusForbidden, "UPSTREAM_FORBIDDEN", "Gemini has been disabled in this account for violation of Terms of Service. status=PERMISSION_DENIED"),
		},
	}
	service := NewAccountUsageService(accountRepo, usageRepo, fetcher, nil, nil, NewUsageCache(), nil)

	oauthClient := &claudeOAuthClientRefreshStub{}
	oauthService := NewOAuthService(nil, oauthClient)
	service.SetOAuthRecoveryServices(oauthService, nil, nil, nil)

	usage, err := service.GetUsage(context.Background(), account.ID)
	require.Nil(t, usage)
	require.Error(t, err)
	require.Equal(t, http.StatusBadGateway, infraerrors.Code(err))

	appErr := infraerrors.FromError(err)
	require.Equal(t, "ACCOUNT_USAGE_AUTH_FAILED", appErr.Reason)
	require.Equal(t, "skipped_permanent_auth_error", appErr.Metadata["auto_recovery"])
	require.Equal(t, "verify_account_permission_or_replace_account", appErr.Metadata["suggested_action"])
	require.Equal(t, 1, fetcher.fetchCalls)
	require.Equal(t, 0, oauthClient.refreshCalls, "permanent auth error should not trigger token refresh")
}
