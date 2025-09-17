package handler

import (
	darer "../db"
	"../meta"
	"../utils"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//上传文件处理器
func UpLoadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传的Html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "Internal Server Error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//1、接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data error:%s\n", err.Error())
			return
		}
		defer file.Close()

		//文件元信息初始化
		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "D:/UpLoad/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		//2、创建新文件
		newfile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file error:%s\n", err.Error())
			return
		}
		defer newfile.Close()

		//3、复制文件
		fileMeta.FileSize, err = io.Copy(newfile, file)
		if err != nil {
			fmt.Printf("Failed save data into file error:%s\n", err.Error())
			return
		}

		//将新创建的文件移动开始的位置(游标重新回到文件头部)
		newfile.Seek(0, 0)
		//计算新创建文件的Hash值，赋值到文件元结构体的FileSha1中
		fileMeta.FileSha1 = util.FileSha1(newfile)
		//fmt.Println("文件的Hash值是：", fileMeta.FileSha1)
		//更新文件元的FileSha1值
		//meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		//更新用户文件表记录
		r.ParseForm()
		username:=r.Form.Get("username")
		suc:= darer.OnUserFileUploadFinished(username,fileMeta.FileSha1,fileMeta.FileName,fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed."))
		}

		//4、重定向
		http.Redirect(w, r, "/file/upload/success", http.StatusFound)
	}
}

//上传文件成功提示页面处理器
func UpLoadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "文件上传成功！")
}

//获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta,err:=meta.GetFileMetaDB(filehash)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

//查询批量文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCut, _ := strconv.Atoi(r.Form.Get("limit"))

	//fileMetas := meta.GetLastFileMetas(limitCut)
	username:=r.Form.Get("username")
	userFiles,err:= darer.QueryUserFileMetas(username,limitCut)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

//下载文件逻辑
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	//fm := meta.GetFileMeta(fsha1)

	fm,err:=meta.GetFileMetaDB(fsha1)
	if err !=nil{
		fmt.Println("下载出错了,错误是："+err.Error())
		return
	}

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//设置下载的请求头信息
	w.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	w.Header().Set("content-disposition", "attachment; filename=\""+fm.FileName+"\"")

	w.Write(data)
}

//FileMetaUpdateHandler:更新文件元信息接口（重命名）
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	//客户端如果传入的不是0，则返回403错误
	if opType != "0" {
		w.WriteHeader(http.StatusForbidden) //403
		return
	}

	//客户端如果不是POST请求，则返回405错误
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//查找fileSha1对应的文件，改名字，并保存到当前的文件元信息结构体中
	//curFileMeta := meta.GetFileMeta(fileSha1)

	curFileMeta,err:=meta.GetFileMetaDB(fileSha1)
	if err!=nil{
		fmt.Println("出错了")
		return
	}

	curFileMeta.FileName = newFileName
	meta.UpdateFileMetaDB(*curFileMeta)

	//将curFileMeta以json格式返回给客户端
	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError) //500
		return
	}

	//如果正常，则返回200
	w.WriteHeader(http.StatusOK)

	//将data数据写到客户端
	w.Write(data)
}

//FileDeleteHandler：删除文件元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	//获取到fileSha1的文件元信息
	//fMeta := meta.GetFileMeta(fileSha1)
	fMeta,err:=meta.GetFileMetaDB(fileSha1)
	if err != nil{
		fmt.Println("出错了")
		return
	}

	//调用os的删除方法删除对应的文件
	os.Remove(fMeta.Location)
	//在文件元结构体中删除fileSha1的文件元信息
	//Bug 数据库中的地址没有删掉
	//meta.RemoveFileMeta(fileSha1)
	suc:=meta.RemoveFileMetaDB(fileSha1)
	if suc{
		w.Write([]byte("删除成功"))
	}else{
		w.Write([]byte("删除失败"))
	}
	//给客户端返回正常的状态信息
	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler : 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表， 返回成功
	suc := darer.OnUserFileUploadFinished(
		username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

// DownloadURLHandler : 生成文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	row, _ := darer.GetFileMeta(filehash)

	// TODO: 判断文件存在OSS，还是Ceph，还是在本地
	if strings.HasPrefix(row.FileAddr.String, "D:/Upload/") {
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			r.Host, filehash, username, token)
		fmt.Println(tmpUrl)
		w.Write([]byte(tmpUrl))
	} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
		// TODO: ceph下载url
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		// oss下载url
		//signedURL := oss.DownloadURL(row.FileAddr.String)
		//w.Write([]byte(signedURL))
	}
}
