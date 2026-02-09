package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
)

// logPathLength is the number of path items to caller
// 1 - file.go:line
// 2 - folder/file.go:line
const logPathLength = 1

func NewContextWithLogger(ctx context.Context, debug bool) (context.Context, func()) {
	//	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
	//		pathItems := strings.Split(file, "/")
	//		if len(pathItems) > 4 {
	//			file = strings.Join(pathItems[len(pathItems)-logPathLength:], "/")
	//		}
	//		return file + ":" + strconv.Itoa(line)
	//	}

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return ""
	}

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Use a diode (ring buffer) for non-blocking logging
	// Size: 1000, Poll interval: 10ms
	wr := diode.NewWriter(os.Stdout, 1000, 5*time.Millisecond, func(missed int) {
		fmt.Printf("Logger Dropped %d messages\n", missed)
	})

	output := zerolog.ConsoleWriter{
		Out:        wr,
		TimeFormat: time.DateTime,
		PartsOrder: []string{
			zerolog.LevelFieldName,
			zerolog.TimestampFieldName,
			zerolog.CallerFieldName,
			zerolog.MessageFieldName,
		},
	}

	logger := zerolog.New(output).
		With().
		Timestamp().
		CallerWithSkipFrameCount(2).
		Logger()

	log.Logger = logger

	// Return context and a cleanup function to close the diode writer
	return log.With().Logger().WithContext(ctx), func() {
		wr.Close()
	}
}

func FromCtx(ctx context.Context) *zerolog.Logger {
	return log.Ctx(ctx)
}
