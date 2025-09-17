package main

import (
	"FileSystem/handler"
	"fmt"
	"net/http"
)

func main() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	//文件上传
	http.HandleFunc("/file/upload/", handler.HTTPInterceptor(handler.UpLoadHandler))
	//处理文件上传成功的重定向操作
	http.HandleFunc("/file/upload/success", handler.HTTPInterceptor(handler.UpLoadSuccessHandler))
	http.HandleFunc("/file/meta", handler.HTTPInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/query", handler.HTTPInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/file/download", handler.HTTPInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/update", handler.HTTPInterceptor(handler.FileMetaUpdateHandler))
	http.HandleFunc("/file/delete", handler.HTTPInterceptor(handler.FileDeleteHandler))
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler))

	//用户注册请求
	http.HandleFunc("/user/signup", handler.SignupHandler)
	//用户登陆
	http.HandleFunc("/user/signin", handler.SignInHandler)

	//查询用户信息
	//http.HandleFunc("/user/info",handler.UserInfoHandler)
	//查询用户信息(增加了拦截器)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	//------------------------分块上传部份-----------------------------------------------------//
	//1、初始化分块信息
	//http.HandleFunc("/file/mpupload/init",hdl.AccessAuth(hdl.InitateMultiparUploadHandler))
	//2、上传分块
	//http.HandleFunc("/file/mpupload/uppart",hdl.AccessAuth(hdl.UploadPartHandler))
	//3、通知分块上传完成
	//http.HandleFunc("/file/mpupload/complete",hdl.AccessAuth(hdl.CompleteUploadPartHandler))
	//4、取消上传分块
	//http.HandleFunc("/file/mpupload/cancel",hdl.AccessAuth(hdl.CancelUploadPartHandler))
	//5、查看分块上传的整体状态
	//http.HandleFunc("/file/mpupload/status",hdl.AccessAuth(hdl.MultiparUploadStatusHandler))

	http.HandleFunc("/file/downloadurl", handler.HTTPInterceptor(handler.DownloadURLHandler))

	//端口监听服务
	err := http.ListenAndServe("", nil)
	if err != nil {
		fmt.Printf("Failed start server error:&s", err.Error())
	}
}
