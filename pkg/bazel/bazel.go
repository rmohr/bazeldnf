package bazel

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

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

func LoadBuild(path string) (*build.File, error) {
	buildfileData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WORSPACE orig: %v", err)
	}
	buildfile, err := build.ParseBuild(path, buildfileData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WORSPACE orig: %v", err)
	}
	return buildfile, nil
}

func WriteBuild(dryRun bool, buildfile *build.File, path string) error {
	if dryRun {
		fmt.Println(build.FormatString(buildfile))
		return nil
	}
	return ioutil.WriteFile(path, build.Format(buildfile), 0666)
}

func WriteWorkspace(dryRun bool, workspace *build.File, path string) error {
	if dryRun {
		fmt.Println(build.FormatString(workspace))
		return nil
	}
	return ioutil.WriteFile(path, build.Format(workspace), 0666)
}

func AddRPMs(workspace *build.File, pkgs []*api.Package) {

	rpms := map[string]*rpmRule{}

	for _, rule := range workspace.Rules("rpm") {
		rpms[rule.Name()] = &rpmRule{rule}
	}

	for _, pkg := range pkgs {
		pkgName := sanitize(pkg.String())
		rule := rpms[pkgName]
		if rule == nil {
			call := &build.CallExpr{X: &build.Ident{Name: "rpm"}}
			rule = &rpmRule{&build.Rule{call, ""}}
			rpms[pkgName] = rule
		}
		rule.SetName(pkgName)
		rule.SetSHA256(pkg.Checksum.Text)
		urls := rule.URLs()
		if len(urls) == 0 {
			rule.SetURLs(pkg.Repository.Mirrors, pkg.Location.Href)
		}
	}

	rules := []*rpmRule{}
	for _, rule := range rpms {
		rules = append(rules, rule)
	}

	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Name() < rules[j].Name()
	})

	workspace.DelRules("rpm", "")
	for _, rule := range rules {
		workspace.Stmt = edit.InsertAtEnd(workspace.Stmt, rule.Call)
	}
}

func AddTree(name string, buildfile *build.File, pkgs []*api.Package, files []string) {
	rpmtrees := map[string]*rpmTree{}

	for _, rule := range buildfile.Rules("rpmtree") {
		rpmtrees[rule.Name()] = &rpmTree{rule}
	}
	buildfile.DelRules("rpmtree", "")

	rpms := []string{}
	for _, pkg := range pkgs {
		pkgName := sanitize(pkg.String())
		rpms = append(rpms, "@"+pkgName+"//rpm")
	}
	sort.SliceStable(rpms, func(i, j int) bool {
		return rpms[i] < rpms[j]
	})

	sort.SliceStable(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	fileMap := map[string][]string{}
	for _, file := range files {
		fileMap[filepath.Dir(file)] = append(fileMap[filepath.Dir(file)], filepath.Base(file))
	}

	dirs := []string{}
	for dir, _ := range fileMap {
		dirs = append(dirs, dir)
	}
	sort.SliceStable(dirs, func(i, j int) bool {
		return dirs[i] < dirs[j]
	})

	rule := rpmtrees[name]
	if rule == nil {
		call := &build.CallExpr{X: &build.Ident{Name: "rpmtree"}}
		rule = &rpmTree{&build.Rule{call, ""}}
		rpmtrees[name] = rule
	}
	rule.SetName(name)
	rule.SetRPMs(rpms)
	rule.SetFiles(dirs, fileMap)

	rules := []*rpmTree{}
	for _, rule := range rpmtrees {
		rules = append(rules, rule)
	}

	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Name() < rules[j].Name()
	})

	for _, rule := range rules {
		buildfile.Stmt = edit.InsertAtEnd(buildfile.Stmt, rule.Call)
	}
}

func PruneRPMs(buildfile *build.File, workspace *build.File) {
	referenced := map[string]struct{}{}
	for _, pkg := range buildfile.Rules("rpmtree") {
		tree := &rpmTree{pkg}
		for _, rpm := range tree.RPMs() {
			referenced[rpm] = struct{}{}
		}
	}
	rpms := workspace.Rules("rpm")
	for _, rpm := range rpms {
		if _, exists := referenced["@"+rpm.Name()+"//rpm"]; !exists {
			workspace.DelRules("rpm", rpm.Name())
		}
	}
}

type rpmRule struct {
	*build.Rule
}

func (r *rpmRule) URLs() []string {
	if urlsAttr := r.Rule.Attr("urls"); urlsAttr != nil {
		if len(urlsAttr.(*build.ListExpr).List) > 0 {
			urls := []string{}
			for _, expr := range urlsAttr.(*build.ListExpr).List {
				urls = append(urls, expr.(*build.StringExpr).Value)
			}
			return urls
		}
	}
	return nil
}

func (r *rpmRule) SetURLs(urls []string, href string) {
	urlsAttr := []build.Expr{}
	for _, url := range urls {
		u := strings.TrimSuffix(url, "/") + "/" + strings.TrimSuffix(href, "/")
		urlsAttr = append(urlsAttr, &build.StringExpr{Value: u})
	}
	r.Rule.SetAttr("urls", &build.ListExpr{List: urlsAttr})
}

func (r *rpmRule) SetName(name string) {
	r.Rule.SetAttr("name", &build.StringExpr{Value: name})
}

func (r *rpmRule) SetSHA256(sha256 string) {
	r.Rule.SetAttr("sha256", &build.StringExpr{Value: sha256})
}

type rpmTree struct {
	*build.Rule
}

func (r *rpmTree) SetName(name string) {
	r.Rule.SetAttr("name", &build.StringExpr{Value: name})
}

func (r *rpmTree) RPMs() []string {
	if rpmAttrs := r.Rule.Attr("rpms"); rpmAttrs != nil {
		if len(rpmAttrs.(*build.ListExpr).List) > 0 {
			rpms := []string{}
			for _, expr := range rpmAttrs.(*build.ListExpr).List {
				rpms = append(rpms, expr.(*build.StringExpr).Value)
			}
			return rpms
		}
	}
	return nil
}

func (r *rpmTree) SetRPMs(rpms []string) {
	rpmsAttr := []build.Expr{}
	for _, rpm := range rpms {
		rpmsAttr = append(rpmsAttr, &build.StringExpr{Value: rpm})
	}
	r.Rule.SetAttr("rpms", &build.ListExpr{List: rpmsAttr})
}

func (r *rpmTree) SetFiles(dirs []string, fileMap map[string][]string) {
	filesMapExpr := &build.DictExpr{}
	for _, dir := range dirs {
		filesListExpr := &build.ListExpr{}
		for _, file := range fileMap[dir] {
			filesListExpr.List = append(filesListExpr.List, &build.StringExpr{Value: file})
		}
		filesMapExpr.List = append(filesMapExpr.List, &build.KeyValueExpr{Key: &build.StringExpr{Value: dir}, Value: filesListExpr})
	}
	r.Rule.SetAttr("files", filesMapExpr)
}

func sanitize(name string) string {
	name = strings.ReplaceAll(name, ":", "__")
	name = strings.ReplaceAll(name, "+", "__plus__")
	return name
}
