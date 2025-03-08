package ui

import (
	"github.com/zrcoder/amisgo"
	"github.com/zrcoder/amisgo/comp"
	"github.com/zrcoder/podFiles/pkg/api"
)

func FileList(app *amisgo.App) comp.Page {
	return page(
		app,
		app.Crud().Name("files").Api(api.Files).
			Columns(
				app.Column().Name("display").Label("${i18n.podFile.fileName}"),
				app.Column().Name("size").Label("${i18n.podFile.fileSize}"),
				app.Column().Name("time").Label("${i18n.podFile.modifyTime}"),
				app.Column().Type("operation").Buttons(
					app.Button().
						Icon("fa fa-download").
						Label("${i18n.podFile.download}").
						ActionType("ajax").
						ConfirmText("${i18n.podFile.confirmDownload}"+"${event.data.item.name}").
						Api("post:"+api.Download+"?file=${event.data.item.name}"),
				),
			).
			Filter(filter(app)).
			OnEvent(
				app.Event().RowClick(
					app.EventActions(
					// TODO
					),
				),
			),
	)
}
