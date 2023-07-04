package main

import (
	"flag"
	"net/http"
	"os"
	"sync/atomic"


	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	lecho "github.com/ziflex/lecho/v3"
)

func init() {
	path, err := os.Getwd()
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	log.Info().Err(err).Msg(`path: ` + path)

}

var clientStart uint64 = 0
var timeStart uint64 = 0

type responseHttp struct {
	ClientStart uint64 `json:"clientStart" validate:""`
	TimeStart   uint64 `json:"timeStart" validate:"required"`
}

func tokenHandler(e echo.Context) error {
	response := responseHttp{ClientStart: clientStart, TimeStart: timeStart}
	e.JSON(http.StatusOK, response)
	atomic.AddUint64(&timeStart, 10)
	atomic.AddUint64(&clientStart, 1)
	return nil
}


func main() {
	log.Info().Msg("Go Time")
	flag.Parse()

	

	e := echo.New()
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.With().Caller().Logger()
	logger := lecho.From(log.Logger)
	e.Logger = logger
	e.Use(lecho.Middleware(lecho.Config{
		Logger: logger}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.GET("/getToken", tokenHandler)

	log.Error().Err(e.Start(":8099")).Msg("")

}
