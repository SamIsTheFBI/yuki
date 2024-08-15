package main

import (
	"fmt"
	"math/rand"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"os"

	"github.com/gin-gonic/gin"
)

type BindFile struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func gen_id(length int) string {
	id := make([]byte, length)

	ch := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	for i := 0; i < length; i++ {
		id[i] += ch[rand.Intn(len(ch))]
	}

	return string(id)
}

func main() {
	router := gin.Default()

	var limit int64 = 8 << 20

	router.MaxMultipartMemory = limit

	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")

		file := filepath.Join("uploads", id)
		c.File(file)
	})

	router.POST("/", func(c *gin.Context) {
		var bindFile BindFile

		if err := c.ShouldBind(&bindFile); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("err: %s", err.Error()))
			return
		}

		file := bindFile.File
		if file.Size > limit {
			c.String(http.StatusBadRequest, fmt.Sprintf("file size limit exceeded!"))
			return
		}

		err := os.Mkdir("uploads", 0755)
		if os.IsNotExist(err) {
			os.Mkdir("uploads", 0755)
		}

		id := gen_id(4) + filepath.Ext(file.Filename)

		dst := filepath.Join("uploads", id)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		hostname := c.Request.Host
		url := hostname + `/` + id
		c.String(http.StatusOK, fmt.Sprintf("%s", url))
	})
	router.Run(":8080")
}
