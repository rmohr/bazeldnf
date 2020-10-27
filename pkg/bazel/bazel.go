package bazel

import (
	"fmt"
	"io/ioutil"

	"github.com/bazelbuild/buildtools/build"
	"github.com/bazelbuild/buildtools/edit"
	"github.com/rmohr/bazeldnf/pkg/api"
)

type Artifact struct {
	rule *build.Rule
}

func LoadWorkspace(path string) (*build.File, error) {
	workspaceData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WORSPACE orig: %v", err)
	}
	workspace, err := build.ParseWorkspace(path, workspaceData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WORSPACE orig: %v", err)
	}
	return workspace, nil
}

func WriteWorkspace(dryRun bool, workspace *build.File, path string) error {
	if dryRun {
		fmt.Println(build.FormatString(workspace))
		return nil
	}
	return ioutil.WriteFile(path, build.Format(workspace), 0666)
}

func PruneRPMs(workspace *build.File) () {
	workspace.DelRules("rpm", "")
}

func AddRPMS(workspace *build.File, pkgs []*api.Package) {
	for _, pkg := range pkgs {
		call := &build.CallExpr{X: &build.Ident{Name: "rpm"}}
		rule := &build.Rule{call, ""}
		rule.SetAttr("name", &build.StringExpr{Value: pkg.String()})
		workspace.Stmt = edit.InsertAfterLastOfSameKind(workspace.Stmt, rule.Call)
	}
}