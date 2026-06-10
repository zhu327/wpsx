---
name: wps-airpage
description: "Create, update, or publish Markdown content into WPS AirPage/智能文档/flexpaper via wpsx. Use for writing/importing reports, publishing local Markdown/code docs, or updating cloud docs; use wps-yundoc for read/search-only document tasks."
---

# WPS AirPage Document Import

通过 `wpsx` CLI 将 Markdown 内容导入到 WPS AirPage（智能文档/flexpaper）。

## Setup

**Endpoint:** `https://openapi.wps.cn/mcp/v2/kso-airpage/message`

**认证:** token 由 `wpsx` 自动注入。首次使用请先运行 `wpsx auth --app-id <ID> --app-secret <SECRET>` 完成 WPS 365 OAuth2 授权。

**调用格式:**

```bash
wpsx call <tool> -p '<json>' https://openapi.wps.cn/mcp/v2/kso-airpage/message -f json
```

**支持的 Markdown 语法:** 标题(# ## ###)、**粗体**、*斜体*、~~删除线~~、`行内代码`、代码块、引用块、有序/无序列表、表格、链接、水平线。

---

## Guardrails

**确认规则:**

- AI 生成内容时（Playbook 3），必须先展示内容让用户确认后再导入
- 更新已有文档时（Playbook 2/4），提醒用户此操作会覆盖原有内容

**错误处理:**

- API 返回错误 → 将错误信息展示给用户，不要吞掉
- file_name 匹配多个文件 → 展示列表让用户选择，或让用户提供 file_id
- Endpoint 不可达 / 401 未授权 → 提示用户先运行 `wpsx auth` 完成 WPS 365 授权

**内容限制:**

- Markdown 内容建议以 `# 标题` 开头，工具自动提取作为文档标题
- 如果内容过大导致 API 超时或失败，考虑分段导入或精简内容

---

## Playbook 1: 创建新文档

**触发:** 用户说"帮我写个文档"、"把这个内容发布成文档"、"创建一个XX文档"、"生成一个需求文档"

**步骤:**

1. **组织 Markdown 内容** — 确保内容以 `# 文档标题` 开头，工具会自动提取一级标题作为文档标题。

2. **导入:**
   ```bash
   wpsx call kso_airpage_import_markdown_data \
     -p '{"markdown_content": "# 文档标题\n\n## 第一章\n\n正文内容..."}' \
     https://openapi.wps.cn/mcp/v2/kso-airpage/message -f json
   ```

3. 将返回的文档链接展示给用户。

---

## Playbook 2: 更新已有文档

**触发:** 用户说"更新一下XX文档"、"把内容写到XX文档里"、"覆盖XX文档的内容"

**步骤:**

1. **定位文档** — 如果用户给了文件名：
   - 可以用 `file_name` 参数直接指定（需文件名唯一）
   - 或者先通过 wps-yundoc skill 搜索获取 `file_id`

2. **导入到指定文件:**
   ```bash
   wpsx call kso_airpage_import_markdown_data \
     -p '{"file_id": "{FILE_ID}", "markdown_content": "# 新标题\n\n更新后的内容..."}' \
     https://openapi.wps.cn/mcp/v2/kso-airpage/message -f json
   ```

**注意:** 如果 `markdown_content` 中没有一级标题，会保持原文档标题不变。

---

## Playbook 3: AI 生成内容并发布

**触发:** 用户说"帮我写一份XX报告/方案/总结，发到文档里"、"生成一份XX并创建文档"

**步骤:**

1. **根据用户要求生成 Markdown 内容** — 按用户意图撰写内容，以 `# 标题` 开头。
2. **先展示给用户确认** — 让用户审阅生成的内容。
3. **确认后导入:**
   ```bash
   wpsx call kso_airpage_import_markdown_data \
     -p '{"markdown_content": "{生成的完整Markdown}"}' \
     https://openapi.wps.cn/mcp/v2/kso-airpage/message -f json
   ```

---

## Playbook 4: 读取文档 → 修改 → 回写

**触发:** 用户说"帮我改一下XX文档的内容"、"在XX文档里加一段"、"优化一下XX文档的表述"

**步骤:**

1. **读取原文档** — 使用 wps-yundoc skill 的 Playbook 2 提取文档的 Markdown 内容和 `file_id`。
2. **修改内容** — 按用户要求编辑 Markdown。
3. **展示修改后的内容给用户确认。**
4. **回写:**
   ```bash
   wpsx call kso_airpage_import_markdown_data \
     -p '{"file_id": "{FILE_ID}", "markdown_content": "{修改后的完整Markdown}"}' \
     https://openapi.wps.cn/mcp/v2/kso-airpage/message -f json
   ```

---

## Playbook 5: 将本地文件/代码文档发布到云端

**触发:** 用户说"把这个 README 发到文档上"、"把本地笔记同步到云文档"、"发布到 AirPage"

**步骤:**

1. **读取本地文件** — 使用 Read 工具获取本地 Markdown/文本文件内容。
2. **导入:**
   ```bash
   wpsx call kso_airpage_import_markdown_data \
     -p '{"markdown_content": "{文件内容}"}' \
     https://openapi.wps.cn/mcp/v2/kso-airpage/message -f json
   ```

---

## 参数速查

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| markdown_content | string | **yes** | 完整 Markdown 内容，建议以 `# 标题` 开头 |
| file_id | string | no | 指定导入到已有文件（精确匹配） |
| file_name | string | no | 按文件名查找已有文件（需唯一） |
| title | string | no | 一般不需要，工具自动从 markdown 提取标题 |

- 不提供 `file_id` 和 `file_name` → 自动创建新文件
- 提供 `file_id` → 导入到指定文件
- 提供 `file_name` → 搜索同名文件，找到多个会报错

---

## 跨 Skill 协作

- **读取已有文档内容** → 使用 wps-yundoc skill 的 Playbook 2 获取文档 Markdown 内容和 file_id
- **编辑已有文档** → 先用 wps-yundoc 读取，修改后用本 skill 的 Playbook 2 回写
- **从其他来源生成文档** → 联动 wps-calendar（日程）、wps-todo（待办）、wps-meeting（会议纪要）等 skill 获取数据，生成 Markdown 后通过 Playbook 1 发布
