package ui

import (
	"bytes"
	"encoding/json"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	chrome_styles "github.com/alecthomas/chroma/v2/styles"
	"strings"
)

func PrettyPrintJson(text string) string {
	builder := &strings.Builder{}
	formatter := formatters.TTY256
	style := chrome_styles.Get("github-dark")

	lexer := lexers.Fallback
	var prettyJSON bytes.Buffer
	if strings.Contains(text, "{") &&
		strings.Contains(text, "}") {
		json.Indent(&prettyJSON, []byte(text), "", "\t")
		lexer = lexers.Get("json")
	}

	iterator, _ := lexer.Tokenise(nil, prettyJSON.String())
	formatter.Format(builder, style, iterator)

	return builder.String()
}
