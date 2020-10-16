package main

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"strings"
)

func putStations(ctx iris.Context) {
	var payload struct {
		RuName   string    `json:"ru_name"`
		District uuid.UUID `json:"district"`
	}

	type office struct {
		IntId int32
	}

	type district struct {
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

	if payload.District == uuid.Nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{".district missing"})
		return
	}

	uid, errNR := uuid.NewRandom()
	if errNR != nil {
		ctx.StatusCode(500)
		ctx.JSON(errorResponse{errNR.Error()})
		return
	}

	var foundOffice, foundDistrict bool

	{
		errTx := doTx(false, func(tx *sql.Tx) error {
			rawOffices, errFA := fetchAll(tx, office{}, `SELECT int_id FROM office WHERE ext_id=$1`, extId)
			if errFA != nil {
				return errFA
			}

			offices := rawOffices.([]office)
			if foundOffice = len(offices) > 0; !foundOffice {
				return nil
			}

			rawDistricts, errFA := fetchAll(
				tx, district{}, `SELECT int_id FROM district WHERE ext_id=$1`, payload.District,
			)
			if errFA != nil {
				return errFA
			}

			districts := rawDistricts.([]district)
			if foundDistrict = len(districts) > 0; !foundDistrict {
				return nil
			}

			_, errEx := tx.Exec(
				`INSERT INTO station(ext_id, office, district, ru_name) VALUES ($1, $2, $3, $4)`,
				uid, offices[0].IntId, districts[0].IntId, payload.RuName,
			)
			return errEx
		})
		if errTx != nil {
			ctx.StatusCode(500)
			ctx.JSON(errorResponse{errTx.Error()})
			return
		}
	}

	if foundOffice {
		if foundDistrict {
			ctx.StatusCode(204)
		} else {
			ctx.StatusCode(404)
			ctx.JSON(errorResponse{"no such district"})
		}
	} else {
		ctx.StatusCode(404)
		ctx.JSON(errorResponse{"no such office"})
	}
}

func getStations(ctx iris.Context) {
	type office struct {
		IntId int32
	}

	type station struct {
		ExtId    uuid.UUID
		District uuid.UUID
		RuName   string
	}

	extId, errPU := uuid.Parse(ctx.Params().Get("ext_id"))
	if errPU != nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{errPU.Error()})
		return
	}

	var found bool
	var stations []station

	{
		errTx := doTx(true, func(tx *sql.Tx) error {
			rawOffices, errFA1 := fetchAll(tx, office{}, `SELECT int_id FROM office WHERE ext_id=$1`, extId)
			if errFA1 != nil {
				return errFA1
			}

			offices := rawOffices.([]office)
			if found = len(offices) > 0; !found {
				return nil
			}

			rawStations, errFA2 := fetchAll(
				tx, station{},
				"SELECT s.ext_id, d.ext_id, s.ru_name "+
					"FROM station s INNER JOIN district d ON d.int_id=s.district WHERE s.office=$1",
				offices[0].IntId,
			)
			if errFA2 != nil {
				return errFA2
			}

			stations = rawStations.([]station)
			return nil
		})
		if errTx != nil {
			ctx.StatusCode(500)
			ctx.JSON(errorResponse{errTx.Error()})
			return
		}
	}

	if found {
		type station struct {
			District uuid.UUID `json:"district"`
			RuName   string    `json:"ru_name"`
		}

		res := make(map[uuid.UUID]station, len(stations))

		for _, row := range stations {
			res[row.ExtId] = station{row.District, row.RuName}
		}

		ctx.JSON(res)
	} else {
		ctx.StatusCode(404)
		ctx.JSON(errorResponse{"no such office"})
	}
}

func postStations(ctx iris.Context) {
	var payload struct {
		RuName   string    `json:"ru_name"`
		District uuid.UUID `json:"district"`
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

	if payload.District == uuid.Nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{".district missing"})
		return
	}

	var foundDistrict, foundStation bool

	{
		errTx := doTx(false, func(tx *sql.Tx) error {
			rawDistricts, errFA := fetchAll(
				tx, row{}, `SELECT int_id FROM district WHERE ext_id=$1`, payload.District,
			)
			if errFA != nil {
				return errFA
			}

			districts := rawDistricts.([]row)
			if foundDistrict = len(districts) > 0; !foundDistrict {
				return nil
			}

			res, errEx := tx.Exec(
				`UPDATE station SET ru_name=$1, district=$2 WHERE ext_id=$3`, payload.RuName, districts[0].IntId, extId,
			)
			if errEx != nil {
				return errEx
			}

			rows, errRA := res.RowsAffected()
			if errRA != nil {
				return errRA
			}

			foundStation = rows > 0

			return nil
		})
		if errTx != nil {
			ctx.StatusCode(500)
			ctx.JSON(errorResponse{errTx.Error()})
			return
		}
	}

	if foundDistrict {
		if foundStation {
			ctx.StatusCode(204)
		} else {
			ctx.StatusCode(404)
			ctx.JSON(errorResponse{"no such station"})
		}
	} else {
		ctx.StatusCode(404)
		ctx.JSON(errorResponse{"no such district"})
	}
}

func deleteStations(ctx iris.Context) {
	extId, errPU := uuid.Parse(ctx.Params().Get("ext_id"))
	if errPU != nil {
		ctx.StatusCode(400)
		ctx.JSON(errorResponse{errPU.Error()})
		return
	}

	var found bool

	{
		errTx := doTx(false, func(tx *sql.Tx) error {
			res, errEx := tx.Exec(`DELETE FROM station WHERE ext_id=$1`, extId)
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
		ctx.JSON(errorResponse{"no such station"})
	}
}
