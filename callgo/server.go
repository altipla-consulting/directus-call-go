package callgo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"sort"
	"time"
)

const levelTrace = slog.Level(-8)

type ServerOption func(r *serverOpts)

func WithSecurityToken(token string) ServerOption {
	return func(r *serverOpts) {
		r.securityToken = token
	}
}

func WithLogger(logger *slog.Logger) ServerOption {
	return func(r *serverOpts) {
		r.logger = logger
	}
}

type ErrorReporter interface {
	ReportError(ctx context.Context, err error)
	ReportPanics(ctx context.Context)
}

func WithErrorReporter(reporter ErrorReporter) ServerOption {
	return func(r *serverOpts) {
		r.reporter = reporter
	}
}

// PingFn is the default ping implementation.
func PingFn(ctx context.Context) (string, error) {
	return "pong", nil
}

type serverOpts struct {
	securityToken string
	logger        *slog.Logger
	reporter      ErrorReporter
}

func NewServer(opts ...ServerOption) (string, http.Handler) {
	cnf := serverOpts{}
	for _, opt := range opts {
		opt(&cnf)
	}

	if cnf.logger == nil {
		cnf.logger = slog.New(slog.Default().Handler())
	}

	Handle(PingFn)

	mux := http.NewServeMux()
	mux.HandleFunc("/__callgo/invoke", invokeHandler(cnf))
	mux.HandleFunc("/__callgo/functions", functionsHandler(cnf))

	return "/__callgo/", mux
}

type invokeRequest struct {
	FnName         string          `json:"fnname"`
	Accountability *Accountability `json:"accountability"`
	Payload        json.RawMessage `json:"payload"`
	Trigger        invokeTrigger   `json:"trigger"`
}

type invokeTrigger struct {
	Event      string          `json:"event"`
	Key        TriggerKey      `json:"key"`
	Keys       []TriggerKey    `json:"keys"`
	Collection string          `json:"collection"`
	Payload    json.RawMessage `json:"payload"`

	// Manual invokations.
	Path string     `json:"path"`
	Body invokeBody `json:"body"`
}

type invokeBody struct {
	Collection string       `json:"collection"`
	Keys       []TriggerKey `json:"keys"`
}

type invokeReply struct {
	Payload     any          `json:"payload,omitempty"`
	Error       string       `json:"error,omitempty"`
	CallGoError *callGoError `json:"callGoError,omitempty"`
}

func invokeHandler(cnf serverOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cnf.reporter != nil {
			defer cnf.reporter.ReportPanics(r.Context())
		}

		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if cnf.securityToken != "" && r.Header.Get("Authorization") != "Bearer "+cnf.securityToken {
			http.Error(w, "wrong authorization token", http.StatusUnauthorized)
			return
		}

		in, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot read request body: %s", err), http.StatusInternalServerError)
			return
		}
		cnf.logger.Log(r.Context(), levelTrace, "CallGo JSON Request", slog.String("body", string(in)))

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
			Path:       ir.Trigger.Path,
		}

		// Clone the single key as part of the list to look only for the list when implementing flows.
		// With this we can access the key both ways in the flow code.
		if !ir.Trigger.Key.IsEmpty() {
			trigger.Keys = append(trigger.Keys, ir.Trigger.Key)
		}

		// Manual invokation. We copy everything from its body
		if ir.Trigger.Body.Collection != "" {
			trigger.Collection = ir.Trigger.Body.Collection
			trigger.Keys = ir.Trigger.Body.Keys
		}

		keys := make([]string, len(trigger.Keys))
		for i, k := range trigger.Keys {
			keys[i] = k.String()
		}

		if ir.Accountability == nil {
			cnf.logger.InfoContext(r.Context(), "Function call", slog.String("fnname", ir.FnName))
		} else {
			cnf.logger.InfoContext(r.Context(), "Function call",
				slog.String("fnname", ir.FnName),
				slog.String("event", ir.Trigger.Event),
				slog.String("collection", ir.Trigger.Collection),
				slog.Any("keys", keys),
				slog.String("user", ir.Accountability.User))
		}

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
		var reply invokeReply
		switch len(out) {
		case 1:
			if err := out[0].Interface(); err != nil {
				emitUserError(cnf, r, w, ir, err.(error))
				return
			}

		case 2:
			if err := out[1].Interface(); err != nil {
				emitUserError(cnf, r, w, ir, err.(error))
				return
			}
			reply.Payload = out[0].Interface()

		default:
			panic("should not reach here")
		}

		resp, err := json.Marshal(reply)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot encode response data: %s", err), http.StatusInternalServerError)
			return
		}
		cnf.logger.DebugContext(r.Context(), "CallGo JSON Response", slog.String("body", string(resp)))

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintln(w, string(resp))
	}
}

func emitUserError(cnf serverOpts, r *http.Request, w http.ResponseWriter, ir invokeRequest, userError error) {
	var known *callGoError
	if errors.As(userError, &known) {
		cnf.logger.ErrorContext(r.Context(), "callgo: function returned error",
			slog.String("code", known.Code),
			slog.String("message", known.Message),
			slog.Any("extensions", known.Extensions),
			slog.String("fnname", ir.FnName))

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(invokeReply{CallGoError: known}); err != nil {
			http.Error(w, fmt.Sprintf("cannot encode error response: %s", err), http.StatusInternalServerError)
			return
		}
		return
	}

	cnf.logger.ErrorContext(r.Context(), "callgo: function unexpected error",
		slog.String("error", userError.Error()),
		slog.String("fnname", ir.FnName))

	if cnf.reporter != nil {
		cnf.reporter.ReportError(r.Context(), userError)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(invokeReply{Error: fmt.Sprint(userError)}); err != nil {
		http.Error(w, fmt.Sprintf("cannot encode error response: %s", err), http.StatusInternalServerError)
		return
	}
}

func functionsHandler(cnf serverOpts) http.HandlerFunc {
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
