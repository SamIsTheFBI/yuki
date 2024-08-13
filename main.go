package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"os"
)

type BindFile struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
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

		// Bind file
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

		dst := filepath.Join("uploads", filepath.Base(file.Filename))
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}

		hostname := c.Request.Host
		url := hostname + `/` + file.Filename
		c.String(http.StatusOK, fmt.Sprintf("%s", url))
	})
	router.Run(":8080")
}
