package main

import (
	"bytes"
	"go/format"
	"io"
	"log"
	"os"
	"strconv"

	flags "github.com/jessevdk/go-flags"
	hschema "github.com/lestrrat/go-jshschema"
	"github.com/web-apps-tech/identity/internal/genutil"
)

const (
	flagParseErr = iota + 1
	schemaParseErr
	processErr
	fileopenErr
)

type options struct {
	Filename string `short:"f" long:"filename" description:"spec filename" required:"true"`
	Output   string `short:"o" long:"output" description:"output filename"`
}

type errWriter struct {
	io.Writer
	err error
}

func (ew *errWriter) Write(b []byte) (int, error) {
	if ew.err != nil {
		return -1, ew.err
	}
	var n int
	n, ew.err = ew.Writer.Write(b)
	return n, ew.err
}

func main() { os.Exit(exec()) }
func exec() int {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		log.Println("parse flag error: ", err)
		return flagParseErr
	}
	hs, err := hschema.ReadFile(opts.Filename)
	if err != nil {
		log.Println("parse schema error: ", err)
		return schemaParseErr
	}
	buf, err := process(hs)
	if err != nil {
		log.Println("processing error: ", err)
		return processErr
	}
	source, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	var out io.Writer
	if opts.Output == "" {
		out = os.Stdout
	} else {
		out, err = os.OpenFile(opts.Output, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fileopenErr
		}
	}
	out.Write(source)
	return 0
}

func process(hs *hschema.HyperSchema) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	genutil.Write(buf, "package identity")
	genutil.Write(buf, "\n\nimport (")
	genutil.Write(buf, "\n", strconv.Quote("database/sql"))
	genutil.Write(buf, "\n", strconv.Quote("net/http"))
	genutil.Write(buf, "\n\n", strconv.Quote("github.com/garyburd/redigo/redis"))
	genutil.Write(buf, "\n)")

	for _, link := range hs.Links {
		if err := generateDo(buf, link); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func generateDo(out io.Writer, link *hschema.Link) error {
	ew := &errWriter{Writer: out}
	genutil.Write(ew, "\n\nfunc do", genutil.Uppercamel(link.Title), "(w http.ResponseWriter, r *http.Request, tx *sql.Tx, kvs redis.Conn) {")
	genutil.Write(ew, "\n}")
	return ew.err
}
