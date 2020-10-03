package main

import (
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
)

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
