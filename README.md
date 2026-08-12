# wikipedia

`wikipedia` 是一个在终端浏览 Wikipedia 条目的命令行工具。

它会获取 Wikipedia 页面 HTML，转换为 Markdown，再使用 Glamour 渲染为适合终端阅读的格式。标题、段落、粗体、斜体、链接、列表、引用、代码块和简单表格会尽量保留；图片、复杂模板、编辑控件和参考文献导航会做简化或过滤。

## 安装

```sh
go install wikipedia/cmd/wikipedia@latest
```

也可以在项目目录直接构建：

```sh
go build -o wikipedia ./cmd/wikipedia
```

## 使用

```sh
wikipedia "苹果"
wikipedia --lang en "Apple"
wikipedia -l en "ISBN"
```

未指定 `--lang` 或 `-l` 时，工具会依次读取 `LC_ALL`、`LC_MESSAGES`、`LANG`，从中提取语言代码；如果环境变量不可用，则使用 `en`。

如果找不到精确条目，工具会调用 Wikipedia 搜索接口提供第一个相似标题，并询问是否打开：

```text
没有找到条目 "ibsn"，你是在找 "ISBN" 吗？(y/N)
```

只有输入 `y` 或 `Y` 才会打开推荐条目。

## 选项

- `-l`, `--lang`：指定 Wikipedia 语言代码，例如 `zh`、`en`、`ja`。
- `-h`, `--help`：显示帮助。

正文输出到标准输出，错误输出到标准错误。终端中会显示格式化样式；重定向或管道使用时，输出仍保持可读，但可能包含终端样式控制符。

## 开发

```sh
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
go build ./...
```
