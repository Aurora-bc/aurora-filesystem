package db

import(
		mab "../db/mysql"
	"database/sql"
	"fmt"
)

//当文件上传完成后，将Meta信息插入到mysql表中
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	//打开数据库连接，准备插入新的数据
	stmt,err:= mab.DBConn().Prepare("insert ignore into un_file (`file_sha1`,`file_name`,`file_size`,`file_addr`,`status`) values(?,?,?,?,1) ")
	if err !=nil{
		fmt.Println("Failed to prepare statement err:"+err.Error())
		return false
	}
	defer stmt.Close()

	//将数据插入到数据表中
	ret,err:=stmt.Exec(filehash,filename,filesize,fileaddr)
	if err != nil{
		fmt.Println(err.Error())
		return false
	}

	//这里是判断数据表中是否存在有filehash值
	if rf,err:=ret.RowsAffected();nil==err{
		if rf<0{
			fmt.Printf("File with hash:%s has been uploaded before",filehash)
		}
		//如果之前的数据表中已经存在Hash值，就返回true
		return true
	}
	//最后，插入失败返回false
	return false
}


// TableFile : 文件表结构体
type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta : 从mysql获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mab.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from un_file " +
			"where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(
		&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			// 查不到对应记录， 返回参数及错误均为nil
			return nil, nil
		} else {
			fmt.Println(err.Error())
			return nil, err
		}
	}
	return &tfile, nil
}

// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := mab.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from un_file " +
			"where status=1 limit ?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cloumns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(cloumns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr,
			&tfile.FileName, &tfile.FileSize)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	fmt.Println(len(tfiles))
	return tfiles, nil
}

// UpdateFileLocation : 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mab.DBConn().Prepare(
		"update un_file set`file_addr`=? where  `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("更新文件location失败, filehash:%s", filehash)
		}
		return true
	}
	return false
}

//OnFileDeleteUserFileFinished：删除用户文件表的记录
func OnFileDeleteUserFileFinished(filesha1 string) bool{
	stmt, err := mab.DBConn().Prepare(
		"delete from un_user_file where `file_sha1`=?")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filesha1)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("删除文件失败, filehash:%s", filesha1)
		}
		return true
	}
	return false
}