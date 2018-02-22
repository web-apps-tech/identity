package genutil

import (
	"io"
	"strings"
	"unicode"
)

func Write(out io.Writer, line ...string) error {
	_, err := out.Write([]byte(strings.Join(line, "")))
	return err
}

func Uppercamel(in string) string {
	var list []string
	for _, s := range strings.Split(in, " ") {
		r := []rune(s)
		r[0] = unicode.ToUpper(r[0])
		list = append(list, string(r))
	}
	return strings.Join(list, "")
}
