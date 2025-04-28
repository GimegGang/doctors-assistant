package logger

import (
	"log/slog"
	"os"
	"regexp"
	"strings"
)

func New(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(NewMaskingHandler())
	case "prod":
		log = slog.New(NewMaskingHandler())
	}

	return log
}

func NewMaskingHandler() *slog.JSONHandler {
	paramRegex := regexp.MustCompile(`(?i)(user_?id)=[0-9]+`)

	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			keyLower := strings.ToLower(a.Key)
			if strings.Contains(keyLower, "user") && strings.Contains(keyLower, "id") {
				return slog.String(a.Key, "***")
			}
			if str, ok := a.Value.Any().(string); ok {
				masked := paramRegex.ReplaceAllString(str, `${1}=***`)
				if masked != str {
					return slog.String(a.Key, masked)
				}
			}

			return a
		},
	})
}
