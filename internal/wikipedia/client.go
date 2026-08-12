package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://%s.wikipedia.org/w/api.php"

// Article is a Wikipedia article containing the rendered page HTML.
type Article struct {
	Title string
	HTML  string
}

// Client retrieves articles and search suggestions from MediaWiki.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	UserAgent  string
	Variant    string
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		UserAgent:  "wikipedia-cli/1.0 (https://www.wikipedia.org/)",
	}
}

func (c *Client) endpoint(lang string) string {
	base := c.BaseURL
	if base == "" {
		base = fmt.Sprintf(defaultBaseURL, lang)
	}
	return base
}

func (c *Client) request(ctx context.Context, lang string, params url.Values, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(lang)+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wikipedia API returned HTTP %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decode wikipedia API response: %w", err)
	}
	return nil
}

type pageResponse struct {
	Parse struct {
		Title  string `json:"title"`
		Text   string `json:"text"`
		PageID int    `json:"pageid"`
	} `json:"parse"`
	Error *struct {
		Info string `json:"info"`
	} `json:"error,omitempty"`
}

func (c *Client) Article(ctx context.Context, lang, query string) (Article, error) {
	params := url.Values{
		"action":        {"parse"},
		"format":        {"json"},
		"formatversion": {"2"},
		"page":          {query},
		"prop":          {"text"},
		"redirects":     {"1"},
	}
	if c.Variant != "" {
		params.Set("variant", c.Variant)
	}
	var result pageResponse
	if err := c.request(ctx, lang, params, &result); err != nil {
		return Article{}, err
	}
	if result.Error != nil {
		return Article{}, fmt.Errorf("wikipedia API error: %s", result.Error.Info)
	}
	if result.Parse.PageID == 0 || strings.TrimSpace(result.Parse.Text) == "" {
		return Article{}, ErrNotFound
	}
	return Article{Title: result.Parse.Title, HTML: strings.TrimSpace(result.Parse.Text)}, nil
}

type searchResponse struct {
	Query struct {
		Search []struct {
			Title string `json:"title"`
		} `json:"search"`
	} `json:"query"`
	Error *struct {
		Info string `json:"info"`
	} `json:"error,omitempty"`
}

func (c *Client) Search(ctx context.Context, lang, query string) ([]string, error) {
	params := url.Values{
		"action":      {"query"},
		"format":      {"json"},
		"list":        {"search"},
		"srsearch":    {query},
		"srlimit":     {"5"},
		"srnamespace": {"0"},
	}
	var result searchResponse
	if err := c.request(ctx, lang, params, &result); err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("wikipedia API error: %s", result.Error.Info)
	}
	results := make([]string, 0, len(result.Query.Search))
	for _, item := range result.Query.Search {
		if strings.TrimSpace(item.Title) != "" {
			results = append(results, item.Title)
		}
	}
	return results, nil
}

var ErrNotFound = fmt.Errorf("article not found")
