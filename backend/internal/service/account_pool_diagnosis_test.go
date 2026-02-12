package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type diagnosisAccountRepo struct {
	AccountRepository
	pages             map[int][]Account
	totalPages        int
	receivedPlatforms []string
}

func (r *diagnosisAccountRepo) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, accountType, status, search string) ([]Account, *pagination.PaginationResult, error) {
	r.receivedPlatforms = append(r.receivedPlatforms, platform)
	batch := r.pages[params.Page]
	pages := r.totalPages
	if pages <= 0 {
		pages = 1
	}
	if len(r.pages) > pages {
		pages = len(r.pages)
	}
	return batch, &pagination.PaginationResult{
		Total:    int64(len(batch)),
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

func TestDiagnoseAccountPool_IncludesErrorAccountsInGroup(t *testing.T) {
	repo := &diagnosisAccountRepo{
		pages: map[int][]Account{
			1: {
				{ID: 1, Platform: PlatformAntigravity, Status: StatusError, Schedulable: true, ErrorMessage: "Access forbidden (403): Verify your account", GroupIDs: []int64{4}},
				{ID: 2, Platform: PlatformAntigravity, Status: StatusError, Schedulable: true, ErrorMessage: "PERMISSION_DENIED", GroupIDs: []int64{4}},
				{ID: 3, Platform: PlatformAntigravity, Status: StatusActive, Schedulable: true, GroupIDs: []int64{9}},
			},
		},
		totalPages: 1,
	}
	svc := &GatewayService{accountRepo: repo}
	groupID := int64(4)

	d, err := svc.DiagnoseAccountPool(context.Background(), &groupID, PlatformAntigravity)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 2, d.Total)
	require.Equal(t, 2, d.Error)
	require.Equal(t, 2, d.AuthError)
	require.Equal(t, 0, d.Schedulable)
	require.Contains(t, d.Hint(), "blocked by upstream authentication/permission")
	require.Equal(t, []string{PlatformAntigravity}, repo.receivedPlatforms)
}

func TestDiagnoseAccountPool_PaginatesAllAccounts(t *testing.T) {
	repo := &diagnosisAccountRepo{
		pages: map[int][]Account{
			1: {
				{ID: 11, Platform: PlatformAntigravity, Status: StatusActive, Schedulable: true, GroupIDs: []int64{4}},
			},
			2: {
				{ID: 12, Platform: PlatformAntigravity, Status: StatusError, Schedulable: true, GroupIDs: []int64{4}},
			},
		},
		totalPages: 2,
	}
	svc := &GatewayService{accountRepo: repo}
	groupID := int64(4)

	d, err := svc.DiagnoseAccountPool(context.Background(), &groupID, PlatformAntigravity)
	require.NoError(t, err)
	require.NotNil(t, d)
	require.Equal(t, 2, d.Total)
	require.Equal(t, 1, d.Active)
	require.Equal(t, 1, d.Error)
	require.Equal(t, 1, d.Schedulable)
	require.Len(t, repo.receivedPlatforms, 2)
}

func TestAccountPoolDiagnosisHint(t *testing.T) {
	t.Run("has_schedulable_no_hint", func(t *testing.T) {
		d := &AccountPoolDiagnosis{Platform: PlatformAntigravity, Total: 2, Schedulable: 1}
		require.Equal(t, "", d.Hint())
	})

	t.Run("no_accounts_configured", func(t *testing.T) {
		d := &AccountPoolDiagnosis{Platform: PlatformAntigravity, Total: 0, Schedulable: 0}
		require.Contains(t, d.Hint(), "no antigravity accounts configured")
	})

	t.Run("all_auth_blocked", func(t *testing.T) {
		d := &AccountPoolDiagnosis{Platform: PlatformAntigravity, Total: 2, Schedulable: 0, Error: 2, AuthError: 2}
		require.Contains(t, d.Hint(), "blocked by upstream authentication/permission")
	})

	t.Run("all_rate_limited", func(t *testing.T) {
		d := &AccountPoolDiagnosis{Platform: PlatformAntigravity, Total: 3, Schedulable: 0, RateLimited: 3}
		require.Contains(t, d.Hint(), "rate-limited")
	})

	t.Run("all_manual_unschedulable", func(t *testing.T) {
		d := &AccountPoolDiagnosis{Platform: PlatformAntigravity, Total: 2, Schedulable: 0, ManualUnschedulable: 2}
		require.Contains(t, d.Hint(), "manually unschedulable")
	})
}

func TestIsAuthRelatedAccountError(t *testing.T) {
	require.True(t, isAuthRelatedAccountError("Access forbidden (403): Verify your account to continue"))
	require.True(t, isAuthRelatedAccountError("PERMISSION_DENIED"))
	require.False(t, isAuthRelatedAccountError("upstream timeout"))
}

func TestAccountBelongsToGroup(t *testing.T) {
	acc := Account{GroupIDs: []int64{1, 4}}
	require.True(t, accountBelongsToGroup(acc, 4))
	require.False(t, accountBelongsToGroup(acc, 8))

	acc = Account{AccountGroups: []AccountGroup{{GroupID: 9}}}
	require.True(t, accountBelongsToGroup(acc, 9))
	require.False(t, accountBelongsToGroup(acc, 0))
}
