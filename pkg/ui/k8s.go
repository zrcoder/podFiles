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
	return app.Crud().Title("${i18n.k8s.namespace}").Name("ns").Api(api.Namespaces).SyncLocation(false).
		Columns(
			app.Column().Name("namespace"), //.Label("${i18n.k8s.name}"),
		).
		Filter(filter(app)).
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
	return app.Crud().Title("${i18n.k8s.pod}").Name("pods").Api(api.Pods).
		Columns(
			app.Column().Name("pod"), //.Label("${i18n.k8s.name}"),
		).
		Filter(filter(app)).
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
	return app.Crud().Name("containers").Api(api.Containers).
		Columns(
			app.Column().Name("container").Label("${i18n.k8s.container}"),
		).
		Filter(filter(app)).
		OnEvent(
			app.Event().RowClick(
				app.EventActions(
					app.EventAction().ActionType("ajax").Api("post:"+api.Containers+"?container=${event.data.item.container}"),
					app.EventAction().ActionType("link").Args(app.EventActionArgs().Link(FilesPage)),
				),
			),
		)
}
