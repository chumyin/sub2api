//go:build unit

package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type antigravityTokenCacheStub struct {
	mu           sync.Mutex
	tokens       map[string]string
	setTTLs      map[string]time.Duration
	getErr       error
	setErr       error
	lockAcquired bool
	lockErr      error
	releaseErr   error
	getCalled    int32
	setCalled    int32
	lockCalled   int32
	unlockCalled int32
}

func newAntigravityTokenCacheStub() *antigravityTokenCacheStub {
	return &antigravityTokenCacheStub{
		tokens:       make(map[string]string),
		setTTLs:      make(map[string]time.Duration),
		lockAcquired: true,
	}
}

func (s *antigravityTokenCacheStub) GetAccessToken(ctx context.Context, cacheKey string) (string, error) {
	atomic.AddInt32(&s.getCalled, 1)
	if s.getErr != nil {
		return "", s.getErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tokens[cacheKey], nil
}

func (s *antigravityTokenCacheStub) SetAccessToken(ctx context.Context, cacheKey string, token string, ttl time.Duration) error {
	atomic.AddInt32(&s.setCalled, 1)
	if s.setErr != nil {
		return s.setErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[cacheKey] = token
	s.setTTLs[cacheKey] = ttl
	return nil
}

func (s *antigravityTokenCacheStub) DeleteAccessToken(ctx context.Context, cacheKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, cacheKey)
	delete(s.setTTLs, cacheKey)
	return nil
}

func (s *antigravityTokenCacheStub) AcquireRefreshLock(ctx context.Context, cacheKey string, ttl time.Duration) (bool, error) {
	atomic.AddInt32(&s.lockCalled, 1)
	if s.lockErr != nil {
		return false, s.lockErr
	}
	return s.lockAcquired, nil
}

func (s *antigravityTokenCacheStub) ReleaseRefreshLock(ctx context.Context, cacheKey string) error {
	atomic.AddInt32(&s.unlockCalled, 1)
	return s.releaseErr
}

type antigravityAccountRepoStub struct {
	AccountRepository
	account      *Account
	getErr       error
	updateErr    error
	getCalled    int32
	updateCalled int32
}

func (r *antigravityAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	atomic.AddInt32(&r.getCalled, 1)
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.account, nil
}

func (r *antigravityAccountRepoStub) Update(ctx context.Context, account *Account) error {
	atomic.AddInt32(&r.updateCalled, 1)
	if r.updateErr != nil {
		return r.updateErr
	}
	r.account = account
	return nil
}

func TestAntigravityTokenProvider_GetAccessToken_Upstream(t *testing.T) {
	provider := &AntigravityTokenProvider{}

	t.Run("upstream account with valid api_key", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAntigravity,
			Type:     AccountTypeUpstream,
			Credentials: map[string]any{
				"api_key": "sk-test-key-12345",
			},
		}
		token, err := provider.GetAccessToken(context.Background(), account)
		require.NoError(t, err)
		require.Equal(t, "sk-test-key-12345", token)
	})

	t.Run("upstream account missing api_key", func(t *testing.T) {
		account := &Account{
			Platform:    PlatformAntigravity,
			Type:        AccountTypeUpstream,
			Credentials: map[string]any{},
		}
		token, err := provider.GetAccessToken(context.Background(), account)
		require.Error(t, err)
		require.Contains(t, err.Error(), "upstream account missing api_key")
		require.Empty(t, token)
	})

	t.Run("upstream account with empty api_key", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAntigravity,
			Type:     AccountTypeUpstream,
			Credentials: map[string]any{
				"api_key": "",
			},
		}
		token, err := provider.GetAccessToken(context.Background(), account)
		require.Error(t, err)
		require.Contains(t, err.Error(), "upstream account missing api_key")
		require.Empty(t, token)
	})

	t.Run("upstream account with nil credentials", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAntigravity,
			Type:     AccountTypeUpstream,
		}
		token, err := provider.GetAccessToken(context.Background(), account)
		require.Error(t, err)
		require.Contains(t, err.Error(), "upstream account missing api_key")
		require.Empty(t, token)
	})
}

func TestAntigravityTokenProvider_GetAccessToken_Guards(t *testing.T) {
	provider := &AntigravityTokenProvider{}

	t.Run("nil account", func(t *testing.T) {
		token, err := provider.GetAccessToken(context.Background(), nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "account is nil")
		require.Empty(t, token)
	})

	t.Run("non-antigravity platform", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAnthropic,
			Type:     AccountTypeOAuth,
		}
		token, err := provider.GetAccessToken(context.Background(), account)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an antigravity account")
		require.Empty(t, token)
	})

	t.Run("unsupported account type", func(t *testing.T) {
		account := &Account{
			Platform: PlatformAntigravity,
			Type:     AccountTypeAPIKey,
		}
		token, err := provider.GetAccessToken(context.Background(), account)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not an antigravity oauth account")
		require.Empty(t, token)
	})
}

func TestAntigravityTokenProvider_LockRace_UsesFreshCachedToken(t *testing.T) {
	cache := newAntigravityTokenCacheStub()
	cache.lockAcquired = false // 模拟锁被其他 worker 持有

	account := &Account{
		ID:       201,
		Platform: PlatformAntigravity,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "stale-token",
			"expires_at":   time.Now().Add(1 * time.Minute).Format(time.RFC3339),
		},
	}

	cacheKey := AntigravityTokenCacheKey(account)
	go func() {
		time.Sleep(50 * time.Millisecond)
		cache.mu.Lock()
		cache.tokens[cacheKey] = "winner-token"
		cache.mu.Unlock()
	}()

	provider := NewAntigravityTokenProvider(nil, cache, nil)
	token, err := provider.GetAccessToken(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, "winner-token", token)
	require.Equal(t, int32(0), atomic.LoadInt32(&cache.setCalled), "命中并发刷新缓存后不应再写回旧 token")
}

func TestAntigravityTokenProvider_LockHeldMiss_UsesFreshDBToken(t *testing.T) {
	cache := newAntigravityTokenCacheStub()
	cache.lockAcquired = false // 模拟锁被其他 worker 持有，但缓存尚未写入

	account := &Account{
		ID:       202,
		Platform: PlatformAntigravity,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "stale-token",
			"expires_at":   time.Now().Add(1 * time.Minute).Format(time.RFC3339),
		},
	}

	freshFromDB := &Account{
		ID:       202,
		Platform: PlatformAntigravity,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "fresh-db-token",
			"expires_at":   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		},
	}
	repo := &antigravityAccountRepoStub{account: freshFromDB}

	provider := NewAntigravityTokenProvider(repo, cache, nil)
	token, err := provider.GetAccessToken(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, "fresh-db-token", token)
	require.GreaterOrEqual(t, atomic.LoadInt32(&repo.getCalled), int32(1), "锁竞争后应回读 DB 最新凭证")

	cacheKey := AntigravityTokenCacheKey(account)
	require.Equal(t, "fresh-db-token", cache.tokens[cacheKey])
}

func TestAntigravityTokenProvider_LockError_UsesShortTTLOnRefreshFailure(t *testing.T) {
	cache := newAntigravityTokenCacheStub()
	cache.lockErr = errors.New("redis unavailable")

	account := &Account{
		ID:       203,
		Platform: PlatformAntigravity,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "token-after-lock-error",
			// 不提供 expires_at，确保进入 needsRefresh 流程
		},
	}

	provider := NewAntigravityTokenProvider(nil, cache, nil)
	token, err := provider.GetAccessToken(context.Background(), account)

	require.NoError(t, err)
	require.Equal(t, "token-after-lock-error", token)

	cacheKey := AntigravityTokenCacheKey(account)
	require.Equal(t, time.Minute, cache.setTTLs[cacheKey], "刷新失败降级场景应使用短 TTL 避免 401 抖动")
}
