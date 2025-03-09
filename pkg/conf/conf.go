package conf

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	sdk "gitee.com/rdor/amis-sdk/v6"
	"github.com/gin-gonic/gin"
	"github.com/zrcoder/amisgo/conf"
)

func init() {
	if os.Getenv("DEV") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		gin.SetMode(gin.DebugMode)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
		gin.SetMode(gin.ReleaseMode)
	}
}

var (
	//go:embed i18n/en-US.json
	enUS json.RawMessage
	//go:embed i18n/zh-CN.json
	zhCN json.RawMessage
)

func Options() []conf.Option {
	return []conf.Option{
		conf.WithLocalSdk(http.FS(sdk.FS)),
		conf.WithTheme(conf.ThemeAng),
		conf.WithLocales(
			conf.Locale{Value: conf.LocaleZhCN, Label: "汉", Dict: zhCN},
			conf.Locale{Value: conf.LocaleEnUS, Label: "En", Dict: enUS},
		),
	}
}
