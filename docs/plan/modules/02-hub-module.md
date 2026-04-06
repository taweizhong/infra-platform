# 模块 02：Hub 集群（配置管理与推送控制面）

## 1. 模块目标

Hub 作为控制面，负责：

- 配置文件管理（CRUD + metadata）
- 审批触发与发布编排
- 将已发布配置推送到各环境 Agent Leader

## 2. 逻辑分层

1. API 层：REST + Swagger。
2. 领域层：配置、发布单、审批流程状态机。
3. 集成层：加密服务、飞书审批、Agent WebSocket。
4. 存储层：通过 DB Proxy 访问底层数据库。

## 3. 配置管理实现

### 3.1 数据模型建议

- ConfigFile：`id/scope/name/content/format/metadata/version`。
- ConfigHistory：保存每次变更快照，用于 diff 与回滚。

### 3.2 Scope 约束

- `global`
- `env:{env}`
- `cluster:{cluster}`
- `app:{app}`
- `instance:{instance}`

### 3.3 写入流程

1. API 参数校验。
2. 敏感字段识别（规则引擎/关键字匹配）。
3. 调加密服务获取密文。
4. 写主表 + 历史表（事务）。

## 4. 推送与同步

### 4.1 Agent 连接

- 仅 Agent Leader 建立 WebSocket。
- 连接握手时带：cluster/env/agent_id/version。

### 4.2 发布动作

1. 选择目标 scope 与环境。
2. 生成发布快照。
3. 向目标 Agent Leader 推送增量事件。
4. Agent ACK 后更新发布状态。

## 5. API 任务拆分

- 环境/集群/应用管理 API。
- 配置 CRUD + 历史 + Diff + 预览 API。
- 发布单 API（创建/审批/发布/回滚/状态）。

## 6. 模块任务（Tasks）

- [ ] H01：建立 Hub 项目骨架与统一中间件（日志、鉴权占位、追踪）。
- [ ] H02：实现 ConfigFile/ConfigHistory 数据表与仓储。
- [ ] H03：实现配置 CRUD API 与 Scope 校验。
- [ ] H04：实现配置预览 API（仅做原始聚合展示）。
- [ ] H05：实现敏感字段加密流程与 metadata.encrypted 写入。
- [ ] H06：实现 Agent WebSocket 连接管理器与 ACK 协议。
- [ ] H07：实现发布单模型、状态流转、回滚能力。
- [ ] H08：实现 Swagger 文档与 OpenAPI 自动生成。
- [ ] H09：实现集成测试（配置发布到 Agent 的端到端链路）。

## 7. 完成定义（DoD）

- 可以在 Hub 创建配置并发布到目标 Agent。
- 生产环境多阶段发布状态机可运行（即使先手动驱动）。
- 所有 API 有清晰错误码与审计日志。

