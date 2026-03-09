package commands

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// wpsScopes 是 WPS 365 MCP 平台所需的授权 scope 列表。
var wpsScopes = strings.Join([]string{
	"kso.mcp_airpage.readwrite",
	"kso.mcp_calendar.readwrite",
	"kso.mcp_dbsheet.readwrite",
	"kso.mcp_mail.readwrite",
	"kso.mcp_meeting.readwrite",
	"kso.mcp_message.readwrite",
	"kso.mcp_scenario.readwrite",
	"kso.mcp_todo.readwrite",
	"kso.mcp_yundoc.readwrite",
}, ",")

// generateState 生成 16 字节随机 hex 字符串作为 OAuth2 state 参数。
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// buildAuthorizationURL 根据参数构造 WPS OAuth2 授权跳转链接。
func buildAuthorizationURL(appID, callbackURL, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", appID)
	params.Set("redirect_uri", callbackURL)
	params.Set("scope", wpsScopes)
	params.Set("state", state)

	return "https://openapi.wps.cn/oauth2/auth?" + params.Encode()
}

// parseCallbackURL 从用户粘贴的回调 URL 中解析 code 和 state。
func parseCallbackURL(rawURL string) (code, state string, err error) {
	parsed, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		return "", "", fmt.Errorf("invalid callback URL: %w", parseErr)
	}

	code = parsed.Query().Get("code")
	if code == "" {
		return "", "", fmt.Errorf("callback URL 中未找到 code 参数，请确认已完成授权")
	}

	state = parsed.Query().Get("state")
	return code, state, nil
}

// exchangeCodeForToken 用 authorization code 换取 access_token，并写入配置文件。
func exchangeCodeForToken(appID, appSecret, code, callbackURL string) error {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", appID)
	form.Set("client_secret", appSecret)
	form.Set("code", code)
	form.Set("redirect_uri", callbackURL)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.PostForm("https://openapi.wps.cn/oauth2/token", form)
	if err != nil {
		return fmt.Errorf("请求 token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取 token 响应失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("获取 token 失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("解析 token 响应失败: %w", err)
	}

	if tokenResp.Code != 0 {
		return fmt.Errorf("获取 token 失败 (code=%d): %s", tokenResp.Code, tokenResp.Msg)
	}
	if tokenResp.AccessToken == "" {
		return fmt.Errorf("获取 token 返回空 access_token，请检查 app_id/app_secret 是否正确")
	}
	if tokenResp.RefreshToken == "" {
		return fmt.Errorf("获取 token 返回空 refresh_token，请联系 WPS 开放平台支持")
	}

	now := time.Now()
	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 7200
	}
	cfg := &WPSConfig{
		AppID:                appID,
		AppSecret:            appSecret,
		AccessToken:          tokenResp.AccessToken,
		RefreshToken:         tokenResp.RefreshToken,
		AccessTokenExpiresAt: now.Add(time.Duration(expiresIn) * time.Second),
	}
	if tokenResp.RefreshExpiresIn > 0 {
		cfg.RefreshTokenExpiresAt = now.Add(time.Duration(tokenResp.RefreshExpiresIn) * time.Second)
	}

	return SaveWPSConfig(cfg)
}

// AuthCmd 创建 auth 子命令。
func AuthCmd() *cobra.Command {
	var appID string
	var appSecret string
	var callbackURL string

	cmd := &cobra.Command{
		Use:   "auth",
		Short: "WPS 365 OAuth2 授权",
		Long: `通过 WPS 365 OAuth2 授权流程获取 access_token。

授权信息将保存到 ~/.config/wps/config.json。
后续调用 https://openapi.wps.cn/mcp/ 时将自动使用该 token。`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if appID == "" {
				return fmt.Errorf("--app-id 不能为空")
			}
			if appSecret == "" {
				return fmt.Errorf("--app-secret 不能为空")
			}
			if strings.TrimSpace(callbackURL) == "" {
				callbackURL = "http://localhost"
			}

			// 1. 生成随机 state
			state, err := generateState()
			if err != nil {
				return err
			}

			// 2. 构造授权链接
			authURL := buildAuthorizationURL(appID, callbackURL, state)

			// 3. 打印授权链接
			fmt.Fprintln(cmd.OutOrStdout(), "请在浏览器中打开以下链接完成授权：")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), authURL)
			fmt.Fprintln(cmd.OutOrStdout())

			// 4. 等待用户输入回调 URL
			fmt.Fprint(cmd.OutOrStdout(), "授权完成后，请将浏览器地址栏中的完整回调 URL 粘贴到此处：\n> ")

			reader := bufio.NewReader(os.Stdin)
			rawCallbackURL, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("读取输入失败: %w", err)
			}
			rawCallbackURL = strings.TrimSpace(rawCallbackURL)

			// 5. 解析 code 和 state
			code, returnedState, err := parseCallbackURL(rawCallbackURL)
			if err != nil {
				return err
			}

			// 6. 校验 state
			if returnedState != state {
				return fmt.Errorf("state 不匹配，可能存在安全风险，请重新运行 auth 命令")
			}

			// 7. 换取 token
			fmt.Fprintln(cmd.OutOrStdout(), "正在获取 access_token...")
			if err := exchangeCodeForToken(appID, appSecret, code, callbackURL); err != nil {
				return err
			}

			// 8. 成功提示
			fmt.Fprintln(cmd.OutOrStdout(), "授权成功！认证信息已保存到 ~/.config/wps/config.json")
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "WPS 应用 APPID（必填）")
	cmd.Flags().StringVar(&appSecret, "app-secret", "", "WPS 应用 APPKEY（必填）")
	cmd.Flags().StringVar(&callbackURL, "callback-url", "http://localhost", "授权回调地址（默认: http://localhost）")

	return cmd
}
