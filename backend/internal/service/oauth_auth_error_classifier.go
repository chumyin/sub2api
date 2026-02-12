package service

import "strings"

func isPermanentOAuthAuthErrorMessage(message string) bool {
	text := strings.ToLower(strings.TrimSpace(message))
	if text == "" {
		return false
	}

	keywords := []string{
		"verify your account to continue",
		"disabled in this account for violation of terms",
		"terms of service",
		"permission_denied",
		"permission denied",
		"access forbidden",
		"account suspended",
		"organization has been disabled",
		"invalid_grant",
		"invalid_client",
		"unauthorized_client",
		"access_denied",
		"token revoked",
	}

	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	return false
}
