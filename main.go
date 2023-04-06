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
		isFlag := ctx.Request.FormValue("frag")

		file, _ := ctx.FormFile("file")

		// 上传了整个文件
		if len(isFlag) == 0 {
			dst := "./uploads/" + file.Filename
			ctx.SaveUploadedFile(file, dst)

			ctx.JSON(http.StatusOK, gin.H{
				"msg": fmt.Sprintf("'%s' uploaded!", file.Filename),
			})
			return
		}

		// 文件碎片
		os.MkdirAll(filepath.Join("uploads", filenameWithoutExt(file.Filename)), os.ModePerm)

		index := ctx.Request.FormValue("index")

		ctx.SaveUploadedFile(file, filepath.Join("uploads", filenameWithoutExt(file.Filename), index))

		ctx.JSON(http.StatusOK, gin.H{
			"msg": fmt.Sprintf("'%s' uploaded!", file.Filename),
		})
	})

	engine.Run(":1122")
}
