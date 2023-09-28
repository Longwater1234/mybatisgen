package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"golang.org/x/term"
)

// ProjectConfig contains connection and java packageName
type ProjectConfig struct {
	TargetPackage string `json:"targetPackage"` //Java style name e.g. com.example.project
	DbCredentials `json:"connection"`
}

// Java database drivers
const (
	MYSQL8 = "mysql-connector-j-8.0.32.jar"
	MYSQL5 = "mysql-connector-java-5.1.49.jar"
)

// last modified date
const versionDate = "2023-09-23"

// DbCredentials for target db
type DbCredentials struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"-"`
	Database string `json:"database"`
	Port     int    `json:"port"`
}

// MyBatisConfig for the generator
type MyBatisConfig struct {
	DbCredentials
	DriverName  string
	PackageName string
	OutFolder   string
	TableList   []string
}

// project file paths
const (
	dbConfigPath  = "./env.json"
	folderOutput  = "./output/"
	assetsDir     = "./assets/"
	TemplateDir   = "./mytemplates"
	genConfigFile = assetsDir + "generatorConfig.xml"
)

// global error handler
func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

var tt *template.Template

func init() {
	templatePath := filepath.Join(TemplateDir, "genConfig.gohtml")
	tt = template.Must(template.New("genConfig.gohtml").ParseFiles(templatePath))
	os.RemoveAll(folderOutput)
	os.MkdirAll(folderOutput, os.ModePerm)
}

func main() {
	f, err := os.Open(dbConfigPath)
	check(err)
	defer f.Close()

	jd := json.NewDecoder(f)
	var config ProjectConfig
	err = jd.Decode(&config)
	check(err)

	fmt.Println("\n+=+=+=\tMyBATIS SPRINGBOOT CODE GENERATOR\t+=+=+=")
	fmt.Printf("\t\t\t(version %s)\n\n", versionDate)
	fmt.Print("Please enter database password (hidden): ")
	dbPassBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	check(err)
	fmt.Println()

	config.Password = string(dbPassBytes)
	dbConn, err := initDbConn(&config.DbCredentials)
	check(err)

	dbVersion, err := GetDbVersion(dbConn)
	check(err)
	log.Println("Detected MySQL version", dbVersion)

	var driverName = MYSQL5
	if strings.HasPrefix(dbVersion, "8.0") {
		driverName = MYSQL8
	}

	tableNames, err := GetTableNames(dbConn)
	check(err)
	log.Println("Done reading all table names from Db")

	absOutputDir, err := filepath.Abs(folderOutput)
	check(err)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, tbl := range tableNames {
		wg.Add(1)
		go GetForeignRelations(dbConn, tbl, &wg, &mu)
	}
	wg.Wait()
	log.Println("Done getting foreignKey relations")

	genConfig := &MyBatisConfig{
		DbCredentials: config.DbCredentials,
		DriverName:    driverName,
		PackageName:   config.TargetPackage,
		OutFolder:     absOutputDir,
		TableList:     tableNames,
	}

	fout, err := os.Create(genConfigFile)
	check(err)

	defer func() {
		fout.Close()
		os.Remove(filepath.Join(genConfigFile)) //remove sensitive File
		log.Println("Removed generatorConfig.xml file")
	}()

	bw := bufio.NewWriter(fout)
	check(tt.Execute(bw, genConfig))
	bw.Flush()
	log.Println("Done creating generatorConfig.xml file")

	time.Sleep(time.Second)
	log.Println("Starting MyBatisGenerator")

	wkDir, _ := filepath.Abs(assetsDir)
	runCmd := filepath.Join(wkDir, "run.sh")
	os.Chmod(runCmd, 0755)

	if runtime.GOOS == "windows" {
		runCmd = filepath.Join(wkDir, "run.cmd")
	}
	cmd := &exec.Cmd{
		Path:   runCmd,
		Dir:    wkDir,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println("Generating. Please wait...")
	cmd.Start()
	check(cmd.Wait())

	log.Println("Done. Now Writing controllers.")

	pkgDir := filepath.Clean(strings.ReplaceAll(config.TargetPackage, ".", "/"))
	innerBaseDir := filepath.Join(absOutputDir, pkgDir)
	GenerateControllers(innerBaseDir, config.TargetPackage)
	GenerateCommonFiles(innerBaseDir, config.TargetPackage)

	log.Println("All is OK! OutputDir:", absOutputDir)
	fmt.Println()
}

// initialize db Connection, with given credentials
func initDbConn(myConn *DbCredentials) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		myConn.User, myConn.Password, myConn.Host, myConn.Port, myConn.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	return db, db.Ping()
}
