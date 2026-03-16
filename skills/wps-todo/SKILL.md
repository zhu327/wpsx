---
name: wps-todo
description: "Manage WPS personal tasks (个人待办/todo). Use when creating, querying, updating, or completing todo tasks."
---

# WPS Todo (个人待办) Operations

通过 `wpsx` CLI 管理 WPS 个人待办任务，支持创建、查询、更新、完成待办。按场景选择对应 Playbook 执行。

## Setup

**Endpoint:** `https://openapi.wps.cn/mcp/v2/kso-todo/message`

**认证:** token 由 `wpsx` 自动注入。首次使用请先运行 `wpsx auth --app-id <ID> --app-secret <SECRET>` 完成 WPS 365 OAuth2 授权。

**调用格式:**

```bash
wpsx call <tool> -p '<json>' https://openapi.wps.cn/mcp/v2/kso-todo/message -f json
```

**时间格式:** 支持 RFC3339 (`2024-01-15T00:00:00+08:00`)、简单格式 (`yyyy-mm-dd hh:mm:ss`，默认北京时间) 或仅日期 (`yyyy-mm-dd`)。

---

## Guardrails

**当前时间:** 涉及"今天"、"本周"、"明天"等相对时间时，使用系统当前时间（北京时间 UTC+8）计算精确时间范围。

**确认规则:**

- 创建待办前，展示标题、截止时间等摘要让用户确认
- 标记完成（`finish_all`）前，先确认目标待办是正确的
- 更新待办前，展示修改内容让用户确认

**空结果处理:**

- 查询返回空 → 告知用户无匹配的待办
- keyword 搜索无结果 → 建议放宽关键词或检查状态过滤（todo/finish）

**分页策略:**

- 用户要求"全部/所有" → 自动翻页直到结束
- 其他情况 → 展示首页结果，告知"还有更多数据，需要继续查看吗？"

**错误处理:**

- API 返回错误 → 将错误信息展示给用户，不要吞掉
- Endpoint 不可达 / 401 未授权 → 提示用户先运行 `wpsx auth` 完成 WPS 365 授权

---

## Playbook 1: 查看待办列表

**触发:** 用户说"我有什么待办"、"看看未完成的任务"、"今天有什么要做的"、"待办清单"

**查看未完成待办:**

```bash
wpsx call kso_todo_get_personal_tasks_list \
  -p '{"status": "todo", "page_size": 20}' \
  {ENDPOINT} -f json
```

**查看已完成待办:**

```bash
wpsx call kso_todo_get_personal_tasks_list \
  -p '{"status": "finish", "page_size": 20}' \
  {ENDPOINT} -f json
```

可选过滤：
- `keyword`: 按标题/描述搜索（最大 50 字符）
- `created_time`: 按创建时间范围 `{"start_time": "...", "end_time": "..."}`
- `due_time`: 按截止时间范围 `{"start_time": "...", "end_time": "..."}`
- `page_token`: 翻页（使用上次返回的 `next_offset`）

整理输出：按截止时间排序，展示标题、状态、截止时间、是否逾期。

---

## Playbook 2: 查看待办详情

**触发:** 用户说"看看这个待办的详情"、"XX任务的具体内容"

```bash
wpsx call kso_todo_get_personal_tasks \
  -p '{"task_id": "{TASK_ID}"}' \
  {ENDPOINT} -f json
```

展示：标题、描述、状态（`all_finished_status` / `my_finished_status`）、创建时间、截止时间、创建者、执行人、提醒配置。

---

## Playbook 3: 创建待办

**触发:** 用户说"帮我建个待办"、"记一下XX"、"添加一个任务"、"提醒我XX"

**基本创建:**

```bash
wpsx call kso_todo_create_personal_tasks \
  -p '{"biz_source": "chatbox", "title": {"subject": "{任务标题}"}}' \
  {ENDPOINT} -f json
```

**带截止时间（相对天数，推荐）:**

```bash
wpsx call kso_todo_create_personal_tasks \
  -p '{"biz_source": "chatbox", "title": {"subject": "{任务标题}"}, "due_in_days": {N}}' \
  {ENDPOINT} -f json
```

`due_in_days` 对照：`0`=今天、`1`=明天、`3`=3天后、`7`=下周

**带截止时间（精确时间）:**

```bash
wpsx call kso_todo_create_personal_tasks \
  -p '{"biz_source": "chatbox", "title": {"subject": "{任务标题}"}, "due_time": "{yyyy-mm-dd hh:mm:ss}"}' \
  {ENDPOINT} -f json
```

**带提醒:**

```bash
wpsx call kso_todo_create_personal_tasks \
  -p '{"biz_source": "chatbox", "title": {"subject": "{任务标题}"}, "due_in_days": 1, "notify_config": {"switch": true, "reminders": [{"before_due_time": 30}]}}' \
  {ENDPOINT} -f json
```

可选参数：
- `description`: 详细描述
- `title.prefix`: 标题前缀
- `executors`: 执行人 ID 列表（不传默认当前用户）
- `status`: `todo`（默认）/ `finish`
- `ext_attrs`: 扩展字段，每个元素包含 name 和 value
- `notify_config`: 提醒设置，包含 switch（推送开关）和 reminders（提醒列表）

**截止时间优先级:** `due_in_days` > `due_time`

**biz_source 说明:** 固定传 `"chatbox"` 即可。此字段标识待办来源，当前场景统一使用 chatbox。

---

## Playbook 4: 完成待办

**触发:** 用户说"完成XX任务"、"标记XX为已完成"、"这个待办做完了"

**步骤:**

1. **定位待办** — 如果不知道 `task_id`，先用 Playbook 1 搜索：
   ```bash
   wpsx call kso_todo_get_personal_tasks_list \
     -p '{"status": "todo", "keyword": "{关键词}"}' \
     {ENDPOINT} -f json
   ```

2. **标记完成:**
   ```bash
   wpsx call kso_todo_update_personal_tasks \
     -p '{"task_id": "{TASK_ID}", "action": "finish_all"}' \
     {ENDPOINT} -f json
   ```

**action 说明:**
- `finish_all`: 整个待办标记完成（推荐，大多数场景使用）
- `unfinish_all`: 整个待办标记未完成
- `finish`: 仅标记当前用户完成（多人协作场景）
- `unfinish`: 仅标记当前用户未完成

---

## Playbook 5: 更新待办

**触发:** 用户说"改一下XX任务的截止时间"、"更新待办标题"、"修改任务描述"、"设置提醒"

**步骤:**

1. **定位待办** — 同 Playbook 4 步骤 1。

2. **更新信息:**
   ```bash
   wpsx call kso_todo_update_personal_tasks \
     -p '{"task_id": "{TASK_ID}", "subject": "{新标题}", "due_time": "{新截止时间}"}' \
     {ENDPOINT} -f json
   ```

只传需要修改的字段。可更新字段：
- `subject`: 标题
- `description`: 描述
- `due_time`: 截止时间
- `notify_config`: 提醒设置

也可同时更新状态和信息：

```bash
wpsx call kso_todo_update_personal_tasks \
  -p '{"task_id": "{TASK_ID}", "action": "finish_all", "subject": "已完成的任务"}' \
  {ENDPOINT} -f json
```

---

## Playbook 6: 待办总结 / 日报素材

**触发:** 用户说"总结一下我的待办"、"今天完成了什么"、"本周任务完成情况"

**步骤:**

1. **查询已完成的待办:**
   ```bash
   wpsx call kso_todo_get_personal_tasks_list \
     -p '{"status": "finish", "due_time": {"start_time": "{起始}", "end_time": "{结束}"}, "page_size": 50}' \
     {ENDPOINT} -f json
   ```

2. **查询未完成的待办:**
   ```bash
   wpsx call kso_todo_get_personal_tasks_list \
     -p '{"status": "todo", "page_size": 50}' \
     {ENDPOINT} -f json
   ```

3. **生成总结** — 已完成任务列表 + 未完成/逾期任务列表 + 完成率。可联动 wps-calendar 和 wps-message skill 获取更完整的工作信息。
