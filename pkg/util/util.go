package util

import "github.com/russross/blackfriday/v2"

const extensions = blackfriday.NoIntraEmphasis |
	blackfriday.Tables |
	blackfriday.FencedCode |
	blackfriday.Autolink |
	blackfriday.Strikethrough |
	blackfriday.SpaceHeadings |
	blackfriday.NoEmptyLineBeforeBlock

func ConvertMarkdownToHTML(text string) string {
	output := blackfriday.Run([]byte(text), blackfriday.WithExtensions(extensions))
	return string(output)
}
