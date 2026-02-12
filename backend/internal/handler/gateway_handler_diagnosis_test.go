package handler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBuildNoAvailableAccountsMessage(t *testing.T) {
	t.Run("append_select_error_without_hint", func(t *testing.T) {
		h := &GatewayHandler{}
		msg := h.buildNoAvailableAccountsMessage(context.Background(), nil, service.PlatformAntigravity, errors.New("select failed"), "No available accounts")
		require.Equal(t, "No available accounts: select failed", msg)
	})

	t.Run("keep_base_message_when_error_nil", func(t *testing.T) {
		h := &GatewayHandler{}
		msg := h.buildNoAvailableAccountsMessage(context.Background(), nil, service.PlatformAntigravity, nil, "No available accounts")
		require.Equal(t, "No available accounts", msg)
	})

	t.Run("non_antigravity_skip_diagnosis", func(t *testing.T) {
		h := &GatewayHandler{}
		msg := h.buildNoAvailableAccountsMessage(context.Background(), nil, service.PlatformGemini, errors.New("no account"), "No available Gemini accounts")
		require.Equal(t, "No available Gemini accounts: no account", msg)
	})
}

func TestShouldLogNoAvailableDiagnosis(t *testing.T) {
	noAvailableDiagnosisLastLog = sync.Map{}
	key := "antigravity:1"

	require.True(t, shouldLogNoAvailableDiagnosis(key))
	require.False(t, shouldLogNoAvailableDiagnosis(key))

	noAvailableDiagnosisLastLog.Store(key, time.Now().Add(-noAvailableDiagnosisLogInterval-time.Second))
	require.True(t, shouldLogNoAvailableDiagnosis(key))
}
