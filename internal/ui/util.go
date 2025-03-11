package ui

import (
	"github.com/zrcoder/amisgo"
	"github.com/zrcoder/amisgo/comp"
	"github.com/zrcoder/amisgo/schema"
)

func page(app *amisgo.App, body ...any) comp.Page {
	return app.Page().
		BodyClassName("bg-light").
		Title("${i18n.name}").
		Toolbar(app.LocaleButtonGroupSelect()).
		Body(body)
}

func crud(app *amisgo.App) comp.Crud {
	return app.Crud().
		SyncLocation(false).
		LoadDataOnce(true).
		AutoFillHeight(schema.Schema{"height": 580}).
		FooterToolbar(
			"statistics",
			"switch-per-page",
			"pagination",
		)
}
