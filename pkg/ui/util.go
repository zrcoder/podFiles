package ui

import (
	"github.com/zrcoder/amisgo"
	"github.com/zrcoder/amisgo/comp"
)

func page(app *amisgo.App, isHome bool, body ...any) comp.Page {
	toolbarBodyButtons := []any{
		app.LocaleButtonGroupSelect(),
	}
	if !isHome {
		toolbarBodyButtons = append(
			[]any{
				app.Button().Icon("fa fa-home").Label("${i18n.home}").ActionType("link").Link("/").ClassName("bg-none border-none"),
				app.Wrapper(),
			},
			toolbarBodyButtons...,
		)
	}
	return app.Page().
		Title("${i18n.name}").
		Toolbar(app.InputGroup().Body(toolbarBodyButtons...).ClassName("my-2")).
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
