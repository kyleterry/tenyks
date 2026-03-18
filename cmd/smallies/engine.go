package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"go.e64ec.com/glerp"
)

// msgHandler pairs a compiled regex with a glerp lambda.
// A nil regex matches every message.
type msgHandler struct {
	re     *regexp.Regexp
	lambda glerp.Expr
}

// engine holds the glerp environment and the handlers registered by scripts.
type engine struct {
	mu       sync.Mutex
	env      *glerp.Environment
	handlers []msgHandler
}

func newEngine() *engine {
	e := &engine{}

	cfg := glerp.DefaultConfig()

	builtins := cfg.Builtins
	forms := cfg.Forms

	// (random-choice list) — returns a uniformly random element from a list.
	builtins["random-choice"] = func(args []glerp.Expr) (glerp.Expr, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("random-choice: expected 1 argument, got %d", len(args))
		}

		lst, ok := args[0].(*glerp.ListExpr)
		if !ok {
			return nil, fmt.Errorf("random-choice: expected a list, got %s", args[0].String())
		}

		elems := lst.Elements()
		if len(elems) == 0 {
			return nil, fmt.Errorf("random-choice: list is empty")
		}

		return elems[rand.IntN(len(elems))], nil
	}

	// (string-split str sep) — splits str on sep and returns a list of strings.
	// Leading/trailing whitespace is trimmed from each part.
	builtins["string-split"] = func(args []glerp.Expr) (glerp.Expr, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("string-split: expected 2 arguments, got %d", len(args))
		}

		s, ok := args[0].(*glerp.StringExpr)
		if !ok {
			return nil, fmt.Errorf("string-split: first argument must be a string, got %s", args[0].String())
		}

		sep, ok := args[1].(*glerp.StringExpr)
		if !ok {
			return nil, fmt.Errorf("string-split: second argument must be a string, got %s", args[1].String())
		}

		parts := strings.Split(s.Value(), sep.Value())
		var sb strings.Builder
		sb.WriteString("(list")
		for _, p := range parts {
			fmt.Fprintf(&sb, " %q", strings.TrimSpace(p))
		}
		sb.WriteString(")")

		// Evaluate the (list ...) literal in the root env to get a ListExpr back.
		results, err := glerp.Eval(sb.String(), e.env)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			return nil, fmt.Errorf("string-split: internal error constructing list")
		}

		return results[0], nil
	}

	forms["on-message"] = e.onMessageForm

	cfg.Builtins = builtins
	cfg.Forms = forms

	e.env = glerp.NewEnvironment(cfg)

	return e
}

// onMessageForm implements the (on-message pattern handler) special form.
//
// pattern is either a regex string or #f (match everything).
// handler must be a lambda accepting four arguments: (nick channel payload matches).
//   - nick: the sender's IRC nick
//   - channel: the destination (channel name or nick for DMs)
//   - payload: the full message text
//   - matches: a list of regex capture groups (empty list when pattern is #f)
//
// The lambda should return a string to send as a reply, or #f / void for no reply.
func (e *engine) onMessageForm(args []glerp.Expr, env *glerp.Environment) (glerp.Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("on-message: expected (pattern handler), got %d args", len(args))
	}

	patternExpr, err := args[0].Eval(env)
	if err != nil {
		return nil, err
	}

	handlerExpr, err := args[1].Eval(env)
	if err != nil {
		return nil, err
	}

	var re *regexp.Regexp
	switch p := patternExpr.(type) {
	case *glerp.StringExpr:
		re, err = regexp.Compile(p.Value())
		if err != nil {
			return nil, fmt.Errorf("on-message: invalid regex %q: %w", p.Value(), err)
		}
	case *glerp.BoolExpr:
		if p.Value() {
			return nil, fmt.Errorf("on-message: pattern must be a string or #f, got #t")
		}
		// #f means match everything; re stays nil
	default:
		return nil, fmt.Errorf("on-message: pattern must be a string or #f, got %T", patternExpr)
	}

	e.mu.Lock()
	e.handlers = append(e.handlers, msgHandler{re: re, lambda: handlerExpr})
	e.mu.Unlock()

	return glerp.Void(), nil
}

// loadDir evaluates every *.glerp file found directly inside dir.
func (e *engine) loadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".scm" {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		log.Printf("smallies: loading script %s", path)

		if err := glerp.EvalFile(path, e.env); err != nil {
			return fmt.Errorf("smallies: loading %s: %w", path, err)
		}
	}

	return nil
}

// dispatch runs all matching handlers for the given message and returns the
// replies that should be sent back. Handlers that return a string contribute
// a reply; void and #f returns are silently ignored.
//
// Each matching lambda is called with four arguments:
//
//	(lambda (nick channel payload matches) ...)
//
// matches is a list of regex capture groups (empty list for catch-all handlers).
func (e *engine) dispatch(nick, channel, payload string) []string {
	e.mu.Lock()
	handlers := make([]msgHandler, len(e.handlers))
	copy(handlers, e.handlers)
	e.mu.Unlock()

	log.Printf("smallies: dispatch payload=%q registered_handlers=%d", payload, len(handlers))

	var replies []string

	for _, h := range handlers {
		var groups []string
		if h.re != nil {
			m := h.re.FindStringSubmatch(payload)
			if m == nil {
				log.Printf("smallies: pattern %q did not match", h.re)
				continue
			}
			log.Printf("smallies: pattern %q matched", h.re)
			groups = m[1:] // capture groups only, skip full match
		}

		// Render the capture groups as a quoted glerp list literal so they
		// can be embedded directly into the Eval source string.
		var matchList strings.Builder
		matchList.WriteString("'(")
		for i, g := range groups {
			if i > 0 {
				matchList.WriteString(" ")
			}

			fmt.Fprintf(&matchList, "%q", g)
		}
		matchList.WriteString(")")

		callEnv := e.env.Extend()
		callEnv.Bind("__fn__", h.lambda)
		src := fmt.Sprintf("(__fn__ %q %q %q %s)", nick, channel, payload, matchList.String())

		log.Printf("smallies: evaluating %s", src)

		results, err := glerp.Eval(src, callEnv)
		if err != nil {
			log.Printf("smallies: handler error: %v", err)
			continue
		}

		if len(results) == 0 {
			continue
		}

		result := results[len(results)-1]
		switch r := result.(type) {
		case *glerp.StringExpr:
			replies = append(replies, r.Value())
		case *glerp.VoidExpr:
			// no reply
		case *glerp.BoolExpr:
			// no reply — #f (or #t) means the handler opted out
		default:
			// Numbers, lists, etc. — use string representation as a fallback
			if s := result.String(); s != "" {
				replies = append(replies, s)
			}
		}
	}

	return replies
}
