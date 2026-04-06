# remade

## 说明

本文件用于记录当前仓库在“重做（remade）”阶段已经完成的核心内容、当前状态与后续建议，便于后续开发者快速接手。

## 当前已完成内容

### 1. 项目结构重建

已按系统设计拆分并建立以下目录：

- `hub/`：控制面（配置管理、发布编排）
- `agent/`：数据面（预留目录，待后续实现）
- `sdk/`：Go SDK
- `deployments/`：部署目录占位
- `docs/plan/`：分模块规划文档
- `cmd/hub/`：Hub 启动入口

### 2. Hub 重做实现（第一版）

已实现一个可运行的 Hub 服务（基于内存存储），包含：

- 环境管理 API
- 集群管理 API（含 status）
- 应用管理 API
- 配置管理 API（CRUD + history + diff + preview）
- 发布单 API（create/get/status/approve/publish/rollback）

并提供：

- `hub/server/store.go`：线程安全内存存储
- `hub/server/server_test.go`：基础单元测试覆盖核心流程

### 3. SDK 重做实现（第一版）

`/sdk/configsdk` 已包含以下模块：

- Config：初始化拉取、合并、热更新轮询、回调
- KV：Put/Get/Delete/List/Watch/WatchPrefix
- Discovery：Register/Deregister/Discover/Watch
- Transport：HTTP 请求封装
- options/types：配置与协议结构

此外新增轻量 TOML 解析组件用于受限环境下运行。

## 当前局限（重要）

1. Hub 当前为**内存存储**，重启后数据丢失。
2. Hub 仅实现了第一版 API 行为，尚未接入审批系统、WebSocket 推送、数据库代理层。
3. SDK TOML 解析为轻量实现，仅覆盖常见场景；复杂 TOML 语法建议后续替换为成熟库。
4. Agent 仅完成目录骨架，核心功能（TreeStore、长轮询引擎、Leader 选举）尚未落地。

## 后续建议

1. 优先落地 Agent（TreeStore + Config/KV Handler），打通端到端链路。
2. Hub 存储层从内存切换至 DB Proxy。
3. 增加 Hub↔Agent 推送链路与 ACK 机制。
4. 补齐 SDK 与 Hub/Agent 的集成测试与压测。
5. 逐步引入发布审批状态机和审计日志。

## 快速运行

```bash
# 启动 Hub
go run ./cmd/hub

# 运行测试
go test ./...
```

