package conf

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	sdk "gitee.com/rdor/amis-sdk/v6"
	"github.com/gin-gonic/gin"
	"github.com/zrcoder/amisgo/conf"
)

const NsBlackListEnv = "NS_BLACK_LIST"

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

var (
	nsPrefixBlackList = []string{"kube-"}
	nsSuffixBlackList []string
	nsBlackList       []string
)

func init() {
	list := strings.Split(os.Getenv(NsBlackListEnv), ",")
	for _, ns := range list {
		ns = strings.TrimSpace(ns)
		if len(ns) == 0 {
			continue
		}
		if ns[0] == '*' {
			nsSuffixBlackList = append(nsSuffixBlackList, ns[1:])
		} else if ns[len(ns)-1] == '*' {
			nsPrefixBlackList = append(nsPrefixBlackList, ns[:len(ns)-1])
		} else {
			nsBlackList = append(nsBlackList, ns)
		}
	}
}

func NsInBlacklist(ns string) bool {
	for _, v := range nsPrefixBlackList {
		if strings.HasPrefix(ns, v) {
			return true
		}
	}
	for _, v := range nsSuffixBlackList {
		if strings.HasSuffix(ns, v) {
			return true
		}
	}
	for _, v := range nsBlackList {
		if ns == v {
			return true
		}
	}
	return false
}
