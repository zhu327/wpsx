package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAuthorizationURL(t *testing.T) {
	appID := "test_app_id"
	callbackURL := "http://localhost"
	state := "abc123"

	authURL := buildAuthorizationURL(appID, callbackURL, state)

	assert.Contains(t, authURL, "https://openapi.wps.cn/oauth2/auth")
	assert.Contains(t, authURL, "client_id=test_app_id")
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "state=abc123")
	assert.Contains(t, authURL, "kso.mcp_yundoc.readwrite")
	assert.Contains(t, authURL, "redirect_uri=")
}

func TestBuildAuthorizationURL_DefaultScope(t *testing.T) {
	authURL := buildAuthorizationURL("id", "http://localhost", "st")

	scopes := []string{
		"kso.mcp_airpage.readwrite",
		"kso.mcp_calendar.readwrite",
		"kso.mcp_dbsheet.readwrite",
		"kso.mcp_mail.readwrite",
		"kso.mcp_meeting.readwrite",
		"kso.mcp_message.readwrite",
		"kso.mcp_scenario.readwrite",
		"kso.mcp_todo.readwrite",
		"kso.mcp_yundoc.readwrite",
	}
	for _, scope := range scopes {
		assert.Contains(t, authURL, scope, "missing scope: %s", scope)
	}
}

func TestGenerateState(t *testing.T) {
	s1, err1 := generateState()
	s2, err2 := generateState()

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Len(t, s1, 32) // 16 bytes hex = 32 chars
	assert.NotEqual(t, s1, s2, "states should be random")
}

func TestParseCallbackURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  string
		wantState string
		wantErr   bool
	}{
		{
			name:      "valid callback URL",
			input:     "http://localhost?code=mycode123&state=mystate",
			wantCode:  "mycode123",
			wantState: "mystate",
		},
		{
			name:    "missing code",
			input:   "http://localhost?state=mystate",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, state, err := parseCallbackURL(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tc.wantCode, code)
			assert.Equal(t, tc.wantState, state)
		})
	}
}
