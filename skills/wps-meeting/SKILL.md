---
name: wps-meeting
description: "Manage WPS online meetings (在线会议), participants, transcripts and minute summaries. Use when creating, querying, or managing meetings, viewing meeting transcripts or summaries."
---

# WPS Meeting (在线会议) Operations

通过 `wpsx` CLI 管理 WPS 在线会议，支持创建/修改/取消会议、管理参与者、获取会议转写和纪要总结。按场景选择对应 Playbook 执行。

## Setup

**Endpoint:** `https://openapi.wps.cn/mcp/v2/kso-meeting/message`

**认证:** token 由 `wpsx` 自动注入。首次使用请先运行 `wpsx auth --app-id <ID> --app-secret <SECRET>` 完成 WPS 365 OAuth2 授权。

**调用格式:**

```bash
wpsx call <tool> -p '<json>' https://openapi.wps.cn/mcp/v2/kso-meeting/message -f json
```

**时间格式:** 支持 `yyyy-mm-dd hh:mm:ss`（默认北京时间）或 Unix 时间戳（秒）。字符串用 `start_date_time_str` / `end_date_time_str`，时间戳用 `start_date_time` / `end_date_time`。

---

## Guardrails

**当前时间:** 涉及"今天"、"本周"等相对时间时，使用系统当前时间（北京时间 UTC+8）计算精确时间范围。

**确认规则:**

- 创建会议前，展示会议主题、时间、参与者摘要让用户确认
- 取消会议（不可逆）前，必须明确获得用户同意
- 修改会议时间前，展示变更内容让用户确认

**空结果处理:**

- 查询会议返回空 → 告知用户该时间段无会议
- 转写/纪要不存在 → 告知用户该会议可能未开启录制（详见 Playbook 8）
- keyword 匹配人名失败 → 提示用户提供更精确的姓名或邮箱

**分页策略:**

- 用户要求"全部/所有" → 自动翻页直到结束
- 其他情况 → 展示首页结果，告知"还有更多数据，需要继续查看吗？"

**错误处理:**

- API 返回错误 → 将错误信息展示给用户，不要吞掉
- Endpoint 不可达 / 401 未授权 → 提示用户先运行 `wpsx auth` 完成 WPS 365 授权

---

## Playbook 1: 查询会议列表

**触发:** 用户说"我最近有什么会议"、"本周的会议"、"看看今天的会议"

```bash
wpsx call kso_meeting_list_meetings \
  -p '{"start_date_time_str": "{起始时间}", "end_date_time_str": "{结束时间}"}' \
  {ENDPOINT} -f json
```

开始和结束时间都必填。整理输出：按时间排序，展示会议主题、时间、状态（进行中/已结束/待开始）、入会码。

---

## Playbook 2: 查看会议详情

**触发:** 用户说"看看XX会议的信息"、"会议详情"、"入会码是什么"

```bash
wpsx call kso_meeting_get_meeting \
  -p '{"meeting_id": "{MEETING_ID}"}' \
  {ENDPOINT} -f json
```

如不知道 `meeting_id`，先用 Playbook 1 查询列表获取。

---

## Playbook 3: 创建会议

**触发:** 用户说"帮我创建一个会议"、"发起一个XX会议"、"预约在线会议"

**基本创建:**

```bash
wpsx call kso_meeting_create_meeting \
  -p '{"subject": "{会议主题}", "start_date_time_str": "{开始时间}", "end_date_time_str": "{结束时间}"}' \
  {ENDPOINT} -f json
```

**带参与者创建:**

```bash
wpsx call kso_meeting_create_meeting \
  -p '{"subject": "{会议主题}", "start_date_time_str": "{开始时间}", "end_date_time_str": "{结束时间}", "participants_keywords": ["{用户1}", "{用户2}"], "participants_roles": ["attendee", "attendee"]}' \
  {ENDPOINT} -f json
```

参与者指定方式（二选一，不能同时用）：
- `participants_keywords`: 用户名/邮箱关键字列表
- `participants_ids`: 用户 ID 列表

**`participants_roles` 必须与参与者列表一一对应**，可选值：`attendee`（参会人）/ `organizer`（组织者）。

**注意:** `subject` 是必填参数。`start_date_time`/`start_date_time_str` 和 `end_date_time`/`end_date_time_str` 是可选参数，但创建实际会议时通常需要提供开始和结束时间。

---

## Playbook 4: 修改会议

**触发:** 用户说"改一下会议时间"、"修改会议主题"、"会议推迟到XX"

```bash
wpsx call kso_meeting_update_meeting \
  -p '{"meeting_id": "{MEETING_ID}", "subject": "{新主题}", "start_date_time_str": "{新开始时间}", "end_date_time_str": "{新结束时间}"}' \
  {ENDPOINT} -f json
```

只传需要修改的字段。如不知道 `meeting_id`，先用 Playbook 1 查询。

---

## Playbook 5: 取消会议

**触发:** 用户说"取消XX会议"、"把会议删了"

```bash
wpsx call kso_meeting_delete_meeting \
  -p '{"meeting_id": "{MEETING_ID}"}' \
  {ENDPOINT} -f json
```

操作不可逆，执行前向用户确认。

---

## Playbook 6: 查看会议参与者

**触发:** 用户说"会议有哪些人参加"、"参会人名单"、"谁参加了XX会议"

```bash
wpsx call kso_meeting_list_meeting_participants \
  -p '{"meeting_id": "{MEETING_ID}"}' \
  {ENDPOINT} -f json
```

---

## Playbook 7: 添加 / 移除会议参与者

**触发:** 用户说"把XX加到会议里"、"邀请XX参加会议"、"把XX从会议中移除"

**添加参与者:**

```bash
wpsx call kso_meeting_add_meeting_participants \
  -p '{"meeting_id": "{MEETING_ID}", "participants_keywords": ["{用户名}"], "participants_roles": ["attendee"]}' \
  {ENDPOINT} -f json
```

`participants_keywords` 和 `participants_ids` 二选一，`participants_roles` 必须与之一一对应。

**移除参与者:**

```bash
wpsx call kso_meeting_delete_meeting_participants \
  -p '{"meeting_id": "{MEETING_ID}", "participants_keywords": ["{用户名}"]}' \
  {ENDPOINT} -f json
```

`participants_keywords` 和 `participants_ids` 二选一。

---

## Playbook 8: 获取会议纪要 / 总结

**触发:** 用户说"会议讲了什么"、"看看会议纪要"、"会议总结"、"帮我看看XX会议的要点"

**步骤:**

1. **查询会议列表获取 minute_id:**
   ```bash
   wpsx call kso_meeting_list_meetings \
     -p '{"start_date_time_str": "{起始}", "end_date_time_str": "{结束}"}' \
     {ENDPOINT} -f json
   ```
   找到 `type: "ended"` 的会议，从其 `transcripts[].id` 取得 `minute_id`（即 `transcript_id`，两者等价）。

   **如果 `transcripts` 为空或不存在** → 该会议没有转写记录（可能未开启录制），告知用户并跳过后续步骤。

2. **获取纪要总结:**
   ```bash
   wpsx call kso_meeting_get_minute_summary \
     -p '{"meeting_id": "{MEETING_ID}", "minute_id": "{MINUTE_ID}"}' \
     {ENDPOINT} -f json
   ```

整理输出：会议主题 + 纪要要点 + 行动事项。

---

## Playbook 9: 获取会议转写内容

**触发:** 用户说"看看会议的完整记录"、"会议转写"、"谁在会上说了什么"

**步骤:**

1. **获取 transcript_id** — 同 Playbook 8 步骤 1。如果 `transcripts` 为空，告知用户该会议没有转写记录。

2. **获取转写内容:**
   ```bash
   wpsx call kso_meeting_get_transcript_content_json \
     -p '{"meeting_id": "{MEETING_ID}", "transcript_id": "{TRANSCRIPT_ID}"}' \
     {ENDPOINT} -f json
   ```

返回结构化 JSON，包含发言人、时间轴、文本。整理输出：按发言人和时间线展示完整会议记录。

---

## Playbook 10: 会议工作总结

**触发:** 用户说"总结我本周的会议"、"会议周报"

**步骤:**

1. **查询时间范围内的会议** — Playbook 1。
2. **对已结束会议批量获取纪要** — 对每个 `type: "ended"` 且有 `transcripts` 的会议，执行 Playbook 8 步骤 2。
3. **生成总结** — 汇总所有会议的主题、参与者、关键结论和行动事项。可联动 wps-calendar、wps-todo skill 生成更完整的周报。
