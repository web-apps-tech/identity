package identity

import (
	"database/sql"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

func doHealthCheck(w http.ResponseWriter, r *http.Request, tx *sql.Tx, kvs redis.Conn) {
}

func doSession(w http.ResponseWriter, r *http.Request, tx *sql.Tx, kvs redis.Conn) {
}
