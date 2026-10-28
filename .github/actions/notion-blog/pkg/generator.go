package notion_blog

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"net/url"

	"github.com/janeczku/go-spinner"
	"github.com/jomei/notionapi"
)

func emphFormat(a *notionapi.Annotations) (s string) {
	s = "%s"
	if a == nil {
		return
	}

	if a.Code {
		return "`%s`"
	}

	switch {
	case a.Bold && a.Italic:
		s = "***%s***"
	case a.Bold:
		s = "**%s**"
	case a.Italic:
		s = "*%s*"
	}

	if a.Underline {
		s = "__" + s + "__"
	} else if a.Strikethrough {
		s = "~~" + s + "~~"
	}

	// TODO: color

	return s
}

func ConvertRich(t notionapi.RichText) string {
	switch t.Type {
	case notionapi.ObjectTypeText:
		if t.Text.Link != nil {
			return fmt.Sprintf(
				emphFormat(t.Annotations),
				fmt.Sprintf("[%s](%s)", t.Text.Content, t.Text.Link.Url),
			)
		}
		return fmt.Sprintf(emphFormat(t.Annotations), t.Text.Content)
	case notionapi.ObjectTypeList:
	}
	return ""
}

func ConvertRichText(t []notionapi.RichText) string {
	buf := &bytes.Buffer{}
	for _, word := range t {
		buf.WriteString(ConvertRich(word))
	}

	return buf.String()
}

// extFromContentType 根据响应头 Content-Type 推断图片后缀。
func extFromContentType(ct string) string {
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = ct[:i]
	}
	switch strings.TrimSpace(strings.ToLower(ct)) {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/avif":
		return ".avif"
	case "image/bmp":
		return ".bmp"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico"
	}
	return ".png"
}

// sanitizeFileName 清理文件名中对文件系统不友好的字符，并限制长度。
func sanitizeFileName(name string) string {
	name = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		}
		return r
	}, name)
	runes := []rune(name)
	if len(runes) > 80 {
		name = string(runes[:80])
	}
	return name
}

func getImage(imgURL string, config BlogConfig) (_ string, err error) {
	// Split image url to get host and file name
	splittedURL, err := url.Parse(imgURL)
	if err != nil {
		return "", fmt.Errorf("malformed url: %s", err)
	}

	// Get file name
	//
	// 注意：Notion/微信等图床的 URL，其 path 通常是多级目录
	// （Notion: /<workspace-uuid>/<object-uuid>/<file>，微信: /sz_mmbiz_png/<hash>/<n>）。
	// 原实现直接拿整段 path 当文件名，os.Create 到 static/images/ 下写多级路径时
	// 因父目录不存在而失败；失败后又把这段 path 原样返回当图片地址，
	// 最终在页面上渲染成指向站点根目录的相对链接（404，图片显示不出来）。
	// 这里只取最后一段做文件名，并用 URL 哈希前缀避免不同目录下同名文件互相覆盖。
	base := path.Base(splittedURL.Path)
	if base == "" || base == "." || base == "/" {
		base = "image"
	}
	if decoded, derr := url.PathUnescape(base); derr == nil && decoded != "" {
		base = decoded
	}
	sum := sha1.Sum([]byte(imgURL))
	name := fmt.Sprintf("%x_%s", sum[:4], sanitizeFileName(base))

	spin := spinner.StartNew(fmt.Sprintf("Getting image `%s`", name))
	defer func() {
		spin.Stop()
		if err != nil {
			fmt.Printf("❌ Getting image `%s`: %s", name, err)
		} else {
			fmt.Printf("✔ Getting image `%s`: Completed", name)
		}
	}()

	resp, err := http.Get(imgURL)
	if err != nil {
		return imgURL, fmt.Errorf("couldn't download image: %s", err)
	}
	defer resp.Body.Close()

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		body, _ := io.ReadAll(resp.Body)
		return imgURL, fmt.Errorf("couldn't download image, server returned an error: %d\n%s", resp.StatusCode, body)
	}

	// 有些图床的 URL 路径不带扩展名（如微信 mmbiz 的 `/640?wx_fmt=png`），
	// 用响应头的 Content-Type 补一个后缀，避免静态文件没有扩展名。
	if path.Ext(name) == "" {
		name += extFromContentType(resp.Header.Get("Content-Type"))
	}

	err = os.MkdirAll(config.ImagesFolder, 0777)
	if err != nil {
		return "", fmt.Errorf("couldn't create images folder: %s", err)
	}

	// Create the file
	out, err := os.Create(filepath.Join(config.ImagesFolder, name))
	if err != nil {
		return imgURL, fmt.Errorf("couldn't create image file: %s", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return imgURL, fmt.Errorf("couldn't save image to file: %s", err)
	}

	return filepath.Join(config.ImagesLink, name), nil
}

func Generate(w io.Writer, page notionapi.Page, blocks []notionapi.Block, config BlogConfig) error {

	// Generate markdown content
	buffer := &bytes.Buffer{}
	GenerateContent(buffer, blocks, config)

	// Parse template file
	t := template.New(path.Base(config.ArchetypeFile))
	t.Funcs(template.FuncMap{
		"add":    func(a, b int) int { return a + b },
		"sub":    func(a, b int) int { return a - b },
		"mul":    func(a, b int) int { return a * b },
		"div":    func(a, b int) int { return a / b },
		"repeat": func(s string, n int) string { return strings.Repeat(s, n) },
		"rich":   ConvertRichText,
	})

	t, err := t.ParseFiles(config.ArchetypeFile)
	if err != nil {
		return fmt.Errorf("error parsing archetype file: %s", err)
	}

	// Dump markdown content into output according to archetype file
	fileArchetype := MakeArchetypeFields(page, config)
	fileArchetype.Content = buffer.String()

	err = t.Execute(w, fileArchetype)
	if err != nil {
		return fmt.Errorf("error filling archetype file: %s", err)
	}
	return nil
}

func GenerateContent(w io.Writer, blocks []notionapi.Block, config BlogConfig, prefixes ...string) {
	if len(blocks) == 0 {
		return
	}

	numberedList := false
	bulletedList := false

	for _, block := range blocks {
		// Add line break after list is finished
		if bulletedList && block.GetType() != notionapi.BlockTypeBulletedListItem {
			bulletedList = false
			fmt.Fprintln(w)
		}
		if numberedList && block.GetType() != notionapi.BlockTypeNumberedListItem {
			numberedList = false
			fmt.Fprintln(w)
		}

		switch b := block.(type) {
		case *notionapi.ParagraphBlock:
			fprintln(w, prefixes, ConvertRichText(b.Paragraph.RichText)+"\n")
			GenerateContent(w, b.Paragraph.Children, config)
		case *notionapi.Heading1Block:
			fprintf(w, prefixes, "# %s", ConvertRichText(b.Heading1.RichText))
		case *notionapi.Heading2Block:
			fprintf(w, prefixes, "## %s", ConvertRichText(b.Heading2.RichText))
		case *notionapi.Heading3Block:
			fprintf(w, prefixes, "### %s", ConvertRichText(b.Heading3.RichText))
		case *notionapi.CalloutBlock:
			if !config.UseShortcodes {
				continue
			}
			if b.Callout.Icon != nil {
				if b.Callout.Icon.Emoji != nil {
					fprintf(w, prefixes, `{{%% callout emoji="%s" %%}}`, *b.Callout.Icon.Emoji)
				} else {
					fprintf(w, prefixes, `{{%% callout image="%s" %%}}`, b.Callout.Icon.GetURL())
				}
			}
			fprintln(w, prefixes, ConvertRichText(b.Callout.RichText))
			GenerateContent(w, b.Callout.Children, config, prefixes...)
			fprintln(w, prefixes, "{{% /callout %}}")

		case *notionapi.BookmarkBlock:
			if !config.UseShortcodes {
				// Simply generate the url link
				fprintf(w, prefixes, "[%s](%s)", b.Bookmark.URL, b.Bookmark.URL)
				continue
			}
			// Parse external page metadata
			og, err := parseMetadata(b.Bookmark.URL, config)
			if err != nil {
				log.Println("error getting bookmark metadata:", err)
			}

			// GenerateContent shortcode with given metadata
			fprintf(w, prefixes,
				`{{< bookmark url="%s" title="%s" img="%s" >}}%s{{< /bookmark >}}`,
				og.URL,
				og.Title,
				og.Image,
				og.Description,
			)

		case *notionapi.QuoteBlock:
			fprintf(w, prefixes, "> %s", ConvertRichText(b.Quote.RichText))
			GenerateContent(w, b.Quote.Children, config,
				append([]string{"> "}, prefixes...)...)
			fprintln(w, prefixes)

		case *notionapi.BulletedListItemBlock:
			bulletedList = true
			fprintf(w, prefixes, "- %s", ConvertRichText(b.BulletedListItem.RichText))
			GenerateContent(w, b.BulletedListItem.Children, config,
				append([]string{"    "}, prefixes...)...)

		case *notionapi.NumberedListItemBlock:
			numberedList = true
			fprintf(w, prefixes, "1. %s", ConvertRichText(b.NumberedListItem.RichText))
			GenerateContent(w, b.NumberedListItem.Children, config,
				append([]string{"    "}, prefixes...)...)

		case *notionapi.ImageBlock:
			src, _ := getImage(b.Image.GetURL(), config)
			fprintf(w, prefixes, "![%s](%s)\n", ConvertRichText(b.Image.Caption), src)

		case *notionapi.CodeBlock:
			if b.Code.Language == "plain text" {
				fprintln(w, prefixes, "```")
			} else {
				fprintf(w, prefixes, "```%s", b.Code.Language)
			}
			fprintln(w, prefixes, ConvertRichText(b.Code.RichText))
			fprintln(w, prefixes, "```")

		case *notionapi.TableBlock:
			rows := b.Table.Children
			if len(rows) == 0 {
				continue
			}
			for index, row1 := range rows {
				row := row1.(*notionapi.TableRowBlock)
				if index == 1 {
					// Markdown 表格的分隔行必须紧跟表头行，但它本身不是数据行。
					// 原实现在第二行「打印分隔行而不打印该行内容」，把首个数据行
					// （如组件表的 Ingest 行）整行吃掉了。
					fprintf(w, prefixes, "|%s", strings.Repeat("---|", len(row.TableRow.Cells)))
				}
				cells := row.TableRow.Cells
				line := ""
				for _, cell := range cells {
					line += fmt.Sprintf("|%s", ConvertRichText(cell))
				}
				line += "|"
				fprintln(w, prefixes, line)
			}
			fprintln(w, prefixes)
		case *notionapi.UnsupportedBlock:
			if b.GetType() != "unsupported" {
				fmt.Println("ℹ Unimplemented block", b.GetType())
			} else {
				fmt.Println("ℹ Unsupported block type")
			}
		default:
			fmt.Println("ℹ Unimplemented block", b.GetType())
		}

	}
}
