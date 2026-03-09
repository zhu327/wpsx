package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// WPSConfig 存储 WPS 365 OAuth2 认证信息。
type WPSConfig struct {
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	AppID                 string    `json:"app_id"`
	AppSecret             string    `json:"app_secret"`
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
}

// wpsConfigDir 返回 ~/.config/wps 目录路径。
func wpsConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "wps"), nil
}

// wpsConfigPath 返回配置文件的完整路径。
func wpsConfigPath() (string, error) {
	dir, err := wpsConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadWPSConfig 从 ~/.config/wps/config.json 读取 WPS 配置。
// 若文件不存在，返回提示用户先授权的错误。
func LoadWPSConfig() (*WPSConfig, error) {
	path, err := wpsConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path) //nolint:gosec // path from user home directory
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("请先运行 `wps auth` 完成授权")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read WPS config: %w", err)
	}

	var cfg WPSConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse WPS config: %w", err)
	}
	return &cfg, nil
}

// SaveWPSConfig 原子写入 WPS 配置到 ~/.config/wps/config.json。
// 使用临时文件 + rename 保证写入原子性。
func SaveWPSConfig(cfg *WPSConfig) error {
	if cfg == nil {
		return fmt.Errorf("WPSConfig cannot be nil")
	}
	dir, err := wpsConfigDir()
	if err != nil {
		return err
	}

	if mkErr := os.MkdirAll(dir, 0o700); mkErr != nil {
		return fmt.Errorf("failed to create WPS config directory: %w", mkErr)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal WPS config: %w", err)
	}

	// 原子写：先写临时文件再 rename
	tmpPath := filepath.Join(dir, "config.json.tmp")
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil { //nolint:gosec // user config file
		return fmt.Errorf("failed to write WPS config temp file: %w", err)
	}

	configPath := filepath.Join(dir, "config.json")
	if err := os.Rename(tmpPath, configPath); err != nil {
		return fmt.Errorf("failed to rename WPS config: %w", err)
	}
	return nil
}

// tokenRefreshBuffer 是提前判定 access_token 过期的缓冲时间。
const tokenRefreshBuffer = 5 * time.Minute

// wpsTokenLockPath 返回 token 刷新的文件锁路径。
func wpsTokenLockPath() (string, error) {
	dir, err := wpsConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.lock"), nil
}

// isAccessTokenExpired 判断 access_token 是否即将过期（提前 5 分钟）。
func isAccessTokenExpired(cfg *WPSConfig) bool {
	return cfg.AccessToken == "" || time.Now().Add(tokenRefreshBuffer).After(cfg.AccessTokenExpiresAt)
}

// GetValidAccessToken 返回有效的 access_token，必要时自动刷新。
// 使用文件锁防止并发重复刷新。
func GetValidAccessToken() (string, error) {
	cfg, err := LoadWPSConfig()
	if err != nil {
		return "", err
	}

	// 快速路径：token 未过期
	if !isAccessTokenExpired(cfg) {
		return cfg.AccessToken, nil
	}

	// 需要刷新：加文件锁
	lockPath, err := wpsTokenLockPath()
	if err != nil {
		return "", err
	}

	// 确保锁文件目录存在
	lockDir := filepath.Dir(lockPath)
	if mkErr := os.MkdirAll(lockDir, 0o700); mkErr != nil {
		return "", fmt.Errorf("failed to create WPS config directory: %w", mkErr)
	}

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600) //nolint:gosec // lock file
	if err != nil {
		return "", fmt.Errorf("failed to open token lock file: %w", err)
	}
	defer func() { _ = lockFile.Close() }()

	// 加独占锁（阻塞直到获取）
	if flockErr := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); flockErr != nil {
		return "", fmt.Errorf("failed to acquire token lock: %w", flockErr)
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }() //nolint:errcheck // best effort unlock

	// Double-check：加锁后重新读配置，可能其他进程已刷新
	cfg, err = LoadWPSConfig()
	if err != nil {
		return "", err
	}
	if !isAccessTokenExpired(cfg) {
		return cfg.AccessToken, nil
	}

	// 执行刷新：refreshAccessToken 通过指针就地修改 cfg 并持久化。
	if err := refreshAccessToken(cfg); err != nil {
		return "", err
	}
	return cfg.AccessToken, nil
}

// tokenResponse 是 WPS OAuth2 token 接口的响应体。
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	Msg              string `json:"msg"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	// 失败时的字段
	Code int `json:"code"`
}

// refreshAccessToken 调用 WPS 刷新接口更新 cfg 中的 token，并保存到磁盘。
func refreshAccessToken(cfg *WPSConfig) error {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", cfg.RefreshToken)
	form.Set("client_id", cfg.AppID)
	form.Set("client_secret", cfg.AppSecret)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.PostForm("https://openapi.wps.cn/oauth2/token", form)
	if err != nil {
		return fmt.Errorf("failed to refresh WPS token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read WPS token response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("WPS token refresh failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse WPS token response: %w", err)
	}

	if tokenResp.Code != 0 {
		return fmt.Errorf("WPS token refresh failed (code=%d): %s", tokenResp.Code, tokenResp.Msg)
	}
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("WPS token refresh returned empty access_token")
	}

	now := time.Now()
	cfg.AccessToken = tokenResp.AccessToken
	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 7200 // 2 hours default per WPS API docs
	}
	cfg.AccessTokenExpiresAt = now.Add(time.Duration(expiresIn) * time.Second)
	if tokenResp.RefreshToken != "" {
		cfg.RefreshToken = tokenResp.RefreshToken
	}
	if tokenResp.RefreshExpiresIn > 0 {
		cfg.RefreshTokenExpiresAt = now.Add(time.Duration(tokenResp.RefreshExpiresIn) * time.Second)
	}

	return SaveWPSConfig(cfg)
}
