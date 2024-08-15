package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"os"

	"github.com/gin-gonic/gin"
)

type BindFile struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type FileMeta struct {
	Filename   string    `json:"filename"`
	Expiration time.Time `json:"expiration"`
}

func gen_id(length int) string {
	id := make([]byte, length)

	ch := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	for i := 0; i < length; i++ {
		id[i] += ch[rand.Intn(len(ch))]
	}

	return string(id[:])
}

func check_expire(id string) (bool, error) {
	meta_path := filepath.Join("metadata", id+`.json`)
	json_data, err := os.ReadFile(meta_path)
	if err != nil {
		return false, err
	}

	var file_meta FileMeta
	err = json.Unmarshal(json_data, &file_meta)
	if err != nil {
		return false, err
	}

	return time.Now().After(file_meta.Expiration), nil
}

func create_meta(id string) {
	meta := FileMeta{
		Filename:   id,
		Expiration: time.Now().Add(24 * time.Hour),
	}

	err := os.Mkdir("metadata", 0755)
	if os.IsNotExist(err) {
		os.Mkdir("metadata", 0755)
	}

	meta_name := id + `.json`
	meta_path := filepath.Join("metadata", meta_name)
	file, err := os.Create(meta_path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(meta)
	if err != nil {
		fmt.Println("error encoding JSON: ", err)
		return
	}

	fmt.Println("metadata file created successfully")
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

		expired, err := check_expire(id)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("an error occurred while retrieving requested file."))
			return
		}

		if expired {
			file := filepath.Join("uploads", id)
			meta := filepath.Join("metadata", id+`.json`)
			os.Remove(file)
			os.Remove(meta)
			c.String(http.StatusBadRequest, fmt.Sprintf("Seems like the file you requested has already expired"))
		} else {
			file := filepath.Join("uploads", id)
			c.File(file)
		}
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
		create_meta(id)
		c.String(http.StatusOK, fmt.Sprintf("%s", url))
	})
	router.Run(":8080")
}
