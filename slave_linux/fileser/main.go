package fileser

import (
	"archive/zip"
	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Run() {

	gin.SetMode(gin.DebugMode)
	s := gin.Default()
	s.GET("/file", File)
	s.POST("/set/file", SetFile)
	s.GET("/zip/file", ZipFile)
	s.Run(":" + conf.SlaveConf.FileServer.Port)
}

func File(ctx *gin.Context) {
	secret := ctx.Query("secret")
	if secret != conf.SlaveConf.FileServer.Secret {
		ctx.JSON(403, "无权下载")
		return
	}
	filePath := ctx.Query("path")
	// 判断是否为文件
	if !IsFile(filePath) {
		ctx.JSON(403, "无法下载目录，只能下载文件")
		return
	}
	fileName := path.Base(filePath)
	//获取文件的后缀(文件类型)
	fileType := path.Ext(filePath)
	ctx.Header("Content-Type", fileType) // 这里是压缩文件类型 .zip
	ctx.Header("Content-Disposition", "attachment;filename=\""+fileName+"\"")
	ctx.File(filePath)

}

func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

// SetFile 上传文件
func SetFile(ctx *gin.Context) {
	secret := ctx.Query("secret")
	logger.Info("secret = ", secret)
	if secret != conf.SlaveConf.FileServer.Secret {
		ctx.JSON(403, "无权上传")
		return
	}
	savePath := ctx.Request.PostFormValue("path")
	logger.Info("savePath = ", savePath)
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(403, "错误文件")
		return
	}
	dst := savePath + "/" + file.Filename
	isHave, _ := utils.PathExists(dst)
	if isHave {
		ctx.JSON(403, "文件存在")
		return
	}
	err = ctx.SaveUploadedFile(file, dst)
	if err != nil {
		ctx.JSON(403, "保存文件失败")
		return
	}
	ctx.JSON(403, "上传成功")
	return
}

// ZipFile 压缩目录 并下载
func ZipFile(ctx *gin.Context) {
	secret := ctx.Query("secret")
	logger.Info("secret = ", secret)
	if secret != conf.SlaveConf.FileServer.Secret {
		ctx.JSON(403, "无权上传")
		return
	}
	dirPath := ctx.Query("path")
	dirPathList := strings.Split(dirPath, "/")
	lastDirPathName := dirPathList[len(dirPathList)-1]
	// 创建 dir_zip 目录
	_ = os.MkdirAll(conf.SlaveConf.ProjectExeStoreHousePath+"dir_zip/", 0777)
	filePath := conf.SlaveConf.ProjectExeStoreHousePath + "dir_zip/" + lastDirPathName + "_" + utils.NowTimeStrT2() + ".zip"
	err := ZipDir(dirPath, filePath)
	if err != nil {
		logger.Error("目录压缩失败!")
		ctx.JSON(403, "目录压缩失败！")
		return
	}

	fileName := path.Base(filePath)
	//获取文件的后缀(文件类型)
	fileType := path.Ext(filePath)
	ctx.Header("Content-Type", fileType) // 这里是压缩文件类型 .zip
	ctx.Header("Content-Disposition", "attachment;filename=\""+fileName+"\"")
	ctx.File(filePath)
}

func ZipDir(src, outFile string) error {
	logger.Info("outFile = ", outFile)
	// 预防：旧文件无法覆盖
	os.RemoveAll(outFile)

	// 创建：zip文件
	zipFile, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 打开：zip文件
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	// 遍历路径信息
	filepath.Walk(src, func(path string, info os.FileInfo, _ error) error {

		// 如果是源路径，提前进行下一个遍历
		if path == src {
			return nil
		}

		// 获取：文件头信息
		header, _ := zip.FileInfoHeader(info)
		header.Name = strings.TrimPrefix(path, src+`/`)

		// 判断：文件是不是文件夹
		if info.IsDir() {
			header.Name += `/`
		} else {
			// 设置：zip的文件压缩算法
			header.Method = zip.Deflate
		}

		// 创建：压缩包头部信息
		writer, _ := archive.CreateHeader(header)
		if !info.IsDir() {
			file, _ := os.Open(path)
			defer file.Close()
			io.Copy(writer, file)
		}
		return nil
	})
	return nil
}
