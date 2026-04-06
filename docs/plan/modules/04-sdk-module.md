# 模块 04：Go SDK（Config / KV / Discovery）

## 1. 模块目标

提供业务应用“一次初始化、内存快速读取、自动热更新”的接入能力。

## 2. SDK 总体结构

- `client`：全局单例生命周期管理。
- `config`：配置拉取、合并、回调、解密。
- `kv`：KV CRUD + Watch。
- `discovery`：可选服务发现。
- `transport`：HTTP 与长轮询基础组件。
- `cache`：内存缓存 + 可选文件缓存。

## 3. Config 实现

### 3.1 初始化

1. 拉取全量 scope 文件。
2. 按文件名归组，按 scope 优先级深度合并 TOML。
3. 反序列化到目标结构体。
4. 若失败：返回错误（调用方应退出并告警）。

### 3.2 热更新

1. 筛选 `hot=true` 文件。
2. 计算合并后文件 MD5。
3. 发起长轮询，获取变化文件。
4. 增量重算受影响文件，并触发回调。

### 3.3 解密

- 若文件 metadata.encrypted=true，调用外部 decryptor。
- 解密失败遵循“保守失败”：不更新运行态配置并上报告警。

## 4. KV 实现

- Put/Get/Delete/List 直连 Agent。
- Watch/WatchPrefix 使用长轮询并自动续接。
- 对瞬时网络错误进行退避重试。

## 5. 模块任务（Tasks）

- [ ] S01：定义 SDK 公共配置项与默认值（超时、重试、回调并发）。
- [ ] S02：实现 Config 全量拉取与 TOML 深度合并。
- [ ] S03：实现 Config.Unmarshal / UnmarshalFile。
- [ ] S04：实现 Config 热更新轮询器与 MD5 管理。
- [ ] S05：实现 OnFileChange / OnAnyChange 回调总线。
- [ ] S06：实现 Decryptor 接口与失败策略。
- [ ] S07：实现 KV 客户端（CRUD + Watch + WatchPrefix）。
- [ ] S08：实现本地文件缓存（可选开关）。
- [ ] S09：实现 SDK 指标与日志埋点（重试次数、轮询延迟等）。
- [ ] S10：提供最小接入样例与迁移指南。

## 6. 完成定义（DoD）

- SDK 初始化可稳定拉起，不依赖业务侧手工拼装配置。
- Config Get/Unmarshal 为纯内存路径，无远程调用。
- 热更新仅针对 hot 文件触发，回调去重且顺序可控。

