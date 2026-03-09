# WPS Cloud Document MCP — Tool Reference

## kso_yundoc_search_yundoc

搜索云文档。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| keyword | string | **yes** | 搜索关键词 |
| page_size | int | no | 返回结果数 |
| type | string | no | 文件类型过滤: `otl`（智能文档）、`docx`、`ksheet`、`pdf` |

**Response 关键字段:**

```json
{
  "items": [
    {
      "file": {
        "id": "file_id",
        "name": "文件名.docx",
        "drive_id": "drive_id",
        "link_url": "https://www.kdocs.cn/l/xxx",
        "version": 10
      }
    }
  ]
}
```

---

## kso_yundoc_get_file_meta

获取文件元信息。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| file_id | string | no* | 文件 ID |
| file_url | string | no* | 文件 URL |
| with_drive | boolean | no | 是否返回 drive_id（推荐设为 true） |

*`file_id` 和 `file_url` 至少提供一个。

---

## kso_yundoc_extract_yundoc_content

提取文档内容（返回 Markdown 格式）。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| drive_id | string | **yes** | Drive ID（从文件元信息获取） |
| file_id | string | no* | 文件 ID |
| file_url | string | no* | 文件 URL |

*`file_id` 和 `file_url` 至少提供一个。

**Response 关键字段:**

```json
{
  "dst_format": "markdown",
  "markdown": "# 标题\n内容...",
  "src_format": "docx",
  "version": "42"
}
```

`src_format` 取值: `otl`（智能文档）、`pdf`、`docx`

---

## kso_yundoc_extract_yundoc_comment

提取文档评论。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| drive_id | string | **yes** | Drive ID |
| file_id | string | no* | 文件 ID |
| file_url | string | no* | 文件 URL |

*`file_id` 和 `file_url` 至少提供一个。

**Response 关键字段:**

```json
{
  "doc": {
    "comments": [
      {
        "author": "张三",
        "date": "2025-07-20T10:00:00Z",
        "blocks": [
          {
            "para": {
              "runs": [
                {"text": "评论内容"}
              ]
            }
          }
        ]
      }
    ]
  }
}
```

---

## 权限角色参考

| role_id | 说明 |
|---------|------|
| viewer | 仅查看 |
| commenter | 可评论 |
| editor | 可编辑 |
| sharer | 可分享 |
| manager | 完全控制 |
