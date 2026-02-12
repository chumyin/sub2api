# Sub2API 分布式部署与负载均衡方案（含 antigravity 稳定性专项）

> 更新时间：2026-02-12

## 1. 问题画像与根因

当前 `antigravity` 认证“报错后刷新恢复”的典型现象，核心是**瞬态错误与状态持久化策略之间的错配**：

- 上游 OAuth/网络/代理抖动导致短时失败；
- 失败被立即落库为 `error`，在下一次刷新成功前暴露为不可用；
- 多实例并行刷新时，出现同一账号重复刷新、重复回写、日志风暴，放大抖动。

这类问题不是单点 bug，而是**分布式一致性 + 异常分级 + 调度节奏**的系统性问题。

## 2. 本轮已落地改造（代码层）

### 2.1 Token Refresh 分布式去重

已在 `TokenRefreshService` 增加：

- **Leader Lock**（基于 `SchedulerCache.TryLockBucket`，Redis `SET NX + TTL`）
- **Startup Jitter**（启动错峰）
- **Cycle Jitter**（每周期错峰）

目标：避免多实例同一时刻冲击上游、重复刷新同一批账号。

### 2.2 配置化能力

新增配置项（并加入默认值与校验）：

- `token_refresh.startup_jitter_seconds`
- `token_refresh.cycle_jitter_seconds`
- `token_refresh.leader_lock_enabled`
- `token_refresh.leader_lock_ttl_seconds`

同时补齐 `deploy/config.example.yaml` 示例。

### 2.3 回归与稳定性验证

已完成：

- `go test ./... -count=1`
- `go test -race -tags unit ./internal/service -run "TestTokenRefreshService_..." -count=1`
- `go vet ./...`

覆盖点包括：leader lock 获取/失败/冲突分支、TTL 兜底、jitter 确定性等。

## 3. 参考 `new-api` 的可借鉴设计

参考 `new-api` 官方文档与仓库，可借鉴的核心思想：

- **节点角色化**：主节点处理写操作，从节点读扩展；
- **节点状态同步**：主从间周期同步配置和缓存；
- **多副本无状态**：应用层横向扩展，状态下沉数据库/缓存。

对 Sub2API 的映射建议：

- API 网关层保持无状态，多副本统一接入；
- 所有账号调度快照、限流状态、leader lock 放 Redis；
- 长周期任务（refresh / cleanup / aggregation）全部改成“分布式锁 + 幂等执行”。

## 4. 目标架构（推荐）

```text
[Clients]
   |
[CDN/WAF]
   |
[L7 Ingress: Nginx/Traefik/Envoy]
   |
+-----------------------------+
|  sub2api pods (N replicas)  |
|  - stateless API handling   |
|  - distributed jobs (lock)  |
+-----------------------------+
   |                |
[PostgreSQL]     [Redis]
   |
[Observability: Prometheus + Logs + Traces]
```

### 关键策略

- **读写分离**：热点读走 Redis，持久写走 PostgreSQL；
- **任务去重**：所有后台任务必须先争抢锁；
- **容错优先**：锁获取失败应“跳过本轮”，避免并发雪崩；
- **连接复用隔离**：按 `account_proxy` 维度隔离连接池，避免坏代理污染全局；
- **请求超时预算**：入口超时 > 服务超时 > 上游超时，确保可预期退避。

## 5. 反向代理 / 聚合网关选型建议

### Nginx（成熟稳定）

适合追求稳定与低维护成本场景：

- upstream 负载均衡（`least_conn` / `ip_hash` 等）
- 高性能反代与连接复用
- 配置静态、可控、对 SRE 友好

### Traefik（云原生友好）

适合 K8s + 自动发现场景：

- 原生服务发现与动态路由
- sticky session、权重路由、熔断中间件
- 变更成本低，灰度便利

### Envoy（高级流量治理）

适合需要细粒度弹性控制：

- outlier detection（异常实例摘除）
- circuit breaking（连接/并发保护）
- 强扩展能力（xDS、服务网格）

> 建议：
> - 1~2 台机器：Nginx 优先；
> - K8s 多环境：Traefik 优先；
> - 高并发多集群复杂治理：Envoy 优先。

## 6. antigravity 专项持久化建议

### 6.1 错误分级持久化

将错误分为：

- **Permanent**：例如 `invalid_grant` / `access_denied` / 明确资质失效
- **Transient**：网络抖动、超时、上游 5xx、代理异常

策略：

- Permanent 才写 `error`；
- Transient 仅记失败计数和最近错误，不降级可用状态；
- 达到阈值再进入 `error`（避免单次抖动误伤）。

### 6.2 状态回写幂等

- 每次 token 更新写 `_token_version`；
- 缓存写入必须比较版本号；
- `SetError` / `ClearError` 增加“状态前置条件”避免并发覆盖。

### 6.3 刷新节流

- 单账号刷新最小间隔（如 30~60s）；
- 平台/账号级并发上限；
- 同账号 refresh in-flight 去重（singleflight/Redis key）。

## 7. 分阶段实施路线

### Phase 1（已完成）

- token refresh leader lock + jitter
- 参数配置化 + 校验
- 单测 + race + 全量回归

### Phase 2（建议下一步）

- 引入统一错误分类器（permanent/transient）
- 引入账号级 refresh 去重键
- 增加 refresh 失败熔断窗口（短期失败快速失败）

### Phase 3（架构增强）

- 多副本部署 + LB 健康检查 + 滚动发布
- Redis 高可用（哨兵/集群）
- K8s HPA + PDB + 优雅终止

## 8. 外部参考（官方文档）

- New API 项目：
  - https://github.com/QuantumNous/new-api
  - https://docs.newapi.pro/en/installation/cluster-deployment
- Redis 分布式锁（`SET ... NX EX`）：
  - https://redis.io/docs/latest/commands/set/
- Nginx 负载均衡：
  - https://nginx.org/en/docs/http/load_balancing.html
- Traefik 负载均衡与 Sticky Sessions：
  - https://doc.traefik.io/traefik/reference/routing-configuration/http/load-balancing/service/
- Envoy 异常实例摘除：
  - https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier
- Kubernetes HPA / PDB：
  - https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
  - https://kubernetes.io/docs/tasks/run-application/configure-pdb/
