package main

import (
	"database/sql"
	"flag"
	"github.com/xeuus/amigo/pkg"
	"log"
	"os"
)

func getEnv(key, def string) string {
	op := os.Getenv(key)
	if op == "" {
		return def
	}
	return op
}

func main()  {
	path := flag.String("path", "migrations", "migrations path relative to current directory")
	flag.Parse()

	db, err := sql.Open(getEnv("DB_DRIVER", "mysql"), getEnv("DB_QUERY", ""))
	if err != nil {
		log.Fatal(err)
	}

	action := flag.Arg(0)
	option := flag.Arg(1)
	if action == "" {
		action = "create"
	}
	amigo.Migrate(*path, action, option, db)
}