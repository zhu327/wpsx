<h1 align="center">wpsx</h1>

<p align="center">
  <strong>A command-line client specifically designed for WPS 365 MCP — born for AI Agents.</strong>
</p>

<p align="center">
  Authorize once. Call any WPS 365 MCP tool from your terminal, scripts, or AI agent pipelines — no manual token management required.
</p>

---

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [WPS 365 Authorization](#wps-365-authorization)
- [Calling WPS 365 MCP](#calling-wps-365-mcp)
- [Commands](#commands)
  - [tools](#tools)
  - [call](#call)
  - [resources / prompts](#resources--prompts)
  - [shell](#shell)
  - [web](#web)
  - [auth](#auth)
- [AI Agent Skills](#ai-agent-skills)
- [Transport Options](#transport-options)
- [Output Formats](#output-formats)
- [Authentication Options](#authentication-options)
- [Server Aliases](#server-aliases)
- [LLM Apps Config Management](#llm-apps-config-management)
- [Server Modes](#server-modes)
  - [Mock Server](#mock-server)
  - [Proxy Mode](#proxy-mode)
  - [Guard Mode](#guard-mode)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

`wpsx` is a CLI built on top of the [MCP Tools](https://github.com/f/mcptools) foundation, extended specifically for the **WPS 365 MCP platform** (`https://365.kdocs.cn/3rd/open/documents/app-integration-dev/mcp-server/introduction`).

It adds:

- **`wpsx auth`** — One-time OAuth2 authorization flow. Credentials are stored at `~/.config/wps/config.json`.
- **Automatic token injection** — When a target URL starts with `https://openapi.wps.cn/mcp/`, `wpsx` automatically reads and injects a valid `Bearer` token. Expired tokens are refreshed transparently using a file-lock–protected refresh flow (safe for concurrent agent processes).
- **All standard MCP Tools commands** — `tools`, `call`, `resources`, `prompts`, `shell`, `web`, `mock`, `proxy`, `guard`, `alias`, `configs`, etc.

---

## Installation

### From Source

```bash
git clone https://github.com/zhu327/wpsx
cd wpsx
go build -o bin/wpsx ./cmd/mcptools
# Optionally move to PATH
mv bin/wpsx /usr/local/bin/wpsx
```

---

## Quick Start

```bash
# 1. Authorize with WPS 365 (one-time setup)
wpsx auth --app-id YOUR_APP_ID --app-secret YOUR_APP_SECRET

# 2. List tools on the WPS 365 MCP platform — token is injected automatically
wpsx tools https://openapi.wps.cn/mcp/v2/kso-yundoc/message

# 3. Call a tool
wpsx call kso_yundoc_search_yundoc --params '{"keyword":"AI Agent"}' https://openapi.wps.cn/mcp/v2/kso-yundoc/message
```

---

## WPS 365 Authorization

`wpsx auth` implements the WPS 365 OAuth2 authorization code flow.

```bash
wpsx auth --app-id <APPID> --app-secret <APPKEY> [--callback-url <url>]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--app-id` | Yes | — | WPS application APPID |
| `--app-secret` | Yes | — | WPS application APPKEY |
| `--callback-url` | No | `http://localhost` | OAuth2 redirect URI configured in WPS developer console |

**Flow:**

1. `wpsx auth` prints an authorization URL. Open it in a browser.
2. Log in and approve the requested permissions (all WPS 365 MCP scopes are requested automatically).
3. After approval, copy the full redirect URL from the browser address bar and paste it into the terminal.
4. `wpsx` exchanges the authorization code for tokens and saves them to `~/.config/wps/config.json`.

```
$ wpsx auth --app-id AK2024xxx --app-secret your_secret

请在浏览器中打开以下链接完成授权：

https://openapi.wps.cn/oauth2/auth?client_id=AK2024xxx&...

授权完成后，请将浏览器地址栏中的完整回调 URL 粘贴到此处：
> http://localhost?code=ga_xxxxx&state=abc123def456...

正在获取 access_token...
授权成功！认证信息已保存到 ~/.config/wps/config.json
```

**What is stored (`~/.config/wps/config.json`):**

```json
{
  "app_id": "AK2024xxx",
  "app_secret": "...",
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "access_token_expires_at": "2026-03-09T14:00:00Z",
  "refresh_token_expires_at": "2027-03-09T12:00:00Z"
}
```

> **Security:** The config file is created with permissions `0600` (owner-readable only). It contains sensitive credentials — do not commit or share it.

---

## Calling WPS 365 MCP

Once authorized, any `wpsx` command targeting `https://openapi.wps.cn/mcp/` will automatically:

1. Read the stored `access_token`.
2. If it expires within 5 minutes, refresh it via the WPS refresh endpoint — using a file lock at `~/.config/wps/token.lock` to prevent concurrent refresh races between multiple agent processes.
3. Inject a `Bearer <token>` header into the request.

No `--auth-header` flag needed:

```bash
# Token is injected automatically
wpsx tools https://openapi.wps.cn/mcp/v2/kso-yundoc/message
wpsx call kso_yundoc_search_yundoc --params '{"keyword":"AI Agent"}' https://openapi.wps.cn/mcp/v2/kso-yundoc/message
```

To override with an explicit token (bypasses auto-inject):

```bash
wpsx tools --auth-header "Bearer explicit_token" https://openapi.wps.cn/mcp/v2/kso-yundoc/message
```

**Auth priority (highest to lowest):**

| Priority | Source |
|----------|--------|
| 1 | `--auth-user username:password` (Basic Auth) |
| 2 | `--auth-header "Bearer ..."` (Custom header) |
| 3 | URL-embedded credentials (`https://user:pass@host/`) |
| 4 | Auto-injected WPS token (WPS MCP URLs only) |

---

## Commands

### `tools`

List all tools available on an MCP server.

```bash
# WPS 365 (token auto-injected)
wpsx tools https://openapi.wps.cn/mcp/v2/kso-yundoc/message

# Any MCP server via stdio
wpsx tools npx -y @modelcontextprotocol/server-filesystem ~

# HTTP with explicit auth
wpsx tools --auth-header "Bearer token" https://other-mcp.example.com
```

Output (default table format):

```
read_file(path:str)
     Read the complete contents of a file.
write_file(path:str, content:str)
     Write content to a file.
list_dir(path:str)
     List directory contents.
```

### `call`

Call a tool, resource, or prompt on an MCP server.

```bash
# Call a WPS tool (token auto-injected)
wpsx call kso_yundoc_search_yundoc --params '{"keyword":"AI Agent"}' https://openapi.wps.cn/mcp/v2/kso-yundoc/message

# Call with pretty JSON output
wpsx call read_file --params '{"path":"README.md"}' --format pretty \
  npx -y @modelcontextprotocol/server-filesystem ~

# Call a resource
wpsx call resource:test://static/resource/1 npx -y @modelcontextprotocol/server-everything -f json
```

### `resources` / `prompts`

```bash
wpsx resources https://openapi.wps.cn/mcp/v2/kso-yundoc/message
wpsx prompts npx -y @modelcontextprotocol/server-everything
```

### `shell`

Start an interactive shell session with an MCP server:

```bash
wpsx shell https://openapi.wps.cn/mcp/v2/kso-yundoc/message
```

```
mcp > tools
some_tool(param:str)
     Description of the tool.

mcp > some_tool {"param":"hello"}
...response...

mcp > /q
```

Shell commands: `tools`, `resources`, `prompts`, `call <entity>`, `format [json|pretty|table]`, `/h` (help), `/q` (quit).

### `web`

Launch a browser-based UI for exploring and calling MCP tools:

```bash
wpsx web https://openapi.wps.cn/mcp/v2/kso-yundoc/message

# Custom port
wpsx web --port 8080 https://openapi.wps.cn/mcp/v2/kso-yundoc/message
```

Opens at `http://localhost:41999` by default. Includes auto-generated parameter forms, JSON/pretty response views, and supports all tool/resource/prompt types.

### `auth`

WPS 365 OAuth2 authorization. See [WPS 365 Authorization](#wps-365-authorization) above.

```bash
wpsx auth --app-id <APPID> --app-secret <APPKEY> [--callback-url <url>]
```

---

## AI Agent Skills

`wpsx` ships with a set of **Claude Agent Skills** — ready-made playbooks that teach AI agents how to call each WPS 365 MCP service correctly. Drop these skills into your Claude project and your agent will know the right endpoint, tool names, parameters, and guardrails for every operation.

Skills live under `skills/` in this repository:

| Skill | MCP Endpoint | Description |
|-------|-------------|-------------|
| [`wps-airpage`](skills/wps-airpage/SKILL.md) | `kso-airpage` | Import Markdown content into WPS AirPage (智能文档). Create, update, or publish documents. |
| [`wps-calendar`](skills/wps-calendar/SKILL.md) | `kso-calendar` | Manage calendar events (日历/日程). Query schedules, create/update events, check free/busy status, manage attendees. |
| [`wps-dbsheet`](skills/wps-dbsheet/SKILL.md) | `kso-dbsheet` | CRUD on WPS smart sheets (数据表/多维表格). Query, insert, and update records and sheet structures. |
| [`wps-mail`](skills/wps-mail/SKILL.md) | `kso-mail` | Read, search, compose, and send WPS email. List inbox, read messages, draft and send. |
| [`wps-meeting`](skills/wps-meeting/SKILL.md) | `kso-meeting` | Manage WPS online meetings (在线会议). Create/modify/cancel meetings, manage participants, get transcripts and summaries. |
| [`wps-message`](skills/wps-message/SKILL.md) | `kso-message` | Read and search WPS IM chat conversations and messages. Summarize chats, check unread/@me messages, search history. |
| [`wps-todo`](skills/wps-todo/SKILL.md) | `kso-todo` | Manage WPS personal tasks (个人待办). Create, query, update, and complete todo items. |
| [`wps-yundoc`](skills/wps-yundoc/SKILL.md) | `kso-yundoc` | Search, read, and share WPS cloud documents (云文档). Work with document content, comments, and metadata. |

### Using Skills

Each skill is a `SKILL.md` file with structured playbooks for common scenarios. To use them with Claude:

1. Copy the `skills/` directory into your project (or reference this repo's skills path in Claude settings).
2. When working with WPS 365 data, Claude will automatically invoke the relevant skill to guide the agent.

Skills handle:
- **Correct endpoints** — each skill points to `https://openapi.wps.cn/mcp/v2/{service}/message`
- **Auth** — token is auto-injected by `wpsx`; skills remind agents to run `wpsx auth` first if needed
- **Guardrails** — confirmation rules, pagination strategy, error handling
- **Cross-skill collaboration** — skills reference each other (e.g., `wps-airpage` + `wps-yundoc` for read-then-rewrite workflows)

---

## Transport Options

### Stdio

Spawns a local process and communicates via stdin/stdout:

```bash
wpsx tools npx -y @modelcontextprotocol/server-filesystem ~
```

### Streamable HTTP (default for HTTP/HTTPS URLs)

Modern MCP transport with session management and streaming support:

```bash
wpsx tools https://openapi.wps.cn/mcp/v2/kso-yundoc/message
wpsx tools https://other-mcp-server.com
```

### HTTP SSE (legacy)

Used when URL ends with `/sse`:

```bash
wpsx tools http://localhost:3001/sse
```

> SSE currently supports MCP protocol version 2024-11-05 only.

---

## Output Formats

Use `--format` / `-f` with any command:

| Format | Flag | Description |
|--------|------|-------------|
| Table (default) | `-f table` | Colorized, man-page style |
| Compact JSON | `-f json` | Single-line JSON |
| Pretty JSON | `-f pretty` | Indented JSON |

```bash
wpsx tools -f pretty https://openapi.wps.cn/mcp/v2/kso-yundoc/message
wpsx call my_tool -f json --params '{}' https://openapi.wps.cn/mcp/v2/kso-yundoc/message
```

---

## Authentication Options

For non-WPS servers requiring auth:

```bash
# Basic auth
wpsx tools --auth-user username:password https://protected-mcp.example.com

# Bearer token
wpsx tools --auth-header "Bearer eyJhbGci..." https://protected-mcp.example.com

# URL-embedded credentials
wpsx tools https://user:password@mcp.example.com
```

---

## Server Aliases

Save long server commands or URLs under short aliases:

```bash
# Add alias
wpsx alias add wps-mcp https://openapi.wps.cn/mcp/v2/kso-yundoc/message
wpsx alias add myfs npx -y @modelcontextprotocol/server-filesystem ~/

# Use alias
wpsx tools wps-mcp
wpsx call read_file --params '{"path":"README.md"}' myfs

# List aliases
wpsx alias list

# Remove alias
wpsx alias remove wps-mcp
```

Aliases are stored in `~/.mcpt/aliases.json`.

---

## LLM Apps Config Management

Manage MCP server configurations across AI clients (Claude, VS Code, Claude Desktop, Windsurf, etc.):

```bash
# Scan for all MCP server configs across supported apps
wpsx configs scan

# List all configured servers
wpsx configs ls

# View a specific app's config
wpsx configs view Claude

# Add a WPS 365 MCP server to Claude and VS Code at once
wpsx configs set Claude,vscode wps-365 https://openapi.wps.cn/mcp/v2/kso-yundoc/message

# Add a server with a custom auth header
wpsx configs set Claude my-api https://api.example.com/mcp \
  --headers "Authorization=Bearer token"

# Remove a server
wpsx configs remove Claude wps-365

# Convert a command to MCP JSON config format
wpsx configs as-json https://openapi.wps.cn/mcp/v2/kso-yundoc/message
# Output: {"url":"https://openapi.wps.cn/mcp/v2/kso-yundoc/message"}
```

Predefined aliases: `vscode`, `vscode-insiders`, `Claude`, `windsurf`, `claude-desktop`, `claude-code`.

---

## Server Modes

### Mock Server

Create a simulated MCP server for testing clients:

```bash
# Single tool
wpsx mock tool hello_world "A greeting tool"

# Multiple entity types
wpsx mock tool hello_world "A greeting tool" \
       prompt welcome "Welcome prompt" "Hello {{name}}, welcome to {{location}}!" \
       resource docs://readme "Documentation" "This is mock content"
```

Logs are written to `~/.mcpt/logs/mock.log`.

### Proxy Mode

Expose shell scripts or inline commands as MCP tools:

```bash
# Register a script as a tool
wpsx proxy tool add_operation "Adds a and b" "a:int,b:int" ./add.sh

# Register an inline command
wpsx proxy tool add_operation "Adds a and b" "a:int,b:int" \
  -e 'echo "total: $(($a+$b))"'

# Start the proxy server
wpsx proxy start
```

Parameters are passed to scripts as environment variables. If output is `data:image/png;base64,...`, it is returned as image content. Logs: `~/.mcpt/logs/proxy.log`.

### Guard Mode

Restrict access to specific tools/resources/prompts via allow/deny patterns:

```bash
# Allow only read operations
wpsx guard --allow 'tools:read_*' npx -y @modelcontextprotocol/server-filesystem ~

# Deny destructive operations
wpsx guard --deny 'tools:write_*,delete_*,create_*' npx -y @modelcontextprotocol/server-filesystem ~

# Combined patterns
wpsx guard --allow 'tools:read_*,prompts:system_*' --deny tools:execute_* \
  npx -y @modelcontextprotocol/server-filesystem ~
```

Pattern syntax: `entity_type:glob_pattern` where `*` is a wildcard.

Use in MCP client configs to sandbox AI agents:

```json
{
  "filesystem": {
    "command": "wpsx",
    "args": ["guard", "--deny", "tools:write_*,create_*,delete_*",
             "npx", "-y", "@modelcontextprotocol/server-filesystem", "/data"]
  }
}
```

Logs: `~/.mcpt/logs/guard.log`.

---

## Contributing

Contributions are welcome. Please open an issue or pull request.

---

## License

MIT License.
