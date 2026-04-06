# 模块 01：统一 TreeStore 与数据模型

## 1. 模块目标

构建统一数据抽象，支撑：

- 配置中心（文件级）
- 分布式 KV（路径级）
- 服务发现（可选命名空间）

统一命名空间：

- `/config/*`
- `/kv/*`
- `/svc/*`

## 2. 数据结构设计

```go
type Entry struct {
    Path      string
    Value     []byte
    Version   int64
    Metadata  map[string]any
    CreatedAt time.Time
    UpdatedAt time.Time
    ExpiredAt *time.Time
}
```

### 2.1 Metadata 规范

- Config 文件：`hot`、`encrypted`、`format`
- KV 节点：`ttl`
- 保留扩展字段：`owner`、`tags`、`deprecated`

### 2.2 版本策略

- Config：不使用全局版本，使用文件内容 MD5 检测。
- KV：每个 path 独立版本递增（Redis INCR + DB 记录）。

## 3. 分层存储实现

### 3.1 L1（内存缓存）

- 读优先命中，降低 Redis/DB 压力。
- 缓存结构：Path 索引 + Prefix 索引（Trie/有序 Map 均可）。

### 3.2 L2（Redis）

- 存储 Entry 主体（JSON 或 Hash）。
- TTL 使用 Redis 原生过期。
- 版本号按 path 维度 INCR。
- 变更通知通过 Pub/Sub 发给 Agent 集群。

### 3.3 L3（本地 DB）

- 持久化最终状态（含 metadata、expired_at）。
- 提供冷启动恢复与 Redis 故障兜底。

## 4. 核心接口

```go
type TreeStore interface {
    Get(ctx context.Context, path string) (Entry, error)
    Set(ctx context.Context, path string, value []byte, metadata map[string]any) (int64, error)
    Delete(ctx context.Context, path string) error
    List(ctx context.Context, prefix string) ([]Entry, error)
    ListRecursive(ctx context.Context, prefix string) ([]Entry, error)
    Watch(ctx context.Context, path string, fromVersion int64) (<-chan Event, error)
    WatchPrefix(ctx context.Context, prefix string, fromVersion int64) (<-chan Event, error)
}
```

## 5. 关键流程

1. 写入：API → L1 → L2（INCR+TTL）→ L3 → Pub/Sub。
2. 读取：L1 命中；未命中回源 L2，再回填 L1；L2 异常回源 L3。
3. 冷启动：L3 预热 L1，再从 L2 增量修正。

## 6. 模块任务（Tasks）

- [ ] T01：定义 Entry、Event、Metadata Schema 与校验器。
- [ ] T02：实现内存存储适配器（线程安全 + Prefix 索引）。
- [ ] T03：实现 Redis 适配器（Set/Get/List、TTL、INCR、Pub/Sub）。
- [ ] T04：实现 DB 适配器（CRUD、过期字段、批量预热查询）。
- [ ] T05：实现三层聚合 TreeStore（读写路径、降级策略）。
- [ ] T06：实现 Watch/WatchPrefix 多路复用器。
- [ ] T07：补齐单测（缓存命中、TTL 过期、并发写、顺序性）。
- [ ] T08：补齐压测脚本（QPS、P99、内存占用）。

## 7. 完成定义（DoD）

- 支持 `/config` 与 `/kv` 全部基础 API。
- 支持 TTL、Prefix 查询、单 path Watch 与 Prefix Watch。
- Redis 或 DB 异常时可降级读取，不出现全量不可用。

