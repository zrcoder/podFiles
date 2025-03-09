package ui

import (
	"github.com/zrcoder/podFiles/pkg/api"

	"github.com/zrcoder/amisgo"
	"github.com/zrcoder/amisgo/comp"
)

const (
	FilesPage = "/files"
)

func Index(app *amisgo.App) comp.Page {
	return page(app, app.Group().Body(nsList(app), podList(app), containerList(app)))
}

func nsList(app *amisgo.App) comp.Crud {
	return crud(app).Name("ns").Api(api.Namespaces).
		Columns(
			app.Column().Name("namespace").Searchable(true).Label("${i18n.k8s.namespace}"),
		).
		OnEvent(
			app.Event().RowClick(
				app.EventActions(
					app.EventAction().ActionType("ajax").Api("post:"+api.Namespaces+"?namespace=${event.data.item.namespace}"),
					app.EventAction().ActionType("reload").ComponentName("pods"),
					app.EventAction().ActionType("reload").ComponentName("containers"),
				),
			),
		)
}

func podList(app *amisgo.App) comp.Crud {
	return crud(app).Name("pods").Api(api.Pods).
		Columns(
			app.Column().Name("pod").Searchable(true).Label("${i18n.k8s.pod}"),
		).
		OnEvent(
			app.Event().RowClick(
				app.EventActions(
					app.EventAction().ActionType("ajax").Api("post:"+api.Pods+"?pod=${event.data.item.pod}"),
					app.EventAction().ActionType("reload").ComponentName("containers"),
				),
			),
		)
}

func containerList(app *amisgo.App) comp.Crud {
	return crud(app).Name("containers").Api(api.Containers).
		Columns(
			app.Column().Name("container").Searchable(true).Label("${i18n.k8s.container}"),
		).
		OnEvent(
			app.Event().RowClick(
				app.EventActions(
					app.EventAction().ActionType("ajax").Api("post:"+api.Containers+"?container=${event.data.item.container}"),
					app.EventAction().ActionType("link").Args(app.EventActionArgs().Link(FilesPage)),
				),
			),
		)
}
