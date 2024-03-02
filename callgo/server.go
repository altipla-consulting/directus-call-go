package callgo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"syscall"
	"time"
)

type serveOpts struct {
	securityToken string
	logger        *slog.Logger
	port          string
}

type ServeOption func(r *serveOpts)

func WithSecurityToken(token string) ServeOption {
	return func(r *serveOpts) {
		r.securityToken = token
	}
}

func WithLogger(logger *slog.Logger) ServeOption {
	return func(r *serveOpts) {
		r.logger = logger
	}
}

func WithPort(port string) ServeOption {
	return func(r *serveOpts) {
		r.port = port
	}
}

// PingFn is the default ping implementation.
func PingFn(ctx context.Context) (string, error) {
	return "pong", nil
}

func Serve(opts ...ServeOption) {
	cnf := serveOpts{
		port: "8080",
	}
	for _, opt := range opts {
		opt(&cnf)
	}

	if cnf.logger == nil {
		cnf.logger = slog.New(slog.Default().Handler())
	}

	Handle(PingFn)

	http.HandleFunc("/__callgo/invoke", invokeHandler(cnf))
	http.HandleFunc("/__callgo/functions", functionsHandler(cnf))

	w := slog.New(cnf.logger.Handler())
	w = w.With("stdlib", "net/http")
	server := &http.Server{
		Addr:     ":" + cnf.port,
		ErrorLog: slog.NewLogLogger(w.Handler(), slog.LevelError),
	}

	go func() {
		cnf.logger.Info("Instance initialized successfully!", slog.String("port", cnf.port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cnf.logger.Error("could not listen and serve", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	signalctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()
	<-signalctx.Done()

	cnf.logger.Debug("Signal received, shutting down server")
	shutdownctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownctx); err != nil {
		cnf.logger.Error("could not shutdown server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	cnf.logger.Info("Server shutdown successfully!")
}

type invokeRequest struct {
	FnName         string          `json:"fnname"`
	Accountability *Accountability `json:"accountability"`
	Payload        json.RawMessage `json:"payload"`
	Trigger        *invokeTrigger  `json:"trigger"`
}

type invokeTrigger struct {
	Event      string          `json:"event"`
	Key        TriggerKey      `json:"key"`
	Keys       []TriggerKey    `json:"keys"`
	Collection string          `json:"collection"`
	Payload    json.RawMessage `json:"payload"`
}

func invokeHandler(cnf serveOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if cnf.securityToken != "" {
			if r.Header.Get("Authorization") != "Bearer "+cnf.securityToken {
				http.Error(w, "wrong authorization token", http.StatusUnauthorized)
				return
			}
		}

		in, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read request body: %s", err), http.StatusInternalServerError)
			return
		}
		cnf.logger.DebugContext(r.Context(), "JSON Request", slog.String("body", string(in)))

		var ir invokeRequest
		if err := json.Unmarshal(in, &ir); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		f, ok := funcs[ir.FnName]
		if !ok {
			cnf.logger.WarnContext(r.Context(), "Function not found", slog.String("fnname", ir.FnName))
			http.Error(w, fmt.Sprintf("function %q not found", ir.FnName), http.StatusNotFound)
			return
		}

		trigger := &RawTrigger{
			Event:      ir.Trigger.Event,
			Keys:       ir.Trigger.Keys,
			Collection: ir.Trigger.Collection,
			Payload:    ir.Trigger.Payload,
		}
		if !ir.Trigger.Key.IsEmpty() {
			trigger.Keys = append(trigger.Keys, ir.Trigger.Key)
		}

		keys := make([]string, len(trigger.Keys))
		for i, k := range trigger.Keys {
			keys[i] = k.String()
		}
		cnf.logger.InfoContext(r.Context(), "Function called",
			slog.String("fnname", ir.FnName),
			slog.String("event", ir.Trigger.Event),
			slog.String("collection", ir.Trigger.Collection),
			slog.Any("keys", keys),
			slog.String("user", ir.Accountability.User))

		ctx := r.Context()
		ctx = context.WithValue(ctx, accountabilityKey, ir.Accountability)
		ctx = context.WithValue(ctx, rawTriggerKey, trigger)
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		args := []reflect.Value{
			reflect.ValueOf(ctx),
		}
		if f.fv.Type().NumIn() == 2 {
			payload := reflect.New(f.fv.Type().In(1).Elem())
			if err := json.Unmarshal(ir.Payload, payload.Interface()); err != nil {
				cnf.logger.ErrorContext(r.Context(), "callgo: cannot decode request payload",
					slog.String("fnname", ir.FnName),
					slog.String("error", err.Error()),
					slog.String("payload", string(ir.Payload)),
					slog.String("target", fmt.Sprintf("%T", payload.Interface())))
				http.Error(w, fmt.Sprintf("cannot decode request payload: %s", err), http.StatusBadRequest)
				return
			}
			args = append(args, payload)
		}

		out := f.fv.Call(args)
		switch len(out) {
		case 1:
			if err := out[0].Interface(); err != nil {
				emitUserError(cnf, r, w, ir, err.(error))
				return
			}
			fmt.Fprintln(w, "{}")

		case 2:
			if err := out[1].Interface(); err != nil {
				emitUserError(cnf, r, w, ir, err.(error))
				return
			}

			out, err := json.Marshal(out[0].Interface())
			if err != nil {
				http.Error(w, fmt.Sprintf("cannot encode response data: %s", err), http.StatusInternalServerError)
				return
			}
			cnf.logger.DebugContext(r.Context(), "JSON Response", slog.String("body", string(out)))

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprintln(w, string(out))

		default:
			panic("should not reach here")
		}
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func emitUserError(cnf serveOpts, r *http.Request, w http.ResponseWriter, ir invokeRequest, err error) {
	cnf.logger.ErrorContext(r.Context(), "callgo: function call error",
		slog.String("error", err.Error()),
		slog.String("fnname", ir.FnName))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(errorResponse{Error: fmt.Sprint(err)}); err != nil {
		http.Error(w, fmt.Sprintf("cannot encode error response: %s", err), http.StatusInternalServerError)
		return
	}
}

func functionsHandler(cnf serveOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if cnf.securityToken != "" {
			if r.Header.Get("Authorization") != "Bearer "+cnf.securityToken {
				http.Error(w, "wrong authorization token", http.StatusUnauthorized)
				return
			}
		}

		var fns []string
		for fn := range funcs {
			fns = append(fns, fn)
		}
		sort.Strings(fns)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(fns); err != nil {
			http.Error(w, fmt.Sprintf("cannot encode function list: %s", err), http.StatusInternalServerError)
			return
		}
	}
}
