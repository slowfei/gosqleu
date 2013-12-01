//	Copyright 2013 slowfei And The Contributors All rights reserved.
//
//	Software Source Code License Agreement (BSD License)
//
//  Create on 2013-11-30
//  Update on 2013-12-01
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	go sql entity utils
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
)

type FieldStruct struct {
	FieldName    string
	FieldType    string
	FieldComment string
}

type TableStruct struct {
	TableName      string
	TableComment   string
	Fields         []FieldStruct
	TableRefernces []string
	GetSet         bool
}

type FileOutStruct struct {
	CreateTime  time.Time
	DBName      string
	PackageName string
	Imports     []string
	Entitys     []TableStruct
}

const (
	APP_NAME = "gosqleu"
	VERSION  = "1.0"
)

var (
	importTime  = false
	connection  = flag.String("conn", "", "database connection information [user:password@tcp(localhost:3306)/dbname]")
	isGetSet    = flag.Bool("getset", true, "is struct field out get set func")
	packageName = flag.String("package", "main", "package name default main")
	filterTable = flag.String("filter-table", "", "filter table name [table1,table]")
	outFilePath = flag.String("path", "", "out file path")
	connRex     = regexp.MustCompile(`\w+(:\w)?@((tcp(\(.+:\d{1,5}\))?)?|(unix\(\/(.*)([^/]*)\w+\)))/\w+((\\?|&+)(.+?)=([^&]*))?`)

	fileTmpl, _ = template.New("entityfile").Funcs(template.FuncMap{"FuncName": FuncName, "ParamLower": ParamLower, "MysqltypeToGotype": MysqltypeToGotype}).Parse(fileLayout)
	fileLayout  = `//
//	create time {{.CreateTime}}
//	` + APP_NAME + ` version ` + VERSION + `

//	database {{.DBName}}
package {{.PackageName}}
{{range .Imports}}
import "{{.}}"
{{end}}{{with .Entitys}}{{range .}}
/*====================================*
 	{{.TableName}} table
 
 	{{.TableComment}}
 *====================================*/
type {{FuncName .TableName}} struct{
	{{with .Fields}}{{range .}}
	{{.FieldName}}	{{MysqltypeToGotype .FieldType}}	// {{.FieldComment}}{{end}}{{end}}
	{{range .TableRefernces}}
	*{{.}}{{end}}
}
{{$TempTableName := .TableName}}{{if .GetSet}}{{with .Fields}}{{range .}}
/* get set {{.FieldComment}} */
func (this *{{$TempTableName}}) Get{{FuncName .FieldName}}() {{MysqltypeToGotype .FieldType}}{
	return this.{{.FieldName}}
}
func (this *{{$TempTableName}}) Set{{FuncName .FieldName}}({{ParamLower .FieldName}} {{MysqltypeToGotype .FieldType}}){
	this.{{.FieldName}} = {{ParamLower .FieldName}}
}
{{end}}{{end}}{{end}}{{end}}

{{end}}
`
)

func FuncName(name string) string {
	spl := strings.Split(name, "_")
	joinName := ""

	for _, s := range spl {
		joinName += strings.Title(s)
	}

	return joinName
}

func ParamLower(name string) string {
	name = strings.ToLower(name)
	name = strings.Replace(name, "_", "", -1)
	return name
}

func MysqltypeToGotype(typeName string) string {
	typeName = strings.ToLower(typeName)
	switch {
	case 0 <= strings.Index(typeName, "bit"):
		return "bool"
	case 0 <= strings.Index(typeName, "int"):
		result := ""
		if 0 <= strings.Index(typeName, "unsigned") {
			result = "u"
		}
		switch {
		case 0 <= strings.Index(typeName, "tinyint"):
			result += "int8"
		case 0 <= strings.Index(typeName, "smallint"):
			result += "int16"
		case 0 <= strings.Index(typeName, "mediumint"):
			result += "int32"
		case 0 <= strings.Index(typeName, "bigint"):
			result += "int64"
		default:
			result += "int"
		}
		return result
	case 0 <= strings.Index(typeName, "real"):
		return "float32"
	case 0 <= strings.Index(typeName, "float"):
		return "float32"
	case 0 <= strings.Index(typeName, "double"):
		return "float64"
	case 0 <= strings.Index(typeName, "decimal"):
		return "float64"
	case 0 <= strings.Index(typeName, "numeric"):
		return "float64"
	case 0 <= strings.Index(typeName, "date"):
		return "time.Time"
	case 0 <= strings.Index(typeName, "time"):
		return "time.Time"
	case 0 <= strings.Index(typeName, "year"):
		return "time.Time"
	case 0 <= strings.Index(typeName, "blob"):
		return "[]byte"
	default:
		return "string"
	}

	return "string"
}

func getFieldInfos(sqldb *sql.DB, tableName string) []FieldStruct {
	rows, err := sqldb.Query(`
select COLUMN_NAME,COLUMN_TYPE,COLUMN_COMMENT from information_schema.COLUMNS 
	where TABLE_SCHEMA = database() and TABLE_NAME = ?
`, tableName)
	if nil != err {
		fmt.Println("Query Tables Error:", err)
		return nil
	}
	defer rows.Close()

	fields := make([]FieldStruct, 0)
	for rows.Next() {
		fs := FieldStruct{}
		rows.Scan(&fs.FieldName, &fs.FieldType, &fs.FieldComment)
		if 0 <= strings.Index(strings.ToLower(fs.FieldType), "time") || 0 <= strings.Index(strings.ToLower(fs.FieldType), "date") {
			importTime = true
		}
		fields = append(fields, fs)
	}

	return fields
}

func getTables(sqldb *sql.DB) []TableStruct {

	rows, err := sqldb.Query(`
select TABLE_NAME as name,TABLE_COMMENT as comment, 
(select GROUP_CONCAT(REFERENCED_TABLE_NAME SEPARATOR ',') from information_schema.REFERENTIAL_CONSTRAINTS where TABLE_NAME=name) as refnames 
from information_schema.TABLES where TABLE_SCHEMA = database();
`)
	if nil != err {
		fmt.Println("Query Tables Error:", err)
		return nil
	}
	defer rows.Close()

	tables := make([]TableStruct, 0)

	for rows.Next() {
		ts := TableStruct{}
		ts.GetSet = *isGetSet
		refstr := ""
		rows.Scan(&ts.TableName, &ts.TableComment, &refstr)

		isFilter := false
		for _, filterName := range strings.Split(*filterTable, ",") {
			if strings.ToLower(filterName) == strings.ToLower(ts.TableName) {
				isFilter = true
				break
			}
		}

		if isFilter {
			continue
		}

		if 0 != len(refstr) {
			ts.TableRefernces = strings.Split(refstr, ",")
		}
		fmt.Printf("Table (%v) ...\n", ts.TableName)
		ts.Fields = getFieldInfos(sqldb, ts.TableName)
		tables = append(tables, ts)
	}

	return tables
}

func getFileOut(sqldb *sql.DB) (FileOutStruct, error) {
	fileOut := FileOutStruct{}

	var dbName string
	e := sqldb.QueryRow("SELECT database()").Scan(&dbName)
	switch {
	case e == sql.ErrNoRows:
		return fileOut, errors.New("No database with name.")
	case e != nil:
		return fileOut, e
	}

	if 0 == len(dbName) {
		return fileOut, errors.New("Get not database name.")
	}

	fileOut.Entitys = getTables(sqldb)
	fileOut.CreateTime = time.Now()
	if importTime {
		fileOut.Imports = []string{"time"}
	}
	fileOut.PackageName = *packageName
	fileOut.DBName = dbName

	return fileOut, nil
}

func main() {
	flag.Parse()

	conn := ""
	if 0 != len(flag.Args()) {
		conn = flag.Arg(0)
	}

	if 0 == len(conn) || !connRex.MatchString(conn) {
		if !connRex.MatchString(*connection) {
			fmt.Println("error: connection = $gosqleu [user:password@tcp(localhost:3306)/dbname]\nor other params:")
			flag.PrintDefaults()
			return
		} else {
			conn = *connection
		}
	}

	db, err := sql.Open("mysql", conn)
	if nil != err {
		fmt.Println("mysql database connection error:", err)
		return
	}
	defer db.Close()

	fileOut, e := getFileOut(db)
	if nil != e {
		fmt.Println(e)
		return
	}

	filePath := *outFilePath
	if 0 == len(filePath) {
		filePath = fileOut.DBName + ".go"
	}
	if '/' == filePath[len(filePath)-1] {
		filePath += fileOut.DBName + ".go"
	}
	fmt.Println("out file...")
	newFile, errFile := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if nil != errFile {
		fmt.Println(errFile)
		return
	}

	tplErr := fileTmpl.Execute(newFile, fileOut)
	if nil != tplErr {
		fmt.Println(tplErr)
		return
	}
	fmt.Println("success.")
	fmt.Println("path: ", newFile.Name())

}
