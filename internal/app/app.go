package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"wikipedia/internal/render"
	"wikipedia/internal/wikipedia"
)

type Config struct {
	Lang  string
	Query string
	In    io.Reader
	Out   io.Writer
	Err   io.Writer
	API   *wikipedia.Client
}

func Run(ctx context.Context, cfg Config) error {
	if cfg.API == nil {
		cfg.API = wikipedia.NewClient()
	}
	if cfg.In == nil {
		cfg.In = os.Stdin
	}
	if cfg.Out == nil {
		cfg.Out = os.Stdout
	}
	if cfg.Err == nil {
		cfg.Err = os.Stderr
	}
	article, err := cfg.API.Article(ctx, cfg.Lang, cfg.Query)
	if err == nil {
		output, renderErr := render.Article(article.Title, article.HTML)
		if renderErr != nil {
			return renderErr
		}
		_, writeErr := io.WriteString(cfg.Out, output)
		return writeErr
	}
	if !errors.Is(err, wikipedia.ErrNotFound) {
		return err
	}
	suggestions, searchErr := cfg.API.Search(ctx, cfg.Lang, cfg.Query)
	if searchErr != nil {
		return fmt.Errorf("search suggestions: %w", searchErr)
	}
	if len(suggestions) == 0 {
		return fmt.Errorf("没有找到条目 %q", cfg.Query)
	}
	candidate := suggestions[0]
	fmt.Fprintf(cfg.Out, "没有找到条目 %q，你是在找 %q 吗？(y/N)\n", cfg.Query, candidate)
	if !confirmed(cfg.In) {
		return fmt.Errorf("没有找到条目 %q", cfg.Query)
	}
	article, err = cfg.API.Article(ctx, cfg.Lang, candidate)
	if err != nil {
		return fmt.Errorf("load suggested article %q: %w", candidate, err)
	}
	output, renderErr := render.Article(article.Title, article.HTML)
	if renderErr != nil {
		return renderErr
	}
	_, writeErr := io.WriteString(cfg.Out, output)
	return writeErr
}

func confirmed(in io.Reader) bool {
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false
	}
	line = strings.TrimSpace(line)
	return len(line) > 0 && (line[0] == 'y' || line[0] == 'Y')
}

func DetectLanguage(getenv func(string) string) string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := getenv(key); value != "" {
			if lang := languageCode(value); lang != "" {
				return lang
			}
		}
	}
	return "en"
}

func DetectVariant(getenv func(string) string, lang string) string {
	if strings.ToLower(lang) != "zh" {
		return ""
	}
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		value := strings.ToLower(strings.TrimSpace(getenv(key)))
		if value == "" {
			continue
		}
		if underscore := strings.IndexByte(value, '_'); underscore >= 0 {
			value = value[underscore+1:]
		} else if dash := strings.IndexByte(value, '-'); dash >= 0 {
			value = value[dash+1:]
		} else {
			continue
		}
		if dot := strings.IndexByte(value, '.'); dot >= 0 {
			value = value[:dot]
		}
		if at := strings.IndexByte(value, '@'); at >= 0 {
			value = value[:at]
		}
		switch value {
		case "cn", "sg":
			return "zh-hans"
		case "tw", "hk", "mo":
			return "zh-hant"
		}
	}
	return "zh-hans"
}

func languageCode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "C") || strings.EqualFold(value, "POSIX") {
		return ""
	}
	if dot := strings.IndexByte(value, '.'); dot >= 0 {
		value = value[:dot]
	}
	if at := strings.IndexByte(value, '@'); at >= 0 {
		value = value[:at]
	}
	if dash := strings.IndexByte(value, '-'); dash >= 0 {
		value = value[:dash]
	}
	if underscore := strings.IndexByte(value, '_'); underscore >= 0 {
		value = value[:underscore]
	}
	value = strings.ToLower(value)
	for _, r := range value {
		if !unicode.IsLetter(r) {
			return ""
		}
	}
	return value
}
