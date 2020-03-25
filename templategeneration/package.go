package templategeneration

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/anz-bank/sysl/pkg/sequencediagram"

	"github.com/anz-bank/sysl/pkg/datamodeldiagram"

	"github.com/anz-bank/sysl/pkg/sysl"
	"github.com/anz-bank/sysl/pkg/syslutil"
)

const (
	md           = ".md"
	ext          = ".svg"
	pageFilename = "README.md"
)

// Package is the second level where apps and endpoints are specified.
type Package struct {
	Parent              *Project
	OutputDir           string
	PackageName         string
	OutputFile          string
	IntegrationDiagrams []*Diagram
	SequenceDiagrams    []*SequenceDiagram // map[appName + endpointName]
	DataModelDiagrams   []*Diagram
}

func (p Package) RegisterIntegrationDiagrams(m *sysl.Module) {

}

func (p Package) RegisterDataModelDiagrams(m *sysl.Module) {

}

// AlphabeticalRows returns an alphabetically sorted list of packages of any project.
func (p Project) AlphabeticalRows() []*Package {
	packages := make([]*Package, 0, len(p.Packages))
	for _, key := range alphabeticalPackage(p.Packages) {
		packages = append(packages, p.Packages[key])
	}
	return packages
}

// RegisterSequenceDiagrams creates sequence Diagrams from the sysl Module in Project.
func (p Project) RegisterSequenceDiagrams() error {
	for _, key := range alphabeticalApps(p.Module.Apps) {
		app := p.Module.Apps[key]
		packageName, appName := GetAppPackageName(app)
		if syslutil.HasPattern(app.Attrs, "ignore") {
			p.Log.Infof("Skipping application %s", app.Name)
			continue
		}
		for _, key2 := range alphabeticalEndpoints(app.Endpoints) {
			endpoint := app.Endpoints[key2]
			if syslutil.HasPattern(endpoint.Attrs, "ignore") {
				p.Log.Infof("Skipping application %s", app.Name)
				continue
			}
			packageD := p.Packages[packageName]
			diagram, err := packageD.SequenceDiagramFromEndpoint(appName, endpoint)
			if err != nil {
				return err
			}
			p.Packages[packageName].SequenceDiagrams = append(packageD.SequenceDiagrams, diagram)
			if p.Packages[packageName].DataModelDiagrams == nil {
				p.Packages[packageName].DataModelDiagrams = []*Diagram{}
			}
		}
	}
	return nil
}

func (p Project) GenerateEndpointDataModel(parentAppName string, t *sysl.Type) string {
	pl := &datamodelCmd{}
	name := "owiehfwoi"
	pl.Project = ""
	pl.Output = path.Join(p.Output, name)
	p.Fs.MkdirAll(pl.Output, os.ModePerm)
	pl.Output += "/" + name + "_datamodel.svg"
	pl.Direct = true
	pl.ClassFormat = "%(classname)"
	spclass := sequencediagram.ConstructFormatParser("", pl.ClassFormat)
	var stringBuilder strings.Builder
	dataParam := &DataModelParam{}
	dataParam.Mod = p.Module
	dataParam.Title = "datagenParams.Title"

	v := datamodeldiagram.MakeDataModelView(spclass, dataParam.Mod, &stringBuilder, dataParam.Title, "")
	vNew := &DataModelView{
		DataModelView: *v,
	}
	return vNew.GenerateDataView(dataParam, parentAppName, t, p)
}

// SequenceDiagramFromEndpoint generates a sequence diagram from a sysl endpoint
func (p Package) SequenceDiagramFromEndpoint(appName string, endpoint *sysl.Endpoint) (*SequenceDiagram, error) {
	call := fmt.Sprintf("%s <- %s", appName, endpoint.Name)
	seq, err := CreateSequenceDiagram(p.Parent.Module, call)
	if err != nil {
		return nil, err
	}
	diagram := &SequenceDiagram{}
	diagram.Parent = &p
	diagram.AppName = appName
	diagram.EndpointName = endpoint.Name
	diagram.OutputFileName__ = appName + endpoint.Name
	diagram.OutputDir = path.Join(p.Parent.Output, p.PackageName)
	diagram.DiagramString = seq
	diagram.Diagramtype = diagram_sequence
	diagram.OutputMarkdownFileName = pageFilename
	diagram.OutputDataModel = []*Diagram{}
	diagram.InputDataModel = []*Diagram{}
	for _, param := range endpoint.Param {
		newDiagram := &Diagram{
			Parent:           &p,
			OutputDir:        path.Join(p.Parent.Output, p.PackageName),
			AppName:          appName,
			DiagramString:    p.Parent.GenerateEndpointDataModel(appName, param.Type),
			OutputFileName__: appName + endpoint.Name + "data-model",
			EndpointName:     endpoint.Name,
		}
		diagram.InputDataModel = append(diagram.InputDataModel, newDiagram)
	}
	return diagram, nil
}
