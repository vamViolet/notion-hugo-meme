---
title: "博客主题从 MemE 切换到 Stack 实录"
description: ""
date: 2026-08-12T12:45:00+08:00
image: ""
math: false
license:
comments: true
draft: false
build:
    list: always
tags : [Hugo, Notion, GitHub-Actions]
categories: 技术
lastmod: 2026-08-13T12:21:00+08:00
---
> 一次完整的话题切换实操记录：从 MemE 换到 hugo-theme-stack，顺带把被钉死的 Hugo 老版本解锁。包含三个真实踩坑和 CI 迭代验证过程。

## 一、背景
博客原本用 MemE 主题（git submodule，钉在 2022 年的 v5.0.0）。MemE v5.0.0 用了旧版 `resources.ToCSS` API，与新版 Hugo 不兼容，导致 CI 工作流被迫把 Hugo 钉在 `0.112.7`——这是个 2023 年的老版本，越来越难维护，新特性也用不了。

目标：换成更现代的主题，并顺带解除旧 Hugo 版本锁。

## 二、选型
候选了三个主流主题，最终选 **hugo-theme-stack**：

- 卡片式文章列表、左侧边栏、原生暗色模式切换
- 维护活跃，文档完善
- 要求 Hugo ≥ 0.157.0 extended，正好顺带升级

## 三、改动清单
### 1. 替换主题 submodule
移除 MemE，新增 Stack 并锁定到稳定 tag：

```bash
git submodule deinit -f themes/meme
git rm -f themes/meme
rm -rf .git/modules/themes/meme
git submodule add -b master https://github.com/CaiJimmy/hugo-theme-stack.git themes/stack
git -C themes/stack checkout v4.0.3
```
### 2. 重写 config.toml
从 1432 行的 MemE 配置精简到约 210 行。关键调整：

- `theme = "stack"`
- `mainSections = ["zh"]`：核心，让首页读 `content/zh/`（Notion Action 同步目录，不改 Action 也不改清理脚本）
- `[permalinks] zh = "/p/:slug/"`：用 Stack 默认 URL 风格，文章 URL 从 `/zh/xxx/` 变 `/p/xxx/`
- `[menu]` 改成 Stack 的 `main` + `social` 格式
- 删除 MemE 专属的 `[params]`、`[outputFormats]`、`[outputs]` 等几百行

### 3. 重写 archetypes/notion.md
Notion Action 同步时套这个模板。改成 Stack 的 front matter 字段，两个关键点：

- `draft: false`：Stack 默认 archetype 是 `draft: true`，不显式改 false 的话同步出的文章会被当草稿隐藏
- `category` → `categories`（复数）：Stack 模板用 `.Params.categories` 判断分类，单数 key 不会显示分类徽章

### 4. 升级 Hugo
工作流里 `0.112.7` → `0.163.3`，`extended: true` 保留。

## 四、踩坑记录
整个过程走了三轮 CI 迭代，每个坑都是一次失败构建。

### 坑 1：计划里的 Hugo 0.157.2 根本不存在
第一轮按计划填 `hugo-version: '0.157.2'`，`peaceiris/actions-hugo` 直接报错：

```
Unable to find a compatible Hugo release asset for this runner.
```
查 Hugo releases 才发现：0.157 线**只有 0.157.0**，之后直接跳到 0.160+，根本不存在 0.157.1/0.157.2。改成实际存在的 `0.163.3`（成熟 .3 补丁版本，远高于 Stack 的 min 0.157.0）解决。

教训：版本号要查 releases 确认存在，别想当然填补丁号。

### 坑 2：菜单图标不在 Stack 内置集
第二轮 Hugo 装上了，但构建报：

```
ERROR Error: icon 'code.svg' is not found under 'assets/icons' folder
ERROR Error: icon 'notes.svg' is not found under 'assets/icons' folder
```
菜单里 `技术` 用了 `icon = "code"`、`随笔` 用了 `icon = "notes"`，但 Stack 只内置 24 个 SVG（`themes/stack/assets/icons/`），不含 code/notes。改成内置图标：技术→`categories`、随笔→`archives`、关于→`user`。

教训：主题里引用的资源（图标、图片）要先确认主题是否自带，没有就得自己往 `assets/` 放。

### 坑 3：languageCode 弃用警告
第三轮构建成功，但有 WARN：

```
WARN deprecated: project config key languageCode was deprecated in Hugo v0.158.0
```
修这个又踩了一层：顶层 `languageCode` 在 0.158 弃用，移到 `[languages.zh].languageCode` 后，0.163 又把这个键改名为 `locale`（连同 `languageName` → `label`）。最终用最新键名：

```toml
[languages]
    [languages.zh]
        locale = "zh-CN"
        label = "中文"
        weight = 1
```
额外风险：显式加 `[languages.zh]` 后，Hugo 可能把 `content/zh/` 当成 zh 语言根目录（而非 section），导致 `mainSections=["zh"]` 失配、首页变空。本地用下载的 hugo 二进制构建验证过：仍是 section，首页 2 篇文章、`/p/` URL 都没变，才敢推 CI。

## 五、验证方法
这次没用本地 `hugo server` 预览，直接靠 CI 验证：

1. `git push` 后用 GitHub API 手动触发 `workflow_dispatch`
1. 轮询 run 状态到 completed
1. 失败就下载日志 zip，定位报错步骤，修复重推
1. 成功后抓线上首页 HTML，核对 Stack 标记、文章链接、CSS/JS 资源是否 200

三轮迭代，每轮约 2–3 分钟，比反复起本地服务更快也更接近真实部署环境。

## 六、最终效果
线上 https://notion.dongxiaoqi.top/ 已确认：

- Stack 外观：左侧边栏、卡片式文章列表、暗色模式切换
- 首页正常显示 2 篇文章，URL 是 `/p/xxx/`
- 文章页正文、标题、代码块、标签、阅读时长都正常
- `/tags/` 三个标签（Hugo / Notion / GitHub-Pages）正常
- CSS（55KB）/ JS（8KB）资源都 200，无样式崩坏
- 旧 `/zh/xxx/` URL 已 404（符合预期，文章少无外链，可接受）

## 七、语言切换器的实现
主题切换上线后，照着 demo 站（demo.stack.cai.im）加了侧边栏的语言切换下拉框。这一节记录实现机制和一个非标准目录结构下踩的坑。

### 机制：hugo.IsMultilingual
Stack 的切换器在 themes/stack/layouts/_partials/sidebar/left.html，是一个 <select> 下拉框，遍历 .Site.Home.AllTranslations 生成 option，每个指向对应语言首页 URL、label 用 .Language.Label。但它被包在一个前置条件里：

```go
{{ if hugo.IsMultilingual }}
    <li id="i18n-switch"> ... <select> ... </li>
{{ end }}
```
也就是说，只有 Hugo 处于多语言模式（[languages] 里定义了 ≥2 种语言）时切换器才渲染。demo 站每种语言都有真实翻译内容，多语言模式天然开启，切换器自然出现。单语博客要显示它，就得显式加第二种语言。

### 坑：content/zh 当 section 用，content/en 进不了英语语言
本站的内容目录结构是非标准的：文章在 content/zh/，靠 mainSections=["zh"] + permalinks zh="/p/:slug/" 让首页读到它、文章落在 /p/。这是为了不动 Notion Action 硬编码的 contentFolder。

加第二种语言时本能地想建 content/en/ 放英语内容——但本地 hugo 实测发现行不通：默认语言 zh 且 defaultContentLanguageInSubdir=false 时，Hugo 从 content/ 根扫描，把 content/ 下所有子目录（含 content/en/）都当成 zh 的 section。结果 content/en/ 永远成不了英语语言根，英语首页 RegularPages=0，空列表。

换成 defaultContentLanguageInSubdir=true 也没用，content/en/ 依旧被默认语言吞成 section（变成 /zh/en/posts/...）。这个坑用调试 partial 打印 .Site.RegularPages 的 type/url 才定位到。

### 解法：by-filename 后缀 .en.md
改用 Hugo 的 by-filename 机制：文件名带 .<lang>.md 后缀的会归对应语言，且继承所在目录的 section。英文欢迎页放在 content/zh/welcome.en.md——.en 后缀让它归英语语言，content/zh/ 目录让它共享 section "zh"，于是 mainSections 和 permalinks 完全不用改。

效果：中文文章仍在 /p/:slug/（URL 零变化），英文欢迎页在 /en/p/welcome/，切换器选项 中文(/) ↔ English(/en/)。

### 配套改动
- config.toml 加 [languages.en]（locale="en-US", label="English", weight=2），开启 IsMultilingual，切换器自动渲染。
- 新建 content/zh/welcome.en.md：英文欢迎页，说明本站主语言为中文、英文版建设中，附返回中文首页的链接。
- 工作流清理脚本加 TRANSLATION_SUFFIXES=(".en.md",) 跳过翻译文件——否则每次 CI 会 git rm 掉欢迎页（Notion Action 不产生 .en.md，它不在 expected 集合里）。已实测：found 2 .md file(s); 0 to delete，欢迎页安全。

### 后续加英文内容
想加英文文章，按同样规则放 content/zh/<name>.en.md 即可，自动出现在 /en/ 列表、URL 为 /en/p/<name>/。若英文文章多了，建议重新评估是否把 Notion 同步改成真正的多语言结构（那是另一个项目）。

## 八、遗留事项
- **分类徽章不显示**：现有两篇文章 front matter 是 `category: 首页`（单数 key），Stack 读 `categories`（复数）。新 archetype 已改复数，下次 Notion 重新同步时自动修好。
- **Notion token 安全**：集成 token 曾在对话中明文出现，建议去 Notion 重新生成并更新 GitHub 仓库的 `NOTION_TOKEN` secret。

## 九、关键命令速查
触发 CI 构建（用仓库存储的 GitHub 凭据，无需 gh CLI）：

```bash
# 取 git credential 里的 token
TOKEN=$(printf "protocol=https\nhost=github.com\n\n" | git credential fill | grep '^password=' | sed 's/^password=//')
# 手动触发 workflow
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -d '{"ref":"master"}' \
  https://api.github.com/repos/vamViolet/notion-hugo-meme/actions/workflows/notion-blog.yml/dispatches
```
下载某次 run 的日志（API 返回 302 到 S3，要跟跳转）：

```bash
REDIR=$(curl -D - -o /dev/null -H "Authorization: Bearer $TOKEN" \
  https://api.github.com/repos/vamViolet/notion-hugo-meme/actions/runs/<RUN_ID>/logs \
  | grep -i '^location:' | sed 's/^[Ll]ocation: //')
curl -o logs.zip "$REDIR"
unzip logs.zip
```

