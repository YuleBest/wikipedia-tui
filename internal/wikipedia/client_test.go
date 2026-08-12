package wikipedia

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientArticleAndSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("list") == "search" {
			_, _ = w.Write([]byte(`{"query":{"search":[{"title":"ISBN"}]}}`))
			return
		}
		if r.URL.Query().Get("redirects") != "1" {
			t.Errorf("redirects = %q, want 1", r.URL.Query().Get("redirects"))
		}
		_, _ = w.Write([]byte(`{"parse":{"pageid":1,"title":"苹果","text":"<p>正文</p>"}}`))
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL
	article, err := client.Article(context.Background(), "zh", "苹果")
	if err != nil {
		t.Fatal(err)
	}
	if article.Title != "苹果" || article.HTML != "<p>正文</p>" {
		t.Fatalf("unexpected article: %+v", article)
	}
	suggestions, err := client.Search(context.Background(), "en", "ibsn")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) != 1 || suggestions[0] != "ISBN" {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
}

func TestClientArticleNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"parse":{"pageid":0}}`))
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL
	_, err := client.Article(context.Background(), "en", "missing")
	if err != ErrNotFound {
		t.Fatalf("Article() error = %v, want ErrNotFound", err)
	}
}
