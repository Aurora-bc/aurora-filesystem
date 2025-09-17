package hdl

import "net/http"

//初始化分块信息
func InitateMultiparUploadHandler(w http.ResponseWriter,r *http.Request){
	//判断是否已经上传过，如果上传过，就秒传
	//生成唯一的上传ID
	//缓存分块初始化信息
}


//上传分块
func UploadPartHandler(w http.ResponseWriter,r *http.Request){

}

//通知分块上传完成
func CompleteUploadPartHandler(w http.ResponseWriter,r *http.Request){

}

//取消上传分块
func CancelUploadPartHandler(w http.ResponseWriter,r *http.Request){

}

//查看分块上传的整体状态
func MultiparUploadStatusHandler(w http.ResponseWriter,r *http.Request){

}