---
name: wps-message
description: "Read, search, and create WPS协作 chat conversations and messages. Use when summarizing chats, checking unread/@me messages, searching chat history, or creating new conversations."
---

# WPS IM Message Operations

通过 `wpsx` CLI 操作 WPS IM 的会话、消息和成员数据。按场景选择对应 Playbook 执行。

## Setup

**Endpoint:** `https://openapi.wps.cn/mcp/v2/kso-message/message`

**认证:** token 由 `wpsx` 自动注入。首次使用请先运行 `wpsx auth --app-id <ID> --app-secret <SECRET>` 完成 WPS 365 OAuth2 授权。

**调用格式:**

```bash
wpsx call <tool> -p '<json>' https://openapi.wps.cn/mcp/v2/kso-message/message -f json
```

**当前时间:** 涉及时间过滤时，使用 `start_time_str` / `end_time_str`，格式为 `yyyy-mm-dd hh:mm:ss`（北京时间）。使用系统当前时间（北京时间 UTC+8）计算"今天"、"本周"等相对时间。

---

## Guardrails

**确认规则:**

- 创建群聊前，展示群名、群主、成员列表让用户确认
- 成员关键字存在重名时，展示候选列表让用户选择

**空结果处理:**

- 搜索消息/会话返回空 → 告知用户无匹配结果，建议调整关键词或时间范围
- 群成员搜索无结果 → 告知用户该群中没有匹配的成员

**分页策略:**

- 用户要求"全部/所有" → 自动翻页直到结束
- 其他情况 → 展示首页结果，告知"还有更多数据，需要继续查看吗？"

**错误处理:**

- API 返回错误 → 将错误信息展示给用户，不要吞掉
- Endpoint 不可达 / 401 未授权 → 提示用户先运行 `wpsx auth` 完成 WPS 365 授权

---

## Playbook 1: 今日消息摘要 / 每日速览

**触发:** 用户说"今天聊了什么"、"帮我看看今天的消息"、"日报素材"、"今天群里有什么动态"

**步骤:**

1. **获取最近会话列表:**
   ```bash
   wpsx call kso_message_get_chat_list \
     -p '{"page_size": 20, "start_time": "{今天日期} 00:00:00"}' \
     {ENDPOINT} -f json
   ```

2. **逐个拉取今天的消息** — 对感兴趣的会话：
   ```bash
   wpsx call kso_message_get_chat_messages \
     -p '{"chat_id": "{CHAT_ID}", "page_size": 50, "start_time_str": "{今天日期} 00:00:00", "order": "asc"}' \
     {ENDPOINT} -f json
   ```

3. **汇总输出** — 按群聊分组，提取关键讨论、决定、待办，忽略表情/闲聊，生成结构化摘要：
   - 每个群: 主题 + 关键结论 + 需要你关注的事项
   - 如果有 @我 的消息，单独高亮

---

## Playbook 2: 查看未读 / 我漏了什么

**触发:** 用户说"我漏了什么消息"、"未读消息有哪些"、"帮我看看没读的"

**步骤:**

1. **获取会话列表**（会返回 `unread_count`）：
   ```bash
   wpsx call kso_message_get_chat_list -p '{"page_size": 20}' {ENDPOINT} -f json
   ```

2. **筛选有未读的会话**（`unread_count > 0`），按未读数排序。

3. **对未读数较多的会话拉取消息:**
   ```bash
   wpsx call kso_message_get_chat_messages \
     -p '{"chat_id": "{CHAT_ID}", "page_size": 50, "filter_unread": "true", "order": "asc"}' \
     {ENDPOINT} -f json
   ```

4. **输出** — "你有 N 个群/人有未读消息" + 每个会话的关键内容摘要。

---

## Playbook 3: 查看 @我 的消息

**触发:** 用户说"谁@我了"、"有没有人找我"、"看看@我的消息"

**步骤:**

1. **获取会话列表** — 检查 `message_notice` 中 `notice_type` 为 `mention` 的会话。

2. **拉取 @我 的消息:**
   ```bash
   wpsx call kso_message_get_chat_messages \
     -p '{"chat_id": "{CHAT_ID}", "page_size": 20, "filter_mention_me": "true", "order": "desc"}' \
     {ENDPOINT} -f json
   ```

3. **输出** — 列出谁在哪个群 @了你，说了什么，是否需要回复。

---

## Playbook 4: 某个群的聊天回顾

**触发:** 用户说"帮我看看XX群聊了什么"、"XX群最近有什么动态"、"回顾一下XX群"

**步骤:**

1. **找到目标群** — 用搜索定位：
   ```bash
   wpsx call kso_message_search_chats \
     -p '{"keyword": "{群名关键词}", "page_size": 5}' \
     {ENDPOINT} -f json
   ```

2. **拉取消息** — 用 `chat_id` 获取历史消息（可加时间范围）：
   ```bash
   wpsx call kso_message_get_chat_messages \
     -p '{"chat_id": "{CHAT_ID}", "page_size": 50, "order": "desc"}' \
     {ENDPOINT} -f json
   ```

3. **输出** — 按时间线整理，提炼讨论主题、关键发言、结论和待办。

---

## Playbook 5: 搜索聊天记录

**触发:** 用户说"谁说过XX"、"找一下关于XX的消息"、"之前讨论XX的记录"

```bash
wpsx call kso_message_search_messages \
  -p '{"keyword": "{关键词}", "page_size": 20, "order": "desc"}' \
  {ENDPOINT} -f json
```

可选过滤：
- `chat_id_list`: 限定在某些群搜索
- `start_time_str` / `end_time_str`: 限定时间范围
- `sender_id_list`: 限定某人发的
- `filter_chat_type_list`: 只搜群聊 `["group"]` 或单聊 `["p2p"]`

展示匹配的消息：发送人 + 群名 + 时间 + 内容。

---

## Playbook 6: 查看群成员

**触发:** 用户说"XX群里有谁"、"群成员列表"、"群里有多少人"

**步骤:**

1. **找到目标群** — 同 Playbook 4 步骤 1。

2. **获取成员列表:**
   ```bash
   wpsx call kso_message_list_chat_members \
     -p '{"chat_id": "{CHAT_ID}", "page_size": 100, "with_total": "true"}' \
     {ENDPOINT} -f json
   ```

3. **输出** — 成员总数 + 成员名单。

---

## Playbook 7: 在群里找人

**触发:** 用户说"XX群里有没有张三"、"帮我在群里找个人"

```bash
wpsx call kso_message_search_chat_members \
  -p '{"chat_id_list": ["{CHAT_ID}"], "keyword": "{人名}", "page_size": 10, "with_member_detail": "true"}' \
  {ENDPOINT} -f json
```

---

## Playbook 8: 某人最近说了什么

**触发:** 用户说"张三最近说了什么"、"看看XX最近的发言"

**步骤:**

1. **找到此人的 ID** — 如果不知道，先在相关群中搜索成员获取 `sender_id`。

2. **搜索此人的消息:**
   ```bash
   wpsx call kso_message_search_messages \
     -p '{"sender_id_list": ["{SENDER_ID}"], "page_size": 20, "order": "desc"}' \
     {ENDPOINT} -f json
   ```

3. **输出** — 按时间列出此人在各群的发言。

---

## Playbook 9: 创建会话 / 建群

**触发:** 用户说"帮我建个群"、"创建一个XX群聊"、"和张三建一个单聊"、"拉个群讨论XX"

**创建单聊 (p2p):**

```bash
wpsx call kso_message_create_chat \
  -p '{"type": "p2p", "member_names": ["{对方姓名}"]}' \
  {ENDPOINT} -f json
```

只需指定对方，系统自动添加当前用户。两人之间重复创建单聊会复用同一个会话。

**创建群聊 (group):**

```bash
wpsx call kso_message_create_chat \
  -p '{"type": "group", "name": "{群名}", "owner_name": "{群主姓名}", "member_names": ["{成员1}", "{成员2}", "..."]}' \
  {ENDPOINT} -f json
```

- `member_names`: 通过姓名/邮箱搜索成员（1-100 人），也可用 `member_ids` 直接传 ID
- `owner_name`: 群主姓名（群聊必填），也可用 `owner_id`
- 如果成员关键字存在重名，会返回候选列表，需用户确认后重新调用
- 相同成员多次创建群聊会产生不同的群

可选参数：
- `is_join_approve`: 进群需审核 (`true`/`false`)
- `is_owner_admin_modify`: 仅群主/管理员可修改群信息
- `is_owner_admin_at_all`: 仅群主/管理员可 @所有人
- `is_enable_nickname`: 群昵称优先展示 (`true`/`false`)

---

## 消息类型说明

返回的消息 `type` 字段常见值：
- `text`: 纯文本，内容在 `content.text.content`
- `rich_text`: 富文本，含 @人、图片等混合内容，在 `content.rich_text.elements` 中解析
- `image`: 图片消息（无法直接展示内容，告知用户"发了一张图片"）

文本中的 @人 格式: `<at id="用户ID">用户名</at>`，`<at id="1">所有人</at>` 表示 @所有人。

## Additional Resources

- 参数详情见 [reference.md](reference.md)

## 已知限制

- **不支持发送消息** — 当前 API 仅支持读取和搜索消息，不支持在会话中发送消息。如用户要求发消息，告知此限制并建议通过 WPS 客户端操作。可通过 wps-mail skill 发送邮件作为替代沟通方式。
