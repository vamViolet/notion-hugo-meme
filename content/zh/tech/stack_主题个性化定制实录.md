---
title: "Stack 主题个性化定制实录"
description: ""
date: 2026-08-13T11:12:00+08:00
image: ""
math: false
license:
comments: true
draft: false
build:
    list: always
tags : [Hugo, Stack, GitHub-Pages]
categories: 技术
lastmod: 2026-08-13T12:21:00+08:00
---
## 一、背景
Stack 主题切换上线后（见《博客主题从 MemE 切换到 Stack 实录》），博客外观与基础链路就绪。此后围绕头像、导航、视觉细节、字体、评论与社交做了一系列个性化定制，本文记录这些改动及其背后的取舍与踩坑。整个过程延续「无本地 hugo、直接靠 CI 验证」的工作方式，每个改动都是一次 push → 触发构建 → curl 线上确认的闭环。

## 二、改动清单
### 1. 博客头像
用户提供了一张《猫和老鼠》指挥家横图（1102×689）。Stack 侧边栏头像经 helper/image.html 的 .Resources.Get 解析，而 .site-logo 的 CSS 没有 object-fit（默认 fill 会拉伸），直接用横图会被拉变形。

处理：用 Pillow 把原图中心裁成正方形 689×689，存 assets/img/avatar.jpeg，config 里 [params.sidebar] avatar = "img/avatar.jpeg"，原图备份为 avatar-source.jpeg。

```python
from PIL import Image
img = Image.open("avatar-source.jpeg")
w, h = img.size
s = min(w, h)
left = (w - s) // 2
img.crop((left, 0, left + s, s)).save("avatar.jpeg")
```
关键点：头像必须在 assets/ 下（Stack 用 Hugo Resources 解析，不是 static/），且文件扩展名要与实际文件一致，否则 .Resources.Get 匹配不到。

### 2. 分类菜单 404 修复
切换后点侧边栏「技术 / 随笔 / 关于」全跳 404。排查发现菜单 url 配的是 /categories/tech/（categoryMap 的小写英文 key），但 Notion Action 把文章的 Notion Category（中文 select 名「技术/随笔/关于」）原样写进 front matter 的 categories: 字段——categoryMap 只决定 content/zh 下的子目录名，不影响 front matter。Hugo 按 front matter 的 categories 生成分类页，实际路径是中文 /categories/技术/，与菜单对不上。

```bash
# 实测验证
curl /categories/tech/   -> 404
curl /categories/技术/   -> 200
```
修复：把菜单三个 url 改成 Hugo 实际生成的中文路径 /categories/技术/、/categories/随笔/、/categories/关于/。随笔暂无文章，/categories/随笔/ 仍 404，等有文章后自动出现，保留菜单项占位。

教训：Notion→Hugo 链路里，分类页 URL 由 front matter 的 categories 决定，不由文件目录决定；改菜单前先 curl 确认 Hugo 实际生成的路径。

### 3. 页脚跳动的心
参考 guanqr.com，在版权行末尾加一颗 Font Awesome 实心心形，用 fa-heartbeat 双跳动画（1.3s 循环）。通过覆盖主题 partial 实现，不动 submodule：

- layouts/_partials/footer/footer.html：版权行追加 <svg class="heart-icon">（FA 心形路径）
- layouts/_partials/footer/custom.html：注入心跳 @keyframes CSS（主题留空的扩展点）

```css
@keyframes heartbeat {
  0%   { transform: scale(1); }
  14%  { transform: scale(1.3); }
  28%  { transform: scale(1); }
  42%  { transform: scale(1.3); }
  70%  { transform: scale(1); }
  100% { transform: scale(1); }
}
```
细节：心形红色 #ff4d4f，1em 尺寸随字号缩放，vertical-align 对齐基线；加 @media (prefers-reduced-motion: reduce) 守卫，系统开了「减少动态效果」时停止动画（无障碍）。

### 4. 字体：中文霞鹜文楷 + 代码 JetBrains Mono
目标是中文用霞鹜文楷、代码用 JetBrains Mono。采用混合方案：JetBrains Mono 走 Google Fonts CDN，霞鹜文楷只设 font-family 名走系统回退——用户本地装了才显示，省去几 MB webfont 的网络开销。

Stack 主题的字体由三个 CSS 变量控制（在 variables.scss）：--base-font-family（界面）、--article-font-family（正文，继承 base）、--code-font-family（代码）。覆盖方式：

- layouts/_partials/head/custom-font.html：把加载的 Google Fonts 从 Lato 换成 JetBrains Mono
- assets/scss/custom.scss：重定义三个 CSS 变量

```css
:root {
  --base-font-family: "LXGW WenKai Screen", "LXGW WenKai", "霞鹜文楷",
      -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif;
  --article-font-family: var(--base-font-family);
  --code-font-family: "JetBrains Mono", "LXGW WenKai Mono",
      Menlo, Monaco, Consolas, monospace;
}
```
生效原理：style.scss 里 @import "custom.scss" 在 variables.scss 之后，我的 :root 重定义排在主题默认值之后，CSS 变量后者覆盖前者。霞鹜文楷回退链：LXGW WenKai Screen → LXGW WenKai → 霞鹜文楷 → 系统无衬线，未装则回退 PingFang SC / 微软雅黑。

踩坑：custom.scss 是静态资源，不经过 Hugo 模板引擎，只能用 CSS 注释，不能写 {{}}（否则 SCSS 编译报错）。

### 5. 评论系统 Giscus + 社交链接
评论选 Giscus（基于 GitHub Discussions，与 GitHub Pages 技术栈契合，免费无广告）。社交在原有 GitHub + RSS 基础上加 CSDN 和 Email。

Giscus 需要 repo-id 和 category-id。这两个不必走 giscus.app 配置页，直接用 GitHub API 取：

```bash
# 启用 Discussions（REST PATCH）
curl -X PATCH -H "Authorization: token $TOKEN" \
  https://api.github.com/repos/vamViolet/notion-hugo-meme \
  -d '{"has_discussions": true}'

# 取 repo-id 和 category-id（GraphQL）
# query: repository(owner,name){ id, discussionCategories(first:20){ nodes{ id name slug } } }
```
拿到 repo-id R_kgDOTllGGw、Announcements 分类 id DIC_kwDOTllGG84DDR7I，写入 config [params.comments.giscus]。mapping=title 按标题关联 discussion，文章 URL 变了也不丢评论；lightTheme/darkTheme 跟随博客明暗主题。

关键点：Stack 单篇文章的 comments front matter 字段会覆盖全局 [params.comments.enabled]。所以光开全局不够，还要把 archetype 和现有文章的 comments: false 全改成 true（9 篇批量 sed）。

社交：Stack 侧边栏图标经 helper/icon.html 的 resources.GetMatch "icons/<name>.svg" 解析，主题内置图标没有 email/csdn，于是自己画了 assets/icons/email.svg（Tabler mail 风格）和 assets/icons/csdn.svg，在 [menu.social] 引用。

## 三、踩坑记录
### 坑 1：Docker Action 生成文件的权限
某次同步新文章，fix-tags 步骤报 PermissionError: [Errno 13] Permission denied。根因：rxrw/notion-blog 是 Docker action，在容器内以 root 生成 content/zh、static/images 文件，归 root 所有；后续 fix-tags / fix-dates 以 runner 用户运行，写或删这些文件就失败。这个潜在 bug 只有在需要改新文件时才暴露——只读的步骤会蒙混过去，所以前面几次构建一直没发现。

修复：在 notion-blog 步骤之后、cleanup 之前加 chmod -R a+rwX content/zh static/images，一次性放开权限。

### 坑 2：Notion created_time 不可控，迁移文章日期全变成今天
迁移 9 篇 Gridea 老文章时，Notion Action 用页面 created_time 作为文章 date。但 created_time 是只读系统字段，API 创建时自动设为 now，无法覆盖。结果 2021/2023 的老文章同步后全变成今天的日期，堆在归档页顶部当「最新」。

解法：加自愈 workflow 步骤——scripts/migrated-dates.json（title→原日期映射）+ scripts/fix-migrated-dates.py，每次 sync 在 fix-tags 之后、commit 之前重写匹配文章的 date/lastmod 回原日期。幂等，每次同步都跑，文件已在正确日期则不动。

### 坑 3：Notion 链接必须是绝对 URL
创建 Notion 页面时，两篇文章 block append 报 Invalid URL for link。排查是源 HTML 里有相对锚点（#fn1 脚注、%E7%9B%AE%E5%BD%95 编码的「目录」TOC 链接）。Notion 的 link 字段只接受绝对 http(s) URL。修复：在转换器 text_seg 里加 valid_url 校验，非法链接保留文字、丢弃 link，避免阻断整批 block 写入。

### 坑 4：mybatisplus 图片 404
2 张 mybatisplus 图片原图在 vamViolet.github.io/post-images/，Action 的 getImage 把文件存到 static/images/post-images/ 子目录时 os.Create 失败（子目录不存在），回退写出裸路径 /post-images/X.png 导致 404。解法：Hugo 把 static/ 全部映射到站点根，直接把图放到 static/post-images/X.png 即可让 /post-images/X.png 解析。文章 Status=Published 不会再被 Action 覆盖，此补丁持久。

## 四、验证方法
全程无本地 hugo，靠 CI 验证：改完 push → workflow_dispatch 触发 → 轮询 run 状态 → 成功后 curl 线上页面抓 HTML/CSS 确认元素和变量生效。

```bash
# 触发构建（用 git credential 里的 token，无需 gh CLI）
TOKEN=$(printf "protocol=https\nhost=github.com\n\n" | git credential fill | sed -n "s/^password=//p")
curl -X POST -H "Authorization: token $TOKEN" \
  https://api.github.com/repos/vamViolet/notion-hugo-meme/actions/workflows/notion-blog.yml/dispatches \
  -d '{"ref":"master"}'

# 验证字体变量在编译后 CSS 里（minify 后有两个 :root 定义，后者覆盖前者）
curl https://notion.dongxiaoqi.top/scss/style.min.<hash>.css | grep -o 'JetBrains Mono'
```
## 五、最终效果
线上 https://notion.dongxiaoqi.top/ 已确认：

- 侧边栏头像（Tom & Jerry 指挥家，正方形裁剪不拉伸）
- 技术 / 关于 菜单可正常打开（中文分类页），随笔待文章
- 页脚版权行末尾跳动的心（红色，1.3s 双跳）
- 中文霞鹜文楷（系统装了才显示）、代码 JetBrains Mono
- 文章页 Giscus 评论框（参数齐全，待装 App）、侧边栏 CSDN / Email / GitHub / RSS 四图标

## 六、遗留事项
- Giscus App 待手动安装：https://github.com/apps/giscus → Install → 授权 notion-hugo-meme 仓库，否则评论框报错
- Notion token 在对话历史明文出现过，建议轮换并更新 GitHub NOTION_TOKEN secret
- 随笔分类待补文章，/categories/随笔/ 目前 404

