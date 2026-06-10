---
name: wps-calendar
description: "Manage WPS calendar events (日历/日程) via wpsx. Use for schedules, creating/updating events, free-busy, attendees, and finding time slots. For online meeting rooms, transcripts, or minutes, use wps-meeting instead."
---

# WPS Calendar (日历) Operations

通过 `wpsx` CLI 操作 WPS 日历，管理日程、查询忙闲、安排会议。按场景选择对应 Playbook 执行。

## Setup

**Endpoint:** `https://openapi.wps.cn/mcp/v2/kso-calendar/message`

Examples below use `{ENDPOINT}` for this URL. In shell snippets, either replace `{ENDPOINT}` with the full URL or set `ENDPOINT="https://openapi.wps.cn/mcp/v2/kso-calendar/message"` and use `$ENDPOINT`.

**认证:** token 由 `wpsx` 自动注入。首次使用请先运行 `wpsx auth --app-id <ID> --app-secret <SECRET>` 完成 WPS 365 OAuth2 授权。

**调用格式:**

```bash
wpsx call <tool> -p '<json>' https://openapi.wps.cn/mcp/v2/kso-calendar/message -f json
```

**时间格式:**

- **字符串格式:** `yyyy-mm-dd hh:mm:ss`（默认北京时间），用于 `list_calendar_events` 和 `list_free_busy`
- **字典格式:** `{"date": "yyyy-mm-dd"}`（全天）或 `{"datetime": "yyyy-mm-dd hh:mm:ss"}`（非全天），用于 `create/update` 和事件定位

---

## Guardrails

**当前时间:** 涉及"今天"、"本周"、"明天"等相对时间时，使用系统当前时间（北京时间 UTC+8）计算精确时间范围。"本周"从周一 00:00:00 到周日 23:59:59。

**确认规则:**

- 创建/更新日程前，先展示操作摘要（标题、时间、参与者）让用户确认
- 安排多人会议时（Playbook 4），找到空闲时间后必须让用户确认再创建

**空结果处理:**

- 查询日程返回空 → 告知用户该时间段无日程
- 忙闲查询返回空 → 告知用户该时段无忙碌记录（可能是查询范围问题）
- keyword 匹配多个日程 → 展示列表让用户选择
- keyword 匹配人名失败 → 提示用户提供更精确的姓名或邮箱

**分页策略:**

- 用户要求"全部/所有" → 自动翻页直到结束
- 其他情况 → 展示首页结果，告知"还有更多数据，需要继续查看吗？"

**错误处理:**

- API 返回错误 → 将错误信息展示给用户，不要吞掉
- Endpoint 不可达 / 401 未授权 → 提示用户先运行 `wpsx auth` 完成 WPS 365 授权

---

## Playbook 1: 查询日程

**触发:** 用户说"我今天有什么日程"、"本周日程"、"看看我的安排"、"这周有什么会"

```bash
wpsx call kso_calendar_list_calendar_events \
  -p '{"start_time": "{起始时间}", "end_time": "{结束时间}"}' \
  {ENDPOINT} -f json
```

- 时间跨度不超过 31 天
- 默认分页 20 条，最大 100，用 `page_token` 翻页
- 周期性会议不要过滤

整理输出：按日期分组，展示日程标题、时间、地点、参与者。

---

## Playbook 2: 查询忙闲

**触发:** 用户说"张三本周什么时候有空"、"看看XX的忙闲"、"找一个大家都空闲的时间"

```bash
wpsx call kso_calendar_list_free_busy \
  -p '{"start_time": "{起始时间}", "end_time": "{结束时间}", "user_keywords": ["{用户名}"]}' \
  {ENDPOINT} -f json
```

**关键约束:**

- **时间跨度不超过 7 天**，超过需拆分多次查询（如 10 天 → 前 7 天 + 后 3 天）
- `end_time` 计算公式: `start_time + N天 - 1秒`（即第 N 天的 23:59:59）
- `user_ids` / `user_keywords` / `room_ids` 至少提供一个
- 用 `user_keywords`（用户名/邮箱）查询比 `user_ids` 更方便

输出：列出忙碌时段，反向推导空闲时段（企业工作时间 9:00-22:00）。

---

## Playbook 3: 创建日程

**触发:** 用户说"帮我建个会议"、"安排一个日程"、"创建一个XX日程"

**非全天日程:**

```bash
wpsx call kso_calendar_create_calendar_event \
  -p '{"summary": "{标题}", "start_time": {"datetime": "{开始时间}"}, "end_time": {"datetime": "{结束时间}"}, "description": "{描述}", "locations": [{"name": "{地点}"}]}' \
  {ENDPOINT} -f json
```

**全天日程:**

```bash
wpsx call kso_calendar_create_calendar_event \
  -p '{"summary": "{标题}", "start_time": {"date": "{日期}"}, "end_time": {"date": "{日期}"}}' \
  {ENDPOINT} -f json
```

可选参数：

- `duration_minutes`: 时长（分钟），不传 end_time 时自动计算
- `reminders`: 提醒，如 `[{"minutes": 15}]`
- `visibility`: `default` / `public` / `private`

时间优先级：仅 start_time → +2h；+duration → 按时长；+end_time → 以 end_time 为准。

---

## Playbook 4: 安排多人会议

**触发:** 用户说"约XX和YY开会"、"帮我安排一个XX会议"、"找个时间大家一起开会"

**步骤:**

1. **确定会议信息** — 主题、时间范围、会议时长、参与者列表。

2. **查询所有人的忙闲** — 执行 Playbook 2，查询组织者和所有参与者的忙闲：
   ```bash
   wpsx call kso_calendar_list_free_busy \
     -p '{"start_time": "{起始}", "end_time": "{结束}", "user_keywords": ["{用户1}", "{用户2}", "..."]}' \
     {ENDPOINT} -f json
   ```

3. **找到共同空闲时间** — 比对所有人的 busy_times，推荐合适的时间段。如果没有合适时间，向用户解释原因。

4. **用户确认后创建日程** — 执行 Playbook 3 创建日程。

5. **添加参与者:**
   ```bash
   wpsx call kso_calendar_batch_create_event_attendees \
     -p '{"event_id": "{EVENT_ID}", "attendees": [{"type": "user", "user_keyword": "{用户1}"}, {"type": "user", "user_keyword": "{用户2}"}]}' \
     {ENDPOINT} -f json
   ```
   通过 `user_keyword` 添加比 `user_id` 更方便。

---

## Playbook 5: 添加参与者

**触发:** 用户说"把XX加到会议里"、"添加参会人"、"邀请XX参加"

**已知 event_id:**

```bash
wpsx call kso_calendar_batch_create_event_attendees \
  -p '{"event_id": "{EVENT_ID}", "attendees": [{"type": "user", "user_keyword": "{用户名}"}]}' \
  {ENDPOINT} -f json
```

**未知 event_id（用关键字搜索）:**

```bash
wpsx call kso_calendar_batch_create_event_attendees \
  -p '{"keyword": "{日程标题关键字}", "attendees": [{"type": "user", "user_keyword": "{用户名}"}]}' \
  {ENDPOINT} -f json
```

- 如果 keyword 匹配到多个日程，将结果展示给用户确认后，用 `event_id` 重新调用
- `is_notification` 默认 true，设为 false 可静默添加

---

## Playbook 6: 查看日程参与者

**触发:** 用户说"这个会议有哪些人"、"参会人有谁"、"看看日程的参与者"

```bash
wpsx call kso_calendar_list_event_attendees \
  -p '{"event_id": "{EVENT_ID}"}' \
  {ENDPOINT} -f json
```

也支持 `keyword` 搜索日程。返回参与者的 name、type（user/group）和 response_status（accepted/declined/tentative/not_responded）。

---

## Playbook 7: 更新日程

**触发:** 用户说"改一下会议时间"、"更新日程标题"、"把会议推迟到XX"

**已知 event_id:**

```bash
wpsx call kso_calendar_update_calendar_event \
  -p '{"event_id": "{EVENT_ID}", "summary": "{新标题}", "new_start_time": {"datetime": "{新开始时间}"}, "new_end_time": {"datetime": "{新结束时间}"}}' \
  {ENDPOINT} -f json
```

**未知 event_id:**

```bash
wpsx call kso_calendar_update_calendar_event \
  -p '{"keyword": "{日程标题关键字}", "summary": "{新标题}"}' \
  {ENDPOINT} -f json
```

- keyword 匹配多个日程时，展示给用户确认后用 event_id 重新调用
- `need_notification` 默认 true，设为 false 不通知参与者
- 只传需要修改的字段，其他保持不变

---

## Playbook 8: 工作总结 / 周报

**触发:** 用户说"总结我本周工作"、"上周我做了什么"、"生成周报"

**步骤:**

1. **查询时间范围内的日程:**
   ```bash
   wpsx call kso_calendar_list_calendar_events \
     -p '{"start_time": "{起始}", "end_time": "{结束}"}' \
     {ENDPOINT} -f json
   ```

2. **整理日程数据** — 按日期分组，提取会议主题、时间、参与人。

3. **生成周报** — 结合日程信息归纳本周工作内容。如需更完整的信息，可联动 wps-message skill 获取聊天记录。

---

## 事件定位策略

操作日程（更新、添加参与者、查看参与者）时，统一遵循以下定位逻辑：

1. 已知 `event_id` → 直接传入
2. 未知 `event_id` → 用 `keyword` 搜索（可配合 `start_time` / `end_time` 缩小范围）
3. keyword 匹配多个结果 → 展示给用户确认 → 用 `event_id` 重新调用

---

## 已知限制

- **不支持删除日程** — 当前 API 不提供删除日程的能力。如用户要求删除/取消日程，告知此限制并建议通过 WPS 客户端操作。
