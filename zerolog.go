package helper

import (
	"io"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errbase"
	json "github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

var (
	zerologHelper     *ZerologHelper
	zerologHelperOnce sync.Once
)

type ZerologHelper struct{}

func Zerolog() *ZerologHelper {
	zerologHelperOnce.Do(func() {
		zerologHelper = &ZerologHelper{}
		zerologHelper.
			SetDefaultGlobalLevel().
			SetDefaultGlobalInterfaceMarshalFunc().
			SetDefaultGlobalErrorStackMarshaller()
	})
	return zerologHelper
}

func (h *ZerologHelper) SetDefaultGlobalLevel() *ZerologHelper {
	return h.SetGlobalLevel(zerolog.TraceLevel)
}

func (h *ZerologHelper) SetGlobalLevel(level zerolog.Level) *ZerologHelper {
	zerolog.SetGlobalLevel(level)
	return h
}

func (h *ZerologHelper) SetDefaultGlobalInterfaceMarshalFunc() *ZerologHelper {
	return h.SetInterfaceMarshalFunc(json.Marshal)
}

func (h *ZerologHelper) SetInterfaceMarshalFunc(fn func(any) ([]byte, error)) *ZerologHelper {
	zerolog.InterfaceMarshalFunc = fn
	return h
}

func (h *ZerologHelper) SetDefaultGlobalErrorStackMarshaller() *ZerologHelper {
	return h.SetErrorStackMarshaller(func(err error) any {
		lastStackTraceErr := err
		for ; err != nil; err = errors.Unwrap(err) {
			_, ok := err.(errbase.StackTraceProvider)
			if ok {
				lastStackTraceErr = err
			}
		}
		safeDetails := errors.GetSafeDetails(lastStackTraceErr).SafeDetails
		if len(safeDetails) == 1 {
			stackMsg, err := parsePII(safeDetails[0])
			if err != nil {
				return safeDetails[0]
			}
			return stackMsg
		}
		res := make([][]StackInfo, 0)
		for _, details := range safeDetails {
			stackMsg, err := parsePII(details)
			if err != nil {
				return safeDetails
			}
			res = append(res, stackMsg)
		}
		return res
	})
}

func (h *ZerologHelper) SetErrorStackMarshaller(fn func(error) any) *ZerologHelper {
	zerolog.ErrorStackMarshaler = fn
	return h
}

func (h *ZerologHelper) NewLogger(out io.Writer) *zerolog.Logger {
	log := zerolog.New(out).With().Timestamp().Caller().Stack().Logger()
	return &log
}

func (h *ZerologHelper) NewConsoleLogger() *zerolog.Logger {
	return h.NewLogger(h.ConsoleWriter())
}

func (h *ZerologHelper) ConsoleWriter() zerolog.ConsoleWriter {
	return zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.DateTime + " -0700"
	})
}

type StackInfo struct {
	Func string `json:"func"`
	File string `json:"file"`
}

func parsePII(detail string) ([]StackInfo, error) {
	s := strings.TrimSpace(detail)
	ss := strings.Split(s, "\n")
	if len(ss)%2 != 0 {
		return nil, errors.New("invalid PII-free strings")
	}
	var stackInfos []StackInfo
	for i := 0; i < len(ss); i += 2 {
		fn := ss[i]
		file := ss[i+1]
		stackInfos = append(stackInfos, StackInfo{
			Func: strings.TrimSpace(fn),
			File: strings.TrimSpace(file),
		})
	}

	return stackInfos, nil
}
