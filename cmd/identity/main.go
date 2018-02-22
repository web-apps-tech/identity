package main

import (
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	flags "github.com/jessevdk/go-flags"
	"github.com/web-apps-tech/identity"
)

type options struct {
	Addr        string `short:"a" long:"addr" default:":8080"`
	MySQLAddr   string `long:"mysql-address" env:"IDENTITY_MYSQL_ADDR" required:"true"`
	MySQLUser   string `long:"mysql-user" env:"IDENTITY_MYSQL_USER" required:"true"`
	MySQLPasswd string `long:"mysql-passwd" env:"IDENTITY_MYSQL_PASSWORD"`
	MySQLDB     string `long:"mysql-db" env:"IDENTITY_MYSQL_DB"`
	RedisAddr   string `long:"redis-address" env:"IDENTITY_REDIS_ADDR" required:"true"`
}

func main() { os.Exit(exec()) }
func exec() int {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		log.Printf("error occured in parsing flags: %s\n", err)
		return 1
	}
	cfg := mysql.Config{
		Net:       "tcp",
		Addr:      opts.MySQLAddr,
		User:      opts.MySQLUser,
		Passwd:    opts.MySQLPasswd,
		DBName:    opts.MySQLDB,
		ParseTime: true,
	}
	s := identity.NewServer(cfg, opts.RedisAddr)
	if err := s.Run(opts.Addr); err != nil {
		log.Printf("error occured in running: %s\n", err)
		return 1
	}
	return 0
}
