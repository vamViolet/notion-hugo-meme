---
banner: ""
title: "notion-hugo-meme 操作手册"
date: 2026-07-29T03:05:00+08:00
tags : [Hugo, Notion, GitHub-Pages]
category: 首页
description: ""
lastmod: 2026-08-12T11:01:00+08:00
author: Github Blog
toc: false
gallery: false
---
# notion-hugo-meme 操作手册
本项目把 Notion 数据库作为内容源，通过 GitHub Actions 自动同步为 Markdown，再用 Hugo 构建成静态博客并部署到 GitHub Pages。本文档梳理项目结构，并给出从「写文章」到「上线/下线」的完整操作流程。

> 仓库地址：https://github.com/vamViolet/notion-hugo-meme 博客地址：https://notion.dongxiaoqi.top/

## 一、整体架构
数据流是单向的、全自动的（每 2 小时触发一次，也可手动触发）：

```
Notion 数据库 ──(rxrw/notion-blog Action)──> content/zh/*.md ──(Hugo build)──> GitHub Pages
                      │                                                │
                      └─ 把 Status=Finished 的文章拉成 md               └─ https://notion.dongxiaoqi.top/
                      └─ 同步成功后把 Status 改成 Published
                      └─ 清理步骤：删除已下线文章的 md
```
三个关键角色：

- **Notion 数据库**：唯一的内容编辑入口。每篇文章是一条记录，`Status` 控制是否上线。
- **GitHub 仓库**：存放 Hugo 站点源码、同步出来的 md、以及自动化工作流。
- **GitHub Pages**：托管构建产物，对外提供博客访问。

## 二、项目结构
```
notion-hugo-meme/
├── .github/workflows/notion-blog.yml   # 核心：同步 + 清理 + 构建 + 部署 的工作流
├── config.toml                          # Hugo 站点配置（站名、域名、主题、菜单等）
├── notionblog.config.json               # notion-blog Action 的配置（数据库 ID、字段映射、过滤规则）
├── archetypes/
│   ├── default.md                       # `hugo new` 默认模板
│   └── notion.md                        # Notion 同步时套用的 front matter 模板
├── content/zh/                          # 同步出来的文章 md（按 Category 分子目录）
│   └── 建筑服务能力开发.md
├── static/
│   ├── CNAME                            # 自定义域名（notion.dongxiaoqi.top）
│   └── images/                          # 文章图片（Action 自动下载到这里）
├── themes/meme/                         # MemE 主题（git submodule，指向 rxrw/hugo-theme-meme）
├── resources/                           # Hugo 构建缓存
└── README.md
```
### 关键配置文件
**`notionblog.config.json`** —— 决定 Action 怎么从 Notion 拉文章：

|字段|含义|
|---|---|
|`filterProp` / `filterValue`|只同步 `Status` 为 `Finished` 的文章|
|`publishedValue`|同步成功后把 Status 改成 `Published`|
|`propertyCategories` / `categoryMap`|`Category` 字段映射：技术→tech、随笔→essay、关于→about，首页→根目录|
|`contentFolder`|同步到的目录：`content/zh`|
|`imagesFolder` / `imagesLink`|图片下载到 `static/images/`，链接前缀 `/images`|

**`config.toml`** —— Hugo 站点配置要点：

- `baseURL = "https://notion.dongxiaoqi.top/"`：自定义域名，构建时不能用 github.io URL 覆盖，否则内链全 404。
- `title = "dongxiaoqi's Blog"`：站点名。
- `theme = "meme"`：使用 MemE 主题（submodule，v5.0.0，2022 年版）。

**`.github/workflows/notion-blog.yml`** —— 两个 Job：

1. `auto-sync-from-notion-to-github`：checkout → 跑 notion-blog Action 拉文章 → 清理已下线文章 → 提交推送。
1. `build-and-deploy`：拉最新提交 → Hugo 0.112.7 extended 构建 → 部署到 GitHub Pages。

## 三、Notion 数据库字段
每篇文章（一条记录）需要这些属性：

|属性|类型|作用|
|---|---|---|
|`Status`|状态（status）|控制上线：`Finished`/`Published` = 在线，其余 = 下线|
|`Category`|单选（select）|分类目录：`首页`/`技术`/`随笔`/`关于`|
|`Tags`|多选（multi_select）|标签|
|`Created by`|创建者（created_by）|Action 读取作者名（硬编码依赖，缺了会报错）|

### Status 状态机
|Status|含义|博客上是否可见|
|---|---|---|
|`In progress`（进行中）|写到一半|否|
|`Finished`|写完待发布 → 下次同步会拉取并上线|是（同步后）|
|`Published`|已上线（同步后 Action 自动置为此状态）|是|
|`Offline`（已下线）|主动下线 → 清理步骤会删掉对应 md|否|

> 注意：博客是纯静态站点，没有登录鉴权，所以「私密」等价于「下线」——只要 Status 不是 Finished/Published，文章就不会出现在博客上。

## 四、日常操作
### 4.1 发布一篇新文章
1. 在 Notion 数据库里 **New** 一条记录，填好 `Name`（标题）、`Category`、`Tags`。
1. 在正文里写内容（支持 Markdown 风格、代码块、表格、图片）。
1. 把 `Status` 设为 **`Finished`**。
1. 等待下次 workflow 触发（最多 2 小时），或去 GitHub 仓库 Actions 页面手动 `Run workflow`。
1. 同步成功后，文章出现在博客，Status 自动变成 `Published`。

### 4.2 修改已发布文章
直接在 Notion 改正文/标题/分类，保持 `Status` 为 `Finished` 或 `Published`，下次同步会更新对应 md。

> 想立即触发同步，又不想等 2 小时：GitHub 仓库 → Actions → `notion-blog` → `Run workflow`。

### 4.3 下线一篇文章
把 `Status` 从 `Finished`/`Published` 改成 **`Offline`**（或 `Todo`/`In progress`）即可。下次 run 的清理步骤会自动 `git rm` 掉对应 md，文章从博客消失。

### 4.4 改分类 / 改标题
- **改分类**：在 Notion 改 `Category`。注意：旧分类目录下的旧 md 不会自动删除，需手动清理或等清理步骤处理（清理是按「当前 Notion 里 Finished/Published 的文章」对账的）。
- **改标题**：会生成新文件名的 md，旧文件名的 md 会成为孤儿，下次清理步骤会自动删除。

## 五、首次部署 / 排错清单
如果同步或部署出问题，按这个顺序排查：

1. **Notion 数据库是否关联了 Integration**：数据库右上角 `...` → `Connections` → 添加你创建的 Integration。没关联会 404。
1. **GitHub Secret 是否正确**：仓库 Settings → Secrets and variables → Actions，需有 `NOTION_TOKEN`（值是 Notion Integration token，`ntn_` 开头）。
1. **Notion 字段是否齐全**：`Name`/`Status`/`Category`/`Created by` 必须存在且类型正确，缺了 Action 会报错（但 Action 会吞掉错误、显示成功，要看 Actions 日志才能发现）。
1. **GitHub Pages 源是否设为 Actions**：仓库 Settings → Pages → Source 选 `GitHub Actions`（不是 gh-pages 分支）。
1. **域名/CNAME**：`static/CNAME` 必须是裸域名 `notion.dongxiaoqi.top`，不能带 `http://`。
1. **Hugo 版本**：工作流固定用 `0.112.7 extended`。MemE v5.0.0 与新版 Hugo 的 `resources.ToCSS` 不兼容，**不要升到 latest**，除非先升主题 submodule。

## 六、已知限制
- **Action 只增不删**：rxrw/notion-blog 只会新增/更新 md，不会删除。下线靠本仓库自加的清理步骤实现。
- **正文只取前 100 个 block**：Action 不做 block 分页，超长文章（>100 个 Notion block）会被截断。
- **错误被吞**：Action 内部出错只 `log.Println` 后 `exit 0`，永远显示「成功」。排查要看 Actions 的完整日志。
- **属性名硬编码**：Action 代码里写死了 `Name`/`Category`/`Created by` 等英文字段名，改不了。
- **Tags 渲染缺陷：**rxrw/notion-blog 用 **tags : {{.Tags}}** 渲染多选标签，Go 切片默认输出 **[a b c]**（空格分隔无逗号），Hugo 会当成一个名为 “a b c” 的畸形标签。本仓库工作流加了 Fix tags front-matter format 步骤，同步后自动改成逗号分隔 **[a, b, c]**，博客 /tags/ 页和文章页才能正确显示多个独立标签。
- **Tag 命名约束：**Notion 的 tag 名**不要含空格**（如用 GitHub-Pages 而非 GitHub Pages）。Action 输出端是空格分隔，含空格的标签名无法无损还原，会被错误拆成多个。中文标签不受影响（如 生产力）。

## 七、相关链接
- 仓库：https://github.com/vamViolet/notion-hugo-meme
- 博客：https://notion.dongxiaoqi.top/
- 主题：https://github.com/rxrw/hugo-theme-meme
- Action：https://github.com/rxrw/notion-blog

