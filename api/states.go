package main

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"strings"
)

func putStates(ctx iris.Context) {
	var payload struct {
		RuName string `json:"ru_name"`
	}

	if errRJ := ctx.ReadJSON(&payload); errRJ != nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{errRJ.Error()})
		return
	}

	if strings.TrimSpace(payload.RuName) == "" {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{".ru_name missing"})
		return
	}

	uid, errNR := uuid.NewRandom()
	if errNR != nil {
		ctx.StatusCode(500)
		ctx.JSON(errorResponse{errNR.Error()})
		return
	}

	{
		errTx := rwTx(func(tx *sql.Tx) error {
			_, errEx := tx.Exec(`INSERT INTO state(ext_id, ru_name) VALUES ($1, $2)`, uid, payload.RuName)
			return errEx
		})
		if errTx != nil {
			ctx.StatusCode(500)
			ctx.JSON(errorResponse{errTx.Error()})
			return
		}
	}

	ctx.StatusCode(204)
}

func getStates(ctx iris.Context) {
	type row struct {
		ExtId  uuid.UUID
		RuName string
	}

	if !ensureSchema() {
		ctx.StatusCode(500)
		return
	}

	rawRows, errFA := fetchAll(db, row{}, "SELECT ext_id, ru_name FROM state")
	if errFA != nil {
		log.WithFields(log.Fields{"error": errFA.Error()}).Error("Query error")
		ctx.StatusCode(500)
		return
	}

	rows := rawRows.([]row)
	res := make(map[uuid.UUID]string, len(rows))

	for _, row := range rows {
		res[row.ExtId] = row.RuName
	}

	_, _ = ctx.JSON(res)
}
