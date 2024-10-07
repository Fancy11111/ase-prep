package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Fancy11111/ase-prep/mock-api/stage"
	"github.com/Fancy11111/ase-prep/mock-api/token"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("exited with error")
	}
}

func run() (err error) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	mux := http.NewServeMux()
	registerHandlers(mux)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := os.Getenv("ADDR")
	fmt.Println(addr)

	if addr == "" {
		addr = "localhost"
	}

	port := os.Getenv("PORT")
	fmt.Println(port)

	if port == "" {
		port = "3000"
	}

	baseAddr := fmt.Sprintf("%s:%s", addr, port)
	ctx = context.WithValue(ctx, "baseAddr", baseAddr)

	/*	otelShutdown, err := setupOTelSDK(ctx)
		if err != nil {
			return
		}
		// Handle shutdown properly so nothing leaks.
		defer func() {
			err = errors.Join(err, otelShutdown(context.Background()))
		}()*/

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	srvErr := make(chan error, 1)
	go func() {
		log.Info().Str("baseAddress", baseAddr).Msg("Starting server")
		srvErr <- server.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	/*go func() {
		signalHandler := make(chan os.Signal, 1)
		signal.Notify(signalHandler, syscall.SIGINT, syscall.SIGTERM)

		signal := <-signalHandler
		log.Info().Any("signal", signal).Msg("Got signal, shutting down")

		server.Close()
	}()*/

	/*if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Err(err).Msg("Error while running server")
	} else {
		log.Info().Msg("Server shut down gracefully")
	}*/

	//cancelCtx()
	return
}

func createHandler(stage stage.StagePointsC) Handler {
	return Handler{
		tm:    token.NewTokenManagerInMemory(),
		stage: stage,
	}
}

func registerHandlers(mux *http.ServeMux) {

	handler := createHandler(stage.NewStagePointC())

	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		h := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, h)
	}

	handleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "pong!")
	})

	handleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, "healthy")
	})

	handleFunc("GET /assignment/{mnr}/token", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Str("mnr", mnr).Msg("")
		handler.getToken(w, r, mnr)
	})

	handleFunc("GET /assignment/{mnr}/token/reset", func(w http.ResponseWriter, r *http.Request) {
		mnr := r.PathValue("mnr")
		log.Info().Any("url", r.URL.Path).Any("method", r.Method).Str("mnr", mnr).Msg("")
		handler.tm.ResetToken(mnr)
		w.WriteHeader(http.StatusOK)
	})

	handleFunc("GET /assignment/{mnr}/stage/{stage}/testcase/{testcase}", func(w http.ResponseWriter, r *http.Request) {
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

	handleFunc("POST /assignment/{mnr}/stage/{stage}/testcase/{testcase}", func(w http.ResponseWriter, r *http.Request) {
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

	handleFunc("GET /assignment/{mnr}/finish", func(w http.ResponseWriter, r *http.Request) {
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

type Handler struct {
	tm    token.TokenManager
	stage stage.Stage[stage.TestCase, stage.Solution]
}

func (h Handler) getToken(w http.ResponseWriter, r *http.Request, mnr string) {
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

func (h Handler) getTestcase(w http.ResponseWriter, r *http.Request, ti TestcaseInfo) {

	valid, err := h.tm.ValidateToken(ti.mnr, ti.token)

	if !valid {
		io.WriteString(w, fmt.Sprintf("%v", err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	testcase := h.stage.CreateTestcase(ti.token, ti.testcase)

	log.Debug().Any("testcase", testcase).Msg("")

	encoded, err := json.Marshal(testcase)
	if err != nil {
		log.Err(err).Msg("Could not marshal testcase")
	}

	log.Debug().Str("encoded", string(encoded)).Msg("encoded testcase")

	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

func (h Handler) postTestResult(w http.ResponseWriter, r *http.Request, ti TestcaseInfo) {
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

	baseUrl := r.Context().Value("baseAddr")

	if ti.testcase >= 10 {
		io.WriteString(w, fmt.Sprintf("%s/assignment/%s/finish?token=%s", baseUrl, ti.mnr, ti.token))
		return
	}

	nextLink := fmt.Sprintf("%s/assignment/%s/stage/%s/testcase/%d?token=%s", baseUrl, ti.mnr, ti.stage, ti.testcase+1, ti.token)
	io.WriteString(w, nextLink)
}

func (h Handler) getFinish(w http.ResponseWriter, r *http.Request) {
	log.Info().Any("url", r.URL.Path).Any("method", r.Method).Msg("")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "TODO")
}
