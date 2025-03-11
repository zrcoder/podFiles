package main

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"
)

//go:embed templates/cluster-resources.yaml
var clusterResources string

//go:embed templates/namespace-resources.yaml
var namespaceResources string

//go:embed templates/install_sh.tpl
var scriptTemplate string

func main() {
	tmpl, err := template.New("script").Parse(scriptTemplate)
	checkError(err)

	data := struct {
		ClusterResources   string
		NamespaceResources string
	}{
		ClusterResources:   clusterResources,
		NamespaceResources: namespaceResources,
	}

	f, err := os.Create("cmd/deploy/install.sh")
	checkError(err)
	defer f.Close()

	checkError(f.Chmod(0o755))

	checkError(tmpl.Execute(f, data))

	fmt.Println("Successfully generated install.sh")
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
