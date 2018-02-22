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
	genutil.Write(buf, "\n", strconv.Quote("bytes"))
	genutil.Write(buf, "\n", strconv.Quote("database/sql"))
	genutil.Write(buf, "\n", strconv.Quote("encoding/json"))
	genutil.Write(buf, "\n", strconv.Quote("net/http"))
	genutil.Write(buf, "\n", strconv.Quote("sync"))
	genutil.Write(buf, "\n\n", strconv.Quote("github.com/garyburd/redigo/redis"))
	genutil.Write(buf, "\n", strconv.Quote("github.com/go-sql-driver/mysql"))
	genutil.Write(buf, "\n", strconv.Quote("github.com/gorilla/mux"))
	genutil.Write(buf, "\n)")
	genutil.Write(buf, "\ntype Server struct {")
	genutil.Write(buf, "\nRouter *mux.Router")
	genutil.Write(buf, "\nEnv *Environment")
	genutil.Write(buf, "\n\nonce sync.Once")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\n\ntype Environment struct {")
	genutil.Write(buf, "\nRDB *sql.DB")
	genutil.Write(buf, "\nKVS redis.Conn")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\n\ntype jsonErr struct {")
	genutil.Write(buf, "\nError string `json:\"error\"`")
	genutil.Write(buf, "\nMessage string `json:\"message\"`")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\n\nfunc NewServer(cfg mysql.Config, redisAddr string) (*Server, error) {")
	genutil.Write(buf, "\ndb, err := sql.Open(\"mysql\", cfg.FormatDSN())")
	genutil.Write(buf, "\nif err != nil {")
	genutil.Write(buf, "\nreturn nil, err")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\nconn, err := redis.Dial(\"tcp\", redisAddr)")
	genutil.Write(buf, "\nif err != nil {")
	genutil.Write(buf, "\nreturn nil, err")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\ns := Server{")
	genutil.Write(buf, "\nRouter: mux.NewRouter(),")
	genutil.Write(buf, "\nEnv: &Environment{")
	genutil.Write(buf, "\nRDB: db,")
	genutil.Write(buf, "\nKVS: conn,")
	genutil.Write(buf, "\n},")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\nreturn &s, nil")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\n\nfunc (s *Server) bindRoutes() {")
	genutil.Write(buf, "\ns.once.Do(func(){")
	for _, link := range hs.Links {
		if err := generateRoutes(buf, link); err != nil {
			return nil, err
		}
	}
	genutil.Write(buf, "\n})")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\n\nfunc (s *Server) Run(addr string) error {")
	genutil.Write(buf, "\nreturn http.ListenAndServe(addr, s.Router)")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\n\nfunc renderJSON(w http.ResponseWriter, status int, v interface{}) {")
	genutil.Write(buf, "\nif err, ok := v.(error); ok {")
	genutil.Write(buf, "\nje := jsonErr{")
	genutil.Write(buf, "\nError: http.StatusText(status),")
	genutil.Write(buf, "\nMessage: err.Error(),")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\nrenderJSON(w, status ,je)")
	genutil.Write(buf, "\nreturn")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\nbuf := bytes.Buffer{}")
	genutil.Write(buf, "\nif err := json.NewEncoder(&buf).Encode(v); err != nil {")
	genutil.Write(buf, "\nrenderJSON(w, http.StatusInternalServerError, err)")
	genutil.Write(buf, "\n}")
	genutil.Write(buf, "\nw.Header().Set(\"Content-Type\", \"application/json\")")
	genutil.Write(buf, "\nw.WriteHeader(status)")
	genutil.Write(buf, "\nbuf.WriteTo(w)")
	genutil.Write(buf, "\n}")

	for _, link := range hs.Links {
		if err := generateHandlers(buf, link); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func generateRoutes(out io.Writer, link *hschema.Link) error {
	ew := &errWriter{Writer: out}
	genutil.Write(ew, "\ns.Router.Handle(\"", link.Href, "\", ", genutil.Uppercamel(link.Title), "Handler(s.Env))")
	return ew.err
}

func generateHandlers(out io.Writer, link *hschema.Link) error {
	ew := &errWriter{Writer: out}
	genutil.Write(ew, "\n\nfunc ")
	genutil.Write(ew, genutil.Uppercamel(link.Title), "Handler(env *Environment) http.Handler {")
	genutil.Write(ew, "\nreturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {")
	if link.Method != "" {
		genutil.Write(ew, "\nif r.Method != ", strconv.Quote(link.Method), " {")
		genutil.Write(ew, "\nw.WriteHeader(http.StatusMethodNotAllowed)")
		genutil.Write(ew, "\nreturn")
		genutil.Write(ew, "\n}")
	}
	genutil.Write(ew, "\ntx, err := env.RDB.BeginTx(r.Context(), nil)")
	genutil.Write(ew, "\nif err != nil {")
	genutil.Write(ew, "\nw.WriteHeader(http.StatusInternalServerError)")
	genutil.Write(ew, "\nreturn")
	genutil.Write(ew, "\n}")
	genutil.Write(ew, "\ndo", genutil.Uppercamel(link.Title), "(w, r, tx, env.KVS)")
	genutil.Write(ew, "\n})")
	genutil.Write(ew, "\n}")
	return ew.err
}
