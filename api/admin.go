package main

import (
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"os"
)

var adminHash []byte

func initAdmin() {
	name := os.Getenv("VOTEAPI_ADMIN_NAME")
	if name == "" {
		log.WithFields(log.Fields{"var": "VOTEAPI_ADMIN_NAME"}).Fatal("Env var missing")
	}

	pass := os.Getenv("VOTEAPI_ADMIN_PASSWORD")
	if pass == "" {
		log.WithFields(log.Fields{"var": "VOTEAPI_ADMIN_PASSWORD"}).Fatal("Env var missing")
	}

	var errGF error
	if adminHash, errGF = bcrypt.GenerateFromPassword([]byte(name+":"+pass), bcrypt.DefaultCost); errGF != nil {
		log.WithFields(log.Fields{
			"cost": bcrypt.DefaultCost, "error": errGF.Error(),
		}).Fatal("Couldn't hash admin credentials")
	}
}

func mustBeAdmin(ctx iris.Context) {
	user, pass, ok := ctx.Request().BasicAuth()
	if ok && bcrypt.CompareHashAndPassword(adminHash, []byte(user+":"+pass)) == nil {
		ctx.Next()
	} else {
		ctx.StatusCode(401)
		ctx.ContentType("text/plain")
		ctx.Write([]byte("Эх, чекисты! Пошли бы вы далеко и надолго."))
	}
}
