package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	mux := http.NewServeMux()
	registerHandlers(mux)

	ctx, cancelCtx := context.WithCancel(context.Background())

	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		signalHandler := make(chan os.Signal, 1)
		signal.Notify(signalHandler, syscall.SIGINT, syscall.SIGTERM)

		<-signalHandler
		cancelCtx()
	}()

	log.Info().Msg("Starting server")
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("Error while running server")
	} else {
		log.Info().Msg("Server shut down gracefully")
	}

}

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "pong!")
	})

	mux.HandleFunc("GET /assignment/{mnr}/token", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Str("mnr", mnr).Msg("")
		getToken(w, r, mnr)
	})

	mux.HandleFunc("GET /assignment/{mnr}/stage/{stage}/testcase/{testcase}", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		stageNr := r.PathValue("stage")
		testcase := r.PathValue("testcase")
		token := r.URL.Query().Get("token")
		log.Info().Any("url", r.URL.Path).
			Any("method", r.Method).
			Str("mnr", mnr).
			Str("stage", stageNr).
			Str("testcase", testcase).
			Str("token", token).
			Msg("")

		getTestcase(w, r, TestcaseInfo{
			mnr:      mnr,
			stage:    stageNr,
			testcase: testcase,
			token:    token,
		})
	})

	mux.HandleFunc("POST /assignment/{mnr}/stage/{stage}/testcase/{testcase}", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		stageNr := r.PathValue("stage")
		testcase := r.PathValue("testcase")
		token := r.URL.Query().Get("token")

		log.Info().Any("url", r.URL.Path).
			Any("method", r.Method).
			Str("mnr", mnr).
			Str("stage", stageNr).
			Str("testcase", testcase).
			Str("token", token).
			Msg("")

		postTestResult(w, r, TestcaseInfo{
			mnr:      mnr,
			stage:    stageNr,
			testcase: testcase,
			token:    token,
		})
	})

	mux.HandleFunc("GET /assignment/{mnr}/finish", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		token := r.URL.Query().Get("token")

		log.Info().Any("url", r.URL.Path).
			Any("method", r.Method).
			Str("mnr", mnr).
			Str("token", token).
			Msg("")

		getFinish(w, r)
	})
	// mux.HandleFunc("")
}

func getToken(w http.ResponseWriter, r *http.Request, mnr string) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "TODO")
}

type TestcaseInfo struct {
	mnr      string
	stage    string
	testcase string
	token    string
}

func getTestcase(w http.ResponseWriter, r *http.Request, testcase TestcaseInfo) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "TODO")
}

func postTestResult(w http.ResponseWriter, r *http.Request, testcase TestcaseInfo) {
	log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "TODO")
}

func getFinish(w http.ResponseWriter, r *http.Request) {
	log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "TODO")
}
