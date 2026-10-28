---
title: "DataBuff介绍与使用指南"
description: ""
date: 2026-10-21T07:45:00+08:00
image: ""
math: false
license:
comments: true
draft: false
build:
    list: always
tags : [APM, Java]
categories: 技术
lastmod: 2026-10-28T06:44:00+08:00
---
# DataBuff 介绍与使用指南
> 参考来源：DataBuff 官方文档（v0.1.7） <https://databuff.ai/docs/zh>
仓库：<https://github.com/databufflabs/databuff> · 许可证：AGPL-3.0

## 一、DataBuff 是什么
DataBuff 是一款**国产开源、AI 原生的 OpenTelemetry APM**（应用性能监控）后端。

一句话定位：**先接入标准遥测数据，再让 AI 读懂你的系统。**

它把两件事做在一起：

||OpenTelemetry APM|AI 原生|
|---|---|---|
|**定位**|标准、可靠的数据底座|直接读遥测数据的智能大脑|
|**价值**|看清 Trace、指标、拓扑、告警|用对话完成查询、巡检、诊断|

### OpenTelemetry APM 三大特点
1. **功能完善**：基于 OpenTelemetry 标准接入，覆盖应用性能监控全链路 —— 故障排查（服务红绿灯）、链路追踪、服务指标（QPS / 延迟 / 错误率 / JVM）、服务拓扑（自动绘制调用关系）。
1. **告警基础能力**：灵活的阈值与突变检测规则、定时评估核心服务指标、记录告警事件。
1. **架构极简**：**仅 3 个核心组件**（接入 + 存储 + 平台），Docker 一条命令即可跑起来，无复杂中间件堆砌，运维成本极低。

### AI 三大亮点
1. **AI 原生，不是外挂聊天框**：将 LLM 能力与 OpenTelemetry APM 数据原生融合，AI 直接查询 Trace、指标、拓扑、告警，而不是脱离上下文猜答案。
1. **功能丰富**：智能问数（自然语言查指标 / Trace / 拓扑 / 告警）、服务巡检（自动发现异常）、故障分析（综合多源数据给出诊断）、MCP 开放（外部 Agent 可调用平台能力）。
1. **AI 架构先进 · 多智能体协同**：AI 大脑统一理解意图并分派最合适的专家，数字专家各司其职（问数、巡检、分析），复杂问题多专家并行协作。

### 适用场景
- 希望快速落地 APM，又不想维护重型平台
- 想让研发 / 运维用对话代替查图表
- 需要开源可私有化的 AI 运维能力
- 正在评估 AI 原生 OpenTelemetry APM 的技术选型

## 二、架构介绍
### 2.1 最小架构：三大核心组件
DataBuff 采用极简的三组件架构，把遥测数据接入、存储、查询与可视化串在一条流水线里，没有多余的中间层：

```
OpenTelemetry / SkyWalking Agent
        │  (OTLP gRPC 4317 / HTTP 4318)
        ▼
   ┌──────────┐
   │  Ingest  │  接收数据：OTLP 接入、Trace 二次处理、指标聚合、Stream Load 写 Doris
   └────┬─────┘
        │  (Stream Load / JDBC)
        ▼
   ┌──────────┐
   │  Doris   │  统一查询存储：列存与时序查询，Trace/Log/Metric 表按天动态分区（默认保留约 30 天）
   └────┬─────┘
        │
        ▼
   ┌──────────┐
   │   Web    │  APM 平台与 AI 专家：REST API、前端、告警引擎、AI 平台与 MCP
   └──────────┘
```
Docker 默认四容器：Doris FE/BE、ingest、web。

### 2.2 组件职责
|组件|职责|
|---|---|
|**Ingest**|OTLP 接入、Trace 二次处理（Span 组装、服务元数据）、指标分钟级聚合、Stream Load 写 Doris|
|**Doris**|列存与时序查询；Trace / Log / Metric 表按天动态分区（默认保留约 30 天）|
|**Web**|REST API、前端、告警引擎、AI 平台与 MCP|

### 2.3 三种信号如何流转与关联
|信号|典型来源|Ingest 处理|Doris 存储（示例）|Web 能力|
|---|---|---|---|---|
|**Trace**|OTel SDK、Java Agent|Span 组装、服务元数据|`trace_dc_span`|调用链、拓扑、火焰图|
|**Metrics**|OTel metrics、JVM/HTTP 等|分钟级聚合|`metric_service*` 等|服务指标、仪表盘|
|**Logs**|OTel logs exporter|解析 OTLP log records|`log_dc_record`|日志检索、Trace 关联|

**三信号关联方式：**

- **Trace ↔ Log**：日志记录可携带 `trace_id` / `span_id`（OTel 语义约定）；在 Trace 详情中可关联查看同 trace 的日志行，日志也可反向落到具体 Span。
- **Trace ↔ Metric**：Ingest 从 Span 派生服务级、接口级分钟指标，与 Trace 共享 `service` / `instance` 维度。
- **统一查询面**：Web 按服务、时间范围聚合三类数据；AI 诊断以 Trace 与指标为上下文，逐步纳入日志。

### 2.4 AI 平台架构（多智能体协同）
DataBuff 的 AI 不是「外挂聊天框」，而是按 AI 原生设计、直接长在 APM 数据之上。

**核心原则：**

|原则|说明|
|---|---|
|数据驱动|AI 回答必须基于真实 APM 数据，不编造|
|专家分工|不同场景由不同专家处理，比单一模型更准|
|大脑编排|用户只面对一个入口，复杂协作在后台完成|
|开放扩展|Skill 定义专家行为，Tool 扩展能力边界|

**三层能力体系：**

|层|作用|举例|
|---|---|---|
|**Expert**|面向用户的智能角色|AI 大脑、智能问数、巡检|
|**Tool**|专家可调用的原子能力|查服务列表、查 Trace、画趋势图|
|**Skill**|约束专家的行为规则|问数口径、巡检流程、路由策略|

新增能力 = 组合 Tool + 编写 Skill + 注册 Expert，不用改核心代码。

**与 APM 的原生融合：** 问「错误率」查的是真实 Doris 指标而非幻觉；问「慢 Trace」拉的是真实链路数据；问「服务关系」画的是真实拓扑。这是「AI 原生 OpenTelemetry APM」与「APM + 聊天框」的本质区别。

**开放生态：** 支持多模型（OpenAI 兼容、Anthropic 等即配即用）；支持 MCP 协议（外部 Agent 如 Cursor / Claude 可通过标准 `/mcp` 端点调用平台 APM 工具，平台也可作 MCP 客户端接入外部服务）；Skill 可定制，内置 Skill 可覆盖。

### 2.5 数据流地址（默认）
|用途|Docker|Kubernetes|
|---|---|---|
|Web UI|`http://<本机IP>:27403`|`http://<节点IP>:32703`|
|默认账号|`admin` / `Databuff@123`|`admin` / `Databuff@123`|
|OTLP gRPC|`<本机IP>:4317`|`<节点IP>:30417`|
|OTLP HTTP|`http://<本机IP>:4318`|`http://<节点IP>:30418`|

K8s 集群内 Agent 上报：`http://ai-apm-ingest:4318`（gRPC `ai-apm-ingest:4317`）。

## 三、核心优势与亮点
> 本章是面向技术选型的浓缩版，便于快速决策。架构基础见第二章，与具体竞品的逐项实测对照见第四章。

### 3.1 六大核心优势
|#|优势|说明|
|---|---|---|
|1|**AI 原生，而非外挂聊天框**|LLM 与 APM 数据原生融合：问「错误率」查的是真实 Doris 指标，问「慢 Trace」拉的是真实链路，不编造、不脱离上下文。这是与「APM + 聊天框」的本质区别。|
|2|**7 大 AI 能力开箱即用**|看得见（自然语言问系统）→ 军团协同（多 Agent 协同）→ 会巡检（服务巡检 + 报告）→ 会诊断（瓶颈 / 根因取证）→ 会修复（运维专家处置）→ 会预测（容量 / 趋势）→ 会答疑（产品答疑专家）。SkyWalking / Pinpoint / Jaeger 均无等价能力。|
|3|**原生 OpenTelemetry / OTLP，多语言全信号**|原生支持 OTLP Trace + Metrics + Logs，一套后端覆盖 Java / Go / Python / Node 等多语言；并兼容 SkyWalking gRPC，可接管现有 SW Agent。|
|4|**APM 纵深：服务/实例/接口三级 + 服务流 + 中间件专页**|从「看见连谁」到「谁拖慢、再点进 Trace」：服务级 / 实例级 / 接口级调用分析均可落到 Trace；服务流按入口展开下游响应贡献度；DB / 缓存 / MQ / 外部服务独立专页。|
|5|**极简部署，3 组件一条命令**|仅 Ingest + Doris + Web，Docker 一条命令即可跑起来，无复杂中间件堆砌；推荐 8C16G 即可起步。|
|6|**开源可私有化，数据在自己手里**|AGPL-3.0 开源，支持私有化部署，遥测数据不出企业网络。|

### 3.2 开放生态亮点
- **多模型支持**：OpenAI 兼容、Anthropic 等即配即用，不绑定单一 LLM 厂商。
- **MCP 协议双向**：DataBuff 既是 MCP 服务端（外部 Agent 如 Cursor / Claude 可通过 `/mcp` 调用平台 APM 工具），也可作 MCP 客户端接入外部服务。
- **Skill 可定制**：新增能力 = 组合 Tool + 编写 Skill + 注册 Expert，不用改核心代码；内置 Skill 可覆盖。
- **多智能体协同**：AI 大脑统一理解意图并分派专家，数字专家各司其职（问数、巡检、分析），复杂问题多专家并行协作。

### 3.3 适合谁 / 不适合谁
||适合 DataBuff|暂不适合（或需并跑）|
|---|---|---|
|**团队**|想让研发 / 运维用对话代替查图表；想要 AI 运维但要求私有化|已深度绑定某 APM 插件 / 手动埋点 API，迁移成本高|
|**技术栈**|多语言微服务、已用或愿意用 OpenTelemetry|纯 Java 且强依赖方法级 Call Tree / 字节码 Profiling|
|**诉求**|AI + APM 纵深 + 告警一体化、快速落地|强依赖可定制仪表盘 / eBPF Profiling（DataBuff 暂不支持）|
|**迁移意愿**|接受改上报地址并跑验证（SW / Jaeger）或换 OTel Agent（Pinpoint）|不愿换任何 Agent|

> 一句话：**DataBuff = OpenTelemetry 标准接入 + AI 原生分析 + 极简私有化部署**。先用现有 Agent 改上报地址并跑验证，是最稳妥的评估路径。

### 3.4 AI 平台的启用前提与能力边界（不配 API Key 能用吗）
这是选型时最常被问到的问题，单独说明。

**关键结论：DataBuff 本身开源免费、不卖 API Key。** 这里的「API Key」指你自己从大模型厂商（OpenAI 兼容接口、Anthropic 等）申请的密钥，用来驱动 AI 平台的 LLM 推理。模型推理由外部 LLM 完成，因此**必须有可用的模型 API Key 才能启用 AI 能力**。

但 APM 基础功能与 AI 平台是**两条相互独立的链路**：官方安装后验证把「配置模型」明确标注为**可选**步骤（第 3 步带「（可选）」），前两步看服务列表、链路、拓扑无需任何模型配置即可完成。

||不配 API Key|配了 API Key|
|---|---|---|
|**APM 基础**：Trace / 指标 / 拓扑 / 服务流 / 中间件专页 / 日志 / 告警|✅ 完整可用|✅ 完整可用|
|**AI 平台**：问数 / 巡检 / 诊断 / 修复 / 预测 / 答疑 / AI 大脑协同|❌ 不可用|✅ 可用|
|**开放扩展**：MCP / Skill / 自定义数字专家|❌ 不可用（依赖 AI 大脑路由）|✅ 可用|

**不配 API Key 仍可用的功能（OpenTelemetry APM 基础能力）：**

- ✅ 链路追踪（Trace 列表 / 详情 / 瀑布图）
- ✅ 服务列表 / 黄金指标（QPS、延迟、错误率、JVM）
- ✅ 全局拓扑 / 服务流 / 服务-实例-接口三级调用分析
- ✅ 中间件专页（DB / 缓存 / MQ / 外部服务）
- ✅ 日志分析 + Trace↔Log 关联
- ✅ 告警中心（阈值规则、突变检测、告警事件）
- ✅ 数据采集、存储、查询

**必须配 API Key 才能用的功能（AI 平台）：**

AI 平台「快速上手」第一步即「配置管理 → 模型配置 → 填入 API Key」，没有这一步无法进入对话：

- ❌ 智能问数（自然语言查指标 / Trace / 拓扑 / 告警）
- ❌ 会巡检（主动发现异常 + 报告）
- ❌ 会诊断 / 会修复 / 会预测
- ❌ 产品答疑专家
- ❌ AI 大脑多智能体协同
- ❌ MCP / Skill / 自定义数字专家

> **推荐路径**：先不配 Key，把 DataBuff 当标准 OpenTelemetry APM 跑通验证（采集、存储、可视化、告警全部跑得起来，零成本）；确认 APM 数据符合预期后，再填入你已有的任意 OpenAI 兼容模型 Key 解锁 AI 能力。两步相互独立、互不阻塞。

## 四、与同类产品对比
以下均为 **同机实测对比**（同一台 `192.168.50.140`，同一份 Demo 数据双跑），覆盖三款主流开源 APM：**SkyWalking 10.4.0**、**Pinpoint 3.1.0**、**Jaeger v1.76.0**，对应 DataBuff v0.1.4。标记：✅ 本环境可验证 · △ 有入口但深度有限 · ❌ 无等价能力。

> DataBuff 的总体优势已在第三章概述，本章聚焦与三款产品的逐项差异。

### 4.1 AI 能力（差距最大的一组）
SkyWalking 无等价 AI 平台；DataBuff 把 7 大能力组织成可配置首页入口，APM 数据直接作 AI 上下文，并支持外接 MCP / Skill 与自定义数字专家。

|AI 能力|SkyWalking 10.4.0|DataBuff|
|---|---|---|
|① 看得见 · 自然语言问系统|❌|✅ 中文问服务 / 拓扑 / 异常趋势，AI 直接读遥测作答|
|② 军团协同 · 多 Agent 协同|❌|✅ 多专家并行取证、串行保上下文；任务可编排复用|
|③ 会巡检 · 服务巡检 + 报告|❌|✅ 一句话巡检，输出带证据与处置建议的报告|
|④ 会诊断 · 瓶颈 / 根因取证|❌|✅ 结合 Trace / 指标 / 拓扑拼诊断证据（非黑盒一句「根因」）|
|⑤ 会修复 · 运维专家处置|❌|✅ 策略允许 + 人工授权下执行修复；危险命令 denylist|
|⑥ 会预测 · 容量 / 趋势|❌|✅ 容量与趋势分析，从事后排障拉到事前预判|
|⑦ 会答疑 · 答疑专家|❌|✅ 检索产品文档与代码，回答部署 / 接入 / 配置问题|
|外部拓展 · MCP / Skill / 自定义专家|❌|✅ 外接 MCP、Skill，并可自定义数字专家扩展排障能力|

### 4.2 APM 应用性能
基础面（拓扑 / 服务列表 / Trace / 日志 / Span↔日志）两侧都有；DataBuff 领先在 **服务级 · 实例级 · 接口级调用分析（含关联 Trace）**、**实例 / 接口级拓扑**、**服务流**、**APM 专页与调用分析 / Trace 联动**及**错误分析纵深**，以及 Log→Trace 落到具体 Span。Profiling 与可定制仪表盘深度是 SkyWalking 明显更强、DataBuff 暂未覆盖的两块。

|APM 能力|SkyWalking|DataBuff|
|---|---|---|
|全局拓扑|✅|✅ 全局拓扑 + 健康色标 + 节点下钻|
|服务列表 / 黄金指标|✅|✅|
|服务级调用分析（上下游指标 + 关联 Trace）|❌|✅ 可直接落到 Trace|
|实例级拓扑 / 实例级调用分析|❌|✅|
|接口级拓扑 / 接口级调用分析|❌|✅|
|服务流（服务级 / 接口级 Trace 链路分析）|❌|✅ 按入口展开下游响应贡献度|
|中间件 / 外部调用专页（DB / 缓存 / MQ）|✅ Dashboard 大盘|✅ 独立 APM 专页 + 联动调用分析 / Trace|
|错误分析（统计 + 接口级）|❌|✅|
|Trace 列表 / 详情 / Span 关联日志|✅|✅|
|Log → Trace|✅|✅ 并可落到具体 Span|
|Profiling（Tracing / AsyncProfiler / eBPF）|✅ 三类均支持|❌ 暂不支持|
|可定制仪表盘|✅|❌ 暂不支持|

### 4.3 告警
差异主要在两头：**怎么配**（后端文件 vs 告警中心）和**配完能干什么**（查事件 / hooks vs 列表 + 智能告警 + 回服务上下文）。

|告警能力|SkyWalking|DataBuff|
|---|---|---|
|规则怎么配|△ 改 OAP `alarm-settings.yml`（或动态配置）；UI 不建规则|✅ 告警中心内配置，产品化入口|
|阈值告警|△ 靠后端 YAML / MQE 表达式维护|✅ 阈值规则可在平台内管理|
|智能告警|❌|✅ 与 APM 指标联动|
|告警事件列表|△|✅（等级 / 服务 / 时间）|
|告警落到服务 / 中间件|△ 触发后多靠 hooks 通知|✅ 列表直接挂服务 / 中间件，可回 APM 下钻|

### 4.4 总体对比
|对比维度|传统 APM（SkyWalking 等）|DataBuff|
|---|---|---|
|AI 能力|无或外挂|**AI 原生，直接读遥测数据**|
|部署复杂度|组件多、资源重|**3 组件，极简部署，一条命令**|
|排障方式|人翻图表|**对话式智能分析**|
|接入标准|私有协议为主|**原生 OpenTelemetry / OTLP**，并可接管现有 SkyWalking / Pinpoint / Jaeger Agent|
|数据归属|私有化|**开源可私有化，数据在自己手里**|

### 4.5 对比 Pinpoint（v0.1.4 vs Pinpoint 3.1.0）
同机双跑：DataBuff 走 OTLP `:4318`，Pinpoint 走官方 quickstart Java Agent（Web `:18080` / Demo `:18085`）。

**AI 能力**：Pinpoint 无 AI 平台，7 大能力与 MCP / Skill / 自定义专家全部 ❌，DataBuff 全部 ✅。

**应用性能（APM）关键差异**：

|APM 能力|Pinpoint 3.1.0|DataBuff|
|---|---|---|
|全局拓扑|✅ Server Map（含 USER→App、吞吐/耗时）|✅ 全局拓扑 + 健康色标 + 节点下钻|
|服务级调用分析（上下游指标 + 关联 Trace）|❌ Scatter/Call Tree 偏单请求，无「上下游贡献」分析页|✅ 上下游调用结构与耗时/贡献；可直接落到 Trace|
|实例级拓扑 / 实例级调用分析|❌ / △|✅ / ✅|
|接口级拓扑 / 接口级调用分析|❌ / △|✅ / ✅|
|服务流（服务级 / 接口级 Trace 链路分析）|❌ Server Map 只回答「连谁 / 多少次」|✅ 按入口展开下游响应贡献度；可关联 Trace|
|中间件 / 外部调用专页（库 / 缓存 / MQ / 外部服务）|△ 拓扑可出现中间件节点，无专页纵深|✅ 独立专页：数据库 / 缓存 / MQ / 外部服务|
|错误分析（统计 + 接口级）|△ 入口在，本 lab API 未就绪|✅ 独立错误分析统计 + 接口级错误下钻|
|Trace 详情|✅ Call Tree / Server Map / Flame Graph；方法级字节码栈深|✅ 调用次序瀑布图 + Span 属性|
|Trace Span 关联日志 / 日志平台 / Log→Trace|❌ 无 OTLP/业务日志平台|✅ 日志分析 + Span Logs + Log→Trace 落到具体 Span|
|Profiling|△ 侧重 Java 调用栈 / Active Thread|❌ 暂不支持|
|仪表盘（可定制 Dashboard）|❌ 无等价可定制大盘|❌ 暂不支持|
|接入协议 / 语言|专有 Java Agent（字节码增强）|OTLP 多语言 + SkyWalking gRPC|

**告警**：Pinpoint 在 Administration 下有 Alarm / Webhook 入口（△）；DataBuff 在告警中心内配置（✅），含智能告警、告警事件列表、告警直接挂服务/中间件可回 APM 下钻。

> **客观边界：** Pinpoint 在 **Java 方法级 Call Tree** 与 **Server Map + Scatter + Apdex 一体视图**上仍然扎实；纯 Java、要方法级调用栈时继续用合理。DataBuff 适合「多语言 OTel + AI + APM 纵深」。从 Pinpoint 迁出需换 Agent（专有 → OTel），迁移成本高于「只改 SkyWalking 上报地址」。

### 4.6 对比 Jaeger（v0.1.4 vs Jaeger v1.76.0）
同一 Demo 分别上报 OTLP：DataBuff `:4318`，Jaeger all-in-one UI `:16686`。

**定位差异**：Jaeger 是**分布式追踪后端**，专注 Trace 存储、搜索与瀑布图；DataBuff 是**完整 APM + AI 平台**。Jaeger 在 APM 服务面（黄金指标、多级拓扑 / 调用分析、服务流、中间件专页、日志、告警）基本空白。

**应用性能（APM）关键差异**：

|APM 能力|Jaeger v1.76.0|DataBuff|
|---|---|---|
|全局拓扑|△ Dependencies（只回答「连谁」）|✅ 全局拓扑 + 健康色标 + 节点下钻（含中间件）|
|服务列表 / 黄金指标|❌ 仅 Search 页服务下拉，无独立列表 / 曲线|✅ 服务列表 + 曲线|
|服务/实例/接口级拓扑 + 调用分析|❌ / ❌ / ❌|✅ / ✅ / ✅（均可落到 Trace）|
|服务流（响应贡献度）|❌ Dependencies 只回答「连谁」|✅ 按入口展开下游响应贡献度|
|中间件 / 外部调用专页|❌|✅ 数据库 / 缓存 / MQ / 外部服务|
|错误分析（统计 + 接口级）|❌ 多靠 Trace 状态筛选|✅ 独立错误分析 + 接口级下钻|
|Trace 列表 / 搜索 / 详情|✅ 成熟搜索 + 经典 Waterfall + Tags + Span Logs|✅ 图表 + 列表多维过滤 + 瀑布图|
|Trace Span 关联日志 / 日志平台 / Log→Trace|△ 仅有 Span Logs（埋点事件）/ ❌ / ❌|✅ / ✅ / ✅（落到具体 Span）|

**告警**：Jaeger **本身不做告警**，要阈值 / 通知需外挂 Prometheus / Alertmanager；DataBuff 把规则配置、智能告警、事件列表与回服务上下文放在同一告警中心。

**AI 能力**：Jaeger 无 AI 平台，7 大能力与 MCP / Skill / 自定义专家全部 ❌，DataBuff 全部 ✅。

> **客观边界：** 已深度绑定 Jaeger 搜索工作流、或只需要轻量 Trace 存储 + 瀑布图（尤其已有 ES / Cassandra 可复用）时，继续用 Jaeger 完全合理。DataBuff 适合「同一 OTLP 数据 + AI + APM 纵深 + 告警」的并跑或渐进切换——改上报地址即可。

### 4.7 三款产品综合对比一览
|对比维度|SkyWalking 10.4.0|Pinpoint 3.1.0|Jaeger v1.76.0|**DataBuff**|
|---|---|---|---|---|
|定位|全栈 APM|Java 深度 APM|分布式追踪后端|**AI 原生 APM**|
|AI 能力（7 大 + MCP/Skill）|❌|❌|❌|**✅ 全部**|
|接入协议|私有 gRPC 为主（可接 OTel）|专有 Java Agent|OTLP（Trace 为主）|**原生 OTLP 全信号 + 兼容 SW gRPC**|
|多语言|Java 强，多语言一般|以 Java 为主|多语言（仅 Trace）|**多语言 Trace+Metrics+Logs**|
|服务/实例/接口级调用分析|❌（服务级有，实例/接口级弱）|❌ / △ / △|❌|**✅ 三级齐全，可落到 Trace**|
|服务流（响应贡献度）|❌|❌|❌|**✅**|
|中间件专页|△ Dashboard 大盘|△ 拓扑节点无专页|❌|**✅ 独立专页**|
|日志平台 + Log↔Trace|△|❌|❌（仅 Span Logs）|**✅ 双向关联**|
|内置告警 / 智能告警|△（YAML 配置）|△|❌（需外挂）|**✅ 告警中心 + 智能告警**|
|Profiling|✅ 三类|△ Java 调用栈|❌|❌ 暂不支持|
|可定制仪表盘|✅|❌|❌|❌ 暂不支持|
|部署复杂度|中（组件多）|中|低（但仅 Trace）|**低（3 组件一条命令）**|
|从其迁出到 DataBuff|改上报地址即可|需换 Agent|改上报地址即可|—|

**选型一句话**：要 Java 方法级深度 / Profiling 选 SkyWalking 或 Pinpoint；只要轻量 Trace 选 Jaeger；要 **AI + 多语言 OTel + APM 纵深 + 告警一体化** 选 DataBuff。三者均可用现有 Agent 改上报地址到 DataBuff 并跑验证（Pinpoint 需换 OTel Agent）。

## 五、部署
> **部署总览**：DataBuff 平台由 **Web、Ingest、Doris（FE/BE）** 三个组件构成（见 2.1），Docker 默认拉起四个容器（Doris FE、Doris BE、ingest、web）。部署流程为：**环境准备 → 选择模式安装（Docker / K8s）→ 安装后验证 → 按需调参 → 接入业务应用**。两种安装模式产出的组件完全一致，区别只在外层编排（Docker Compose vs K8s 工作负载）。

- **数据面**：业务应用通过 OTel Agent 上报到 Ingest（`4317` gRPC / `4318` HTTP），Ingest 经 Stream Load 写入 Doris，Web 从 Doris 查询展示（见 2.5 地址表）。
- **持久化**：观测数据全部落在安装目录的 `data/` 下；**重装会清空 `data/`，升级不会**——这是选择操作命令时最关键的一条分界线（见 5.5）。

### 5.1 环境准备
- **Docker 部署**：Docker + Docker Compose
- **K8s 部署**：Kubernetes 集群 + kubectl
- **推荐配置**：8C16G（8 核 CPU / 16GB 内存）
- Doris BE 在 x86*64 / amd64 上依赖 AVX2 指令集；无 AVX2 时安装脚本会 `exit 1`，PoC 应急可设 `DATABUFF*SKIP*AVX2*CHECK=1` 跳过（生产请换支持 AVX2 的机器或 arm64）。

### 5.2 Docker 一键安装（推荐入门）
```bash
# 全新安装（会清理旧安装目录与 data/）
curl -fsSL https://databuff.ai/databuff/ai-apm-install.sh | bash

# 指定版本安装
curl -fsSL https://databuff.ai/databuff/ai-apm-install.sh | bash -s -- --version 0.1.7
# 或
APM_VERSION=0.1.7 curl -fsSL https://databuff.ai/databuff/ai-apm-install.sh | bash
```
安装完成后终端会输出 Web UI、账号和 OTLP 地址：

```
Web UI : http://YOUR_HOST:27403
Ingest : http://YOUR_HOST:4318/v1/traces
Status : Databuff is ready
```
**离线安装**（无法访问镜像仓库时）：

```bash
tar -zxvf databuff-ai-apm-offline-<version>-<arch>.tar.gz
cd databuff-ai-apm-offline-<version>-<arch>
sudo ./install.sh
```
**常用启停命令：**

```bash
cd /opt/databuff-ai-apm
./start.sh
./stop.sh
```
**安装 Demo（可选，持续上报 Trace 用于验证）：**

```bash
curl -fsSL https://databuff.ai/databuff/ai-apm-demo-install.sh | bash
```
### 5.3 Kubernetes 安装
```bash
curl -fsSL https://databuff.ai/databuff/ai-apm-k8s-install.sh | bash
```
默认命名空间 `databuff`，安装完成后输出 Web UI、命名空间和访问方式（见 2.5 节地址表）。集群内 Agent 上报 `http://ai-apm-ingest:4318`，NodePort 供集群外使用。

**离线镜像（节点无法拉取镜像时）：**

```bash
curl -fsSL https://databuff.ai/databuff/ai-apm-k8s-download-images.sh | bash
```
### 5.4 安装后验证
1. 打开 Web UI，用默认账号 `admin` / `Databuff@123` 登录
1. **应用性能 → 服务列表**，确认 Demo / 自有服务有数据
1. （可选）**配置管理 → 模型配置**，输入 API Key 启用 AI

### 5.5 升级与卸载
> 已有环境换版本请用升级命令，**不要用 `install.sh` 重装**——重装会清空 `data/` 中的观测数据。

```bash
# 平台在线升级（就地升级，保留 data/ 观测数据）
curl -fsSL https://databuff.ai/databuff/ai-apm-update.sh | bash

# 平台离线升级
tar -zxvf databuff-ai-apm-offline-<version>-<arch>.tar.gz
cd databuff-ai-apm-offline-<version>-<arch>
sudo ./update.sh

# Demo 在线升级
curl -fsSL https://databuff.ai/databuff/ai-apm-demo-update.sh | bash
```
> v0.1.7 含 Schema 迁移 V007（新增 `metric_platform`，`schema_version` → 7），升级脚本会自动跑迁移；**K8s 本版不支持原地升级，需卸载后重装**。

**卸载：**

```bash
# 平台（卸载前备份 data/ 以保留观测数据）
cd /opt/databuff-ai-apm
./stop.sh
cd .. && sudo rm -rf /opt/databuff-ai-apm

# Demo（不会删除平台及 data/）
cd /opt/databuff-ai-apm-demo
./stop.sh
cd .. && sudo rm -rf /opt/databuff-ai-apm-demo
```
### 5.6 关键参数配置
> Docker 推荐写入 `docker-compose.override.yml`（安装脚本自动合并，**升级不会覆盖**该文件）；K8s 编辑 `kubectl -n databuff edit configmap ai-apm-config` 后 `rollout restart`。改密后需同时重启 `ai-apm-web` 与 `ai-apm-ingest`。

**修改 Web 登录账号密码：**

|环境变量|默认值|
|---|---|
|`APM_SECURITY_SEED_USERNAME`|`admin`|
|`APM_SECURITY_SEED_PASSWORD`|`Databuff@123`|

**配置 Doris 用户名密码（Web / Ingest 连 Doris 的账号，非 Web 登录账号）：**

|环境变量|默认值|
|---|---|
|`DORIS_USER`|`root`|
|`DORIS_PASSWORD`|空|

顺序：先在 Doris 内改密或建用户，再改应用侧环境变量。

**配置多 FE / 多 BE 地址（v0.1.7 新增，逗号分隔 + 健康轮询）：**

```yaml
# Docker：写入 docker-compose.override.yml
services:
  ai-apm-web:
    environment:
      DORIS_FE_HOST: "192.168.1.10,192.168.1.11"
      DORIS_FE_QUERY_PORT: "9030"
  ai-apm-ingest:
    environment:
      DORIS_FE_HOST: "192.168.1.10,192.168.1.11"
      DORIS_FE_QUERY_PORT: "9030"
      DORIS_BE_HTTP_HOST: "192.168.1.20,192.168.1.21"   # 推荐，Stream Load 直连 BE:8040
      DORIS_BE_HTTP_PORT: "8040"
```
端口用途速查：

|端口（默认）|环境变量|谁在用|做什么|
|---|---|---|---|
|9030|`DORIS_FE_QUERY_PORT`|Web、Ingest|FE JDBC 查询（MySQL 协议读数）|
|8030|`DORIS_FE_HTTP_PORT`|主要是 Ingest|FE HTTP；未配 BE 时 Stream Load 先打 FE 再 307 到 BE|
|8040|`DORIS_BE_HTTP_PORT`|仅 Ingest|BE HTTP；配了 BE 时 Stream Load 直连 BE（推荐）|

地址写法：单主机 `fe1`；多主机共享端口 `fe1,fe2`；每条自带端口 `fe1:19030,fe2:19031`。多 FE 时 JDBC 自动变为 `jdbc:mysql:loadbalance://...`。

### 5.7 存储时长调整
想保留 14 天时，连接 Doris 执行（无需重启 FE / BE）：

```sql
USE databuff;
ALTER TABLE trace_dc_span SET ("dynamic_partition.start" = "-14");
ALTER TABLE log_dc_record SET ("dynamic_partition.start" = "-14");
ALTER TABLE metric_service SET ("dynamic_partition.start" = "-14");
-- 其余 metric_* 表同理
```
### 5.8 部署模式选型与资源规划
|维度|Docker 单机（5.2）|Kubernetes（5.3）|
|---|---|---|
|适用场景|体验、PoC、小规模（单实例）|生产、需要弹性伸缩 / 高可用编排|
|编排方式|Docker Compose，默认四容器|命名空间 `databuff` 内工作负载|
|升级|在线/离线升级脚本（5.5）|**v0.1.7 不支持原地升级，需卸载重装**（5.5 注意）|
|集群外上报地址|`http://YOUR_HOST:4318`|NodePort；集群内用 `http://ai-apm-ingest:4318`|
|存储扩展|单机 `data/`，纵向扩容为主|可结合外部 Doris 集群横向扩展|

**资源规划建议：**

- **基线配置 8C16G**（5.1 推荐）；其中 Doris（尤其 BE）是资源大头，内存不足时 BE 容易反复重启。
- **磁盘**：观测数据按天动态分区，默认保留约 30 天（2.2）；预估方式 = 日均遥测量 × 保留天数 × 冗余系数（建议 1.5）。可按 5.7 调整保留时长来控制磁盘增速。
- **网络**：确保业务应用所在网络可达 Ingest 的 4317/4318 端口；Web UI（默认 27403）按需限制来源 IP。
- **CPU 架构**：x86_64 需支持 AVX2（5.1）；arm64 无此限制。

**生产部署建议（在基线之上）：**

1. **改默认密码**：首件事改掉 `admin / Databuff@123`（5.6），并为 Doris 配置独立账号密码（`DORIS_USER` / `DORIS_PASSWORD`）。
1. **Doris 外置/多副本**：把 Doris 从平台内置容器换成外部集群，并按 5.6 配置多 FE / 多 BE 地址（`fe1,fe2`、`DORIS_BE_HTTP_HOST` 直连 BE），提高存储与查询的可用性。
1. **配置持久化**：所有改动写入 `docker-compose.override.yml`（升级不覆盖）；K8s 侧改 ConfigMap 后 `rollout restart`。
1. **接入规划**：平台就绪后再接入业务（第六、七章）；Spring Boot 用 OTel Java Agent（6.2–6.4），Spring Cloud 按服务逐个接入并统一 `service.name`（7.2）。

### 5.9 常见部署问题排查
|现象|原因|处理|
|---|---|---|
|安装脚本 `exit 1`，提示 AVX2|Doris BE 依赖 AVX2 指令集|换支持 AVX2 的机器或 arm64；PoC 应急可 `DATABUFF_SKIP_AVX2_CHECK=1`（5.1）|
|安装后无数据、服务列表为空|业务应用未正确连到 Ingest|检查 Agent 导出地址是否指向 `4317/4318`（6.3）；K8s 内应使用 `http://ai-apm-ingest:4318`|
|BE 容器反复重启|内存不足或磁盘压力|核对 8C16G 基线；检查磁盘水位与 `data/` 所在分区|
|Web 登录报错 / 一直 401|改过 `APM_SECURITY_SEED_*` 但未重启|改密后需同时重启 `ai-apm-web` 与 `ai-apm-ingest`（5.6）|
|Web/Ingest 连不上 Doris|Doris 账号密码或 FE 地址不对|按 5.6 核对 `DORIS_USER` / `DORIS_PASSWORD` 与多 FE 地址写法|
|升级后数据丢失|误用 `install.sh` 重装|重装会清空 `data/`；换版本必须走升级命令（5.5），且升级前备份 `data/`|
|端口被占用|27403 / 4317 / 4318 冲突|排查占用进程；集群外访问走 NodePort 时检查端口放行|

通用排查顺序：`./stop.sh && ./start.sh` 重启组件 → 看 Web UI 是否可登录 → 看 Ingest 端口是否可达 → 看容器/工作负载日志 → 最后才考虑卸载重装（务必先备份 `data/`）。

### 5.10 数据备份与迁移
- **要备份什么**：安装目录下的 `data/`（全部观测数据）+ `docker-compose.override.yml`（自定义配置）。两者都在 `/opt/databuff-ai-apm/` 下。
- **常规备份**：停服后直接打包目录，操作前先 `./stop.sh`：

```bash
cd /opt/databuff-ai-apm && ./stop.sh
sudo tar -czf /backup/databuff-data-$(date +%F).tar.gz data/ docker-compose.override.yml
./start.sh
```
- **迁移到新机器**：新机按 5.2 全新安装同版本 → 停服 → 用备份的 `data/` 覆盖新环境的 `data/` → 恢复 `docker-compose.override.yml` → `./start.sh`。**新旧环境版本需一致**，跨版本请先在旧环境升级（5.5）再迁移。
- **卸载前必备份**：执行 5.5 的卸载操作会删除整个安装目录，观测数据无法找回，先完成上述打包。

## 六、在 Spring Boot 项目中使用
DataBuff 通过 **OpenTelemetry Java Agent** 给 Spring Boot 应用做**零代码埋点**，把 Trace 和 Metrics 经 OTLP 上报到 DataBuff。适合已有 Spring Boot 项目、希望最少改代码即可接入 APM 的场景。

### 6.1 前置条件
- JDK 8+ 的 Spring Boot 应用（可运行 JAR 或本地 `spring-boot:run`）
- 已部署 DataBuff，且应用进程能访问 Ingest 的 OTLP 端口（默认 4317 gRPC / 4318 HTTP）

### 6.2 下载 OpenTelemetry Java Agent
```bash
curl -L -o opentelemetry-javaagent.jar \
  https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar
```
Agent 会自动识别 Spring MVC、WebFlux、Jdbc、Kafka 等常见框架，**无需修改业务代码**。

### 6.3 配置应用
`application.yml`（Agent 会读取 `spring.application.name` 作为默认服务名）：

```yaml
spring:
  application:
    name: my-spring-service
```
设置 OTLP 导出环境变量（将 `<ingest-host>` 换成实际地址，本机 Docker 安装通常为 `localhost`）：

```bash
export OTEL_SERVICE_NAME=my-spring-service
export OTEL_TRACES_EXPORTER=otlp
export OTEL_METRICS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_ENDPOINT="http://<ingest-host>:4318"
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
```
使用 gRPC 时：

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT="http://<ingest-host>:4317"
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
```
> **提示**：Docker 与本机 Spring Boot 同机运行时 `<ingest-host>` 用 `localhost`；应用在容器内、DataBuff 在宿主机时用宿主机 IP 或 `host.docker.internal`（macOS / Windows Docker Desktop）。

### 6.4 启动应用（带 Java Agent）
运行 JAR：

```bash
java -javaagent:./opentelemetry-javaagent.jar \
  -jar target/my-spring-service.jar
```
Maven 本地开发：

```bash
./mvnw spring-boot:run \
  -Dspring-boot.run.jvmArguments="-javaagent:./opentelemetry-javaagent.jar"
```
也可用 JVM 系统属性代替部分环境变量：

```bash
java -javaagent:./opentelemetry-javaagent.jar \
  -Dotel.service.name=my-spring-service \
  -Dotel.exporter.otlp.endpoint=http://<ingest-host>:4318 \
  -Dotel.exporter.otlp.protocol=http/protobuf \
  -Dotel.traces.exporter=otlp \
  -Dotel.metrics.exporter=otlp \
  -jar target/my-spring-service.jar
```
启动后访问几个 HTTP 接口（如 `/actuator/health` 或业务 API）以产生 Trace 和 Metrics。

### 6.5 采样与导出频率（生产可选）
```bash
export OTEL_TRACES_SAMPLER=parentbased_traceidratio
export OTEL_TRACES_SAMPLER_ARG=0.1          # 约 10% Trace
export OTEL_METRIC_EXPORT_INTERVAL=60000    # 指标 60s 导出一次
```
### 6.6 在 DataBuff 中验证
1. 打开 Web UI `http://<ingest-host>:27403`，使用默认账号登录
1. **应用性能 → 服务列表**，确认出现 `my-spring-service`
1. **应用性能 → 链路追踪**，查看刚才请求产生的 Trace
1. 在服务详情中查看 JVM / HTTP 等指标曲线

> 若无数据，请检查：Ingest 端口是否可达、服务名是否与列表一致、应用日志中是否有 OTLP 导出错误。

### 6.7 进阶：Jenkins 构建 + Nacos 注册/配置中心部署
实际生产中，Spring Boot 项目常以 **Jenkins 流水线构建并发布**，并以 **Nacos 作为服务注册中心与配置中心**。本节给出该组合下接入 DataBuff 的完整方案。

> **核心结论**：Nacos 负责「服务注册与配置下发」，DataBuff 负责「可观测性采集与 AI 分析」，二者**互不冲突、各司其职**。OTel Java Agent 会自动埋点 Nacos 客户端调用（服务发现、配置拉取）以及经由 Nacos 注册发现的服务间调用，无需任何额外埋点。接入 DataBuff 不需要改 Nacos，也不需要改业务代码——只需在构建产物里带上 Agent 并在启动时注入 OTLP 环境变量。

### 6.7.1 架构与数据流
```
                 Jenkins (构建/发布)
                      │ 打包 jar + 内置 opentelemetry-javaagent.jar
                      ▼
        ┌──────────────────────────────┐
        │  Spring Boot 应用 (带 OTel Agent) │
        │  ├─ Nacos 注册中心: 注册/发现服务  │  ← Agent 自动埋点服务发现调用
        │  ├─ Nacos 配置中心: 拉取动态配置   │  ← Agent 自动埋点配置拉取
        │  └─ 业务接口 / Feign / JDBC / Redis│  ← Agent 自动埋点
        └───────────────┬──────────────┘
                        │ OTLP (4317 gRPC / 4318 HTTP)
                        ▼
                DataBuff Ingest → Doris → Web (APM + AI)
```
### 6.7.2 前置条件
- Spring Boot 项目已接入 Nacos（`spring-cloud-starter-alibaba-nacos-discovery` / `nacos-config`）
- Jenkins 可拉取代码、有 JDK 与 Maven（或 Gradle）构建环境
- 已部署 DataBuff，Jenkins 部署的目标机器能访问 Ingest OTLP 端口
- OTel Java Agent 版本与 Nacos 客户端兼容（OTel Agent 对 Spring Cloud Alibaba / Nacos 1.x、2.x 均有自动埋点支持）

### 6.7.3 第一步：把 OTel Agent 纳入构建产物
**方案 A（推荐）：Agent 放在固定部署路径，Jenkins 不打包它**

Agent 与业务 jar 解耦，便于统一升级。Jenkins 只负责构建业务 jar，部署脚本把 Agent 放到如 `/opt/otel/opentelemetry-javaagent.jar`。

**方案 B：Agent 随产物一起发布**

在项目根目录准备 `otel/` 目录存放 Agent，Jenkins 打包时一并归档。`Jenkinsfile` 示例：

```
pipeline {
    agent any
    tools { jdk 'JDK17'; maven 'Maven3' }
    environment {
        SVC_NAME      = 'order-service'
        IMAGE_TAG     = "${env.BUILD_NUMBER}"
        OTEL_ENDPOINT = 'http://10.0.0.10:4318'   // DataBuff Ingest 地址
    }
    stages {
        stage('Checkout') {
            steps { checkout scm }
        }
        stage('Build') {
            steps {
                sh 'mvn -B clean package -DskipTests'
            }
        }
        stage('Prepare OTel Agent') {
            steps {
                // 统一下载 OTel Java Agent，随产物一起发布
                sh '''mkdir -p otel
                      curl -fsSL -o otel/opentelemetry-javaagent.jar \
                        https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar'''
            }
        }
        stage('Build Image') {
            steps {
                sh "docker build -t ${SVC_NAME}:${IMAGE_TAG} ."
                sh "docker tag  ${SVC_NAME}:${IMAGE_TAG} ${SVC_NAME}:latest"
            }
        }
        stage('Deploy') {
            steps {
                // 推送并滚动更新（示例：远端 docker 部署）
                sh "docker save ${SVC_NAME}:${IMAGE_TAG} | gzip > ${SVC_NAME}.tar.gz"
                // 远端拉起时注入 OTEL_* 环境变量（见 6.7.6）
            }
        }
    }
}
```
### 6.7.4 第二步：Dockerfile（内置 Agent + Nacos 配置）
```
FROM eclipse-temurin:17-jre-jammy
WORKDIR /app

# 内置 OTel Java Agent
COPY otel/opentelemetry-javaagent.jar /app/opentelemetry-javaagent.jar

# 业务 jar
COPY target/order-service.jar /app/app.jar

# DataBuff OTLP 导出配置（运行时可被 -e 覆盖）
ENV OTEL_TRACES_EXPORTER=otlp \
    OTEL_METRICS_EXPORTER=otlp \
    OTEL_LOGS_EXPORTER=otlp \
    OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf \
    OTEL_EXPORTER_OTLP_ENDPOINT=http://10.0.0.10:4318 \
    OTEL_TRACES_SAMPLER=parentbased_traceidratio \
    OTEL_TRACES_SAMPLER_ARG=0.1 \
    OTEL_METRIC_EXPORT_INTERVAL=60000

# service.name 优先用环境变量注入，回退到 spring.application.name
ENV OTEL_SERVICE_NAME=order-service

ENTRYPOINT ["java", "-javaagent:/app/opentelemetry-javaagent.jar", "-jar", "/app/app.jar"]
```
### 6.7.5 第三步：Nacos 配置无需改动，仅需保留服务名
DataBuff 的服务名取自 `OTEL_SERVICE_NAME`（优先）或 `spring.application.name`。Nacos 的服务注册名通常也是 `spring.application.name`，因此**让二者保持一致**，DataBuff 上的服务就能和 Nacos 控制台里的服务对上号，便于排查。

`application.yml`（Nacos 注册 + 配置中心，DataBuff 不需要新增任何配置）：

```yaml
spring:
  application:
    name: order-service          # 与 OTEL_SERVICE_NAME 保持一致
  cloud:
    nacos:
      discovery:
        server-addr: ${NACOS_ADDR:127.0.0.1:8848}
        namespace: ${NACOS_NS:public}
      config:
        server-addr: ${NACOS_ADDR:127.0.0.1:8848}
        namespace: ${NACOS_NS:public}
        file-extension: yaml
```
> **提示**：如果希望 DataBuff 上的服务名带环境/命名空间后缀（如 `order-service-prod`），把 `OTEL_SERVICE_NAME` 单独设为该值即可，不影响 Nacos 注册名。

### 6.7.6 第四步：部署启动（注入环境变量）
**容器部署（Jenkins 推送到目标主机后拉起）：**

```bash
docker run -d --name order-service \
  -e OTEL_SERVICE_NAME=order-service \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=http://10.0.0.10:4318 \
  -e NACOS_ADDR=nacos.internal:8848 \
  -e NACOS_NS=prod \
  -p 8081:8081 \
  order-service:latest
```
**物理机 / 虚拟机部署（systemd 或脚本拉起）：**

```bash
export OTEL_SERVICE_NAME=order-service
export OTEL_TRACES_EXPORTER=otlp
export OTEL_METRICS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_ENDPOINT=http://10.0.0.10:4318
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf

java -javaagent:/opt/otel/opentelemetry-javaagent.jar \
  -jar /opt/apps/order-service.jar
```
### 6.7.7 OTel Agent 对 Nacos 场景的自动埋点覆盖
OTel Java Agent 会自动识别并埋点以下与 Nacos 相关的调用，**无需手动配置**：

|场景|自动埋点内容|在 DataBuff 中的体现|
|---|---|---|
|服务注册 / 心跳|Nacos 客户端向 Nacos Server 的注册与心跳调用|Trace 中可见 `nacos` 相关 Span|
|服务发现（Feign / RestTemplate / LoadBalancer）|消费方经 Nacos 拿到实例后的下游 HTTP 调用|跨服务调用链 + 全局拓扑自动绘制|
|配置拉取 / 监听|Nacos Config 客户端拉取配置、长轮询监听|Trace 中可见配置获取 Span|
|OpenFeign 调用|经负载均衡的 Feign 远程调用|端到端跨服务链路，含目标服务名|
|数据库 / Redis / MQ|JDBC、Lettuce/Jedis、Kafka/RabbitMQ|中间件专页 + 慢调用 / 错误下钻|

### 6.7.8 常见问题与排错
|现象|排查方向|
|---|---|
|DataBuff 服务列表看不到服务|检查 `OTEL_EXPORTER_OTLP_ENDPOINT` 是否可达；应用启动日志搜 `opentelemetry` / `otlp` 看导出错误|
|服务名是 `unknown_service`|未设 `OTEL_SERVICE_NAME` 且 Agent 未读到 `spring.application.name`（Java Agent 读 `spring.application.name` 需应用已加载该配置），建议显式设 `OTEL_SERVICE_NAME`|
|跨服务链路断开 / 拓扑不连|确认调用方与被调方都挂了 OTel Agent 且都上报到同一 DataBuff；Feign / RestTemplate 走 HTTP 时上下文自动透传，若用自定义 HTTP 客户端需确认未剥离 `traceparent` 头|
|Nacos 配置不生效|与 DataBuff 无关，检查 `NACOS_ADDR` / `namespace` / `data-id`；DataBuff 不干预 Nacos 配置链路|
|采样后 Trace 太少|调高 `OTEL_TRACES_SAMPLER_ARG`（如 `0.5` 或 `1.0`）或临时设 `OTEL_TRACES_SAMPLER=always_on` 排障|
|Agent 与 Nacos 客户端版本冲突|升级 OTel Java Agent 到最新稳定版；极少数情况下需用 `OTEL_INSTRUMENTATION_<NAME>_ENABLED=false` 关闭冲突插件|

### 6.7.9 验证
1. Jenkins 构建发布后，应用启动日志应出现 `[otel.agent]` 或 `opentelemetry-javaagent` 字样
1. 打开 DataBuff Web UI → **应用性能 → 服务列表**，应看到 `order-service`（与 Nacos 服务名一致）
1. 触发一次经由该服务的请求，**链路追踪**中查看 Trace：应包含业务 Span、下游 Feign 调用 Span、DB/Redis Span
1. **全局拓扑**应自动绘制 `order-service` 与其上下游服务的调用关系
1. 在 **AI 平台** 问数：「`order-service` 最近 1 小时 P99 延迟多少？瓶颈在哪个下游服务？」

## 七、在 Spring Cloud 项目中使用
Spring Cloud 项目本身基于 Spring Boot，每个微服务都是一个 Spring Boot 应用，因此接入 DataBuff 的方式与上一章一致：**每个微服务都挂载 OpenTelemetry Java Agent，统一上报到同一个 DataBuff Ingest**。DataBuff 会自动从跨服务的 Trace 中绘制全局拓扑与服务间调用关系。

### 7.1 接入思路
1. 每个微服务（gateway / 订单 / 用户 / 库存等）启动时都加上 `-javaagent:./opentelemetry-javaagent.jar`
1. 每个服务通过 `OTEL_SERVICE_NAME` 或 `spring.application.name` 设置独立服务名
1. 所有服务的 `OTEL_EXPORTER_OTLP_ENDPOINT` 指向**同一个** DataBuff Ingest 地址
1. Agent 自动埋点 Spring Cloud Gateway、OpenFeign、RestTemplate、Ribbon、Eureka/Nacos 注册中心调用、JDBC、Redis、Kafka、RabbitMQ 等，跨服务 Trace 上下文（trace propagation）自动透传

### 7.2 完整示例：多服务统一接入
假设有 `cloud-gateway`、`order-service`、`user-service` 三个服务，DataBuff Ingest 地址为 `http://10.0.0.10:4318`。

**每个服务各自的 `application.yml`：**

```yaml
# cloud-gateway
spring:
  application:
    name: cloud-gateway

# order-service
spring:
  application:
    name: order-service

# user-service
spring:
  application:
    name: user-service
```
**统一的启动方式（每个服务都一样，仅服务名不同）：**

```bash
# cloud-gateway
java -javaagent:./opentelemetry-javaagent.jar \
  -Dotel.service.name=cloud-gateway \
  -Dotel.exporter.otlp.endpoint=http://10.0.0.10:4318 \
  -Dotel.exporter.otlp.protocol=http/protobuf \
  -Dotel.traces.exporter=otlp \
  -Dotel.metrics.exporter=otlp \
  -jar cloud-gateway.jar

# order-service
java -javaagent:./opentelemetry-javaagent.jar \
  -Dotel.service.name=order-service \
  -Dotel.exporter.otlp.endpoint=http://10.0.0.10:4318 \
  -Dotel.exporter.otlp.protocol=http/protobuf \
  -Dotel.traces.exporter=otlp \
  -Dotel.metrics.exporter=otlp \
  -jar order-service.jar

# user-service 同理
```
启动后，一次经由 Gateway → order-service → user-service 的外部请求会形成一条完整跨服务调用链，DataBuff **应用性能 → 全局拓扑**会自动绘制三个服务的调用关系，**链路追踪**中可看到端到端瀑布图。

### 7.3 Dockerfile / 容器化部署模板
若微服务以容器方式运行，把 Agent 打进镜像并设置 `ENTRYPOINT`：

```
FROM eclipse-temurin:8-jre-jammy
WORKDIR /app

# 下载 OTel Java Agent
ARG OTEL_AGENT_VERSION=latest
RUN curl -L -o opentelemetry-javaagent.jar \
  https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/${OTEL_AGENT_VERSION}/download/opentelemetry-javaagent.jar

COPY target/*.jar app.jar

# 通过环境变量配置（运行时注入）
ENV OTEL_TRACES_EXPORTER=otlp \
    OTEL_METRICS_EXPORTER=otlp \
    OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf \
    OTEL_EXPORTER_OTLP_ENDPOINT=http://10.0.0.10:4318

ENTRYPOINT ["java", "-javaagent:/app/opentelemetry-javaagent.jar", "-jar", "/app/app.jar"]
```
部署时通过服务名环境变量区分各微服务：

```bash
docker run -e OTEL_SERVICE_NAME=order-service -p 8081:8081 order-service:latest
docker run -e OTEL_SERVICE_NAME=user-service  -p 8082:8082 user-service:latest
```
### 7.4 Kubernetes 中部署 Spring Cloud 微服务
将 OTLP 地址指向集群内 Ingest Service `ai-apm-ingest:4318`，并在 Deployment 中注入 Agent 与环境变量：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-service
  namespace: production
spec:
  selector:
    matchLabels:
      app: order-service
  template:
    metadata:
      labels:
        app: order-service
    spec:
      containers:
        - name: order-service
          image: order-service:latest
          ports:
            - containerPort: 8080
          env:
            - name: OTEL_SERVICE_NAME
              value: order-service
            - name: OTEL_TRACES_EXPORTER
              value: otlp
            - name: OTEL_METRICS_EXPORTER
              value: otlp
            - name: OTEL_EXPORTER_OTLP_PROTOCOL
              value: http/protobuf
            # 集群内直连 Ingest
            - name: OTEL_EXPORTER_OTLP_ENDPOINT
              value: http://ai-apm-ingest.databuff:4318
          # 通过 initContainer 或已打进镜像的方式提供 javaagent
          volumeMounts:
            - name: otel-agent
              mountPath: /otel
              readOnly: true
          # 假设镜像内 ENTRYPOINT 已带 -javaagent，否则在此 command/args 中补上
      volumes:
        - name: otel-agent
          emptyDir: {}
      initContainers:
        - name: download-otel-agent
          image: curlimages/curl:8.5.0
          command:
            - sh
            - -c
            - curl -L -o /otel/opentelemetry-javaagent.jar https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar
          volumeMounts:
            - name: otel-agent
              mountPath: /otel
```
> 集群外访问用 NodePort `http://<节点IP>:30418`；集群内微服务直连 `http://ai-apm-ingest.databuff:4318` 更稳。

### 7.5 已有 SkyWalking / Pinpoint / Jaeger Agent 的微服务
无需更换 Agent：DataBuff 可直接接管现有 Agent 的上报数据（改上报地址即可），实现「并跑」或渐进切换。详见官方 [SkyWalking 接入](https://databuff.ai/docs/zh/manual/skywalking-ingestion) 与各迁移指南。

### 7.6 验证
1. 打开 Web UI → **应用性能 → 服务列表**，应看到 gateway / order-service / user-service 等多个服务
1. **应用性能 → 全局拓扑**，确认服务间调用关系自动绘制（如 gateway → order-service → user-service）
1. **应用性能 → 链路追踪**，发起一次外部请求，查看端到端跨服务调用链瀑布图
1. 在 **AI 平台** 试用问数：「查询最近 1 小时 order-service 的错误率」或「order-service 调用了哪些下游服务」

## 八、附录：v0.1.7 版本亮点
- **平台自监控与自排障**：新增部署状态页（安装部署 → 部署状态），总览入站 TPS、写出失败、Doris 磁盘、查询失败；ingest / web / Doris 分 tab 看图；每个指标标题可点开看计算口径与可调环境变量。
- **AI 一键巡检与修复**：产品答疑专家可读指标清单、查平台自监控数据，输出 HTML 巡检报告；发现写出丢弃等问题可沿链路诊断，并按建议登录备份配置、改参数、重启 ingest、复查归零。
- **答疑专家 Doris 只读查询**：`POST /webapi/api/v1/ai/doris/query` 与内置工具 `platform.queryDoris`。
- **非 HTTP 接口分析完善**：DB / MQ / RPC 等接口下钻与慢调用、错误路径覆盖。
- **多 FE / 多 BE**：`DORIS_FE_HOST` / `DORIS_BE_HTTP_HOST` 支持逗号分隔地址列表，健康轮询失败自动换节点。
- **链路查询加速**：span 查询按分钟键裁剪并跳过 `call_spans` COUNT；按 TraceID 过滤时 Trace 图从 span 聚合。
- **告警与 MCP**：监控规则评估支持 `lt` / `lte` / `gte` 比较符与 warning 阈值；远程 MCP 在 transport 之后正确应用 headers。

## 参考链接
- 官方文档：<https://databuff.ai/docs/zh>
- GitHub 仓库：<https://github.com/databufflabs/databuff>
- 在线 Demo：<https://demo.databuff.ai/>
- Docker 安装：<https://databuff.ai/docs/zh/guide/docker>
- K8s 安装：<https://databuff.ai/docs/zh/guide/kubernetes>
- Spring Boot OTLP 接入：<https://databuff.ai/docs/zh/guide/spring-boot-otlp>
- OTLP 接入：<https://databuff.ai/docs/zh/guide/otel-otlp-ingestion>
- 遥测数据流与存储：<https://databuff.ai/docs/zh/architecture/telemetry-pipeline>
- AI 平台架构：<https://databuff.ai/docs/zh/architecture/ai-platform>
- 对比 SkyWalking：<https://databuff.ai/docs/zh/comparison/databuff-vs-skywalking>
- v0.1.7 发布说明：<https://databuff.ai/docs/zh/releases/0.1.7>

