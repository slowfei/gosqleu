go sql entity utils
=======
sql table out struct go file
```
/*=====================*
 	表名注释
 *=====================*/
type TableName struct {
	Id   uint64
	Name string
}

/* get set 字段注释 */
func (this *TableName) GetId() uint64 {
	return this.Id
}
func (this *TableName) SetId(id uint64) {
	this.Id = id
}

/* get set 字段注释 */
func (this *TableName) GetName() string {
	return this.Name
}
func (this *TableName) SetName(name string) {
	this.Name = name
}
```


### Install

	go get github.com/go-sql-driver/mysql
	
    go get github.com/slowfei/gosqleu

### Use

	$ gosqleu root:pwd@/dbname
	
	$ gosqleu -h
		-conn="": database connection information [user:password@tcp(localhost:3306)/dbname]
  		-filter-table="": filter table name [table1,table]
 		-getset=true: is struct field out get set func
  		-package="main": package name default main
 		-path="": out file path


