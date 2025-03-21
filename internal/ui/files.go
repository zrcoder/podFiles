package ui

import (
	"github.com/zrcoder/amisgo"
	"github.com/zrcoder/amisgo/comp"
	"github.com/zrcoder/podFiles/internal/api"
)

func FileList(app *amisgo.App) comp.Page {
	return page(
		app,
		app.Service().Name("files").Api(api.Files).Body(
			app.Flex().ClassName("h-10").AlignItems("center").Justify("flex-start").Items(
				app.Button().Icon("fa fa-home").Label("${i18n.home}").
					ActionType("link").Link("/").Reload("/"),
				app.Wrapper(),
				app.Breadcrumb().Source("${breadItems}"),
				app.Wrapper(),
				app.Button().Icon("fa fa-folder-open").Label("..").VisibleOn("${inSubDir}").
					ActionType("ajax").Api("post:"+api.Files+"?back=true").Reload("files"),
				app.Wrapper(),
				app.Button().Icon("fa fa-upload").Label("${i18n.podFile.upload}").
					ActionType("drawer").Drawer(
					app.Drawer().Name("upload").Position("bottom").
						Actions().
						Body(
							app.InputFile().Receiver(api.Upload).Drag(true).UseChunk(false),
							app.Flex().Justify("center").Items(
								app.Button().Label("${i18n.podFile.done}").ActionType("reload").Target("files").Close("upload"),
							),
						),
				),
			),

			crud(app).ClassName("mt-2").Source("${files}").
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
				),
		),
	)
}
