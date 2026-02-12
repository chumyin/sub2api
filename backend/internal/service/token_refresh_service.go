package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const tokenRefreshErrorRecoveryMaxPages = 20

// TokenRefreshService OAuth token自动刷新服务
// 定期检查并刷新即将过期的token
type TokenRefreshService struct {
	accountRepo      AccountRepository
	refreshers       []TokenRefresher
	cfg              *config.TokenRefreshConfig
	cacheInvalidator TokenCacheInvalidator
	schedulerCache   SchedulerCache // 用于同步更新调度器缓存，解决 token 刷新后缓存不一致问题

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewTokenRefreshService 创建token刷新服务
func NewTokenRefreshService(
	accountRepo AccountRepository,
	oauthService *OAuthService,
	openaiOAuthService *OpenAIOAuthService,
	geminiOAuthService *GeminiOAuthService,
	antigravityOAuthService *AntigravityOAuthService,
	cacheInvalidator TokenCacheInvalidator,
	schedulerCache SchedulerCache,
	cfg *config.Config,
) *TokenRefreshService {
	s := &TokenRefreshService{
		accountRepo:      accountRepo,
		cfg:              &cfg.TokenRefresh,
		cacheInvalidator: cacheInvalidator,
		schedulerCache:   schedulerCache,
		stopCh:           make(chan struct{}),
	}

	// 注册平台特定的刷新器
	s.refreshers = []TokenRefresher{
		NewClaudeTokenRefresher(oauthService),
		NewOpenAITokenRefresher(openaiOAuthService),
		NewGeminiTokenRefresher(geminiOAuthService),
		NewAntigravityTokenRefresher(antigravityOAuthService),
	}

	return s
}

// Start 启动后台刷新服务
func (s *TokenRefreshService) Start() {
	if !s.cfg.Enabled {
		log.Println("[TokenRefresh] Service disabled by configuration")
		return
	}

	s.wg.Add(1)
	go s.refreshLoop()

	log.Printf("[TokenRefresh] Service started (check every %d minutes, refresh %v hours before expiry)",
		s.cfg.CheckIntervalMinutes, s.cfg.RefreshBeforeExpiryHours)
}

// Stop 停止刷新服务
func (s *TokenRefreshService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	log.Println("[TokenRefresh] Service stopped")
}

// refreshLoop 刷新循环
func (s *TokenRefreshService) refreshLoop() {
	defer s.wg.Done()

	// 计算检查间隔
	checkInterval := time.Duration(s.cfg.CheckIntervalMinutes) * time.Minute
	if checkInterval < time.Minute {
		checkInterval = 5 * time.Minute
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// 启动时立即执行一次检查
	s.processRefresh()

	for {
		select {
		case <-ticker.C:
			s.processRefresh()
		case <-s.stopCh:
			return
		}
	}
}

// processRefresh 执行一次刷新检查
func (s *TokenRefreshService) processRefresh() {
	ctx := context.Background()

	// 计算刷新窗口
	refreshWindow := time.Duration(s.cfg.RefreshBeforeExpiryHours * float64(time.Hour))

	accounts, err := s.listRefreshCandidates(ctx)
	if err != nil {
		log.Printf("[TokenRefresh] Failed to list accounts: %v", err)
		return
	}

	totalAccounts := len(accounts)
	oauthAccounts := 0 // 可刷新的OAuth账号数
	needsRefresh := 0  // 需要刷新的账号数
	refreshed, failed := 0, 0

	for i := range accounts {
		account := &accounts[i]

		// 遍历所有刷新器，找到能处理此账号的
		for _, refresher := range s.refreshers {
			if !refresher.CanRefresh(account) {
				continue
			}

			oauthAccounts++
			forceRecoveryRefresh := isRecoverableOAuthErrorState(account)

			// 检查是否需要刷新
			if !forceRecoveryRefresh && !refresher.NeedsRefresh(account, refreshWindow) {
				break // 不需要刷新，跳过
			}

			needsRefresh++
			if forceRecoveryRefresh {
				log.Printf("[TokenRefresh] Account %d (%s) is in recoverable error state, forcing refresh", account.ID, account.Name)
			}

			// 执行刷新
			if err := s.refreshWithRetry(ctx, account, refresher); err != nil {
				log.Printf("[TokenRefresh] Account %d (%s) failed: %v", account.ID, account.Name, err)
				failed++
			} else {
				log.Printf("[TokenRefresh] Account %d (%s) refreshed successfully", account.ID, account.Name)
				refreshed++
			}

			// 每个账号只由一个refresher处理
			break
		}
	}

	// 始终打印周期日志，便于跟踪服务运行状态
	log.Printf("[TokenRefresh] Cycle complete: total=%d, oauth=%d, needs_refresh=%d, refreshed=%d, failed=%d",
		totalAccounts, oauthAccounts, needsRefresh, refreshed, failed)
}

func (s *TokenRefreshService) listRefreshCandidates(ctx context.Context) ([]Account, error) {
	activeAccounts, err := s.accountRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	errorAccounts, err := s.listRecoverableOAuthErrorAccounts(ctx)
	if err != nil {
		log.Printf("[TokenRefresh] Failed to list recoverable error accounts: %v", err)
		return activeAccounts, nil
	}
	if len(errorAccounts) == 0 {
		return activeAccounts, nil
	}

	merged := make([]Account, 0, len(activeAccounts)+len(errorAccounts))
	seen := make(map[int64]struct{}, len(activeAccounts)+len(errorAccounts))

	for i := range activeAccounts {
		account := activeAccounts[i]
		merged = append(merged, account)
		seen[account.ID] = struct{}{}
	}

	for i := range errorAccounts {
		account := errorAccounts[i]
		if _, exists := seen[account.ID]; exists {
			continue
		}
		merged = append(merged, account)
		seen[account.ID] = struct{}{}
	}

	return merged, nil
}

func (s *TokenRefreshService) listRecoverableOAuthErrorAccounts(ctx context.Context) ([]Account, error) {
	params := pagination.PaginationParams{Page: 1, PageSize: 100}
	accounts := make([]Account, 0, 8)

	for page := 0; page < tokenRefreshErrorRecoveryMaxPages; page++ {
		batch, pager, err := s.accountRepo.ListWithFilters(ctx, params, "", AccountTypeOAuth, StatusError, "")
		if err != nil {
			return nil, err
		}

		for i := range batch {
			if isRecoverableOAuthErrorState(&batch[i]) {
				accounts = append(accounts, batch[i])
			}
		}

		if pager == nil || params.Page >= pager.Pages || len(batch) == 0 {
			break
		}
		if page == tokenRefreshErrorRecoveryMaxPages-1 && pager.Pages > params.Page {
			log.Printf("[TokenRefresh] Reached recoverable error scan page limit: scanned=%d total_pages=%d", tokenRefreshErrorRecoveryMaxPages, pager.Pages)
		}
		params.Page++
	}

	return accounts, nil
}

func isRecoverableOAuthErrorState(account *Account) bool {
	if account == nil || account.Status != StatusError || account.Type != AccountTypeOAuth {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(account.ErrorMessage))
	if message == "" {
		return false
	}
	if isPermanentOAuthAuthErrorMessage(message) {
		return false
	}

	return strings.Contains(message, "missing_project_id:") ||
		strings.Contains(message, "authentication failed (401)") ||
		strings.Contains(message, "invalid or expired credentials") ||
		strings.Contains(message, "token refresh failed")
}

// refreshWithRetry 带重试的刷新
func (s *TokenRefreshService) refreshWithRetry(ctx context.Context, account *Account, refresher TokenRefresher) error {
	var lastErr error

	for attempt := 1; attempt <= s.cfg.MaxRetries; attempt++ {
		newCredentials, err := refresher.Refresh(ctx, account)

		// 如果有新凭证，先更新（即使有错误也要保存 token）
		if newCredentials != nil {
			// 记录刷新版本时间戳，用于解决缓存一致性问题
			// TokenProvider 写入缓存前会检查此版本，如果版本已更新则跳过写入
			newCredentials["_token_version"] = time.Now().UnixMilli()

			account.Credentials = newCredentials
			if saveErr := s.accountRepo.Update(ctx, account); saveErr != nil {
				return fmt.Errorf("failed to save credentials: %w", saveErr)
			}
		}

		if err == nil {
			if isRecoverableOAuthErrorState(account) {
				if clearErr := s.accountRepo.ClearError(ctx, account.ID); clearErr != nil {
					log.Printf("[TokenRefresh] Failed to clear error status for account %d: %v", account.ID, clearErr)
				} else {
					account.Status = StatusActive
					account.ErrorMessage = ""
					if err := s.accountRepo.ClearRateLimit(ctx, account.ID); err != nil {
						log.Printf("[TokenRefresh] Failed to clear rate limit for account %d: %v", account.ID, err)
					}
					if err := s.accountRepo.ClearTempUnschedulable(ctx, account.ID); err != nil {
						log.Printf("[TokenRefresh] Failed to clear temp unschedulable for account %d: %v", account.ID, err)
					}
					log.Printf("[TokenRefresh] Account %d: cleared recoverable OAuth error", account.ID)
				}
			}
			// 对所有 OAuth 账号调用缓存失效（InvalidateToken 内部根据平台判断是否需要处理）
			if s.cacheInvalidator != nil && account.Type == AccountTypeOAuth {
				if err := s.cacheInvalidator.InvalidateToken(ctx, account); err != nil {
					log.Printf("[TokenRefresh] Failed to invalidate token cache for account %d: %v", account.ID, err)
				} else {
					log.Printf("[TokenRefresh] Token cache invalidated for account %d", account.ID)
				}
			}
			// 同步更新调度器缓存，确保调度获取的 Account 对象包含最新的 credentials
			// 这解决了 token 刷新后调度器缓存数据不一致的问题（#445）
			if s.schedulerCache != nil {
				if err := s.schedulerCache.SetAccount(ctx, account); err != nil {
					log.Printf("[TokenRefresh] Failed to sync scheduler cache for account %d: %v", account.ID, err)
				} else {
					log.Printf("[TokenRefresh] Scheduler cache synced for account %d", account.ID)
				}
			}
			return nil
		}

		// Antigravity 账户：不可重试错误直接标记 error 状态并返回
		if account.Platform == PlatformAntigravity && isNonRetryableRefreshError(err) {
			errorMsg := fmt.Sprintf("Token refresh failed (non-retryable): %v", err)
			if setErr := s.accountRepo.SetError(ctx, account.ID, errorMsg); setErr != nil {
				log.Printf("[TokenRefresh] Failed to set error status for account %d: %v", account.ID, setErr)
			}
			return err
		}

		lastErr = err
		log.Printf("[TokenRefresh] Account %d attempt %d/%d failed: %v",
			account.ID, attempt, s.cfg.MaxRetries, err)

		// 如果还有重试机会，等待后重试
		if attempt < s.cfg.MaxRetries {
			// 指数退避：2^(attempt-1) * baseSeconds
			backoff := time.Duration(s.cfg.RetryBackoffSeconds) * time.Second * time.Duration(1<<(attempt-1))
			time.Sleep(backoff)
		}
	}

	// Antigravity 账户：其他错误仅记录日志，不标记 error（可能是临时网络问题）
	// 其他平台账户：重试失败后标记 error
	if account.Platform == PlatformAntigravity {
		log.Printf("[TokenRefresh] Account %d: refresh failed after %d retries: %v", account.ID, s.cfg.MaxRetries, lastErr)
	} else {
		errorMsg := fmt.Sprintf("Token refresh failed after %d retries: %v", s.cfg.MaxRetries, lastErr)
		if err := s.accountRepo.SetError(ctx, account.ID, errorMsg); err != nil {
			log.Printf("[TokenRefresh] Failed to set error status for account %d: %v", account.ID, err)
		}
	}

	return lastErr
}

// isNonRetryableRefreshError 判断是否为不可重试的刷新错误
// 这些错误通常表示凭证已失效或配置确实缺失，需要用户重新授权
// 注意：missing_project_id 错误只在真正缺失（从未获取过）时返回，临时获取失败不会返回此错误
func isNonRetryableRefreshError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	nonRetryable := []string{
		"invalid_grant",       // refresh_token 已失效
		"invalid_client",      // 客户端配置错误
		"unauthorized_client", // 客户端未授权
		"access_denied",       // 访问被拒绝
		"missing_project_id",  // 缺少 project_id
	}
	for _, needle := range nonRetryable {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}
