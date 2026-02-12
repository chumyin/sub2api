package handler

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestShouldApplyFailoverDelay(t *testing.T) {
	t.Run("non_antigravity_account_skip_delay", func(t *testing.T) {
		account := &service.Account{Platform: service.PlatformOpenAI}
		ok := shouldApplyFailoverDelay(account, &service.UpstreamFailoverError{StatusCode: http.StatusServiceUnavailable})
		require.False(t, ok)
	})

	t.Run("antigravity_503_apply_delay", func(t *testing.T) {
		account := &service.Account{Platform: service.PlatformAntigravity}
		ok := shouldApplyFailoverDelay(account, &service.UpstreamFailoverError{StatusCode: http.StatusServiceUnavailable})
		require.True(t, ok)
	})

	t.Run("antigravity_401_skip_delay", func(t *testing.T) {
		account := &service.Account{Platform: service.PlatformAntigravity}
		ok := shouldApplyFailoverDelay(account, &service.UpstreamFailoverError{StatusCode: http.StatusUnauthorized})
		require.False(t, ok)
	})

	t.Run("antigravity_403_skip_delay", func(t *testing.T) {
		account := &service.Account{Platform: service.PlatformAntigravity}
		ok := shouldApplyFailoverDelay(account, &service.UpstreamFailoverError{StatusCode: http.StatusForbidden})
		require.False(t, ok)
	})
}
