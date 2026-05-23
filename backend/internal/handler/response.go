package handler

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": data})
}

func OKWithMeta(c *gin.Context, data interface{}, meta gin.H) {
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": data, "meta": meta})
}

func Err(c *gin.Context, code int, httpStatus int, msg string) {
	c.JSON(httpStatus, gin.H{"code": code, "message": msg, "data": nil})
}
