package ui

import (
	"github.com/zrcoder/amisgo"
	"github.com/zrcoder/amisgo/comp"
)

func page(app *amisgo.App, body ...any) comp.Page {
	return app.Page().
		Title("${i18n.name}").
		Toolbar(app.LocaleButtonGroupSelect()).ClassName("my-2").
		Body(body)
}

func crud(app *amisgo.App) comp.Crud {
	return app.Crud().
		SyncLocation(false).
		LoadDataOnce(true).
		HeaderToolbar(
			"switch-per-page",
			"pagination",
		)
}
