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
						VisibleOn("${type!=='link'}").
						Icon("fa fa-download").
						Label("${i18n.podFile.download}").
						ActionType("download").
						Api("post:"+api.Download+"?file=${name}&type=${type}"),
					app.Button().
						VisibleOn("${type==='dir'}").
						Icon("fa fa-folder-open").
						Label("${i18n.podFile.open}").
						ActionType("ajax").
						Api("post:"+api.Files+"?dir=${name}").
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
