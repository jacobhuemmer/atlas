package rest

import (
	"html"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	ext "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var jiraMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

// adfFromMarkdown builds Jira rich-text nodes instead of storing Markdown
// punctuation as one literal text node.
func adfFromMarkdown(markdown string) map[string]any {
	source := []byte(strings.TrimSpace(markdown))
	if len(source) == 0 {
		return nil
	}
	doc := jiraMarkdown.Parser().Parse(text.NewReader(source))
	return map[string]any{"type": "doc", "version": 1, "content": adfBlocks(doc, source)}
}

func adfBlocks(parent ast.Node, source []byte) []any {
	blocks := []any{}
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		switch n := node.(type) {
		case *ast.Paragraph, *ast.TextBlock:
			blocks = append(blocks, map[string]any{"type": "paragraph", "content": adfInlines(node, source, nil)})
		case *ast.Heading:
			blocks = append(blocks, map[string]any{"type": "heading", "attrs": map[string]any{"level": n.Level}, "content": adfInlines(node, source, nil)})
		case *ast.List:
			typ := "bulletList"
			attrs := map[string]any{}
			if n.IsOrdered() {
				typ = "orderedList"
				if n.Start > 1 {
					attrs["order"] = n.Start
				}
			}
			items := []any{}
			for item := node.FirstChild(); item != nil; item = item.NextSibling() {
				content := adfBlocks(item, source)
				if len(content) == 0 {
					content = []any{map[string]any{"type": "paragraph", "content": []any{}}}
				}
				items = append(items, map[string]any{"type": "listItem", "content": content})
			}
			list := map[string]any{"type": typ, "content": items}
			if len(attrs) != 0 {
				list["attrs"] = attrs
			}
			blocks = append(blocks, list)
		case *ast.FencedCodeBlock:
			block := map[string]any{"type": "codeBlock", "content": adfCodeLines(n.Lines(), source)}
			if language := string(n.Language(source)); language != "" {
				block["attrs"] = map[string]any{"language": language}
			}
			blocks = append(blocks, block)
		case *ast.CodeBlock:
			blocks = append(blocks, map[string]any{"type": "codeBlock", "content": adfCodeLines(n.Lines(), source)})
		case *ast.Blockquote:
			blocks = append(blocks, map[string]any{"type": "blockquote", "content": adfBlocks(node, source)})
		case *ast.ThematicBreak:
			blocks = append(blocks, map[string]any{"type": "rule"})
		case *ext.Table:
			rows := []any{}
			for row := node.FirstChild(); row != nil; row = row.NextSibling() {
				cells := []any{}
				for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
					typ := "tableCell"
					if _, ok := row.(*ext.TableHeader); ok {
						typ = "tableHeader"
					}
					cells = append(cells, map[string]any{"type": typ, "content": []any{map[string]any{"type": "paragraph", "content": adfInlines(cell, source, nil)}}})
				}
				rows = append(rows, map[string]any{"type": "tableRow", "content": cells})
			}
			blocks = append(blocks, map[string]any{"type": "table", "content": rows})
		case *ast.HTMLBlock:
			blocks = append(blocks, map[string]any{"type": "paragraph", "content": adfText(strings.TrimSpace(linesText(n.Lines(), source)), nil)})
		default:
			if node.HasChildren() {
				blocks = append(blocks, adfBlocks(node, source)...)
			}
		}
	}
	return blocks
}

func adfCodeLines(lines *text.Segments, source []byte) []any {
	value := strings.TrimSuffix(linesText(lines, source), "\n")
	return adfText(value, nil)
}

func linesText(lines *text.Segments, source []byte) string {
	var out strings.Builder
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		out.Write(segment.Value(source))
	}
	return out.String()
}

func adfInlines(parent ast.Node, source []byte, marks []any) []any {
	content := []any{}
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		switch n := node.(type) {
		case *ast.Text:
			content = append(content, adfText(html.UnescapeString(string(n.Value(source))), marks)...)
			if n.HardLineBreak() {
				content = append(content, map[string]any{"type": "hardBreak"})
			} else if n.SoftLineBreak() {
				content = append(content, adfText(" ", marks)...)
			}
		case *ast.String:
			content = append(content, adfText(html.UnescapeString(string(n.Value)), marks)...)
		case *ast.CodeSpan:
			content = append(content, adfInlines(node, source, withMark(marks, map[string]any{"type": "code"}))...)
		case *ast.Emphasis:
			typ := "em"
			if n.Level >= 2 {
				typ = "strong"
			}
			content = append(content, adfInlines(node, source, withMark(marks, map[string]any{"type": typ}))...)
		case *ast.Link:
			mark := map[string]any{"type": "link", "attrs": map[string]any{"href": string(n.Destination)}}
			content = append(content, adfInlines(node, source, withMark(marks, mark))...)
		case *ast.AutoLink:
			mark := map[string]any{"type": "link", "attrs": map[string]any{"href": string(n.URL(source))}}
			content = append(content, adfText(string(n.Label(source)), withMark(marks, mark))...)
		case *ast.Image:
			mark := map[string]any{"type": "link", "attrs": map[string]any{"href": string(n.Destination)}}
			content = append(content, adfInlines(node, source, withMark(marks, mark))...)
		case *ext.Strikethrough:
			content = append(content, adfInlines(node, source, withMark(marks, map[string]any{"type": "strike"}))...)
		case *ext.TaskCheckBox:
			box := "☐ "
			if n.IsChecked {
				box = "☑ "
			}
			content = append(content, adfText(box, marks)...)
		case *ast.RawHTML:
			content = append(content, adfText(string(n.Text(source)), marks)...)
		default:
			content = append(content, adfInlines(node, source, marks)...)
		}
	}
	return content
}

func withMark(marks []any, mark any) []any {
	result := append([]any(nil), marks...)
	return append(result, mark)
}

func adfText(value string, marks []any) []any {
	if value == "" {
		return nil
	}
	node := map[string]any{"type": "text", "text": value}
	if len(marks) > 0 {
		node["marks"] = marks
	}
	return []any{node}
}

// markdownFromADF keeps jira get's description usable as Markdown after Jira
// stores it as rich text.
func markdownFromADF(value any) string {
	if plain, ok := value.(string); ok {
		return plain
	}
	return markdownBlocks(asList(asMap(value)["content"]))
}

func markdownBlocks(blocks []any) string {
	parts := []string{}
	for _, raw := range blocks {
		if rendered := markdownBlock(asMap(raw)); rendered != "" {
			parts = append(parts, rendered)
		}
	}
	return strings.Join(parts, "\n\n")
}

func markdownBlock(block map[string]any) string {
	switch str(block, "type") {
	case "paragraph":
		return markdownInlines(asList(block["content"]))
	case "heading":
		level := 1
		switch n := asMap(block["attrs"])["level"].(type) {
		case float64:
			if n >= 1 && n <= 6 {
				level = int(n)
			}
		case int:
			if n >= 1 && n <= 6 {
				level = n
			}
		}
		return strings.Repeat("#", level) + " " + markdownInlines(asList(block["content"]))
	case "bulletList", "orderedList":
		return markdownList(block, 0)
	case "codeBlock":
		language := str(asMap(block["attrs"]), "language")
		return "```" + language + "\n" + textFromADF(block["content"]) + "\n```"
	case "blockquote":
		quoted := markdownBlocks(asList(block["content"]))
		return "> " + strings.ReplaceAll(quoted, "\n", "\n> ")
	case "rule":
		return "---"
	case "table":
		return markdownTable(block)
	default:
		return textFromADF(block)
	}
}

func markdownList(list map[string]any, depth int) string {
	lines := []string{}
	start := 1
	switch n := asMap(list["attrs"])["order"].(type) {
	case float64:
		if n > 0 {
			start = int(n)
		}
	case int:
		if n > 0 {
			start = n
		}
	}
	for i, raw := range asList(list["content"]) {
		item := asMap(raw)
		marker := "- "
		if str(list, "type") == "orderedList" {
			marker = strconv.Itoa(start+i) + ". "
		}
		parts := []string{}
		for _, child := range asList(item["content"]) {
			block := asMap(child)
			if typ := str(block, "type"); typ == "bulletList" || typ == "orderedList" {
				parts = append(parts, markdownList(block, depth+1))
			} else {
				parts = append(parts, markdownBlock(block))
			}
		}
		first := ""
		if len(parts) > 0 {
			first = parts[0]
		}
		first = strings.Replace(first, "☐ ", "[ ] ", 1)
		first = strings.Replace(first, "☑ ", "[x] ", 1)
		indent := strings.Repeat("  ", depth)
		lines = append(lines, indent+marker+first)
		for _, part := range parts[1:] {
			lines = append(lines, part)
		}
	}
	return strings.Join(lines, "\n")
}

func markdownTable(table map[string]any) string {
	rows := []string{}
	for i, raw := range asList(table["content"]) {
		cells := []string{}
		for _, cell := range asList(asMap(raw)["content"]) {
			cells = append(cells, strings.ReplaceAll(markdownBlocks(asList(asMap(cell)["content"])), "|", "\\|"))
		}
		rows = append(rows, "| "+strings.Join(cells, " | ")+" |")
		if i == 0 {
			separators := make([]string, len(cells))
			for j := range separators {
				separators[j] = "---"
			}
			rows = append(rows, "| "+strings.Join(separators, " | ")+" |")
		}
	}
	return strings.Join(rows, "\n")
}

func markdownInlines(inlines []any) string {
	var out strings.Builder
	for _, raw := range inlines {
		inline := asMap(raw)
		if str(inline, "type") == "hardBreak" {
			out.WriteString("  \n")
			continue
		}
		value := str(inline, "text")
		if value == "" {
			value = markdownInlines(asList(inline["content"]))
		}
		for _, rawMark := range asList(inline["marks"]) {
			mark := asMap(rawMark)
			switch str(mark, "type") {
			case "strong":
				value = "**" + value + "**"
			case "em":
				value = "*" + value + "*"
			case "code":
				value = "`" + value + "`"
			case "strike":
				value = "~~" + value + "~~"
			case "link":
				value = "[" + value + "](" + str(asMap(mark["attrs"]), "href") + ")"
			}
		}
		out.WriteString(value)
	}
	return out.String()
}
