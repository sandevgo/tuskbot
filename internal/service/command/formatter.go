package command

import (
	"fmt"
	"strings"
)

type ResponseFormatter struct{}

func NewResponseFormatter() *ResponseFormatter {
	return &ResponseFormatter{}
}

func (f *ResponseFormatter) Info(title string) string {
	return fmt.Sprintf("⚙️️ **%s**\n\n", title)
}

func (f *ResponseFormatter) Success(message string) string {
	return fmt.Sprintf("✅ **%s**\n", message)
}

func (f *ResponseFormatter) Error(operation string, err error) string {
	return fmt.Sprintf("❌ **Command Error**\n\n**Issue**: %s\n", err.Error())
}

func (f *ResponseFormatter) Label(label, value string) string {
	return fmt.Sprintf("**%s**  ›  `%s`\n", label, value)
}

func (f *ResponseFormatter) Usage(command string) string {
	return fmt.Sprintf("**Usage**:\n```%s```\n", command)
}

func (f *ResponseFormatter) Examples(examples []string) string {
	var sb strings.Builder
	sb.WriteString("**Examples**:\n")
	for _, ex := range examples {
		sb.WriteString(fmt.Sprintf("`%s`\n", ex))
	}
	return sb.String()
}

func (f *ResponseFormatter) List(items []string) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("› %s\n", item))
	}
	return sb.String()
}

func (f *ResponseFormatter) Tip(text string) string {
	return fmt.Sprintf("**Tip**: %s\n", text)
}

func (f *ResponseFormatter) Section(emoji, title, content string) string {
	return fmt.Sprintf("%s **%s**\n%s\n", emoji, title, content)
}

func (f *ResponseFormatter) Combine(sections ...string) string {
	return strings.Join(sections, "\n")
}
