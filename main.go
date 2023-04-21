package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.Default()

	engine.SetTrustedProxies(nil)

	engine.Use(cors.Default())

	// 程序运行的文件夹
	exe, _ := os.Executable()
	exedir := filepath.Dir(exe)
	println(exedir)

	// 检查文件是否已上传或者上传了多少个分片
	engine.POST("/check", func(ctx *gin.Context) {
		type CheckPaylod struct {
			Hash     string `json:"hash"`
			FileName string `json:"fileName"`
		}

		var json CheckPaylod
		if err := ctx.ShouldBindJSON(&json); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		hash := json.Hash
		savePath := filepath.Join(exedir, "files", hash+filepath.Ext(json.FileName))

		_, err := os.Stat(savePath)

		// 存在这个文件了
		if err == nil {
			ctx.JSON(http.StatusOK, gin.H{
				"exist":  1,
				"chunks": []string{},
				"path":   filepath.Join("files", hash+filepath.Ext(json.FileName)),
			})
			return
		}

		// 查看有没有切片
		chunksPath := filepath.Join(exedir, "temp", hash)
		_, err = os.Stat(chunksPath)

		// 存在切片
		if err == nil {
			files, _ := os.ReadDir(chunksPath)

			chunks := []int64{}
			for _, file := range files {
				index, err := strconv.ParseInt(file.Name(), 10, 8)
				if err != nil {
					log.Fatal(err.Error())
					continue
				}
				chunks = append(chunks, index)
			}

			ctx.JSON(http.StatusOK, gin.H{
				"exist":  2,
				"chunks": chunks,
				"path":   "",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"exist":  0,
			"chunks": []string{},
			"path":   "",
		})
	})

	// 上传文件
	engine.POST("/upload", func(ctx *gin.Context) {
		file, _ := ctx.FormFile("file")

		// 上传了整个文件
		if ctx.Request.FormValue("frag") != "yes" {
			savePath := filepath.Join(exedir, "files", ctx.Request.FormValue("hash")+filepath.Ext(ctx.Request.FormValue("fileName")))

			os.MkdirAll(filepath.Join(exedir, "files"), os.ModePerm)

			ctx.SaveUploadedFile(file, savePath)

			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"path":    filepath.Join("files", ctx.Request.FormValue("hash")+filepath.Ext(ctx.Request.FormValue("fileName"))),
			})
			return
		}

		// 文件碎片目录
		os.MkdirAll(filepath.Join(exedir, "temp", ctx.Request.FormValue("hash")), os.ModePerm)
		// 文件碎片保存路径
		savePath := filepath.Join(exedir, "temp", ctx.Request.FormValue("hash"), ctx.Request.FormValue("index"))

		ctx.SaveUploadedFile(file, savePath)

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"path":    "",
		})
	})

	engine.POST("/merge", func(ctx *gin.Context) {
		type MergePaylod struct {
			Hash     string `json:"hash"`
			FileName string `json:"fileName"`
		}

		var json MergePaylod
		if err := ctx.ShouldBindJSON(&json); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		mergePath := filepath.Join(exedir, "temp", json.Hash)

		_, err := os.Stat(mergePath)
		// 没有这个合集
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"success": false,
				"path":    "",
				"message": "没找到文件碎片文件夹",
			})
			return
		}

		os.MkdirAll(filepath.Join(exedir, "files"), os.ModePerm)
		savePath := filepath.Join(exedir, "files", json.Hash+filepath.Ext(json.FileName))

		finFile, err := os.Create(savePath)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"success": false,
				"path":    "",
				"message": "创建合并文件失败",
			})
			return
		}
		defer finFile.Close()

		fs, err := os.ReadDir(mergePath)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"success": false,
				"path":    "",
				"message": "读取碎片文件夹失败",
			})
			return
		}

		sort.Slice(fs, func(i, j int) bool {
			index1, _ := strconv.Atoi(fs[i].Name())
			index2, _ := strconv.Atoi(fs[j].Name())

			return index1 < index2
		})

		for _, f := range fs {
			file, _ := os.Open(filepath.Join(mergePath, f.Name()))
			defer file.Close()

			io.Copy(finFile, file)
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"path":    filepath.Join("files", json.Hash+filepath.Ext(json.FileName)),
		})

		os.RemoveAll(mergePath)
	})

	engine.GET("/files/:path", func(ctx *gin.Context) {
		if path := ctx.Param("path"); path != "" {
			target := filepath.Join(exedir, "files", path)
			ctx.Header("Content-Description", "File Transfer")
			ctx.Header("Content-Transfer-Encoding", "binary")
			ctx.Header("Content-Disposition", "attachment; filename="+path)
			ctx.Header("Content-Type", "application/octet-stream")
			ctx.File(target)
		} else {
			ctx.Status(http.StatusNotFound)
		}
	})

	engine.Run("0.0.0.0:1122")
}
