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
		errTx := doTx(false, func(tx *sql.Tx) error {
			_, errEx := tx.Exec(`INSERT INTO state(ext_id, ru_name) VALUES ($1, $2)`, uid, payload.RuName)
			return errEx
		})
		if errTx != nil {
			ctx.StatusCode(500)
			ctx.JSON(errorResponse{errTx.Error()})
			return
		}
	}

	ctx.StatusCode(201)
	ctx.JSON(struct {
		Id uuid.UUID `json:"id"`
	}{uid})
}

func getStates(ctx iris.Context) {
	type row struct {
		ExtId  uuid.UUID
		RuName string
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

func postStates(ctx iris.Context) {
	var payload struct {
		RuName string `json:"ru_name"`
	}

	extId, errPU := uuid.Parse(ctx.Params().Get("ext_id"))
	if errPU != nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{errPU.Error()})
		return
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

	var found bool

	{
		errTx := doTx(false, func(tx *sql.Tx) error {
			res, errEx := tx.Exec(`UPDATE state SET ru_name=$1 WHERE ext_id=$2`, payload.RuName, extId)
			if errEx != nil {
				return errEx
			}

			rows, errRA := res.RowsAffected()
			if errRA != nil {
				return errRA
			}

			found = rows > 0

			return nil
		})
		if errTx != nil {
			ctx.StatusCode(500)
			ctx.JSON(errorResponse{errTx.Error()})
			return
		}
	}

	if found {
		ctx.StatusCode(204)
	} else {
		ctx.StatusCode(404)
		ctx.JSON(errorResponse{"no such state"})
	}
}

func deleteStates(ctx iris.Context) {
	extId, errPU := uuid.Parse(ctx.Params().Get("ext_id"))
	if errPU != nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{errPU.Error()})
		return
	}

	var found bool

	{
		errTx := doTx(false, func(tx *sql.Tx) error {
			res, errEx := tx.Exec(`DELETE FROM state WHERE ext_id=$1`, extId)
			if errEx != nil {
				return errEx
			}

			rows, errRA := res.RowsAffected()
			if errRA != nil {
				return errRA
			}

			found = rows > 0

			return nil
		})
		if errTx != nil {
			ctx.StatusCode(500)
			ctx.JSON(errorResponse{errTx.Error()})
			return
		}
	}

	if found {
		ctx.StatusCode(204)
	} else {
		ctx.StatusCode(404)
		ctx.JSON(errorResponse{"no such state"})
	}
}
