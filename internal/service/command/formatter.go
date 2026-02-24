package command

import (
	"fmt"
	"strings"
)

type ResponseFormatter struct{}

func NewResponseFormatter() *ResponseFormatter {
	return &ResponseFormatter{}
}

func (f *ResponseFormatter) Header(emoji, title string) string {
	return fmt.Sprintf("%s **%s**\n%s\n", emoji, title, strings.Repeat("━", len(title)))
}

func (f *ResponseFormatter) Info(label, value string) string {
	return fmt.Sprintf("ℹ️ **%s**: %s\n", label, value)
}

func (f *ResponseFormatter) Success(message string) string {
	return fmt.Sprintf("✅ **%s**\n", message)
}

func (f *ResponseFormatter) Error(operation string, err error) string {
	return fmt.Sprintf("❌ **Command Error**\n%s\n\n**Issue**: %s\n", strings.Repeat("━", 15), err.Error())
}

func (f *ResponseFormatter) Usage(command string) string {
	return fmt.Sprintf("📖 **Usage**: `%s`\n", command)
}

func (f *ResponseFormatter) Examples(examples []string) string {
	var sb strings.Builder
	sb.WriteString("💡 **Examples**: \n")
	for _, ex := range examples {
		sb.WriteString(fmt.Sprintf("- `%s`\n", ex))
	}
	return sb.String()
}

func (f *ResponseFormatter) List(emoji string, items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("%s %s\n", emoji, item))
	}
	return sb.String()
}

func (f *ResponseFormatter) Tip(text string) string {
	return fmt.Sprintf("💡 **Tip**: %s\n", text)
}

func (f *ResponseFormatter) Section(emoji, title, content string) string {
	return fmt.Sprintf("%s **%s**\n%s\n", emoji, title, content)
}

func (f *ResponseFormatter) CodeBlock(language, code string) string {
	return fmt.Sprintf("```%s\n%s\n```\n", language, code)
}

func (f *ResponseFormatter) Combine(sections ...string) string {
	return strings.Join(sections, "\n")
}
