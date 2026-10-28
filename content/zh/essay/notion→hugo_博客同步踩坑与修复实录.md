---
title: "Notion→Hugo 博客同步踩坑与修复实录"
description: ""
date: 2026-10-28T08:36:00+08:00
image: ""
math: false
license:
comments: true
draft: false
build:
    list: always
tags : [随笔]
categories: 随笔
lastmod: 2026-10-28T08:36:00+08:00
---
# Notion→Hugo 博客同步踩坑与修复实录
我的博客是「Notion 写作 → GitHub Actions 自动同步 → Hugo 构建 → GitHub Pages」这一套。日常体验很爽：在 Notion 里写完，把状态一改，两小时后文章就上线了。

直到前几天，一篇文章发出来只有一半，另一篇文章的图片全是裂图。排查下来发现，这个用了很久的同步 Action 里藏着三个 bug，都是**静默失败**——构建一路绿灯，页面看着也"正常"，只是内容悄悄少了。

这篇记录完整的排查过程和修复方案，希望能帮到用同款方案的人。

## 一、流水线长什么样
```
Notion 数据库 ──(rxrw/notion-blog Action)──> content/zh/*.md ──(Hugo)──> GitHub Pages
      │                                              │
      └─ Status=Finished 的文章被拉成 md              └─ https://notion.dongxiaoqi.top/
      └─ 同步成功后 Status 自动改成 Published
```
关键机制：

|机制|说明|
|---|---|
|触发|每 2 小时定时 + 手动 `workflow_dispatch`|
|捞取条件|`Status = Finished`，同步成功后自动置为 `Published`|
|幂等性|已 `Published` 的文章不会再被同步 —— **更新内容时必须把状态改回 `Finished`**|
|清理|工作流会删掉"既非 Finished 也非 Published"的文章 md（下线用）|

## 二、现象一：文章只显示到第 4.5 章
### 复现
一篇 900 多行的长文（DataBuff 使用指南），线上只渲染到「4.5 对比 Pinpoint」的第一段就结束了。

### 排查：三方对齐
第一步是把三个地方的内容摆在一起看：

|对象|内容|
|---|---|
|线上 HTML|断在「4.5 …DataBuff 全部 ✅。」|
|仓库 `content/zh/databuff介绍与使用指南.md`|242 行，末行就是这一句|
|Notion 原页面|**309 个顶层块**，完整（含五~八章、参考链接）|

线上和仓库 md 一致，Notion 完整 ⇒ 问题出在 **Notion → md 的同步环节**，不是 Hugo、不是部署。

### 定位：断点是第 100 个块
我写脚本把 Notion 页面的块按顺序编号 dump 出来，发现断掉的那段话正好是**第 100 个顶层块**：

```
98  heading_3   4.5 对比 Pinpoint（v0.1.4 vs Pinpoint 3.1.0）
99  paragraph   同机双跑：DataBuff 走 OTLP :4318 …
100 paragraph   AI 能力：Pinpoint 无 AI 平台，7 大能力与 MCP / Skill …  ← 线上到此为止
101 paragraph   APM 应用性能：…
```
100 是个太可疑的数字：Notion 的 `GET /v1/blocks/{id}/children` 接口**单次最多返回 100 个块**。

### 根因：只取了第一页，没有翻页
问题出在 Action 的 `internal/parser.go`：

```
func recursiveGetChildren(client *notionapi.Client, blockID notionapi.BlockID) (blocks []notionapi.Block, err error) {
	res, err := client.Block.GetChildren(context.Background(), blockID, &notionapi.Pagination{
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}

	blocks = res.Results        // ← 拿完就走，完全没看 res.HasMore / res.NextCursor
	...
}
```
Notion 的分页响应里带了 `has_more` 和 `next_cursor`，需要循环取完。原实现只取第一页，**超出 100 块的部分被静默丢弃**，而且不报任何错。

顺便发现数据库查询也是同样的问题（`PageSize: 100` 不翻页），文章数超过 100 篇就会漏文章。

### 修复
```
	// Notion 的 children 接口单次最多返回 100 个块，必须跟随 next_cursor 翻页。
	var cursor notionapi.Cursor
	for {
		var res *notionapi.GetChildrenResponse
		res, err = client.Block.GetChildren(context.Background(), blockID, &notionapi.Pagination{
			PageSize:    100,
			StartCursor: cursor,
		})
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, res.Results...)

		if !res.HasMore || res.NextCursor == "" {
			break
		}
		cursor = notionapi.Cursor(res.NextCursor)
	}
```
数据库查询同理，用 `DatabaseQueryResponse` 的 `NextCursor` 循环累积结果。

## 三、现象二：每张表格的第一行数据都不见了
同一篇文章里，对比表格总是从第二行开始：

- 「2.2 组件职责」少了 `Ingest` 行
- 「2.3 三种信号」少了 `Trace` 行
- 「3.1 六大核心优势」从 `#2` 开始

而 Notion 页面里这些行都在。

### 根因：分隔行"顶掉"了数据行
Markdown 表格的**分隔行**（`|---|---|`）必须紧跟表头，但它本身不是数据行。原实现是这么写的（`pkg/generator.go`）：

```
for index, row1 := range rows {
	row := row1.(*notionapi.TableRowBlock)
	if index == 1 && !headerLine {
		fprintf(w, prefixes, "|%s", strings.Repeat("---|", len(row.TableRow.Cells)))
		headerLine = true
	} else {
		...打印该行...        // ← index == 1 时走不到这里
	}
}
```
`index == 1` 时只打印分隔行、**不打印这一行的内容**，于是每张表的首个数据行被整行吃掉。

### 修复
```
if index == 1 {
	// 分隔行本身不是数据行：先补分隔行，再照常输出当前行
	fprintf(w, prefixes, "|%s", strings.Repeat("---|", len(row.TableRow.Cells)))
}
...照常打印该行...
```
## 四、现象三：图片全部加载不出来
两篇已发布文章的图片全是裂图。

### 线上图片地址长这样
```
/496b8658-988f-4779-9b2e-628ed43ab891/5778eacc-e04e-4d67-95e6-eee1258d2848/893df6b39d41f1e094eabbf7e055db21.png
/sz_mmbiz_png/cgOOdN0fZFNIEUpjILQtKpQuicBjibHG7F0GpWoALE4Zp0zhkfR7miblLFehzoUqqaWnHFen5rwicSEqPO6icyzib3POtYRXn0RniaIX3WCXvODXt0/640
```
注意：**没有域名、没有协议头**，是站点根相对路径 → 必然 404。

第一个是 Notion 图床（`prod-files-secure.s3...`）的 URL 路径，第二个是微信 `mmbiz.qpic.cn` 的路径。它们共同点是：**URL 路径是多级目录**。

### 根因：拿整段 URL path 当文件名，失败后还把它当成图片地址
```
func getImage(imgURL string, config BlogConfig) (_ string, err error) {
	splittedURL, _ := url.Parse(imgURL)
	name := splittedURL.Path        // ← 整段 path：/workspace-uuid/object-uuid/file.png
	...
	out, err := os.Create(filepath.Join(config.ImagesFolder, name))
	if err != nil {
		return name, fmt.Errorf("couldn't create image file: %s", err)   // ← 失败时返回这段 path
	}
	...
	return filepath.Join(config.ImagesLink, name), nil
}
```
两个问题叠加：

1. `os.Create("static/images/" + "/a/b/c.png")` → 父目录 `a/b` 不存在 → **建文件失败**（只 `MkdirAll` 了 `static/images`）；
1. 失败分支把这段 path 当返回值返回，而调用方 **丢掉了错误**：

   ```go    src, _ := getImage(b.Image.GetURL(), config)    fprintf(w, prefixes, "![%s](%s)\n", ..., src)    ```    于是 md 里就写成了 `![](/496b8658-.../x.png)`。

反过来，如果图片 URL 的路径只有一段（老来源，如 `/20210525105147724.JPG`），`os.Create` 就能成功 —— 这就是为什么**只有部分文章的图片挂了**。

### 全库扫描：分层验证这个规律
我把数据库里所有文章的图片扫了一遍（含 callout/column 里的嵌套图片），结果和上面的规律完全吻合：

|文章|图片|URL 路径段数|结果|
|---|---|---|---|
|robocopy 批量文件复制优化技巧|3（Notion）|3 段|❌ 全挂|
|[大厂场景系列02] 抢红包|1（微信）|3 段|❌ 挂|
|MyBatisPlus 注解方式实现多表关联查询|2（外站）|2 段|❌ 全挂|
|雪花算法的基本理念和简单实现|2|1 段|✅ 正常|
|skyWalking 自动建表-逻辑梳理|10|1 段|✅ 正常|
|关于 ShardingSphere-JDBC 的简单使用|2|1 段|✅ 正常|

### 修复
```
	base := path.Base(splittedURL.Path)          // 只取最后一段
	if base == "" || base == "." || base == "/" {
		base = "image"
	}
	if decoded, derr := url.PathUnescape(base); derr == nil && decoded != "" {
		base = decoded
	}
	sum := sha1.Sum([]byte(imgURL))               // 不同目录下的同名文件（image.png 之类）很常见
	name := fmt.Sprintf("%x_%s", sum[:4], sanitizeFileName(base))
```
再加两处补强：

```
	// 微信图床的 URL 路径不带扩展名（/640?wx_fmt=png），用响应头补后缀
	if path.Ext(name) == "" {
		name += extFromContentType(resp.Header.Get("Content-Type"))
	}
```
```
	// 下载/建文件失败时，回退成原始绝对 URL，而不是会被当成站点内路径的残缺 path
	return imgURL, fmt.Errorf("couldn't create image file: %s", err)
```
修复后实际落盘效果：

|原 URL 路径|修复后|
|---|---|
|`/496b8658-…/5778eacc-…/893df….png`|`/images/b5b26322_893df6b39d41f1e094eabbf7e055db21.png`|
|`/sz_mmbiz_png/cgOOdN…/640`|`/images/001dcc53_640.png`|
|`/post-images/1623814772810.png`|`/images/2a5db996_1623814772810.png`|

## 五、为什么选择"内置 Action"而不是 Fork
这个 Action（`rxrw/notion-blog`，GPL-3.0）是第三方仓库，直接改不了。两条路：

|方案|取舍|
|---|---|
|Fork 一份再改|干净、可独立维护；但要维护第二个仓库|
|**内置到博客仓库**|仓库内 `.github/actions/notion-blog/`，工作流改成 `uses: ./.github/actions/notion-blog`；无需 fork，版本随仓库走|

因为它是 `runs: using: docker` 类型的 Action，GitHub Runner 会**用仓库里的 Dockerfile 现场构建**，本机连 Go 都不用装。我选了内置，并加了 `VENDORED.md` 记录上游提交号、三处改动和升级方法。

工作流只改一行：

```yaml
      - name: notion-blog
        uses: ./.github/actions/notion-blog     # 原来是 rxrw/notion-blog@master
```
## 六、验证结果
|项|修复前|修复后|
|---|---|---|
|DataBuff 指南 md|242 行（截断）|**941 行**（完整）|
|DataBuff 线上页面|39 KB，只到 4.5 章|**141 KB**，五、部署 / 参考链接 全在|
|表格首行|每张表都少第一行|恢复正常|
|文章图片|6 张 404|6 张全部 **200**，Content-Type 正确|
|抢红包文章|未同步|正常发布|

## 七、可复用的排查套路
1. **三方对齐**：线上渲染结果 / 仓库生成物 / 数据源，先确定"问题出在哪一段链路"，再往下钻。三者一比，责任段立刻清楚。
1. **用序号定位边界**：把块、条目编号 dump 出来。断在「第 100 个」这种整数上，基本就指向分页没做。
1. **规律即线索**：'有些图挂了、有些没挂'时，别逐个看，**把不变量列出来对比**（这里是"URL 路径有几段"），规律会自己浮出来。
1. **全库扫描代替个案修复**：修完不要只验证报障的那两篇，扫一遍全库，把同类问题一次列清楚。
1. **警惕"吞掉错误"的调用**：`src, _ := getImage(...)` 这种写法把失败伪装成成功，比崩溃难查十倍。
1. **静默失败要变成响亮的失败**：分页、下载、建文件这类操作失败时必须留下痕迹，否则构建全绿而内容残缺。

## 八、给自己留的操作备忘
**更新一篇已发布文章：**

1. 在 Notion 里改内容；
1. 把该页 `Status`：`Published` → `Finished`（其它状态会被清理步骤删掉 md）；
1. 等定时任务（≤2 小时）或手动 `Run workflow`；
1. 同步完成后状态自动回到 `Published`，站点重新部署。

**注意：**

- 改标题或换分类会改变 md 的路径，旧文件由工作流的清理步骤自动删除；
- 更新不会改变文章的 `date`，不会因此排到最新；
- 不要直接改仓库里的 md，下次同步会被整份覆盖。

