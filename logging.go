package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/kyleterry/tenyks/service"
)

const spaces = 40

var Logger log15.Logger

func LogFormat(r *log15.Record) []byte {
	buf := &bytes.Buffer{}

	buf.WriteString(r.Time.Format(time.RFC3339))
	buf.WriteByte(' ')

	switch r.Lvl {
	case log15.LvlDebug:
		buf.WriteString("[debug] ")
	case log15.LvlInfo:
		buf.WriteString("[info]  ")
	case log15.LvlWarn:
		buf.WriteString("[warn]  ")
	case log15.LvlError:
		buf.WriteString("[error] ")
	case log15.LvlCrit:
		buf.WriteString("[crit]  ")
	}

	ctx := make([]string, 0, len(r.Ctx)/2)

	for i := 0; i < len(r.Ctx); i += 2 {
		k := r.Ctx[i]
		v := r.Ctx[i+1]

		if k == "package" {
			buf.WriteString(fmt.Sprintf("%-11s", fmt.Sprintf("[%v]", v)))
		} else {
			ctx = append(ctx, fmt.Sprintf("%v=%v", k, v))
		}
	}

	if len(r.Msg) > 0 {
		buf.WriteByte(' ')
		buf.WriteString(r.Msg)
	}

	if len(ctx) > 0 {
		if len(r.Msg) > spaces-1 {
			buf.WriteByte(' ')
		} else {
			buf.WriteString(strings.Repeat(" ", spaces-len(r.Msg)))
		}

		buf.WriteString(strings.Join(ctx, " "))
	}

	buf.WriteByte('\n')

	return buf.Bytes()
}

func setupLogger(c config.LogConfig) {
	handler := log15.StreamHandler(os.Stdout, log15.FormatFunc(LogFormat))

	if !c.Debug {
		handler = log15.LvlFilterHandler(log15.LvlInfo, handler)
	}

	lg := log15.New()
	lg.SetHandler(handler)

	Logger = lg.New(log15.Ctx{"package": "main"})
	service.Logger = lg.New(log15.Ctx{"package": "service"})
	irc.Logger = lg.New(log15.Ctx{"package": "irc"})
}
