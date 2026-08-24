---
title: "DataBuff 介绍与使用指南"
description: ""
date: 2026-08-24T10:50:00+08:00
image: ""
math: false
license:
comments: true
draft: false
build:
    list: always
tags : [APM, Java]
categories: 技术
lastmod: 2026-08-24T11:42:00+08:00
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
|**Doris**|列存与时序查询；Trace / Log / Metric 表按天动态分区（默认保留约 30 天）|
|**Web**|REST API、前端、告警引擎、AI 平台与 MCP|

### 2.3 三种信号如何流转与关联
|信号|典型来源|Ingest 处理|Doris 存储（示例）|Web 能力|
|---|---|---|---|---|
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
|专家分工|不同场景由不同专家处理，比单一模型更准|
|大脑编排|用户只面对一个入口，复杂协作在后台完成|
|开放扩展|Skill 定义专家行为，Tool 扩展能力边界|

**三层能力体系：**

|层|作用|举例|
|---|---|---|
|**Tool**|专家可调用的原子能力|查服务列表、查 Trace、画趋势图|
|**Skill**|约束专家的行为规则|问数口径、巡检流程、路由策略|

新增能力 = 组合 Tool + 编写 Skill + 注册 Expert，不用改核心代码。

**与 APM 的原生融合：** 问「错误率」查的是真实 Doris 指标而非幻觉；问「慢 Trace」拉的是真实链路数据；问「服务关系」画的是真实拓扑。这是「AI 原生 OpenTelemetry APM」与「APM + 聊天框」的本质区别。

**开放生态：** 支持多模型（OpenAI 兼容、Anthropic 等即配即用）；支持 MCP 协议（外部 Agent 如 Cursor / Claude 可通过标准 `/mcp` 端点调用平台 APM 工具，平台也可作 MCP 客户端接入外部服务）；Skill 可定制，内置 Skill 可覆盖。

### 2.5 数据流地址（默认）
|用途|Docker|Kubernetes|
|---|---|---|
|默认账号|`admin` / `Databuff@123`|`admin` / `Databuff@123`|
|OTLP gRPC|`<本机IP>:4317`|`<节点IP>:30417`|
|OTLP HTTP|`http://<本机IP>:4318`|`http://<节点IP>:30418`|

K8s 集群内 Agent 上报：`http://ai-apm-ingest:4318`（gRPC `ai-apm-ingest:4317`）。

## 三、核心优势与亮点
> 本章是面向技术选型的浓缩版，便于快速决策。架构基础见第二章，与具体竞品的逐项实测对照见第四章。

### 3.1 六大核心优势
|#|优势|说明|
|---|---|---|
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
|阈值告警|△ 靠后端 YAML / MQE 表达式维护|✅ 阈值规则可在平台内管理|
|智能告警|❌|✅ 与 APM 指标联动|
|告警事件列表|△|✅（等级 / 服务 / 时间）|
|告警落到服务 / 中间件|△ 触发后多靠 hooks 通知|✅ 列表直接挂服务 / 中间件，可回 APM 下钻|

### 4.4 总体对比
|对比维度|传统 APM（SkyWalking 等）|DataBuff|
|---|---|---|
|部署复杂度|组件多、资源重|**3 组件，极简部署，一条命令**|
|排障方式|人翻图表|**对话式智能分析**|
|接入标准|私有协议为主|**原生 OpenTelemetry / OTLP**，并可接管现有 SkyWalking / Pinpoint / Jaeger Agent|
|数据归属|私有化|**开源可私有化，数据在自己手里**|

### 4.5 对比 Pinpoint（v0.1.4 vs Pinpoint 3.1.0）
同机双跑：DataBuff 走 OTLP `:4318`，Pinpoint 走官方 quickstart Java Agent（Web `:18080` / Demo `:18085`）。

**AI 能力**：Pinpoint 无 AI 平台，7 大能力与 MCP / Skill / 自定义专家全部 ❌，DataBuff 全部 ✅。


