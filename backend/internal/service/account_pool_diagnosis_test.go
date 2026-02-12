package service

import "testing"

import "github.com/stretchr/testify/require"

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
