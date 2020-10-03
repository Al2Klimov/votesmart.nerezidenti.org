package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"os"
)

var db *sql.DB

func initDb() {
	dsn, ok := os.LookupEnv("VOTEAPI_DB")
	if !ok {
		logrus.WithFields(logrus.Fields{"var": "VOTEAPI_DB"}).Fatal("Env var missing")
	}

	{
		var errOp error
		if db, errOp = sql.Open("postgres", dsn); errOp != nil {
			logrus.WithFields(logrus.Fields{
				"var": "VOTEAPI_DB", "driver": "postgres", "error": errOp.Error(),
			}).Fatal("Bad database DSN")
		}
	}

	onTerm.ToDo = append(onTerm.ToDo, func() {
		_ = db.Close()
	})
}
