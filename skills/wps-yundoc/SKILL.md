---
name: wps-yundoc
description: "Search, read, extract, summarize, and inspect WPS/KDocs cloud documents (云文档/金山文档) via wpsx. Use for finding docs, reading content, comments, metadata, or document Q&A. Use wps-airpage for creating/updating document content."
---

# WPS Cloud Document Operations

通过 `wpsx` CLI 操作 WPS/KDocs 云文档，支持搜索、读取内容、查看评论、获取元信息。按场景选择对应 Playbook 执行。

## Setup

**Endpoint:** `https://openapi.wps.cn/mcp/v2/kso-yundoc/message`

Examples below use `{ENDPOINT}` for this URL. In shell snippets, either replace `{ENDPOINT}` with the full URL or set `ENDPOINT="https://openapi.wps.cn/mcp/v2/kso-yundoc/message"` and use `$ENDPOINT`.

**认证:** token 由 `wpsx` 自动注入。首次使用请先运行 `wpsx auth --app-id <ID> --app-secret <SECRET>` 完成 WPS 365 OAuth2 授权。

**调用格式:**

```bash
wpsx call <tool> -p '<json>' https://openapi.wps.cn/mcp/v2/kso-yundoc/message -f json
```

---

## Guardrails

**确认规则:**

- 读取文档内容是只读操作，无需确认
- 如需编辑文档内容，请使用 wps-airpage skill（本 skill 仅负责读取）

**空结果处理:**

- 搜索返回空 → 告知用户无匹配文档，建议调整关键词
- 搜索返回多个结果 → 展示文件名和链接列表，让用户选择
- 文档内容提取失败 → 可能是文件格式不支持（内容提取仅支持 otl/docx/pdf；ksheet 可能可搜索但不一定可用本 playbook 提取正文）

**错误处理:**

- API 返回错误 → 将错误信息展示给用户，不要吞掉
- `drive_id` 缺失 → 先用 `get_file_meta` 补查（Playbook 4）
- Endpoint 不可达 / 401 未授权 → 提示用户先运行 `wpsx auth` 完成 WPS 365 授权

---

## Playbook 1: 找文档

**触发:** 用户说"帮我找/搜一下 XX 文档"、"有没有关于 XX 的文档"、"搜索 XX"

```bash
wpsx call kso_yundoc_search_yundoc -p '{"keyword": "{用户关键词}"}' {ENDPOINT} -f json
```

可选参数：
- `page_size`: 返回结果数
- `type`: 文件类型过滤 — `otl`（智能文档）/ `docx` / `ksheet` / `pdf`

将结果中的 `items[].file.name` 和 `link_url` 展示给用户，让用户选择目标文档。

---

## Playbook 2: 读取文档内容

**触发:** 用户说"帮我看看这个文档写了什么"、"读一下 XX 文档"、"总结一下 XX 文档"、"文档里有什么内容"

**步骤:**

1. **定位文档** — 如果用户给了文档名，用搜索定位：
   ```bash
   wpsx call kso_yundoc_search_yundoc -p '{"keyword": "{文档名}"}' {ENDPOINT} -f json
   ```
   从结果中取 `file.id` 和 `file.drive_id`。

2. **如果搜索结果没有 drive_id**，补查 metadata（执行 Playbook 4）。

3. **提取内容:**
   ```bash
   wpsx call kso_yundoc_extract_yundoc_content -p '{"drive_id": "{DRIVE_ID}", "file_id": "{FILE_ID}"}' {ENDPOINT} -f json
   ```
   返回的 `markdown` 字段即为文档内容，直接展示或按用户要求做摘要/翻译/问答。

---

## Playbook 3: 查看文档评论

**触发:** 用户说"看看文档评论"、"大家对这个文档怎么看"、"评审意见是什么"、"评论汇总"

**步骤:**

1. **定位文档** — 同 Playbook 2 步骤 1-2，获取 `drive_id` + `file_id`。

2. **提取评论:**
   ```bash
   wpsx call kso_yundoc_extract_yundoc_comment -p '{"drive_id": "{DRIVE_ID}", "file_id": "{FILE_ID}"}' {ENDPOINT} -f json
   ```

3. **整理输出** — 按评论人分组，展示评论时间和内容。如用户要求，做评审意见汇总。

---

## Playbook 4: 获取文档元信息

**触发:** 需要获取文档的 `drive_id`、文件类型、版本等元信息时使用。通常在 Playbook 2/3 中搜索结果缺少 `drive_id` 时作为补充步骤调用。

```bash
wpsx call kso_yundoc_get_file_meta -p '{"file_id": "{FILE_ID}", "with_drive": true}' {ENDPOINT} -f json
```

返回文件的完整元信息，包括 `drive_id`、文件类型、版本号等。

---

## 跨 Skill 协作

| 场景 | 协作方式 |
|------|----------|
| 编辑/创建文档 | 本 skill 负责**读取**，创建和写入请使用 **wps-airpage** skill |
| 读取 → 修改 → 回写 | 本 skill Playbook 2 读取内容 → 修改 Markdown → wps-airpage Playbook 2 回写 |
| 文档内容导入数据表 | 本 skill 读取文档 → 解析为结构化数据 → **wps-dbsheet** skill 写入 |
| 生成周报/总结文档 | 联动 **wps-calendar**（日程）+ **wps-todo**（待办）+ **wps-meeting**（会议纪要）获取数据 → wps-airpage 发布 |

---

## Additional Resources

- 完整参数说明和返回值结构见 [reference.md](reference.md)
