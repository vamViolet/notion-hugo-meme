# 内置说明（vendored）

本目录是 **[rxrw/notion-blog](https://github.com/rxrw/notion-blog)** 的内置副本，
上游提交：`0d0181a36937aed7bb570e4785643c89335c7737`
（"Add support for table, video. Upgrade notionapi to latest version.", GPL-3.0 —— 见同目录 `LICENSE`）。

内置原因：上游 Action 会把长文静默截断，且表格会丢首个数据行（详见下）。放在仓库里便于
本地修订，不再依赖上游是否修复。

## 相对上游的改动

| 文件 | 改动 |
|---|---|
| `internal/parser.go` · `recursiveGetChildren` | 原来只调一次 `GetChildren(PageSize: 100)` 且不看 `HasMore`/`NextCursor`，顶层块超过 100 的页面**只同步前 100 块**（Observed：文章在正文中途断掉，构建却“成功”）。现改为 join 游标循环取全部页。 |
| `internal/parser.go` · `ParseAndGenerate` | 数据库查询同样是 `PageSize: 100` 不翻页，文章数超过 100 会漏文章。现同样按 `NextCursor` 翻页。 |
| `pkg/generator.go` · `TableBlock` 分支 | Markdown 分隔行必须紧跟表头，但**不是数据行**。原实现在 `index == 1` 时只打印分隔行而不打印该行内容，导致每张表的首个数据行被整行吞掉。现改为先补分隔行、再照常输出该行。 |

## 升级方式

上游有更新时，重新拉取对应提交覆盖本目录，再把上表三处改动重新打一遍即可。

## 排查记录

- 现象：博客文章只显示到某一章，页面 HTML 与仓库 `content/zh/*.md` 都断在同一处，Notion 原页面内容完整。
- 定位：断点正好是页面第 **100** 个顶层块 → 命中第一处改动；表格首行缺失命中第三处改动。
