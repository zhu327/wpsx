---
name: wps-mail
description: "Read, search, summarize, draft, reply/forward, and send WPS邮件 via wpsx. Use for inbox/收件箱, unread mail, 查邮件, 发邮件, 回复, 转发, mail search, or email summaries; confirm before sending."
---

# WPS Mail (邮件) Operations

通过 `wpsx` CLI 操作 WPS 邮件，查收邮件、搜索邮件、创建草稿并发送。按场景选择对应 Playbook 执行。

## Setup

**Endpoint:** `https://openapi.wps.cn/mcp/v2/kso-mail/message`

Examples below use `{ENDPOINT}` for this URL. In shell snippets, either replace `{ENDPOINT}` with the full URL or set `ENDPOINT="https://openapi.wps.cn/mcp/v2/kso-mail/message"` and use `$ENDPOINT`.

**认证:** token 由 `wpsx` 自动注入。首次使用请先运行 `wpsx auth --app-id <ID> --app-secret <SECRET>` 完成 WPS 365 OAuth2 授权。

**调用格式:**

```bash
wpsx call <tool> -p '<json>' https://openapi.wps.cn/mcp/v2/kso-mail/message -f json
```

**时间格式:** 使用 `start_time_str` / `end_time_str`，格式为 `yyyy-mm-dd hh:mm:ss`（北京时间）。也支持秒级时间戳 `start_time` / `end_time`，两者同时提供时时间戳优先。

---

## Guardrails

**当前时间:** 涉及"今天"、"本周"等相对时间时，使用系统当前时间（北京时间 UTC+8）计算精确时间范围。

**确认规则:**

- 发送邮件前（`send_draft`），必须先展示草稿内容（收件人、主题、正文摘要）让用户确认
- 回复/转发邮件前，展示组织好的内容让用户确认
- 禁止使用虚构的邮箱地址

**空结果处理:**

- 搜索/查询返回空 → 告知用户无匹配邮件，建议调整关键词或时间范围
- keyword 匹配收件人返回多个候选 → 展示列表让用户选择

**分页策略:**

- 用户要求"全部/所有" → 自动翻页直到结束
- 其他情况 → 展示首页结果，告知"还有更多数据，需要继续查看吗？"

**错误处理:**

- API 返回错误 → 将错误信息展示给用户，不要吞掉
- Endpoint 不可达 / 401 未授权 → 提示用户先运行 `wpsx auth` 完成 WPS 365 授权

---

## Playbook 1: 查看最近邮件 / 收件箱

**触发:** 用户说"看看我的邮件"、"最近收到什么邮件"、"今天有什么邮件"、"收件箱有什么"

```bash
wpsx call kso_mail_list_letter \
  -p '{"page_size": 20}' \
  {ENDPOINT} -f json
```

可选过滤：
- `start_time_str` / `end_time_str`: 限定时间范围
- `filter`: `["unread"]`（未读）或 `["flagged"]`（星标）
- `page_token`: 翻页

整理输出：按时间倒序展示邮件主题、发件人、时间、是否已读/有附件。

---

## Playbook 2: 查看未读邮件

**触发:** 用户说"有没有未读邮件"、"帮我看看没读的邮件"、"未读邮件有哪些"

```bash
wpsx call kso_mail_list_letter \
  -p '{"page_size": 20, "filter": ["unread"]}' \
  {ENDPOINT} -f json
```

输出：未读邮件数量 + 每封邮件的主题、发件人、时间摘要。

---

## Playbook 3: 读取邮件详情

**触发:** 用户说"帮我看看这封邮件写了什么"、"读一下XX邮件"、"邮件内容是什么"

**步骤:**

1. **定位邮件** — 如果用户给了关键词，先搜索：
   ```bash
   wpsx call kso_mail_search_letter \
     -p '{"keyword": "{关键词}", "page_size": 10}' \
     {ENDPOINT} -f json
   ```
   从结果 `items[].id` 取 `message_id`。如果多个结果，让用户选择。

2. **获取邮件详情:**
   ```bash
   wpsx call kso_mail_get_letter \
     -p '{"message_id": "{MESSAGE_ID}"}' \
     {ENDPOINT} -f json
   ```

3. **整理输出** — 展示主题、发件人、收件人、时间、正文内容。如用户要求，做摘要/翻译/提取待办。

---

## Playbook 4: 搜索邮件

**触发:** 用户说"找一下关于XX的邮件"、"谁发过XX邮件"、"搜索XX相关邮件"

```bash
wpsx call kso_mail_search_letter \
  -p '{"keyword": "{关键词}", "page_size": 20}' \
  {ENDPOINT} -f json
```

可选参数：
- `type`: 搜索范围 — `subject`（主题，默认）/ `sender`（发件人）/ `receiver`（收件人）/ `body`（正文）/ `all`（全部）
- `start_time_str` / `end_time_str`: 限定时间范围
- `filter`: `["unread"]` 或 `["flagged"]`

展示匹配邮件：主题 + 发件人 + 时间 + 正文预览。

---

## Playbook 5: 发送邮件

**触发:** 用户说"帮我发封邮件给XX"、"写邮件"、"发邮件给XX"、"通过邮件发给XX"

**步骤:**

1. **确认邮件信息** — 收件人、主题、正文。如用户未提供完整信息，逐项确认。

2. **创建草稿:**
   ```bash
   wpsx call kso_mail_create_draft \
     -p '{"subject": "{主题}", "to_recipients": [{收件人}], "body": "{正文内容}"}' \
     {ENDPOINT} -f json
   ```

   **收件人格式（二选一）：**
   - 已知邮箱: `{"name": "张三", "email_address": "zhangsan@wps.cn"}`
   - 不知道邮箱: `{"keyword": "张三"}`（系统自动搜索匹配）

   可选：
   - `cc_recipients`: 抄送人（格式同收件人）
   - `bcc_recipients`: 密送人（格式同收件人）

3. **关键字匹配多人时** — 系统会返回候选列表，向用户确认后重新创建草稿。

4. **展示草稿预览** — 将草稿的收件人、主题、正文摘要展示给用户确认。用户确认后再发送。

5. **发送草稿:**
   ```bash
   wpsx call kso_mail_send_draft \
     -p '{"message_id": "{MESSAGE_ID}"}' \
     {ENDPOINT} -f json
   ```

   `message_id` 来自 `create_draft` 的返回值。

**注意:** 必须先 `create_draft` 再 `send_draft`，不能直接发送。收件人禁止使用虚构邮箱地址。邮件发送不可撤回，务必确认后再执行。

---

## Playbook 6: 回复 / 转发邮件

**触发:** 用户说"帮我回复这封邮件"、"转发这封邮件给XX"

**步骤:**

1. **获取原邮件详情** — 执行 Playbook 3 获取 `message_id`、发件人、主题、正文。

2. **组织回复/转发内容:**
   - 回复: 收件人为原发件人，主题加 `Re:` 前缀，正文包含引用原文
   - 转发: 收件人为目标用户，主题加 `Fwd:` 前缀，正文包含原邮件内容

3. **创建草稿并发送** — 执行 Playbook 5 的步骤 2-5。

**注意:** 当前通过手动构造 `Re:` / `Fwd:` 前缀实现回复/转发，邮件线程关联取决于 API 支持程度。如需精确线程关联，确认 API 是否支持 `in_reply_to` 参数。

---

## Playbook 7: 邮件摘要 / 日报素材

**触发:** 用户说"帮我总结今天的邮件"、"这周邮件摘要"、"邮件里有什么重要的事"

**步骤:**

1. **获取时间范围内的邮件:**
   ```bash
   wpsx call kso_mail_list_letter \
     -p '{"page_size": 50, "start_time_str": "{起始时间}", "end_time_str": "{结束时间}"}' \
     {ENDPOINT} -f json
   ```

2. **对重要邮件获取详情** — 对 `body_preview` 不够完整的、有附件的或标记重要的邮件，执行 `get_letter` 获取完整内容。

3. **生成摘要** — 按发件人或主题分组，提炼关键信息、待办事项、需要回复的邮件。
