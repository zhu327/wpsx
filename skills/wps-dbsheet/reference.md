# WPS DBSheet MCP — Tool Reference

## kso_dbsheet_get_file_schema

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| keyword | string | no* | 文件名关键字搜索 |
| fileID | string | no* | 文件 ID（优先级最高） |
| fileURL | string | no* | 文件 URL |

*三者至少提供一个。优先级: fileID > fileURL > keyword。

---

## kso_dbsheet_list_sheet_records

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| fileID | string | no* | 文件 ID |
| fileURL | string | no* | 文件 URL |
| sheetID | string | no** | 工作表 ID |
| sheetName | string | no** | 工作表名称 |
| fieldNames | array | no | 指定返回的字段名列表 |
| fieldIDs | array | no | 指定返回的字段 ID 列表 |
| pageSize | number | no | 每页记录数，默认 100 |
| pageToken | string | no | 分页 token，首次不传 |

*fileID/fileURL 二选一，优先 fileID。
**sheetID/sheetName 二选一，优先 sheetID。

---

## kso_dbsheet_create_sheet

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| fileID | string | no* | 文件 ID |
| fileURL | string | no* | 文件 URL |
| sheetName | string | **yes** | 工作表名称 |
| fieldNames | array | no | 字段名列表，不传则创建空表 |

---

## kso_dbsheet_create_sheet_records

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| fileID | string | no* | 文件 ID |
| fileURL | string | no* | 文件 URL |
| sheetID | string | no** | 工作表 ID |
| sheetName | string | no** | 工作表名称 |
| records | array | **yes** | 记录列表，每项含 `fields_value` (raw JSON string) |

**records 格式:**
```json
[
  {"fields_value": "{\"字段1\": \"值1\", \"字段2\": \"值2\"}"},
  {"fields_value": "{\"字段1\": \"值3\", \"字段2\": \"值4\"}"}
]
```

不可填写的自动字段: auto_number, created_by, created_time 等。

---

## kso_dbsheet_update_sheet_records

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| fileID | string | no* | 文件 ID |
| fileURL | string | no* | 文件 URL |
| sheetID | string | no** | 工作表 ID |
| sheetName | string | no** | 工作表名称 |
| records | array | **yes** | 记录列表，每项含 `id` 和 `fields_value` |

**records 格式:**
```json
[
  {"id": "recordID1", "fields_value": "{\"字段1\": \"新值1\"}"},
  {"id": "recordID2", "fields_value": "{\"字段1\": \"新值2\"}"}
]
```

只需传要更新的字段。

---

## kso_dbsheet_update_sheet

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| fileID | string | no* | 文件 ID |
| fileURL | string | no* | 文件 URL |
| sheetID | string | no** | 工作表 ID |
| sheetName | string | no** | 工作表名称 |
| newSheetName | string | **yes** | 新的工作表名称 |

---

## 通用说明

- **fileID vs fileURL**: 二选一，优先使用 fileID
- **sheetID vs sheetName**: 二选一，优先使用 sheetID
- **fields_value**: 必须是 raw JSON **字符串**，不是 JSON 对象
