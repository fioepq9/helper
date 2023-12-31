package helper

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	gormHelper     *GormHelper
	gormHelperOnce sync.Once
)

type GormHelper struct {
	Config *gorm.Config
}

func Gorm(options ...func(*GormHelper)) *GormHelper {
	gormHelperOnce.Do(func() {
		gormHelper = &GormHelper{
			Config: &gorm.Config{
				Logger: NewGormZerologLogger(),
			},
		}
	})
	for _, opt := range options {
		opt(gormHelper)
	}
	return gormHelper
}

func (h *GormHelper) MySQL(
	host string,
	port int,
	user string,
	passwd string,
	dbname string,
) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, passwd, host, port, dbname,
	)
	return gorm.Open(mysql.Open(dsn), h.Config)
}

type GormZerologLogger struct {
	BackupWriter              io.Writer
	Level                     zerolog.Level
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func NewGormZerologLogger(options ...func(logger *GormZerologLogger)) *GormZerologLogger {
	log := &GormZerologLogger{
		Level:                     zerolog.TraceLevel,
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
	}

	for _, opt := range options {
		opt(log)
	}

	return log
}

func (l *GormZerologLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := GormZerologLogger{
		SlowThreshold: l.SlowThreshold,
	}
	switch level {
	case logger.Silent:
		newLogger.Level = zerolog.Disabled
	case logger.Error:
		newLogger.Level = zerolog.ErrorLevel
	case logger.Warn:
		newLogger.Level = zerolog.WarnLevel
	case logger.Info:
		newLogger.Level = zerolog.InfoLevel
	default:
		newLogger.Level = zerolog.TraceLevel
	}
	return &newLogger
}

func (l *GormZerologLogger) Info(ctx context.Context, msg string, data ...any) {
	log := zerolog.Ctx(ctx).Level(l.Level)
	log.Info().Msgf(msg, data...)
}

func (l *GormZerologLogger) Warn(ctx context.Context, msg string, data ...any) {
	log := zerolog.Ctx(ctx).Level(l.Level)
	log.Warn().Msgf(msg, data...)
}

func (l *GormZerologLogger) Error(ctx context.Context, msg string, data ...any) {
	log := zerolog.Ctx(ctx).Level(l.Level)
	log.Error().Msgf(msg, data...)
}

func (l *GormZerologLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	log := zerolog.Ctx(ctx).Level(l.Level)

	sql, rows := fc()
	elapsed := time.Since(begin)

	var evt *zerolog.Event
	if err != nil {
		switch {
		case l.IgnoreRecordNotFoundError && errors.Is(err, gorm.ErrRecordNotFound):
			evt = log.Trace().Str(zerolog.MessageFieldName, err.Error())
		default:
			evt = log.Err(err)
		}
	} else if elapsed > l.SlowThreshold {
		evt = log.Warn().Str(zerolog.MessageFieldName, fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold))
	} else {
		evt = log.Trace()
	}

	if rows != -1 {
		evt = evt.Int64("rows", rows)
	}

	if l.BackupWriter != nil {
		// write to back up writer
		l.BackupWriter.Write([]byte(sql + ";\n"))
	}

	evt.Str("sql", sql).Dur("elapsed", elapsed).Send()
}
