package main

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// uploadImage 上传图片
func uploadImage(c *gin.Context) {
	// 文件体积限制为15MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 15<<20)

	file, err := c.FormFile("file")
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}

	src, err := file.Open()
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	defer src.Close()

	content := []byte{}
	for {
		tmp := make([]byte, 32*1024)
		n, err := src.Read(tmp)
		content = append(content, tmp[:n]...)
		if err != nil {
			if err == io.EOF {
				break
			}
			c.Error(err).SetType(gin.ErrorTypePrivate)
			return
		}
	}
	query := "INSERT INTO image (name, data) VALUES (?, ?)"
	db := c.MustGet("DB").(*sqlx.DB)
	res, err := db.Exec(query, file.Filename, content)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": id})
}

// showImage 查看图片
func showImage(c *gin.Context) {
	imageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	query := "SELECT data FROM image WHERE id = ?"
	var image []byte
	err = db.QueryRow(query, imageID).Scan(&image)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	c.Data(200, "image/jpeg", image)
}

// deleteImage 删除图片
func deleteImage(c *gin.Context) {
	imageID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	db := c.MustGet("DB").(*sqlx.DB)
	_, err = db.Exec("DELETE FROM image WHERE id = ?", imageID)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": "删除成功"})
}
