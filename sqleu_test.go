package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestFileTmpl(t *testing.T) {

	fs := FieldStruct{}
	fs.FieldName = "Uid"
	fs.FieldType = "uint64"
	fs.FieldComment = "类型ID"
	fs2 := FieldStruct{}
	fs2.FieldName = "U_Name"
	fs2.FieldType = "string"
	fs2.FieldComment = "名称"

	ts := TableStruct{}
	ts.TableComment = "表的注释\n有换行"
	ts.TableName = "TestTable"
	ts.TableRefernces = []string{"TestRef", "TestRef2"}
	ts.Fields = []FieldStruct{fs, fs2}
	ts.GetSet = true

	ts2 := TableStruct{}
	ts2.TableComment = "表的注释"
	ts2.TableName = "TestTable2"
	ts2.TableRefernces = []string{"TestRef", "TestRef2"}
	ts2.Fields = []FieldStruct{fs, fs2}
	ts2.GetSet = true

	fos := FileOutStruct{}
	fos.CreateTime = time.Now()
	fos.DBName = "TestDB"
	fos.Imports = []string{"time"}
	fos.PackageName = "main"
	fos.Entitys = []TableStruct{ts, ts2}

	err := fileTmpl.Execute(os.Stdout, fos)
	if nil != err {
		fmt.Println(err)
	}
}

func TestSqlConn(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@/testsqleu")
	if nil != err {
		fmt.Println("mysql database connection error:", err)
		return
	}
	defer db.Close()

	fileOut, err := getFileOut(db)
	if nil != err {
		fmt.Println(err)
	} else {
		err := fileTmpl.Execute(os.Stdout, fileOut)
		if nil != err {
			fmt.Println(err)
		}
	}
}
