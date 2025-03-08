package ui

import (
	"github.com/zrcoder/amisgo"
	"github.com/zrcoder/amisgo/comp"
)

func page(app *amisgo.App, body any) comp.Page {
	return app.Page().
		Title(app.Button().Icon("fa fa-home").Label("${i18n.name}").ActionType("link").Link("/").ClassName("bg-none")).
		Toolbar(
			app.InputGroup().Body(
				app.ThemeButtonGroupSelect(),
				app.Wrapper(),
				app.LocaleButtonGroupSelect(),
			),
		).
		Body(body)
}

func filter(app *amisgo.App) comp.Form {
	return app.Form().Title("").WrapWithPanel(false).Body(
		app.InputText().Name("keyword"),                        //.Label("${i18n.podFile.keywords}"),
		app.SubmitAction().Icon("fa fa-search").Primary(true),  //.Label("${i18n.podFile.search}")
		app.Action().Icon("fa fa-refresh").ActionType("reset"), //.Label("${i18n.podFile.reset}")
	).Actions()
}
