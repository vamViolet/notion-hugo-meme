# 域名到期处理备忘

> 适用场景：自定义域名 `notion.dongxiaoqi.top`（属于 `dongxiaoqi.top`）到期后，如何让博客继续可访问。
>
> 本文档是**运维备忘**，不发布到博客。到期前主动改最稳；到期后老地址会直接 DNS 解析失败，只能从 GitHub 仓库改配置恢复。

## 当前域名相关配置清单（4 处文件 + 1 处 GitHub 设置）

| # | 位置 | 当前值 | 作用 |
|---|------|--------|------|
| 1 | `config.toml` 第 4 行 `baseURL` | `https://notion.dongxiaoqi.top/` | Hugo 构建时所有内链/CSS/图片的基础 URL |
| 2 | `static/CNAME` | `notion.dongxiaoqi.top` | 告诉 GitHub Pages 这个站点的自定义域名 |
| 3 | GitHub 仓库 `Settings → Pages → Custom domain` | `notion.dongxiaoqi.top` | GitHub 侧的自定义域名绑定 |
| 4 | `.github/workflows/notion-blog.yml` 构建步骤 | 不覆盖 baseURL（用 config.toml 的） | 见下，回退时通常不用改 |
| 5 | `config.toml` 第 107 行 `website` | `https://io-oi.me/` | 作者网站链接（主题原作者残值，与域名无关，可顺手清理） |

> 关键认知：`Status=Offline` 等只控制**博客上线与否**，控制不了**域名解析**。域名到期是 DNS 层面的事，必须改上面这些。

---

## 方案 A：回退到 GitHub 默认域名（推荐，零成本）

放弃自定义域名，用 GitHub Pages 自带的免费地址 `https://vamViolet.github.io/notion-hugo-meme/`。

### 步骤

**1. 改 `config.toml` 第 4 行**
```toml
# 改前
baseURL = "https://notion.dongxiaoqi.top/"
# 改后
baseURL = "https://vamViolet.github.io/notion-hugo-meme/"
```
> 结尾必须有 `/`，路径是 `/仓库名/`。baseURL 写错会导致全站内链、CSS、图片 404。

**2. 删除 `static/CNAME`**
```bash
git rm static/CNAME
```
> CNAME 留着会让 Pages 继续找 `notion.dongxiaoqi.top`，到期后白屏。

**3. GitHub 仓库设置移除自定义域名**
浏览器打开：`https://github.com/vamViolet/notion-hugo-meme/settings/pages`
- `Custom domain` 一栏清空 `notion.dongxiaoqi.top`（或点 Remove），Save。
- `Source` 保持 `GitHub Actions`。

**4. 工作流 `.github/workflows/notion-blog.yml`**
通常**不用改**。当前构建是 `hugo --minify --destination public`，会用 config.toml 里的 baseURL。
> 如果想防止 baseURL 写错、让 Actions 自动注入正确 URL，可把构建改成：
> `hugo --minify --destination public --baseURL "$PAGE_URL"`
> （`$PAGE_URL` 由 `actions/configure-pages` 的 `id: pages` 输出）。非必须。

**5.（顺手）清理 `config.toml` 第 107 行作者网站残值**
```toml
# 改前（主题原作者的网站，不是你的）
website = "https://io-oi.me/"
# 改后
website = ""
```

**6. 提交并推送**
```bash
git add -A
git commit -m "chore: 回退到 GitHub 默认域名（dongxiaoqi.top 到期）"
git push origin master
```

### 改完后访问
- 新地址：**`https://vamViolet.github.io/notion-hugo-meme/`**
- 推送后等 workflow 跑完（约 1-2 分钟）。
- 老地址 `https://notion.dongxiaoqi.top/` 自然作废（DNS 失败）。

---

## 方案 B：换一个新的自定义域名

假设新域名是 `newdomain.com`，博客用子域 `blog.newdomain.com`。

### 步骤

**1. DNS 设置（在 newdomain.com 的域名商后台）**
加一条 CNAME 记录：
```
类型   主机记录   记录值
CNAME  blog      vamViolet.github.io.
```
> 注意记录值末尾的 `.`（根域标记），部分域名商会自动补。等 DNS 生效（几分钟到数小时）。

**2. 改 `config.toml` 第 4 行**
```toml
baseURL = "https://blog.newdomain.com/"
```

**3. 改 `static/CNAME`**
```
blog.newdomain.com
```
> 必须是裸子域名，不带 `https://`。

**4. GitHub 仓库设置绑定新域名**
`https://github.com/vamViolet/notion-hugo-meme/settings/pages`
- `Custom domain` 填 `blog.newdomain.com`，Save。
- 勾选 `Enforce HTTPS`（GitHub 会自动签证书，等几分钟）。

**5. 提交并推送**（同方案 A 第 6 步）

### 改完后访问
- 新地址：**`https://blog.newdomain.com/`**

---

## 排错

| 现象 | 原因 | 处理 |
|------|------|------|
| 全站白屏或样式丢失 | `baseURL` 写错（少了 `/仓库名/` 或没带结尾 `/`） | 按方案 A 第 1 步核对 |
| 访问 github.io 仍跳到老域名 | `static/CNAME` 没删 / Pages 设置没清 | 重做第 2、3 步，清浏览器缓存 |
| 博客内容没更新 | workflow 没跑或失败 | Actions 页面手动 `Run workflow`，看日志 |
| HTTPS 证书报错（方案 B） | DNS 还没生效或证书未签发 | 等 DNS 生效后重新 Enforce HTTPS |
| 老链接 404 | 域名已失效，正常现象 | 用新地址访问 |

## 验证清单（改完逐项确认）
- [ ] `config.toml` baseURL 已改
- [ ] `static/CNAME` 已删（方案 A）或已改新值（方案 B）
- [ ] GitHub Pages 设置已更新，Source = GitHub Actions
- [ ] 推送后 workflow 绿色 success
- [ ] 新地址能打开、文章列表正常、文章页内链不 404
- [ ] 老地址确认已失效（可选）
