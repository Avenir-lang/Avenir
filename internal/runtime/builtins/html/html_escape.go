package html

import (
	"strings"
)

var htmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&#39;",
)

func escapeHTML(s string) string {
	return htmlEscaper.Replace(s)
}

func escapeAttrValue(s string) string {
	return htmlEscaper.Replace(s)
}
