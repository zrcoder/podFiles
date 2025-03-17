//go:build !local_sdk

package conf

import (
	"github.com/zrcoder/amisgo/conf"
)

func Options() []conf.Option {
	return commonOptions()
}
