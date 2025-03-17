//go:build local_sdk

package conf

import (
	"net/http"

	sdk "gitee.com/rdor/amis-sdk/v6"
	"github.com/zrcoder/amisgo/conf"
)

func Options() []conf.Option {
	return append(commonOptions(), conf.WithLocalSdk(http.FS(sdk.FS)))
}
