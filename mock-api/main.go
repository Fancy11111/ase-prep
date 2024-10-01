package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Fancy11111/ase-prep/mock-api/stage"
	"github.com/Fancy11111/ase-prep/mock-api/token"
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

func createHandler[T any, S any](stage stage.Stage[T, S]) Handler[T, S] {
	return Handler[T, S]{
		tm:    token.NewTokenManagerInMemory(),
		stage: stage,
	}
}

func registerHandlers(mux *http.ServeMux) {

	handler := createHandler(stage.NewStagePointC())
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "pong!")
	})

	mux.HandleFunc("GET /assignment/{mnr}/token", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Str("mnr", mnr).Msg("")
		handler.getToken(w, r, mnr)
	})

	mux.HandleFunc("GET /assignment/{mnr}/token/reset", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Str("mnr", mnr).Msg("")
		handler.tm.ResetToken(mnr)
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /assignment/{mnr}/stage/{stage}/testcase/{testcase}", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		stageNr := r.PathValue("stage")
		testcase := r.PathValue("testcase")
		token := r.URL.Query().Get("token")

		testcaseNr, err := strconv.Atoi(testcase)
		if err != nil {
			log.Err(err).Str("testcase", testcase).Msg("Could not parse testcase number")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Could not parse testcase number")
			return
		}

		log.Info().Any("url", r.URL.Path).
			Any("method", r.Method).
			Str("mnr", mnr).
			Str("stage", stageNr).
			Str("testcase", testcase).
			Str("token", token).
			Msg("")

		handler.getTestcase(w, r, TestcaseInfo{
			mnr:      mnr,
			stage:    stageNr,
			testcase: testcaseNr,
			token:    token,
		})
	})

	mux.HandleFunc("POST /assignment/{mnr}/stage/{stage}/testcase/{testcase}", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		stageNr := r.PathValue("stage")
		testcase := r.PathValue("testcase")
		token := r.URL.Query().Get("token")

		testcaseNr, err := strconv.Atoi(testcase)
		if err != nil {
			log.Err(err).Str("testcase", testcase).Msg("Could not parse testcase number")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Could not parse testcase number")
			return
		}

		log.Info().Any("url", r.URL.Path).
			Any("method", r.Method).
			Str("mnr", mnr).
			Str("stage", stageNr).
			Str("testcase", testcase).
			Str("token", token).
			Msg("")

		handler.postTestResult(w, r, TestcaseInfo{
			mnr:      mnr,
			stage:    stageNr,
			testcase: testcaseNr,
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

		handler.getFinish(w, r)
	})
	// mux.HandleFunc("")
}

type Handler[T any, S any] struct {
	tm    token.TokenManager
	stage stage.Stage[stage.TestCase, stage.Solution]
}

func (h Handler[T, S]) getToken(w http.ResponseWriter, r *http.Request, mnr string) {
	w.WriteHeader(http.StatusOK)

	token, err := h.tm.GetToken(mnr)
	if err == nil {
		io.WriteString(w, token)
	}
}

type TestcaseInfo struct {
	mnr      string
	stage    string
	testcase int
	token    string
}

func (h Handler[T, S]) getTestcase(w http.ResponseWriter, r *http.Request, ti TestcaseInfo) {

	valid, err := h.tm.ValidateToken(ti.mnr, ti.token)

	if !valid {
		io.WriteString(w, fmt.Sprintf("%v", err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	testcase := h.stage.CreateTestcase(ti.token, ti.testcase)

	encoded, err := json.Marshal(testcase)
	if err != nil {
		log.Err(err).Msg("Could not marshal testcase")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

func (h Handler[T, S]) postTestResult(w http.ResponseWriter, r *http.Request, ti TestcaseInfo) {
	log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")

	valid, err := h.tm.ValidateToken(ti.mnr, ti.token)

	if !valid {
		io.WriteString(w, fmt.Sprintf("%v", err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	defer r.Body.Close()
	solution := stage.Solution{}
	err = json.NewDecoder(r.Body).Decode(&solution)

	if err != nil {
		log.Err(err).Msg("Could not unmarshal solution")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Could not parse solution")
		return
	}

	h.stage.ValidateSolution(ti.token, ti.testcase, solution)

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "TODO: next testcase link")
}

func (h Handler[T, S]) getFinish(w http.ResponseWriter, r *http.Request) {
	log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "TODO")
}
