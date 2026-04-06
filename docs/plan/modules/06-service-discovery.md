# 模块 06：服务注册/发现（可选能力）

## 1. 模块目标

为非 K8s 原生场景提供服务注册与发现，支持心跳续约与实例变更通知。

## 2. 数据模型

```go
type ServiceInstance struct {
    ID        string
    Name      string
    Address   string
    Port      int
    Metadata  map[string]string
    TTL       int
    UpdatedAt time.Time
}
```

建议存储路径：`/svc/{service}/{instanceID}`。

## 3. 核心流程

1. Register：写入实例信息与 TTL。
2. Heartbeat：刷新 TTL 与更新时间。
3. Discover：按服务名读取健康实例。
4. Watch：订阅服务名前缀变化。

## 4. 与 KV 的复用

- 直接复用 TreeStore Set/Get/List/WatchPrefix。
- 复用长轮询引擎。
- 复用 TTL 过期机制。

## 5. 模块任务（Tasks）

- [ ] D01：定义服务实例模型与 API 协议。
- [ ] D02：实现 Register/Deregister 接口。
- [ ] D03：实现 Heartbeat 接口（续 TTL）。
- [ ] D04：实现 Discover 接口（过滤过期实例）。
- [ ] D05：实现 Watch 接口并输出实例增删事件。
- [ ] D06：实现 SDK Discovery 模块与自动续约器。
- [ ] D07：补齐异常测试（漏心跳、网络分区、重复注册）。

## 6. 完成定义（DoD）

- 独立部署环境可不依赖 K8s 完成服务发现。
- 实例下线在 TTL 窗口后可被正确剔除。
- Watch 能在秒级通知实例变化。

