package render

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	xhtml "golang.org/x/net/html"
)

func extractTables(root *xhtml.Node) map[string]string {
	tables := make(map[string]string)
	counter := 0
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		for child := node.FirstChild; child != nil; {
			next := child.NextSibling
			if child.Type == xhtml.ElementNode && child.Data == "table" {
				token := fmt.Sprintf("WIKIPEDIA_TABLE_%d", counter)
				counter++
				if table := tableMarkdown(child); table != "" {
					tables[token] = table
					node.InsertBefore(&xhtml.Node{Type: xhtml.TextNode, Data: token}, child)
					node.RemoveChild(child)
					child = next
					continue
				}
			}
			walk(child)
			child = next
		}
	}
	walk(root)
	return tables
}

func tableMarkdown(table *xhtml.Node) string {
	rows := tableRows(table)
	if len(rows) == 0 {
		return ""
	}
	caption := strings.TrimSpace(textContent(findChild(table, "caption")))
	isInfobox := hasClass(table, "infobox")
	if isInfobox {
		return infoboxMarkdown(caption, rows)
	}
	return gridMarkdown(caption, rows)
}

type tableRow struct {
	cells []tableCell
}

type tableCell struct {
	text    string
	header  bool
	colspan int
}

func tableRows(table *xhtml.Node) []tableRow {
	var rows []tableRow
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node.Type == xhtml.ElementNode && node.Data == "tr" {
			row := tableRow{}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Type != xhtml.ElementNode || (child.Data != "th" && child.Data != "td") {
					continue
				}
				if hasStyle(child, "display:none") || hasClass(child, "infobox-full-data") {
					continue
				}
				colspan := 1
				for _, attr := range child.Attr {
					if attr.Key == "colspan" {
						if parsed, err := strconv.Atoi(attr.Val); err == nil && parsed > 0 {
							colspan = parsed
						}
					}
				}
				text := cleanCell(textContent(child))
				if text == "" {
					continue
				}
				row.cells = append(row.cells, tableCell{text: text, header: child.Data == "th" || hasClass(child, "infobox-label") || hasClass(child, "infobox-header"), colspan: colspan})
			}
			if len(row.cells) > 0 {
				rows = append(rows, row)
			}
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(table)
	return rows
}

func infoboxMarkdown(caption string, rows []tableRow) string {
	var b strings.Builder
	if caption != "" {
		b.WriteString("**" + escapeInline(caption) + "**\n\n")
	}
	b.WriteString("| 项目 | 内容 |\n| --- | --- |\n")
	for _, row := range rows {
		if len(row.cells) == 1 {
			b.WriteString("| **" + escapeCell(row.cells[0].text) + "** | |\n")
			continue
		}
		label := row.cells[0].text
		value := row.cells[1].text
		b.WriteString("| " + escapeCell(label) + " | " + escapeCell(value) + " |\n")
	}
	return strings.TrimSpace(b.String())
}

func gridMarkdown(caption string, rows []tableRow) string {
	width := 0
	for _, row := range rows {
		if len(row.cells) > width {
			width = len(row.cells)
		}
	}
	if width == 0 {
		return ""
	}
	var b strings.Builder
	if caption != "" {
		b.WriteString("**" + escapeInline(caption) + "**\n\n")
	}
	for i, row := range rows {
		b.WriteString("|")
		for j := 0; j < width; j++ {
			cell := ""
			if j < len(row.cells) {
				cell = escapeCell(row.cells[j].text)
			}
			b.WriteString(" " + cell + " |")
		}
		b.WriteString("\n")
		if i == 0 {
			b.WriteString("|")
			for j := 0; j < width; j++ {
				b.WriteString(" --- |")
			}
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(b.String())
}

func cleanCell(value string) string {
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return whitespace.ReplaceAllString(strings.TrimSpace(value), " ")
}

func escapeCell(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "|", "\\|"), "\n", " ")
}

func escapeInline(value string) string {
	return strings.ReplaceAll(value, "*", "\\*")
}

func textContent(node *xhtml.Node) string {
	var b strings.Builder
	var walk func(*xhtml.Node)
	walk = func(current *xhtml.Node) {
		if current.Type == xhtml.TextNode {
			b.WriteString(current.Data)
			return
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	if node != nil {
		walk(node)
	}
	return b.String()
}

func findChild(node *xhtml.Node, tag string) *xhtml.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xhtml.ElementNode && child.Data == tag {
			return child
		}
	}
	return nil
}

func hasClass(node *xhtml.Node, class string) bool {
	for _, attr := range node.Attr {
		if attr.Key == "class" {
			for _, item := range strings.Fields(attr.Val) {
				if item == class {
					return true
				}
			}
		}
	}
	return false
}

func hasStyle(node *xhtml.Node, style string) bool {
	for _, attr := range node.Attr {
		if attr.Key == "style" && strings.Contains(strings.ReplaceAll(attr.Val, " ", ""), style) {
			return true
		}
	}
	return false
}
