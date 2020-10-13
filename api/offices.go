package main

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"strings"
)

func putOffices(ctx iris.Context) {
	var payload struct {
		RuName string `json:"ru_name"`
	}

	type row struct {
		IntId int16
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

	if !ensureSchema() {
		ctx.StatusCode(500)
		return
	}

	uid, errNR := uuid.NewRandom()
	if errNR != nil {
		ctx.StatusCode(500)
		ctx.JSON(errorResponse{errNR.Error()})
		return
	}

	var found bool

	{
		errTx := doTx(false, func(tx *sql.Tx) error {
			rawRows, errFA := fetchAll(tx, row{}, `SELECT int_id FROM state WHERE ext_id=$1`, extId)
			if errFA != nil {
				return errFA
			}

			rows := rawRows.([]row)
			if found = len(rows) > 0; !found {
				return nil
			}

			_, errEx := tx.Exec(
				`INSERT INTO office(ext_id, state, ru_name) VALUES ($1, $2, $3)`, uid, rows[0].IntId, payload.RuName,
			)
			return errEx
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

func getOffices(ctx iris.Context) {
	type state struct {
		IntId int16
	}

	type office struct {
		ExtId  uuid.UUID
		RuName string
	}

	extId, errPU := uuid.Parse(ctx.Params().Get("ext_id"))
	if errPU != nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{errPU.Error()})
		return
	}

	if !ensureSchema() {
		ctx.StatusCode(500)
		return
	}

	var found bool
	var offices []office

	{
		errTx := doTx(true, func(tx *sql.Tx) error {
			rawStates, errFA1 := fetchAll(tx, state{}, `SELECT int_id FROM state WHERE ext_id=$1`, extId)
			if errFA1 != nil {
				return errFA1
			}

			states := rawStates.([]state)
			if found = len(states) > 0; !found {
				return nil
			}

			rawOffices, errFA2 := fetchAll(
				tx, office{}, "SELECT ext_id, ru_name FROM office WHERE state=$1", states[0].IntId,
			)
			if errFA2 != nil {
				return errFA2
			}

			offices = rawOffices.([]office)
			return nil
		})
		if errTx != nil {
			ctx.StatusCode(500)
			ctx.JSON(errorResponse{errTx.Error()})
			return
		}
	}

	if found {
		res := make(map[uuid.UUID]string, len(offices))

		for _, row := range offices {
			res[row.ExtId] = row.RuName
		}

		ctx.JSON(res)
	} else {
		ctx.StatusCode(404)
		ctx.JSON(errorResponse{"no such state"})
	}
}

func postOffices(ctx iris.Context) {
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

	if !ensureSchema() {
		ctx.StatusCode(500)
		return
	}

	var found bool

	{
		errTx := doTx(false, func(tx *sql.Tx) error {
			res, errEx := tx.Exec(`UPDATE office SET ru_name=$1 WHERE ext_id=$2`, payload.RuName, extId)
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
		ctx.JSON(errorResponse{"no such office"})
	}
}

func deleteOffices(ctx iris.Context) {
	extId, errPU := uuid.Parse(ctx.Params().Get("ext_id"))
	if errPU != nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{errPU.Error()})
		return
	}

	if !ensureSchema() {
		ctx.StatusCode(500)
		return
	}

	var found bool

	{
		errTx := doTx(false, func(tx *sql.Tx) error {
			res, errEx := tx.Exec(`DELETE FROM office WHERE ext_id=$1`, extId)
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
		ctx.JSON(errorResponse{"no such office"})
	}
}
