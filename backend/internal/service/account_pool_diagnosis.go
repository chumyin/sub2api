package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type AccountPoolDiagnosis struct {
	Platform            string
	Total               int
	Active              int
	Schedulable         int
	Error               int
	AuthError           int
	RateLimited         int
	TempUnschedulable   int
	Overloaded          int
	ManualUnschedulable int
}

func (s *GatewayService) DiagnoseAccountPool(ctx context.Context, groupID *int64, platform string) (*AccountPoolDiagnosis, error) {
	if s == nil || s.accountRepo == nil {
		return nil, nil
	}

	accounts, err := s.listAccountsForDiagnosis(ctx, groupID, platform)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	diag := &AccountPoolDiagnosis{Platform: platform}
	for _, account := range accounts {
		if platform != "" && account.Platform != platform {
			continue
		}
		diag.Total++
		if account.Status == StatusActive {
			diag.Active++
		}
		if account.Status == StatusError {
			diag.Error++
			if isAuthRelatedAccountError(account.ErrorMessage) {
				diag.AuthError++
			}
		}
		if !account.Schedulable {
			diag.ManualUnschedulable++
		}
		if account.RateLimitResetAt != nil && now.Before(*account.RateLimitResetAt) {
			diag.RateLimited++
		}
		if account.TempUnschedulableUntil != nil && now.Before(*account.TempUnschedulableUntil) {
			diag.TempUnschedulable++
		}
		if account.OverloadUntil != nil && now.Before(*account.OverloadUntil) {
			diag.Overloaded++
		}
		if account.IsSchedulable() {
			diag.Schedulable++
		}
	}

	return diag, nil
}

func (s *GatewayService) listAccountsForDiagnosis(ctx context.Context, groupID *int64, platform string) ([]Account, error) {
	platform = strings.TrimSpace(platform)
	params := pagination.PaginationParams{Page: 1, PageSize: 100}
	accounts := make([]Account, 0, 32)

	for {
		batch, page, err := s.accountRepo.ListWithFilters(ctx, params, platform, "", "", "")
		if err != nil {
			return nil, err
		}
		for _, account := range batch {
			if groupID != nil && *groupID > 0 && !accountBelongsToGroup(account, *groupID) {
				continue
			}
			accounts = append(accounts, account)
		}
		if page == nil || params.Page >= page.Pages || len(batch) == 0 {
			break
		}
		params.Page++
	}

	return accounts, nil
}

func accountBelongsToGroup(account Account, groupID int64) bool {
	if groupID <= 0 {
		return false
	}
	for _, id := range account.GroupIDs {
		if id == groupID {
			return true
		}
	}
	for _, binding := range account.AccountGroups {
		if binding.GroupID == groupID {
			return true
		}
	}
	return false
}

func (d *AccountPoolDiagnosis) Hint() string {
	if d == nil || d.Schedulable > 0 {
		return ""
	}

	platformLabel := strings.TrimSpace(d.Platform)
	if platformLabel == "" {
		platformLabel = "current"
	}

	if d.Total == 0 {
		return fmt.Sprintf("no %s accounts configured", platformLabel)
	}
	if d.Error == d.Total && d.AuthError == d.Total {
		return fmt.Sprintf("all %s accounts are blocked by upstream authentication/permission; re-verify or replace accounts", platformLabel)
	}
	if d.Error == d.Total {
		return fmt.Sprintf("all %s accounts are in error status; check account errors in admin panel", platformLabel)
	}
	if d.RateLimited == d.Total {
		return fmt.Sprintf("all %s accounts are rate-limited; retry later", platformLabel)
	}
	if d.ManualUnschedulable == d.Total {
		return fmt.Sprintf("all %s accounts are manually unschedulable", platformLabel)
	}
	if d.Active == 0 {
		return fmt.Sprintf("all %s accounts are inactive", platformLabel)
	}
	return fmt.Sprintf("all %s accounts are temporarily unavailable", platformLabel)
}

func isAuthRelatedAccountError(message string) bool {
	text := strings.ToLower(strings.TrimSpace(message))
	if text == "" {
		return false
	}

	keywords := []string{
		"access forbidden",
		"permission denied",
		"permission_denied",
		"verify your account",
		"authentication failed",
		"invalid credentials",
		"invalid token",
		"invalid_grant",
		"token expired",
		"unauthorized",
		"terms of service",
		"organization has been disabled",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
