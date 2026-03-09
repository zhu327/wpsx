---
name: wps-dbsheet
description: "CRUD operations on WPS smart sheets (数据表/多维表格/dbsheet). Use when querying, inserting, or updating table records and sheet structures."
---

# WPS Smart Sheet (DBSheet) Operations

通过 `wpsx` CLI 操作 WPS 数据表（多维表格/dbsheet），实现表结构查询和数据 CRUD。

## Setup

**Endpoint:** `https://openapi.wps.cn/mcp/v2/kso-dbsheet/message`

**认证:** token 由 `wpsx` 自动注入。首次使用请先运行 `wpsx auth --app-id <ID> --app-secret <SECRET>` 完成 WPS 365 OAuth2 授权。

**调用格式:**

```bash
wpsx call <tool> -p '<json>' https://openapi.wps.cn/mcp/v2/kso-dbsheet/message -f json
```

---

## Guardrails

**确认规则:**

- 创建记录前，展示要录入的字段和值让用户确认
- 更新记录前，展示修改前后的对比让用户确认
- 批量操作前，明确告知影响范围

**空结果处理:**

- keyword 搜索不到文件 → 建议用户检查文件名或直接提供 fileID/fileURL
- 查询记录返回空 → 告知用户该表暂无数据

**分页策略:**

- 用户要求"全部/所有" → 自动翻页直到结束
- 其他情况 → 展示首页结果，告知"还有更多数据，需要继续查看吗？"

**错误处理:**

- API 返回错误 → 将错误信息展示给用户，不要吞掉
- 字段名不匹配 → 检查是否与表结构中的字段名完全一致（先执行 Playbook 1）
- Endpoint 不可达 / 401 未授权 → 提示用户先运行 `wpsx auth` 完成 WPS 365 授权

---

## Playbook 1: 查看表结构 / 找数据表

**触发:** 用户说"看看XX表有哪些字段"、"有没有XX相关的表"、"这个表的结构是什么"、"帮我找个表"

```bash
wpsx call kso_dbsheet_get_file_schema \
  -p '{"keyword": "{关键词}"}' \
  https://openapi.wps.cn/mcp/v2/kso-dbsheet/message -f json
```

也可用 `fileID` 或 `fileURL` 精确定位。返回文件基本信息 + 所有工作表名称/ID/字段结构。

**重要:** 几乎所有后续操作都需要先获取 `fileID`、`sheetID` 和字段名，所以这一步通常是必须的前置操作。

---

## Playbook 2: 查询表数据

**触发:** 用户说"看看表里有什么数据"、"查一下XX记录"、"列一下XX表的内容"、"有多少条数据"

**步骤:**

1. **获取表结构** — 如果还没有 `fileID` 和 `sheetID`，先执行 Playbook 1。

2. **查询记录:**
   ```bash
   wpsx call kso_dbsheet_list_sheet_records \
     -p '{"fileID": "{FILE_ID}", "sheetID": "{SHEET_ID}", "pageSize": 100}' \
     https://openapi.wps.cn/mcp/v2/kso-dbsheet/message -f json
   ```

   可选参数：
   - `fieldNames`: 只返回指定字段，如 `["姓名", "状态"]`
   - `pageSize`: 每页记录数，默认 100
   - `pageToken`: 翻页 token，首次不传，后续从上次返回值获取

3. **整理输出** — 用表格形式展示给用户。如果数据量大，汇总关键信息或按用户要求筛选。

---

## Playbook 3: 新增记录

**触发:** 用户说"帮我加一条记录"、"录入一条数据"、"往表里添加XX"、"记一下XX"

**步骤:**

1. **获取表结构** — 先执行 Playbook 1 获取 `fileID`、`sheetID` 和**字段名列表**（必须用准确的字段名）。

2. **确认要录入的数据** — 根据字段结构，向用户确认每个字段的值。

3. **创建记录:**
   ```bash
   wpsx call kso_dbsheet_create_sheet_records \
     -p '{"fileID": "{FILE_ID}", "sheetID": "{SHEET_ID}", "records": [{"fields_value": "{\"字段1\": \"值1\", \"字段2\": \"值2\"}"}]}' \
     https://openapi.wps.cn/mcp/v2/kso-dbsheet/message -f json
   ```

   **注意:**
   - `fields_value` 是一个 **raw JSON 字符串**（字符串里嵌套 JSON），不是对象
   - 自动类型字段（auto_number、created_by、created_time 等）不能手动填入
   - 支持批量添加多条记录

---

## Playbook 4: 更新记录

**触发:** 用户说"帮我改一下XX的数据"、"把XX的状态改成YY"、"更新一下记录"

**步骤:**

1. **先查询** — 执行 Playbook 2 找到要修改的记录，取得记录的 `id`。

2. **确认修改内容** — 向用户确认要改哪些字段和新值。

3. **更新记录:**
   ```bash
   wpsx call kso_dbsheet_update_sheet_records \
     -p '{"fileID": "{FILE_ID}", "sheetID": "{SHEET_ID}", "records": [{"id": "{RECORD_ID}", "fields_value": "{\"字段1\": \"新值\"}"}]}' \
     https://openapi.wps.cn/mcp/v2/kso-dbsheet/message -f json
   ```

   - 只需传要更新的字段，其他字段保持不变
   - 支持批量更新多条记录

---

## Playbook 5: 创建新工作表

**触发:** 用户说"帮我建一个新表"、"创建一个XX工作表"、"加个 sheet"

**步骤:**

1. **确认所属文件** — 需要 `fileID`，如果用户没给，先用 Playbook 1 找到目标文件。

2. **创建工作表:**
   ```bash
   wpsx call kso_dbsheet_create_sheet \
     -p '{"fileID": "{FILE_ID}", "sheetName": "{表名}", "fieldNames": ["字段1", "字段2", "字段3"]}' \
     https://openapi.wps.cn/mcp/v2/kso-dbsheet/message -f json
   ```

   - `fieldNames` 可选，不传则创建空表

---

## Playbook 6: 重命名工作表

**触发:** 用户说"把XX表改个名"、"重命名工作表"

```bash
wpsx call kso_dbsheet_update_sheet \
  -p '{"fileID": "{FILE_ID}", "sheetID": "{SHEET_ID}", "newSheetName": "{新名称}"}' \
  https://openapi.wps.cn/mcp/v2/kso-dbsheet/message -f json
```

---

## Playbook 7: 从其他来源导入数据到表

**触发:** 用户说"把文档里的表格导入到数据表"、"从XX整理成表格"、"帮我把这些数据录到表里"

**步骤:**

1. **获取源数据** — 可能来自：
   - 云文档内容（联动 wps-yundoc skill 读取）
   - 本地文件（Read 工具读取）
   - 用户直接提供的文本

2. **解析为结构化数据** — 将内容解析为字段名 + 记录列表。

3. **创建表（如需要）** — Playbook 5 创建工作表并定义字段。

4. **批量写入** — Playbook 3 批量创建记录。

---

## fields_value 格式说明

`fields_value` 是创建/更新记录时最关键的参数，它是一个 **JSON 字符串**（不是对象）：

```json
{"fields_value": "{\"任务名称\": \"完成设计稿\", \"负责人\": \"张三\", \"状态\": \"进行中\"}"}
```

- key 必须与表中字段名**完全一致**（建议先用 Playbook 1 获取准确字段名）
- 自动类型字段（auto_number、created_by、created_time 等）不要传入

**常见错误:**

```
✅ 正确: {"fields_value": "{\"任务名称\": \"完成设计稿\", \"负责人\": \"张三\"}"}
❌ 错误: {"fields_value": {"任务名称": "完成设计稿", "负责人": "张三"}}
```

`fields_value` 的值必须是 JSON **字符串**（用引号包裹的 JSON），不是 JSON 对象。

## Additional Resources

- 参数详情见 [reference.md](reference.md)
