package middleware

import (
	"net/http"
	"time"

	"github.com/poteto-go/poteto"
	"github.com/poteto-go/poteto/constant"
)

// Example for Logrus (https://github.com/sirupsen/logrus)
// log := logrus.New()
// logConfig := middleware.DefaultRequestLoggerConfig
// logConfig.LogHandleFunc = func(ctx poteto.Context, rlv middleware.RequestLoggerValues) error {
//   if rlv.Error == nil {
//     log.WithFields(logrus.Fields{
//       "method":    rlv.Method,
//       "routePath": rlv.RoutePath,
//       "status":    rlv.Status,
//     }).Info("request")
//   } else {
//     log.WithFields(logrus.Fields{
//       "method":    rlv.Method,
//       "routePath": rlv.RoutePath,
//       "status":    rlv.Status,
//     }).Error("request")
//   }
//   return nil
// }
// p.Register(middleware.RequestLoggerWithConfig(logConfig))

type LogHandlerFunc func(ctx poteto.Context, rlv RequestLoggerValues) error

type RequestLoggerConfig struct {
	HasStatus        bool
	HasMethod        bool
	HasRoutePath     bool
	HasRequestID     bool
	HasUserAgent     bool
	HasRemoteIP      bool
	HasRealIP        bool
	HasHost          bool
	HasContentLength bool
	OpenHeaders      []string
	HasError         bool
	HasStartTime     bool
	HasEndTime       bool
	HasDuration      bool
	LogHandleFunc    LogHandlerFunc
}

var DefaultRequestLoggerConfig = RequestLoggerConfig{
	HasStatus:        true,
	HasMethod:        true,
	HasRoutePath:     true,
	HasRequestID:     true,
	HasUserAgent:     true,
	HasRemoteIP:      true,
	HasRealIP:        true,
	HasHost:          true,
	HasContentLength: true,
	OpenHeaders:      []string{},
	HasError:         true,
	HasStartTime:     true,
	HasEndTime:       true,
	HasDuration:      true,
}

type RequestLoggerValues struct {
	Status        int
	Method        string
	RoutePath     string
	RequestId     string
	UserAgent     string
	RemoteIP      string
	RealIP        string
	Host          string
	ContentLength string
	Headers       map[string][]string
	Error         error
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
}

func RequestLoggerWithConfig(config RequestLoggerConfig) poteto.MiddlewareFunc {
	headers := []string{}
	for i, v := range config.OpenHeaders {
		headers[i] = http.CanonicalHeaderKey(v)
	}

	return func(next poteto.HandlerFunc) poteto.HandlerFunc {
		return func(ctx poteto.Context) error {
			req := ctx.GetRequest()
			res := ctx.GetResponse()

			startTime := time.Now()

			rlv := RequestLoggerValues{}

			err := next(ctx)

			if config.HasStatus {
				rlv.Status = res.Status
			}

			if config.HasMethod {
				rlv.Method = req.Method
			}

			if config.HasRoutePath {
				rlv.RoutePath = ctx.GetPath()
			}

			if config.HasRequestID {
				rlv.RequestId = ctx.RequestId()
			}

			if config.HasUserAgent {
				rlv.UserAgent = req.UserAgent()
			}

			if config.HasRemoteIP {
				rlv.RemoteIP, err = ctx.GetRemoteIP()
				if err != nil {
					panic(err)
				}
			}

			if config.HasRealIP {
				rlv.RealIP, err = ctx.RealIP()
				if err != nil {
					panic(err)
				}
			}

			if config.HasHost {
				rlv.Host = req.Host
			}

			if config.HasContentLength {
				rlv.ContentLength = res.Header().Get(constant.HeaderContentLength)
			}

			if len(config.OpenHeaders) > 0 {
				rlv.Headers = map[string][]string{}
				for _, header := range headers {
					if values, ok := req.Header[header]; ok {
						rlv.Headers[header] = values
					}
				}
			}

			if config.HasError && err != nil {
				rlv.Error = err
			}

			if config.HasStartTime {
				rlv.StartTime = startTime
			}

			if config.HasEndTime || config.HasDuration {
				endTime := time.Now()
				if config.HasEndTime {
					rlv.EndTime = endTime
				}
				if config.HasDuration {
					rlv.Duration = endTime.Sub(startTime)
				}
			}

			if errOnLog := config.LogHandleFunc(ctx, rlv); errOnLog != nil {
				return errOnLog
			}

			return err
		}
	}
}
