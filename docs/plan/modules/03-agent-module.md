# 模块 03：Agent 集群（数据面与长轮询引擎）

## 1. 模块目标

Agent 提供 SDK 直连的数据面能力：

- 配置读取与热更新长轮询
- KV 读写、列表、Watch
- 可选服务注册/发现

并通过 Leader 与 Hub 同步配置变更。

## 2. 组件拆分

1. Leader Manager（选主与续约）。
2. Hub Sync（WebSocket 消费与本地写入）。
3. TreeStore（L1/L2/L3 聚合）。
4. Long Poll Engine（Config+KV 共用基础设施）。
5. API Handlers（Config/KV/Discovery）。

## 3. 配置读取链路

1. SDK 请求 `/api/v1/config`。
2. Agent 按 scope 拉取文件列表。
3. 返回 `scopes -> files[]` 结构，包含 content + metadata。
4. SDK 侧执行多级合并。

## 4. 配置热更新长轮询

### 4.1 变更依据

- 对 `metadata.hot=true` 文件，比较 SDK 上报 MD5 与当前文件 MD5。
- 若不一致立即返回变更文件。
- 一致则 hold（最长 timeout）。

### 4.2 唤醒机制

- Hub 推送后 Agent Leader 写入 Redis。
- Redis Pub/Sub 通知同集群 Agent。
- 各 Agent 唤醒关联 watcher，重新比对 MD5。

## 5. KV 长轮询

- 请求带 `from_version`。
- 若当前版本更高，立即返回事件。
- 否则 hold，直到超时或收到写入事件。

## 6. 模块任务（Tasks）

- [ ] A01：实现 Leader 选举接口抽象（Lease/Redis Lock 可插拔）。
- [ ] A02：实现 Hub WebSocket 客户端（断线重连、重放保护）。
- [ ] A03：实现 ConfigHandler（全量读取接口）。
- [ ] A04：实现 ConfigWatchHandler（MD5 长轮询协议）。
- [ ] A05：实现 KVHandler（Put/Get/Delete/List）。
- [ ] A06：实现 KVWatchHandler（path/prefix watch）。
- [ ] A07：实现通用 LongPoll 引擎（会话、超时、唤醒、清理）。
- [ ] A08：实现 Redis Pub/Sub 通知桥接。
- [ ] A09：实现冷启动流程（L3 预热 + L2 修正 + Hub 同步）。
- [ ] A10：实现 Agent 侧压测（并发长轮询、KV 写入风暴）。

## 7. 完成定义（DoD）

- 单 Agent 支持配置长轮询与 KV watch 秒级通知。
- 多 Agent 下通过 Pub/Sub 保持可接受的一致性延迟。
- Leader 切换不影响已有 SDK 读请求可用性。

