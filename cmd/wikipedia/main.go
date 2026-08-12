package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"wikipedia/internal/app"
	"wikipedia/internal/wikipedia"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "用法：%s [选项] 查询词\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "在终端浏览 Wikipedia 条目。")
		flag.PrintDefaults()
	}
	lang := flag.String("lang", "", "Wikipedia 语言代码，例如 zh 或 en")
	shortLang := flag.String("l", "", "Wikipedia 语言代码（-lang 的缩写）")
	flag.Parse()
	if *lang != "" && *shortLang != "" && *lang != *shortLang {
		fmt.Fprintln(os.Stderr, "--lang 和 -l 的值不能冲突")
		os.Exit(2)
	}
	if *lang == "" {
		*lang = *shortLang
	}
	if *lang == "" {
		*lang = app.DetectLanguage(os.Getenv)
	}
	if *lang == "" {
		fmt.Fprintln(os.Stderr, "语言代码不能为空")
		os.Exit(2)
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	client := wikipedia.NewClient()
	client.Variant = app.DetectVariant(os.Getenv, *lang)
	if err := app.Run(context.Background(), app.Config{Lang: *lang, Query: flag.Arg(0), API: client}); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
