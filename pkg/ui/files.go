package ui

import (
	"github.com/zrcoder/amisgo"
	"github.com/zrcoder/amisgo/comp"
	"github.com/zrcoder/podFiles/pkg/api"
)

func FileList(app *amisgo.App) comp.Page {
	return page(
		app,
		false,
		crud(app).Name("files").Api(api.Files).
			Columns(
				app.Column().Name("name").Label("${i18n.podFile.fileName}").Searchable(true),
				app.Column().Name("size").Label("${i18n.podFile.fileSize}"),
				app.Column().Name("time").Label("${i18n.podFile.modifyTime}"),
				app.Column().Type("operation").Buttons(
					app.Button().
						Icon("fa fa-download").
						Label("${i18n.podFile.download}").
						ActionType("ajax").
						ConfirmText("${i18n.podFile.confirmDownload}"+"${event.data.item.name}").
						Api("post:"+api.Download+"?file=${event.data.item.name}"),
					app.Button().
						VisibleOn("${isDir}").
						Icon("fa fa-folder-open").
						Label("${i18n.podFile.open}").
						ActionType("ajax").
						Api(api.Files).
						Reload("files"),
				),
			).
			OnEvent(
				app.Event().RowClick(
					app.EventActions(
					// TODO
					),
				),
			),
	)
}
