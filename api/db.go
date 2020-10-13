package main

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
)

var db *sql.DB

var schemaImport struct {
	sync.Mutex

	done uint32
}

func initDb() {
	dsn, ok := os.LookupEnv("VOTEAPI_DB")
	if !ok {
		log.WithFields(log.Fields{"var": "VOTEAPI_DB"}).Fatal("Env var missing")
	}

	{
		var errOp error
		if db, errOp = sql.Open("postgres", dsn); errOp != nil {
			log.WithFields(log.Fields{
				"var": "VOTEAPI_DB", "driver": "postgres", "error": errOp.Error(),
			}).Fatal("Bad database DSN")
		}
	}

	onTerm.ToDo = append(onTerm.ToDo, func() {
		_ = db.Close()
	})
}

func ensureSchema() bool {
	if atomic.LoadUint32(&schemaImport.done) == 0 {
		schemaImport.Lock()
		defer schemaImport.Unlock()

		if atomic.LoadUint32(&schemaImport.done) == 0 {
			if errIS := doTx(false, importSchema); errIS != nil {
				log.WithFields(log.Fields{"error": errIS.Error()}).Error("Couldn't create database schema")
				return false
			}

			atomic.StoreUint32(&schemaImport.done, 1)
		}
	}

	return true
}

func importSchema(tx *sql.Tx) error {
	{
		_, errEx := tx.Exec(`CREATE TABLE IF NOT EXISTS state (
	int_id  SMALLSERIAL PRIMARY KEY,
	ext_id  UUID NOT NULL UNIQUE,
	ru_name VARCHAR(255) NOT NULL
)`)
		if errEx != nil {
			return errEx
		}
	}

	_, errEx := tx.Exec(`CREATE TABLE IF NOT EXISTS office (
	int_id  SERIAL PRIMARY KEY,
	ext_id  UUID NOT NULL UNIQUE,
	state   SMALLINT NOT NULL REFERENCES state(int_id),
	ru_name VARCHAR(255) NOT NULL
)`)
	return errEx
}

func doTx(ro bool, f func(tx *sql.Tx) error) error {
	for {
		tx, errBg := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: ro})
		if errBg != nil {
			return errBg
		}

		if errTx := f(tx); errTx != nil {
			_ = tx.Rollback()

			if retryTx(errTx) {
				continue
			}

			return errTx
		}

		if errCm := tx.Commit(); errCm != nil {
			_ = tx.Rollback()

			if retryTx(errCm) {
				continue
			}

			return errCm
		}

		return nil
	}
}

func retryTx(err error) bool {
	errPq, ok := err.(*pq.Error)
	return ok && errPq.Code == "40001"
}

func fetchAll(db interface {
	Query(string, ...interface{}) (*sql.Rows, error)
}, rowType interface{}, query string, args ...interface{}) (interface{}, error) {
	rows, errQr := db.Query(query, args...)
	if errQr != nil {
		return nil, errQr
	}

	defer rows.Close()

	blankRow := reflect.ValueOf(rowType)
	res := reflect.MakeSlice(reflect.SliceOf(blankRow.Type()), 0, 0)
	idx := -1
	scanDest := make([]interface{}, blankRow.NumField())

	for {
		if rows.Next() {
			res = reflect.Append(res, blankRow)
			idx++

			row := res.Index(idx)

			for i := range scanDest {
				scanDest[i] = row.Field(i).Addr().Interface()
			}

			if errSc := rows.Scan(scanDest...); errSc != nil {
				return nil, errSc
			}
		} else if errNx := rows.Err(); errNx == nil {
			break
		} else {
			return nil, errNx
		}
	}

	return res.Interface(), nil
}
