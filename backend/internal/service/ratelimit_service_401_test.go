//go:build unit

package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type rateLimitAccountRepoStub struct {
	mockAccountRepoForGemini
	setErrorCalls    int
	tempCalls        int
	updateExtraCalls int
	lastErrorMsg     string
	lastTempReason   string
	lastTempUntil    *time.Time
	extraByID        map[int64]map[string]any
	setErrorIDs      []int64
	tempUnschedIDs   []int64
	getByIDFallback  map[int64]*Account
}

func (r *rateLimitAccountRepoStub) SetError(ctx context.Context, id int64, errorMsg string) error {
	r.setErrorCalls++
	r.lastErrorMsg = errorMsg
	r.setErrorIDs = append(r.setErrorIDs, id)
	return nil
}

func (r *rateLimitAccountRepoStub) SetTempUnschedulable(ctx context.Context, id int64, until time.Time, reason string) error {
	r.tempCalls++
	r.tempUnschedIDs = append(r.tempUnschedIDs, id)
	r.lastTempReason = reason
	r.lastTempUntil = &until
	if r.getByIDFallback != nil {
		if account, ok := r.getByIDFallback[id]; ok {
			account.TempUnschedulableUntil = &until
			account.TempUnschedulableReason = reason
		}
	}
	return nil
}

func (r *rateLimitAccountRepoStub) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	r.updateExtraCalls++
	if r.extraByID == nil {
		r.extraByID = make(map[int64]map[string]any)
	}
	if _, ok := r.extraByID[id]; !ok {
		r.extraByID[id] = make(map[string]any)
	}
	for key, value := range updates {
		r.extraByID[id][key] = value
	}
	if r.getByIDFallback != nil {
		if account, ok := r.getByIDFallback[id]; ok {
			if account.Extra == nil {
				account.Extra = make(map[string]any)
			}
			for key, value := range updates {
				account.Extra[key] = value
			}
		}
	}
	return nil
}

func (r *rateLimitAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	if r.getByIDFallback != nil {
		if account, ok := r.getByIDFallback[id]; ok {
			return account, nil
		}
	}
	if r.extraByID != nil {
		if extra, ok := r.extraByID[id]; ok {
			copiedExtra := make(map[string]any, len(extra))
			for key, value := range extra {
				copiedExtra[key] = value
			}
			return &Account{ID: id, Extra: copiedExtra}, nil
		}
	}
	return nil, errors.New("account not found")
}

type tokenCacheInvalidatorRecorder struct {
	accounts []*Account
	err      error
}

func (r *tokenCacheInvalidatorRecorder) InvalidateToken(ctx context.Context, account *Account) error {
	r.accounts = append(r.accounts, account)
	return r.err
}

func TestRateLimitService_HandleUpstreamError_OAuth401MarksError(t *testing.T) {
	tests := []struct {
		name     string
		platform string
	}{
		{name: "gemini", platform: PlatformGemini},
		{name: "antigravity", platform: PlatformAntigravity},
		{name: "openai", platform: PlatformOpenAI},
		{name: "anthropic", platform: PlatformAnthropic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &rateLimitAccountRepoStub{}
			invalidator := &tokenCacheInvalidatorRecorder{}
			service := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			service.SetTokenCacheInvalidator(invalidator)
			account := &Account{
				ID:       100,
				Platform: tt.platform,
				Type:     AccountTypeOAuth,
				Credentials: map[string]any{
					"temp_unschedulable_enabled": true,
					"temp_unschedulable_rules": []any{
						map[string]any{
							"error_code":       401,
							"keywords":         []any{"unauthorized"},
							"duration_minutes": 30,
							"description":      "custom rule",
						},
					},
				},
			}

			shouldDisable := service.HandleUpstreamError(context.Background(), account, 401, http.Header{}, []byte("unauthorized"))

			require.False(t, shouldDisable)
			require.Equal(t, 0, repo.setErrorCalls)
			require.Equal(t, 1, repo.tempCalls)
			require.Equal(t, 1, repo.updateExtraCalls)
			require.Contains(t, repo.lastTempReason, oauth401PlatformLabel(tt.platform)+" OAuth 401 temporary cooldown")
			if extra := repo.extraByID[account.ID]; extra != nil {
				countKey, _ := oauth401CounterKeys(tt.platform)
				require.Equal(t, 1, parseExtraInt(extra[countKey]))
			}
			require.Len(t, invalidator.accounts, 1)
		})
	}
}

func TestRateLimitService_HandleUpstreamError_OAuth401InvalidatorError(t *testing.T) {
	repo := &rateLimitAccountRepoStub{}
	invalidator := &tokenCacheInvalidatorRecorder{err: errors.New("boom")}
	service := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	service.SetTokenCacheInvalidator(invalidator)
	account := &Account{
		ID:       101,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}

	shouldDisable := service.HandleUpstreamError(context.Background(), account, 401, http.Header{}, []byte("unauthorized"))

	require.False(t, shouldDisable)
	require.Equal(t, 0, repo.setErrorCalls)
	require.Equal(t, 1, repo.tempCalls)
	require.Equal(t, 1, repo.updateExtraCalls)
	require.Len(t, invalidator.accounts, 1)
}

func TestRateLimitService_HandleUpstreamError_NonOAuth401(t *testing.T) {
	repo := &rateLimitAccountRepoStub{}
	invalidator := &tokenCacheInvalidatorRecorder{}
	service := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	service.SetTokenCacheInvalidator(invalidator)
	account := &Account{
		ID:       102,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
	}

	shouldDisable := service.HandleUpstreamError(context.Background(), account, 401, http.Header{}, []byte("unauthorized"))

	require.True(t, shouldDisable)
	require.Equal(t, 1, repo.setErrorCalls)
	require.Empty(t, invalidator.accounts)
}

func TestRateLimitService_HandleUpstreamError_AntigravityOAuth401ThresholdMarksError(t *testing.T) {
	repo := &rateLimitAccountRepoStub{extraByID: make(map[int64]map[string]any)}
	invalidator := &tokenCacheInvalidatorRecorder{}
	service := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	service.SetTokenCacheInvalidator(invalidator)
	countKey, tsKey := oauth401CounterKeys(PlatformAntigravity)
	account := &Account{
		ID:       103,
		Platform: PlatformAntigravity,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "token",
		},
		Extra: map[string]any{
			countKey: oauth401ErrorThreshold - 1,
			tsKey:    time.Now().Add(-1 * time.Minute).Unix(),
		},
	}
	repo.getByIDFallback = map[int64]*Account{account.ID: account}

	shouldDisable := service.HandleUpstreamError(context.Background(), account, 401, http.Header{}, []byte("unauthorized"))

	require.True(t, shouldDisable)
	require.Equal(t, 1, repo.tempCalls)
	require.Equal(t, 1, repo.setErrorCalls)
	require.Equal(t, 1, repo.updateExtraCalls)
	require.Len(t, invalidator.accounts, 1)
	require.Contains(t, repo.lastErrorMsg, "after retries")
	if extra := repo.extraByID[account.ID]; extra != nil {
		require.Equal(t, oauth401ErrorThreshold, parseExtraInt(extra[countKey]))
	}
}

func TestRateLimitService_HandleUpstreamError_OAuth401CounterWindowReset(t *testing.T) {
	repo := &rateLimitAccountRepoStub{extraByID: make(map[int64]map[string]any)}
	invalidator := &tokenCacheInvalidatorRecorder{}
	service := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	service.SetTokenCacheInvalidator(invalidator)
	countKey, tsKey := oauth401CounterKeys(PlatformGemini)
	account := &Account{
		ID:       104,
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "token",
		},
		Extra: map[string]any{
			countKey: oauth401ErrorThreshold,
			tsKey:    time.Now().Add(-2 * oauth401Window).Unix(),
		},
	}
	repo.getByIDFallback = map[int64]*Account{account.ID: account}

	shouldDisable := service.HandleUpstreamError(context.Background(), account, 401, http.Header{}, []byte("unauthorized"))

	require.False(t, shouldDisable)
	require.Equal(t, 1, repo.tempCalls)
	require.Equal(t, 0, repo.setErrorCalls)
	require.Equal(t, 1, repo.updateExtraCalls)
	require.Len(t, invalidator.accounts, 1)
	if extra := repo.extraByID[account.ID]; extra != nil {
		require.Equal(t, 1, parseExtraInt(extra[countKey]))
	}
}

func TestRateLimitService_ResetOAuth401State(t *testing.T) {
	repo := &rateLimitAccountRepoStub{extraByID: make(map[int64]map[string]any)}
	service := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	countKey, tsKey := oauth401CounterKeys(PlatformOpenAI)
	account := &Account{
		ID:       105,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			countKey:                   2,
			tsKey:                      time.Now().Unix(),
			"openai_oauth_401_count":   2,
			"openai_oauth_401_last_ts": time.Now().Unix(),
		},
	}

	err := service.ResetOAuth401State(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, 1, repo.updateExtraCalls)
	require.Equal(t, 0, parseExtraInt(account.Extra[countKey]))
	require.Equal(t, 0, parseExtraInt(account.Extra[tsKey]))
	require.Equal(t, 0, parseExtraInt(account.Extra["openai_oauth_401_count"]))
	require.Equal(t, 0, parseExtraInt(account.Extra["openai_oauth_401_last_ts"]))
}
