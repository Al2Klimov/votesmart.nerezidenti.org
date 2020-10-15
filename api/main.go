package main

import (
	"context"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type errorResponse struct {
	Error string `json:"error"`
}

var onTerm struct {
	sync.RWMutex

	ToDo []func()
}

func main() {
	initLogging()
	initAdmin()
	initDb()
	go wait4term()

	app := iris.Default()

	app.Put("/v1/states", mustBeAdmin, ensureSchema, putStates)
	app.Get("/v1/states", ensureSchema, getStates)
	app.Post("/v1/states/{ext_id:string}", mustBeAdmin, ensureSchema, postStates)
	app.Delete("/v1/states/{ext_id:string}", mustBeAdmin, ensureSchema, deleteStates)
	app.Put("/v1/states/{ext_id:string}/offices", mustBeAdmin, ensureSchema, putOffices)
	app.Get("/v1/states/{ext_id:string}/offices", ensureSchema, getOffices)
	app.Post("/v1/offices/{ext_id:string}", mustBeAdmin, ensureSchema, postOffices)
	app.Delete("/v1/offices/{ext_id:string}", mustBeAdmin, ensureSchema, deleteOffices)
	app.Put("/v1/districts", mustBeAdmin, ensureSchema, putDistricts)
	app.Get("/v1/districts", ensureSchema, getDistricts)
	app.Post("/v1/districts/{ext_id:string}", mustBeAdmin, ensureSchema, postDistricts)
	app.Delete("/v1/districts/{ext_id:string}", mustBeAdmin, ensureSchema, deleteDistricts)

	onTerm.Lock()
	onTerm.ToDo = append(onTerm.ToDo, func() {
		_ = app.Shutdown(context.Background())
	})
	onTerm.Unlock()

	_ = app.Run(iris.Addr("[::]:8080"), iris.WithoutStartupLog, iris.WithoutInterruptHandler)
}

func initLogging() {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		log.SetFormatter(&log.JSONFormatter{})
	}

	log.SetLevel(log.TraceLevel)
	log.SetOutput(os.Stdout)

	golog.InstallStd(log.StandardLogger())
	golog.SetLevel("debug")
}

func wait4term() {
	ch := make(chan os.Signal, 1)

	{
		signals := [2]os.Signal{syscall.SIGTERM, syscall.SIGINT}
		signal.Notify(ch, signals[:]...)
		log.WithFields(log.Fields{"signals": signals}).Trace("Listening for signals")
	}

	log.WithFields(log.Fields{"signal": <-ch}).Info("Terminating")

	onTerm.Lock()

	for _, f := range onTerm.ToDo {
		f()
	}

	os.Exit(0)
}
