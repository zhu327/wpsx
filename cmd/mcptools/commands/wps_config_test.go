package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWPSConfig_FileNotExist(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := LoadWPSConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "请先运行")
}

func TestSaveAndLoadWPSConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := &WPSConfig{
		AppID:                 "test_app_id",
		AppSecret:             "test_app_secret",
		AccessToken:           "access_token_value",
		RefreshToken:          "refresh_token_value",
		AccessTokenExpiresAt:  time.Now().Add(2 * time.Hour).Truncate(time.Second),
		RefreshTokenExpiresAt: time.Now().Add(365 * 24 * time.Hour).Truncate(time.Second),
	}

	err := SaveWPSConfig(cfg)
	require.NoError(t, err)

	loaded, err := LoadWPSConfig()
	require.NoError(t, err)

	assert.Equal(t, cfg.AppID, loaded.AppID)
	assert.Equal(t, cfg.AppSecret, loaded.AppSecret)
	assert.Equal(t, cfg.AccessToken, loaded.AccessToken)
	assert.Equal(t, cfg.RefreshToken, loaded.RefreshToken)
	assert.WithinDuration(t, cfg.AccessTokenExpiresAt, loaded.AccessTokenExpiresAt, time.Second)
	assert.WithinDuration(t, cfg.RefreshTokenExpiresAt, loaded.RefreshTokenExpiresAt, time.Second)
}

func TestSaveWPSConfig_CreatesDirectory(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := &WPSConfig{AppID: "id", AppSecret: "secret"}

	err := SaveWPSConfig(cfg)
	require.NoError(t, err)

	configPath := filepath.Join(tmpHome, ".config", "wps", "config.json")
	_, statErr := os.Stat(configPath)
	assert.NoError(t, statErr)
}

func TestGetValidAccessToken_NotExpired(t *testing.T) {
	// Arrange: token 未过期
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := &WPSConfig{
		AppID:                "id",
		AppSecret:            "secret",
		AccessToken:          "valid_token",
		RefreshToken:         "refresh_token",
		AccessTokenExpiresAt: time.Now().Add(1 * time.Hour),
	}
	require.NoError(t, SaveWPSConfig(cfg))

	// Act
	token, err := GetValidAccessToken()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "valid_token", token)
}

func TestGetValidAccessToken_NoConfig(t *testing.T) {
	// Arrange: 没有配置文件
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Act
	_, err := GetValidAccessToken()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "请先运行")
}
