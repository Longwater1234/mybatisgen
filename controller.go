package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

// Represents table Entity
type javaClass struct {
	PkCamelName    string       // name of PK in camel Case
	PkPascalName   string       //name of PK in Pascal Case
	PkType         string       // java Type of primary Key
	PascalName     string       // current table name
	CamelCase      string       // current table name
	TargetPackage  string       // e.g. com.example.projectName
	FkRelationList []FkRelation // foreign key relations
}

// Mapping of tableType to TemplateFile
const (
	NormalTable = "DemoController.txt"
	ViewTable   = "ViewDemoController.txt"
)

// for testing table type
var viewReg = regexp.MustCompile(`^View[A-Z]+?`)

const filePattern = "%sController.java"

// GenerateControllers for all tables
func GenerateControllers(baseDir, targetPackage string) {
	controllerDir := filepath.Join(baseDir, "controller")
	os.Mkdir(controllerDir, os.ModePerm)
	modelFiles, err := os.ReadDir(filepath.Join(baseDir, "model"))
	check(err)
	for _, modelFile := range modelFiles {
		writeControllerFile(modelFile, controllerDir, targetPackage)
	}
}

// write content and save to file
func writeControllerFile(modelFile os.DirEntry, controllerDir, targetPackage string) {
	var fullName = modelFile.Name()
	cleanName := fullName[0:strings.Index(fullName, ".")] //without extension
	var fkList, exists = fkRelationMap[cleanName]

	if exists {
		//for each foreign column, attach its Primary key name
		for i, relation := range fkList {
			foreignPkItem := pkTableMap[relation.RefTablePascal]
			relation.RefPkPascal = strcase.ToCamel(foreignPkItem.PkName)
			fkList[i] = relation
		}
	}

	var pkItem = pkTableMap[cleanName]
	singleClass := &javaClass{
		PkType:         pkItem.PkType,
		PkCamelName:    pkItem.PkName,
		PkPascalName:   strcase.ToCamel(pkItem.PkName),
		PascalName:     cleanName,
		CamelCase:      strcase.ToLowerCamel(cleanName),
		TargetPackage:  targetPackage,
		FkRelationList: fkList,
	}

	var tmplFileName = NormalTable
	if viewReg.MatchString(cleanName) {
		tmplFileName = ViewTable
	}

	templatePath := filepath.Join(TemplateDir, tmplFileName)
	tc := template.Must(template.New(tmplFileName).ParseFiles(templatePath))
	fout, err := os.Create(filepath.Join(controllerDir, fmt.Sprintf(filePattern, cleanName)))
	check(err)
	defer fout.Close()

	bw := bufio.NewWriter(fout)
	defer bw.Flush()
	tc.Execute(bw, singleClass)
}

// GenerateCommonFiles for paging and generic Response
func GenerateCommonFiles(baseDir string, targetPackage string) {
	commonDir := filepath.Join(baseDir, "common")
	os.Mkdir(commonDir, os.ModePerm)

	//write CommonResponse file
	tmplResponse := filepath.Join(TemplateDir, "CustomResponse.txt")
	tcc := template.Must(template.New("CustomResponse.txt").ParseFiles(tmplResponse))
	outResFile, err := os.Create(filepath.Join(commonDir, "CustomResponse.java"))
	check(err)
	defer outResFile.Close()
	tcc.Execute(outResFile, targetPackage)

	//write PageInfo file
	tmplPageInfo := filepath.Join(TemplateDir, "PageInfo.txt")
	tpp := template.Must(template.New("PageInfo.txt").ParseFiles(tmplPageInfo))
	outPageInfo, err := os.Create(filepath.Join(commonDir, "PageInfo.java"))
	check(err)
	defer outPageInfo.Close()
	tpp.Execute(outPageInfo, targetPackage)
}
