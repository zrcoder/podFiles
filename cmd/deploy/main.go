package main

import (
	_ "embed"
	"os"
	"text/template"
)

//go:embed tpl/apply_sh.tpl
var scriptTemplate string

//go:embed tpl/cluster-resources.yaml
var clusterResources string

//go:embed tpl/namespace-resources.yaml
var namespaceResources string

//go:embed tpl/ingress.yaml
var ingress string

type ScriptData struct {
	ClusterResources   string
	NamespaceResources string
	Ingress            string
}

func main() {
	tmpl, err := template.New("apply").Parse(scriptTemplate)
	checkError(err)

	data := ScriptData{
		ClusterResources:   clusterResources,
		NamespaceResources: namespaceResources,
		Ingress:            ingress,
	}

	f, err := os.Create("cmd/deploy/apply.sh")
	checkError(err)
	defer f.Close()

	err = tmpl.Execute(f, data)
	checkError(err)

	err = os.Chmod("cmd/deploy/apply.sh", 0o755)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
