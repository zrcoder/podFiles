package main

import (
	"fmt"

	"github.com/zrcoder/podFiles/conf"
	"github.com/zrcoder/podFiles/internal/api"
	"github.com/zrcoder/podFiles/internal/auth"
	"github.com/zrcoder/podFiles/internal/ui"

	"github.com/zrcoder/amisgo"
)

func main() {
	fmt.Println("Starting...")
	app := amisgo.New(conf.Options()...)
	app.Mount("/", ui.Index(app))
	app.Mount(ui.FilesPage, ui.FileList(app), auth.K8s)
	app.Handle(api.Prefix, api.New())
	fmt.Println("Serving on http://localhost:8080")
	app.Run("0.0.0.0:8080")
}
