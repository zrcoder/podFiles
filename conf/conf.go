package conf

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zrcoder/amisgo/conf"
)

var (
	//go:embed i18n/en-US.json
	enUS json.RawMessage
	//go:embed i18n/zh-CN.json
	zhCN json.RawMessage
)

func commonOptions() []conf.Option {
	return []conf.Option{
		conf.WithTheme(conf.ThemeAng),
		conf.WithLocales(
			conf.Locale{Value: conf.LocaleZhCN, Label: "汉", Dict: zhCN},
			conf.Locale{Value: conf.LocaleEnUS, Label: "En", Dict: enUS},
		),
	}
}

const (
	nsBlackListEnv   = "NS_BLACK_LIST"
	servicePrefixEnv = "SVC_PREFIX"
	kubeConfigEnv    = "KUBECONFIG"
)

var (
	nsPrefixBlackList = []string{"kube-"}
	nsSuffixBlackList []string
	nsBlackList       []string
)

func init() {
	if os.Getenv("DEV") != "" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		gin.SetMode(gin.DebugMode)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
		gin.SetMode(gin.ReleaseMode)
	}

	list := strings.Split(os.Getenv(nsBlackListEnv), ",")
	slog.Debug("nsBlackList", "list", list)
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

func KubeConfigPath() string {
	return os.Getenv(kubeConfigEnv)
}
