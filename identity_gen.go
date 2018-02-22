package identity

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
	Env    *Environment

	once sync.Once
}

type Environment struct {
	RDB *sql.DB
	KVS redis.Conn
}

type jsonErr struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func NewServer(cfg mysql.Config, redisAddr string) (*Server, error) {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	conn, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		return nil, err
	}
	s := Server{
		Router: mux.NewRouter(),
		Env: &Environment{
			RDB: db,
			KVS: conn,
		},
	}
	return &s, nil
}

func (s *Server) bindRoutes() {
	s.once.Do(func() {
		s.Router.Handle("/", HealthCheckHandler(s.Env))
		s.Router.Handle("/session", SessionHandler(s.Env))
	})
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.Router)
}

func renderJSON(w http.ResponseWriter, status int, v interface{}) {
	if err, ok := v.(error); ok {
		je := jsonErr{
			Error:   http.StatusText(status),
			Message: err.Error(),
		}
		renderJSON(w, status, je)
		return
	}
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		renderJSON(w, http.StatusInternalServerError, err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func HealthCheckHandler(env *Environment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tx, err := env.RDB.BeginTx(r.Context(), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		doHealthCheck(w, r, tx, env.KVS)
	})
}

func SessionHandler(env *Environment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tx, err := env.RDB.BeginTx(r.Context(), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		doSession(w, r, tx, env.KVS)
	})
}
