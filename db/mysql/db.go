package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var (
	db  *sql.DB
	err error
)

func init() {
	db,_ = sql.Open("mysql","root:root@tcp(127.0.0.1:3306)/unionsdk?charset=utf8")

	//设置DB活跃连接数
	db.SetConnMaxLifetime(1000)
	//设置DB连接测试
	err = db.Ping()

	if err !=nil{
		fmt.Println("Failed to connect to mysql err:"+err.Error())
		//强制结束进程，退出
		os.Exit(1)
	}
}


//DBConn:返回数据库连接对象，提供给外部可以访问
func DBConn() *sql.DB{
	return db
}


func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}