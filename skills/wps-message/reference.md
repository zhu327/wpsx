# WPS Message MCP — Tool Reference

## kso_message_get_chat_list

获取最近会话列表。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| page_size | number | no | 每页数量，默认 20，最大 100 |
| page_token | string | no | 分页 token |
| start_time | string | no | 起始时间，RFC3339 或 `yyyy-mm-dd hh:mm:ss`（北京时间） |
| end_time | string | no | 结束时间，同上 |

**Response 关键字段:**
- `items[].id` — 会话 ID
- `items[].chat.name` — 会话名称
- `items[].chat.type` — `p2p`(单聊) / `group`(群聊)
- `items[].unread_count` — 未读消息数
- `items[].message_notice` — 强提醒信息（@人/待办/会议等）
- `next_page_token` — 翻页 token

**强提醒类型 (notice_type):** mention, todo, meeting, chatplacard(公告), chatvote(投票), p2pfile(快传), autotask, urgent(加急), call(语音), mentionall(@所有人)

---

## kso_message_get_chat_messages

获取指定会话的历史消息。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| chat_id | string | **yes** | 会话 ID |
| page_size | number | **yes** | 每页数量，范围 [1, 50]，默认 20 |
| page_token | string | no | 分页 token |
| start_time | number | no | 起始时间（秒级时间戳） |
| end_time | number | no | 结束时间（秒级时间戳） |
| start_time_str | string | no | 起始时间字符串，优先级低于 start_time |
| end_time_str | string | no | 结束时间字符串，优先级低于 end_time |
| order | string | no | 排序: `asc` / `desc`（默认） |
| filter_unread | string | no | 只返回未读: `"true"` / `"false"` |
| filter_mention_me | string | no | 只返回 @我: `"true"` / `"false"` |

**Response 关键字段:**
- `items[].id` — 消息 ID
- `items[].type` — 消息类型: text / rich_text / image 等
- `items[].content` — 消息内容（结构随 type 不同）
- `items[].ctime` — 创建时间（毫秒时间戳）
- `items[].sender.name` — 发送人姓名
- `items[].mentions` — @人员列表

---

## kso_message_search_chats

按关键字搜索会话。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| keyword | string | **yes** | 搜索关键字 |
| page_size | number | **yes** | 每页数量，范围 [1, 50] |
| page_token | string | no | 分页 token |
| filter_chat_type_list | array | no | 会话类型过滤: `["p2p"]`, `["group"]` |
| with_total | string | no | 是否返回总数 |
| with_group_ext_attrs | string | no | 是否返回群扩展信息 |

---

## kso_message_search_messages

按多维条件搜索消息。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| page_size | number | **yes** | 每页数量，范围 [1, 50] |
| page_token | string | no | 分页 token |
| keyword | string | no* | 搜索关键字 |
| chat_id_list | array | no* | 会话 ID 列表（最多 50） |
| sender_id_list | array | no* | 发送者 ID 列表（最多 50） |
| start_time | number | no* | 起始时间（秒级时间戳） |
| end_time | number | no* | 结束时间（秒级时间戳） |
| start_time_str | string | no | 起始时间字符串 |
| end_time_str | string | no | 结束时间字符串 |
| filter_chat_type_list | array | no | 会话类型过滤 |
| msg_type_list | array | no | 消息类型过滤 |
| filter_msg_tag_list | array | no | 消息标签过滤 |
| order | string | no | 排序: `asc` / `desc` |
| filter_unread | string | no | 只返回未读 |
| with_chat | string | no | 展开会话信息 |

*keyword、chat_id_list、sender_id_list、时间范围四者至少提供一个。

---

## kso_message_list_chat_members

获取会话成员列表。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| chat_id | string | **yes** | 会话 ID |
| page_size | number | no | 每页数量，默认 20，最大 100 |
| page_token | string | no | 分页 token |
| type | string | no | 成员类型 |
| with_total | string | no | 返回总数 |
| with_group_ext_attrs | string | no | 群成员扩展字段 |
| with_ext_attrs | string | no | 自定义扩展字段 |

---

## kso_message_search_chat_members

在会话中搜索成员。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| chat_id_list | array | **yes** | 会话 ID 列表（最多 50） |
| page_size | number | **yes** | 每页数量，范围 [1, 50] |
| page_token | string | no | 分页 token |
| keyword | string | no | 搜索关键字 |
| with_member_detail | string | no | 返回成员详情 |

---

## kso_message_create_chat

创建会话（一对一或群聊）。

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| type | string | **yes** | `p2p`(单聊) / `group`(群聊) |
| member_names | array | no* | 成员姓名列表（搜索匹配） |
| member_ids | array | no* | 成员 ID 列表 |
| owner_name | string | no** | 群主姓名（群聊必填） |
| owner_id | string | no** | 群主 ID（群聊必填） |
| name | string | no | 群聊名称 |
| avatar | string | no | 群头像 |
| is_join_approve | string | no | 进群需审核 |
| is_owner_admin_modify | string | no | 仅群主管理员可改群信息 |
| is_owner_admin_at_all | string | no | 仅群主管理员可@所有人 |
| is_enable_nickname | string | no | 群昵称优先展示 |

*member_names/member_ids 二选一。
**群聊时 owner_name/owner_id 二选一必填。
