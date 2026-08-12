package render

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdownConvertsTable(t *testing.T) {
	got, err := HTMLToMarkdown(`<table><caption>营养</caption><tr><th>项目</th><th>内容</th></tr><tr><td>热量</td><td>200 kJ</td></tr></table>`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"**营养**", "| 项目 | 内容 |", "| --- | --- |", "| 热量 | 200 kJ |"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Markdown %q does not contain %q", got, want)
		}
	}
}

func TestHTMLToMarkdownConvertsInfobox(t *testing.T) {
	got, err := HTMLToMarkdown(`<table class="infobox nowrap"><caption>生苹果</caption><tr><th class="infobox-header" colspan="2">营养成分</th></tr><tr><th class="infobox-label">热量</th><td class="infobox-data">200 kJ（48 kcal）</td></tr><tr><th class="infobox-label">糖类</th><td class="infobox-data">12.76 g</td></tr></table>`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"**生苹果**", "营养成分", "热量", "200 kJ", "糖类", "12.76 g"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Markdown %q does not contain %q", got, want)
		}
	}
}

func TestMarkdownToTerminalRendersTable(t *testing.T) {
	got, err := MarkdownToTerminal("| 项目 | 内容 |\n| --- | --- |\n| 热量 | 200 kJ |", 80)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"项目", "内容", "200 kJ"} {
		if !strings.Contains(stripANSI(got), want) {
			t.Fatalf("unexpected table output: %q", got)
		}
	}
}

func stripANSI(value string) string {
	for {
		start := strings.Index(value, "\x1b[")
		if start < 0 {
			return value
		}
		end := start + 2
		for end < len(value) && (value[end] < 'A' || value[end] > 'z') {
			end++
		}
		if end < len(value) {
			end++
		}
		value = value[:start] + value[end:]
	}
}
