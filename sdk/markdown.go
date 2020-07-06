package sdk

import (
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/russross/blackfriday/v2"
)

const extensions = blackfriday.NoIntraEmphasis |
	blackfriday.Tables |
	blackfriday.FencedCode |
	blackfriday.Autolink |
	blackfriday.Strikethrough |
	blackfriday.SpaceHeadings |
	blackfriday.NoEmptyLineBeforeBlock

// ConvertMarkdownToHTML will convert markdown to html
func ConvertMarkdownToHTML(text string) string {
	input := strings.ReplaceAll(text, "\r", "")
	output := blackfriday.Run([]byte(input), blackfriday.WithExtensions(extensions))
	return strings.TrimSpace(string(output))
}

// ConvertHTMLToMarkdown will convert HTML to Markdown
func ConvertHTMLToMarkdown(html string) (string, error) {
	conv := md.NewConverter("", true, nil)
	conv.Use(plugin.GitHubFlavored())
	markdown, err := conv.ConvertString(html)
	if err != nil {
		return "", err
	}
	return markdown, nil
}
