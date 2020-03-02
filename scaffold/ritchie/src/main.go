package main

import (
	"encoding/json"
	"github.com/fatih/color"
	"github.com/thoas/go-funk"
	"os"
	"ritchie/pkg/file/fileutil"
	"ritchie/pkg/ritchie/pathutil"
	"ritchie/pkg/ritchie/tree"
	"strings"
)

func main() {

	name := os.Getenv("FORMULA_NAME")
	description := os.Getenv("FORMULA_DESCRIPTION")
	mainPaths := pathutil.BuildMainPaths()
	if !pathutil.RightDir(mainPaths) {
		return
	}

	nameList := splitName(name)
	generateFiles(nameList, mainPaths, 0)
	changeMakeFile()

	treeFile, err := fileutil.ReadFile(mainPaths.TreeFile)
	verifyError(err)
	var jsonTree tree.Tree
	verifyError(json.Unmarshal(treeFile, &jsonTree))
	jsonTree = changeTreeFile(nameList, mainPaths, 0, jsonTree)

	//rascunho

	//commands := funk.Filter(jsonTree.Commands, func(command tree.Command) bool {
	//	return command.Parent == "root_github"
	//})
	//
	//jsonTree.Commands = commands.([]tree.Command)

	jsonResult, _ := json.MarshalIndent(jsonTree, "", "  ")
	verifyError(fileutil.WriteFile("tree/tree2.json", jsonResult))
	//fim rascunho

	color.Green("Generate formula:" + name + " with description:" + description + " .")

}

func changeTreeFile(nameList []string, mainPaths pathutil.MainPaths, i int, treeJson tree.Tree) tree.Tree {
	var dir = "-1"
	command := funk.Filter(treeJson.Commands, func(command tree.Command) bool {
		return command.Usage == nameList[i]
	}).([]tree.Command)

	if len(command) == 0 {
		if i > 0 {
			dir = "root_" + strings.Join(nameList[0:i], "_")
		} else {
			dir = "root"
		}
	}
	if len(nameList)-1 == i {
		if dir != "-1" {
			path := strings.Join(nameList, "/")
			commands := append(treeJson.Commands, tree.Command{
				Usage: nameList[i],
				Help:  "",
				Formula: &tree.Formula{
					Path:    path,
					Bin:     nameList[i] + "-${so}",
					Config:  nil,
					RepoUrl: "http://ritchie-cli-bucket152849730126474.s3-website-sa-east-1.amazonaws.com/formulas/" + path,
				},
				Parent: dir,
			})
			treeJson.Commands = commands
			return treeJson
		}
	} else {
		if dir != "-1" {
			commands := append(treeJson.Commands, tree.Command{
				Usage:   nameList[i],
				Help:    os.Getenv("FORMULA_DESCRIPTION"),
				Formula: nil,
				Parent:  dir,
			})
			treeJson.Commands = commands
			return changeTreeFile(nameList, mainPaths, i+1, treeJson)
		}
	}
	return treeJson
}

func changeMakeFile() {
	//todo
}

func splitName(name string) []string {
	return funk.Filter(strings.Split(name, " "), func(input string) bool {
		return input != ""
	}).([]string)
}

func generateFiles(nameList []string, mainPaths pathutil.MainPaths, i int) {
	dir := strings.Join(nameList[0:i+1], "/")
	color.Green("create dir:" + dir)
	verifyError(fileutil.CreateIfNotExists(dir))
	if len(nameList)-1 == i {
		createConfigFile(dir, mainPaths)
		createSrcDir(dir, mainPaths, nameList[i])
	} else {
		generateFiles(nameList, mainPaths, i+1)
	}
}

func createSrcDir(dir string, mainPaths pathutil.MainPaths, name string) {
	srdDir := dir + "/src"
	verifyError(fileutil.CreateIfNotExists(srdDir))
	createMainFile(srdDir, mainPaths)
	createGoModFile(srdDir, mainPaths)
	createMakeFile(srdDir, mainPaths, name)
	verifyError(fileutil.CreateIfNotExists(srdDir + "/pkg/hello"))
	createHelloFile(srdDir, mainPaths)
}

func createMakeFile(dir string, mainPaths pathutil.MainPaths, name string) {
	templateFile, err := fileutil.ReadFile(mainPaths.RitchieScaffoldTemplate + "/template-Makefile")
	verifyError(err)
	templateFile = []byte(
		strings.ReplaceAll(
			string(templateFile),
			"{{name}}",
			name,
		),
	)
	verifyError(fileutil.WriteFile(dir+"/Makefile", templateFile))
}

func createGoModFile(dir string, mainPaths pathutil.MainPaths) {
	templateFile, err := fileutil.ReadFile(mainPaths.RitchieScaffoldTemplate + "/template-go.mod")
	verifyError(err)
	verifyError(fileutil.WriteFile(dir+"/go.mod", templateFile))
}

func createHelloFile(dir string, mainPaths pathutil.MainPaths) {
	templateFile, err := fileutil.ReadFile(mainPaths.RitchieScaffoldTemplate + "/template-hello.txt")
	verifyError(err)
	verifyError(fileutil.WriteFile(dir+"/pkg/hello/hello.go", templateFile))
}

func createMainFile(dir string, mainPaths pathutil.MainPaths) {
	templateFile, err := fileutil.ReadFile(mainPaths.RitchieScaffoldTemplate + "/template-main.txt")
	verifyError(err)
	verifyError(fileutil.WriteFile(dir+"/main.go", templateFile))
}

func createConfigFile(dir string, mainPaths pathutil.MainPaths) {
	templateFile, err := fileutil.ReadFile(mainPaths.RitchieScaffoldTemplate + "/template-config.json")
	verifyError(err)
	templateFile = []byte(
		strings.ReplaceAll(
			string(templateFile),
			"{{description}}",
			os.Getenv("FORMULA_DESCRIPTION"),
		),
	)
	verifyError(fileutil.WriteFile(dir+"/config.json", templateFile))
}

func verifyError(err error) {
	if err != nil {
		panic(err)
	}
}
