package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	engine.Use(cors.Default())

	// 程序运行的文件夹
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	println(dir)

	// 上传文件
	engine.POST("/upload", func(ctx *gin.Context) {
		file, _ := ctx.FormFile("file")

		// 最终文件路径
		filePath := filepath.Join("files", ctx.Request.FormValue("hash")+filepath.Ext(ctx.Request.FormValue("fileName")))

		_, err := os.Stat(filePath)

		if err == nil {
			ctx.JSON(http.StatusOK, gin.H{
				"msg": fmt.Sprintf("file exist in %s", filePath),
			})
			return
		}

		// 上传了整个文件
		if ctx.Request.FormValue("frag") != "yes" {
			os.MkdirAll(filepath.Join("files"), os.ModePerm)

			ctx.SaveUploadedFile(file, filePath)

			ctx.JSON(http.StatusOK, gin.H{
				"msg": filePath,
			})
			return
		}

		// 文件碎片目录
		os.MkdirAll(filepath.Join("temp", ctx.Request.FormValue("hash")), os.ModePerm)

		// 碎片存储路径
		ctx.SaveUploadedFile(file, filepath.Join("temp", ctx.Request.FormValue("hash"), ctx.Request.FormValue("index")))

		ctx.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("'%s' uploaded!", file.Filename),
		})
	})

	engine.Run(":1122")
}
