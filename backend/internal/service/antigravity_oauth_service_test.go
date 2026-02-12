package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolveDefaultTierID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		loadRaw map[string]any
		want    string
	}{
		{
			name:    "nil loadRaw",
			loadRaw: nil,
			want:    "",
		},
		{
			name: "missing allowedTiers",
			loadRaw: map[string]any{
				"paidTier": map[string]any{"id": "g1-pro-tier"},
			},
			want: "",
		},
		{
			name:    "empty allowedTiers",
			loadRaw: map[string]any{"allowedTiers": []any{}},
			want:    "",
		},
		{
			name: "tier missing id field",
			loadRaw: map[string]any{
				"allowedTiers": []any{
					map[string]any{"isDefault": true},
				},
			},
			want: "",
		},
		{
			name: "allowedTiers but no default",
			loadRaw: map[string]any{
				"allowedTiers": []any{
					map[string]any{"id": "free-tier", "isDefault": false},
					map[string]any{"id": "standard-tier", "isDefault": false},
				},
			},
			want: "",
		},
		{
			name: "default tier found",
			loadRaw: map[string]any{
				"allowedTiers": []any{
					map[string]any{"id": "free-tier", "isDefault": true},
					map[string]any{"id": "standard-tier", "isDefault": false},
				},
			},
			want: "free-tier",
		},
		{
			name: "default tier id with spaces",
			loadRaw: map[string]any{
				"allowedTiers": []any{
					map[string]any{"id": "  standard-tier  ", "isDefault": true},
				},
			},
			want: "standard-tier",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resolveDefaultTierID(tc.loadRaw)
			if got != tc.want {
				t.Fatalf("resolveDefaultTierID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestIsNonRetryableAntigravityOAuthError(t *testing.T) {
	require.False(t, isNonRetryableAntigravityOAuthError(nil))

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "retryable", err: errors.New("network timeout"), want: false},
		{name: "invalid_grant", err: errors.New("invalid_grant"), want: true},
		{name: "uppercase_invalid_grant", err: errors.New("INVALID_GRANT: revoked"), want: true},
		{name: "mixed_case_access_denied", err: errors.New("Access_Denied"), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isNonRetryableAntigravityOAuthError(tt.err))
		})
	}
}

func TestWaitWithContext_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	startedAt := time.Now()
	err := waitWithContext(ctx, 3*time.Second)
	elapsed := time.Since(startedAt)

	require.ErrorIs(t, err, context.Canceled)
	require.Less(t, elapsed, 200*time.Millisecond)
}

func TestWaitWithContext_Timeout(t *testing.T) {
	startedAt := time.Now()
	err := waitWithContext(context.Background(), 15*time.Millisecond)
	elapsed := time.Since(startedAt)

	require.NoError(t, err)
	require.GreaterOrEqual(t, elapsed, 10*time.Millisecond)
}
