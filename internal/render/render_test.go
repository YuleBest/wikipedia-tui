package render

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdownFiltersChrome(t *testing.T) {
	got, err := HTMLToMarkdown(`<div><h2>标题</h2><p><strong>粗体</strong> 和 <a href="https://example.com">链接</a></p><div class="mw-editsection">编辑</div><div class="reflist">脚注</div><ul><li>一</li><li>二</li></ul></div>`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"标题", "**粗体**", "链接", "- 一", "- 二"} {
		if !strings.Contains(got, want) {
			t.Fatalf("Markdown %q does not contain %q", got, want)
		}
	}
	for _, unwanted := range []string{"编辑", "脚注", "https://example.com"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("Markdown %q contains filtered text %q", got, unwanted)
		}
	}
}

func TestArticleRendersMarkdown(t *testing.T) {
	got, err := Article("测试", `<p>正文 <strong>重点</strong></p>`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "测试") || !strings.Contains(got, "正文") || !strings.Contains(got, "重点") {
		t.Fatalf("unexpected terminal output: %q", got)
	}
}
