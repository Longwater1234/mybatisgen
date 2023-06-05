package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
)

// FkRelation is for foreign columns in given table
type FkRelation struct {
	RefTableCamel  string // ref table camelCase (parent)
	RefTablePascal string // ref table PascalCase (parent)
	RefPkPascal    string // PK of ref table  (parent)
	PascalName     string // current table PascalCase
	CamelCase      string // current table camelCase
	PkCamelCase    string // PK of current table
}

// TablePrimaryKey name and type
type TablePrimaryKey struct {
	PkType string
	PkName string
}

// map of TableName to its ForeignRelations
var fkRelationMap = make(map[string][]FkRelation)

// map of TableName to its PrimaryKey struct
var pkTableMap = make(map[string]TablePrimaryKey)

// common NySQL to JAVA typings for PK
var sqlJavaTypes = map[string]string{
	"varchar": "String",
	"bigint":  "Long",
	"int":     "Integer",
	"binary":  "String",
	"char":    "String",
}

// will match the foreign columns
var foreignRegex = regexp.MustCompile("FOREIGN KEY \\(`\\w+`\\) REFERENCES `(\\w+)`")

// will match the primary column
var primaryNameRegex = regexp.MustCompile("PRIMARY KEY \\(`(\\w+)`\\),?")

// will match dataType of PK
var primaryPattern = "`%s` (\\w+)\\("

// GetTableNames from given database
func GetTableNames(dbConn *sql.DB) ([]string, error) {
	rows, err := dbConn.QueryContext(context.Background(), "SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tableList []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tableList = append(tableList, name)
	}
	return tableList, nil
}

// GetForeignRelations within given table
func GetForeignRelations(dbConn *sql.DB, tableName string, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	if strings.HasPrefix(tableName, "view_") {
		return
	}
	var fkRelations []FkRelation
	createStmt, err := showCreateStmt(dbConn, tableName)
	check(err)
	camelPkName := getTablePrimaryKey(tableName, createStmt, mu)
	br := bufio.NewScanner(strings.NewReader(createStmt))
	for br.Scan() {
		matchArr := foreignRegex.FindStringSubmatch(br.Text())
		if len(matchArr) > 0 {
			relation := &FkRelation{
				RefTableCamel:  strcase.ToLowerCamel(matchArr[1]),
				RefTablePascal: strcase.ToCamel(matchArr[1]), //actually, it's Pascal Case
				PascalName:     strcase.ToCamel(tableName),
				CamelCase:      strcase.ToLowerCamel(tableName),
				PkCamelCase:    camelPkName,
				RefPkPascal:    "",
			}
			fkRelations = append(fkRelations, *relation)
		}
	}
	pascalTableName := strcase.ToCamel(tableName)
	mu.Lock()
	fkRelationMap[pascalTableName] = fkRelations
	mu.Unlock()
}

// Get table's primaryKey name and type, returns camelCase PK name
func getTablePrimaryKey(tableName, createStmt string, mu *sync.Mutex) string {
	pkItem := new(TablePrimaryKey)
	br := bufio.NewScanner(strings.NewReader(createStmt))
	for br.Scan() {
		matchArr := primaryNameRegex.FindStringSubmatch(br.Text())
		if len(matchArr) > 0 {
			regPKType := regexp.MustCompile(fmt.Sprintf(primaryPattern, matchArr[1]))
			ioutil.WriteFile("haha/"+tableName+".sql", []byte(createStmt), 0644)
			typeResults := regPKType.FindStringSubmatch(createStmt)
			if len(typeResults) > 0 {
				var sqlType = typeResults[1]
				pkItem.PkType = sqlJavaTypes[sqlType]
				pkItem.PkName = strcase.ToLowerCamel(matchArr[1])
			}
		}
	}
	pascalTableName := strcase.ToCamel(tableName) //actually, it's Pascal Case
	mu.Lock()
	pkTableMap[pascalTableName] = *pkItem
	mu.Unlock()
	return pkItem.PkName
}

// Get DDL of the given table
func showCreateStmt(dbConn *sql.DB, tableName string) (string, error) {
	row := dbConn.QueryRowContext(context.Background(), fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName))
	var c1, c2 string
	err := row.Scan(&c1, &c2)
	return c2, err
}

// GetDbVersion helps decide which driver to use
func GetDbVersion(db *sql.DB) (string, error) {
	var version string
	err := db.QueryRowContext(context.Background(), "SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", err
	}

	return version, nil
}
