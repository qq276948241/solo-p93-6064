# 社区家电回收预约系统 (后端)

基于 Go + Gin + GORM + MySQL 实现的社区家电回收预约后端服务。

## 技术栈

- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL
- **认证**: JWT (双 token，居民/管理员分离)
- **密码加密**: bcrypt
- **配置**: godotenv

## 项目结构

```
project93/
├── cmd/server/main.go          # 程序入口
├── internal/
│   ├── config/                 # 配置加载
│   ├── dto/                    # 请求/响应数据结构
│   ├── handler/                # HTTP 处理器
│   ├── model/                  # 数据模型
│   ├── pkg/
│   │   ├── database/           # 数据库连接
│   │   ├── jwt/                # JWT 工具
│   │   ├── middleware/         # 认证中间件
│   │   └── response/           # 统一响应与错误码
│   └── service/                # 业务逻辑层
├── sql/schema.sql              # 数据库初始化脚本
├── .env.example                # 环境变量示例
└── go.mod
```

## 快速开始

### 1. 初始化数据库

```bash
mysql -u root -p < sql/schema.sql
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 修改数据库连接等配置
```

### 3. 运行服务

```bash
go run cmd/server/main.go
```

或者编译运行：
```bash
go build -o server.exe ./cmd/server
./server.exe
```

服务默认运行在 `http://localhost:8080`

## 默认账号

首次启动会自动创建默认管理员账号：
- 用户名: `admin`
- 密码: `admin123`

## 统一响应格式

```json
{
    "code": 0,
    "msg": "success",
    "data": {}
}
```

- `code=0` 表示成功，非 0 表示失败
- HTTP 状态码同步反映错误类型（不全是 200）

## 主要错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40001 | 参数错误 |
| 40002 | 未授权 |
| 40003 | 无权限 |
| 40004 | 资源不存在 |
| 50001 | 服务器内部错误 |
| 50002 | 数据库错误 |
| 10001 | 用户已存在 |
| 10002 | 用户不存在 |
| 10003 | 密码错误 |
| 10004 | Token 无效 |
| 10005 | Token 已过期 |
| 20001 | 时段已满 |
| 20002 | 时段不存在 |
| 20003 | 时段已过期 |
| 20004 | 家电类型无效 |
| 20005 | 预约不存在 |
| 20006 | 预约已完成无法修改 |
| 20007 | 预约已取消无法修改 |

## 接口文档

### 公共接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/appliance-types` | 获取家电类型列表 |
| GET | `/api/public/slots?week_start=YYYY-MM-DD` | 查询排班时段（无需登录） |

### 居民接口

#### 注册 & 登录

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/resident/register` | 居民注册 |
| POST | `/api/resident/login` | 居民登录，返回 token |

**注册请求体**:
```json
{
    "phone": "13800138000",
    "password": "123456"
}
```

**登录请求体**: 同上

**响应**:
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "token": "eyJhbGciOi..."
    }
}
```

#### 预约管理（需要居民 Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/resident/slots?week_start=YYYY-MM-DD` | 查询本周排班时段 |
| POST | `/api/resident/appointments` | 创建预约 |
| GET | `/api/resident/appointments` | 我的预约列表 |
| POST | `/api/resident/appointments/:id/cancel` | 取消预约 |

**创建预约请求体**:
```json
{
    "slot_id": 1,
    "phone": "13800138000",
    "address": "XX小区XX号楼XX单元XX室",
    "appliance_type_id": 1,
    "appliance_weight": 50.5,
    "remark": "冰箱在二楼，需要帮忙搬"
}
```

**排班时段响应**:
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "week_start": "2026-06-22",
        "week_end": "2026-06-28",
        "slots": [
            {
                "id": 1,
                "slot_date": "2026-06-22",
                "start_time": "09:00",
                "end_time": "11:00",
                "capacity": 10,
                "booked_count": 3,
                "available": 7,
                "is_full": false
            }
        ]
    }
}
```

### 管理员接口

#### 登录

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/admin/login` | 管理员登录 |

**请求体**:
```json
{
    "username": "admin",
    "password": "admin123"
}
```

#### 管理接口（需要管理员 Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/appliance-types` | 家电类型列表 |
| GET | `/api/admin/slots?week_start=YYYY-MM-DD` | 查询排班 |
| GET | `/api/admin/appointments` | 预约列表（支持筛选） |
| PUT | `/api/admin/appointments/:id/status` | 修改预约状态 |
| GET | `/api/admin/statistics` | 本月回收统计 |

**预约列表查询参数**:
- `page`: 页码，默认 1
- `page_size`: 每页数量，默认 20
- `start_date`: 创建开始日期 (YYYY-MM-DD)
- `end_date`: 创建结束日期 (YYYY-MM-DD)
- `status`: 状态筛选 (1=待上门, 2=已完成, 3=已取消)

**修改状态请求体**:
```json
{
    "status": 2,
    "remark": "已完成回收"
}
```

**统计响应**:
```json
{
    "code": 0,
    "msg": "success",
    "data": {
        "month": "2026-06",
        "total_weight": 1250.50,
        "total_count": 58,
        "appliance_list": [
            {
                "appliance_type_id": 1,
                "appliance_type": "冰箱",
                "count": 20,
                "total_weight": 600.00
            },
            {
                "appliance_type_id": 2,
                "appliance_type": "洗衣机",
                "count": 15,
                "total_weight": 375.50
            }
        ]
    }
}
```

## Token 使用方式

在需要认证的接口请求头中添加：
```
Authorization: Bearer <token>
```

居民 Token 和管理员 Token 不互通，各自使用独立的密钥签名。

## 预约状态流转

```
待上门 (1) ---> 已完成 (2)
    |
    +-------> 已取消 (3)
```

- 居民可主动取消「待上门」的预约
- 管理员可将「待上门」改为「已完成」或「已取消」
- 已完成/已取消的预约不能再修改状态
- 取消预约时会自动恢复时段容量

## 注意事项

1. 排班系统会自动生成未来 7 天的时段（每天 3 个时间段，默认容量 10）
2. 预约创建使用数据库事务 + 行锁，保证并发安全
3. 所有涉及金额/重量的字段使用 DECIMAL 类型存储
4. 建议在生产环境修改 JWT 密钥和默认管理员密码
