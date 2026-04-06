# 基础设施平台实施计划总览

## 1. 文档目的

本文档用于把“配置中心 + 服务注册/发现 + 分布式 KV”统一基础设施服务落地为可执行工程计划，覆盖：

- 架构分层与模块边界
- 各模块的实现方案与接口约定
- 分阶段里程碑与交付清单
- 模块级任务（Task）定义与验收标准

> 说明：本计划优先保障 **可用性、性能、稳定性**，并按“先跑通主链路，再增强一致性/安全/治理能力”的节奏推进。

---

## 2. 实施范围

### 2.1 核心范围（本期必须）

1. Hub：配置管理、发布编排、审批集成、推送能力。
2. Agent：TreeStore（三层存储）、配置读取与热更新长轮询、KV 读写与 Watch。
3. SDK（Go）：Config/KV 核心 API，配置初始化、合并、热更新、回调。
4. 基础运维能力：部署清单、日志、指标、健康检查、故障降级流程。

### 2.2 可选范围（按需启用）

1. 服务注册/发现模块。
2. 本地文件缓存（SDK L0.5）。
3. 多阶段生产发布编排（灰度/生产/稳定/灾备完整链路）。

### 2.3 非目标（本期不做）

1. 完整前端管理台。
2. 强安全体系（mTLS/ACL 细粒度授权）。
3. 跨集群 KV 强一致同步。

---

## 3. 文档结构

- `docs/plan/modules/01-treestore-and-data-model.md`
- `docs/plan/modules/02-hub-module.md`
- `docs/plan/modules/03-agent-module.md`
- `docs/plan/modules/04-sdk-module.md`
- `docs/plan/modules/05-release-and-approval.md`
- `docs/plan/modules/06-service-discovery.md`
- `docs/plan/modules/07-observability-deployment-security.md`
- `docs/plan/modules/08-project-roadmap-and-tasks.md`

---

## 4. 实施策略（建议）

1. **Phase 1（最小可用）**：配置中心主链路 + KV 主链路 + SDK 基础读取。
2. **Phase 2（高可用）**：Agent 集群能力、Redis Pub/Sub 协同、长轮询优化。
3. **Phase 3（治理增强）**：审批发布状态机、版本历史与回滚、可选服务发现。
4. **Phase 4（生产强化）**：性能压测、故障演练、观测告警、SLO 收敛。

---

## 5. 统一验收基线

每个模块任务均需满足以下最小验收条件：

- 有明确 API/接口定义（输入、输出、错误码）。
- 有单元测试或集成测试方案。
- 有监控指标与日志字段约定。
- 有回滚或降级策略说明。
- 文档中有“完成定义（DoD）”。

