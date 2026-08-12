package render

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/charmbracelet/glamour"
	"golang.org/x/net/html"
)

var whitespace = regexp.MustCompile(`[ \t]+`)
var controlChars = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)

// HTMLToMarkdown removes page chrome and converts Wikipedia HTML to Markdown.
func HTMLToMarkdown(source string) (string, error) {
	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return "", fmt.Errorf("parse article HTML: %w", err)
	}
	removeNodes(doc)
	tables := extractTables(doc)
	root := doc
	if body := findElement(doc, "body"); body != nil {
		root = body
	}
	var cleaned bytes.Buffer
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if err := html.Render(&cleaned, child); err != nil {
			return "", fmt.Errorf("serialize article HTML: %w", err)
		}
	}
	markdown, err := htmltomarkdown.ConvertString(cleaned.String(), converter.WithDomain("https://en.wikipedia.org"))
	if err != nil {
		return "", fmt.Errorf("convert HTML to Markdown: %w", err)
	}
	for token, table := range tables {
		markdown = strings.ReplaceAll(markdown, token, "\n\n"+table+"\n\n")
		markdown = strings.ReplaceAll(markdown, strings.ReplaceAll(token, "_", "\\_"), "\n\n"+table+"\n\n")
	}
	return strings.TrimSpace(controlChars.ReplaceAllString(markdown, "")), nil
}

func removeNodes(node *html.Node) {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if child.Type == html.ElementNode {
			class := ""
			for _, attr := range child.Attr {
				if attr.Key == "class" {
					class = attr.Val
				}
			}
			if child.Data == "a" {
				removeNodes(child)
				unwrapNode(node, child)
				child = next
				continue
			}
			if child.Data == "img" || child.Data == "script" || child.Data == "style" || hasFilteredClass(class) || child.Data == "sup" && strings.Contains(class, "reference") {
				node.RemoveChild(child)
				child = next
				continue
			}
		}
		removeNodes(child)
		child = next
	}
}

func unwrapNode(parent, node *html.Node) {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		node.RemoveChild(child)
		parent.InsertBefore(child, node)

		child = next
	}
	parent.RemoveChild(node)
}
func hasFilteredClass(class string) bool {
	for _, item := range strings.Fields(class) {
		switch item {
		case "mw-editsection", "mw-references-wrap", "reflist", "navbox", "metadata", "toc":
			return true
		}
	}
	return false
}

func findElement(node *html.Node, tag string) *html.Node {
	if node.Type == html.ElementNode && node.Data == tag {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, tag); found != nil {
			return found
		}
	}
	return nil
}

// MarkdownToTerminal renders Markdown with Glamour using a fixed readable width.
func MarkdownToTerminal(markdown string, width int) (string, error) {
	if width <= 0 {
		width = 100
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", fmt.Errorf("create terminal renderer: %w", err)
	}
	output, err := renderer.Render(markdown)
	if err != nil {
		return "", fmt.Errorf("render Markdown: %w", err)
	}
	return output, nil
}

func Article(title, htmlSource string) (string, error) {
	markdown, err := HTMLToMarkdown(htmlSource)
	if err != nil {
		return "", err
	}
	if title != "" {
		markdown = "# " + title + "\n\n" + markdown
	}
	return MarkdownToTerminal(markdown, 100)
}

func LegacyArticle(title, text string) string {
	text = cleanText(text)
	if title == "" {
		return text + "\n"
	}
	return title + "\n" + strings.Repeat("=", max(6, len([]rune(title)))) + "\n\n" + text + "\n"
}

func cleanText(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = whitespace.ReplaceAllString(strings.TrimSpace(line), " ")
		if line != "" || (len(cleaned) > 0 && cleaned[len(cleaned)-1] != "") {
			cleaned = append(cleaned, line)
		}
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
