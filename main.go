package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var (
	token string
	db    *gorm.DB
)

func init() {
	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		log.Fatal("MYSQL_DSN is empty")
	}
	var err error
	db, err = gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("connect database failed: %v\n", err)

	token = os.Getenv("CLAVIS_TOKEN")
	if token == "" {
		log.Fatal("CLAVIS_TOKEN is empty")
	}
	log.Printf("load token success\n")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid token"})
			c.Abort()
			return
		}

		tokenStr := authHeader[7:]
		if tokenStr != token {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func main() {
	r := gin.Default()

	protected := r.Group("/", AuthMiddleware())

	protected.GET("/get", func(c *gin.Context) {
		key := c.Query("key")
		if key == "" {
			log.Printf("key is empty\n")
			c.JSON(http.StatusOK, gin.H{
				"code": GET_ERROR,
				"msg":  "key is empty",
			})
			return
		}

		kv := &KV{}
		err := db.Where("`key` = ?", key).First(kv).Error
		if err == gorm.ErrRecordNotFound {
			log.Printf("key not exists\n")
			c.JSON(http.StatusOK, gin.H{
				"code": GET_ERROR,
				"msg":  "key not exists",
			})
			return
		} else if err != nil {
			log.Printf("database query error: %v\n", err)
			c.JSON(http.StatusOK, gin.H{
				"code": GET_ERROR,
				"msg":  "database query error",
			})
			return
		}

		log.Printf("get key success\n")
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "get key success",
			"data": kv.Value,
		})
	})

	protected.POST("/set", func(c *gin.Context) {
		type Request struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		req := &Request{}
		if err := c.ShouldBindJSON(req); err != nil {
			log.Printf("json parse error: %s\n", err)
			c.JSON(http.StatusOK, gin.H{
				"code": GET_ERROR,
				"msg":  "json parse error",
			})
			return
		}

		key := req.Key
		if key == "" {
			log.Printf("key is empty\n")
			c.JSON(http.StatusOK, gin.H{
				"code": GET_ERROR,
				"msg":  "key is empty",
			})
			return
		}

		value := req.Value

		kv := &KV{}
		err := db.Where("`key` = ?", key).First(kv).Error
		if err == nil {
			kv.Value = value
			err := db.Where("`key` = ?", key).Updates(kv).Error
			if err != nil {
				log.Printf("update key error: %v\n", err)
				c.JSON(http.StatusOK, gin.H{
					"code": GET_ERROR,
					"msg":  "update key error",
				})
				return
			} else {
				log.Printf("update key success\n")
				c.JSON(http.StatusOK, gin.H{
					"code": 0,
					"msg":  "update key success",
				})
				return
			}
		} else if err == gorm.ErrRecordNotFound {
			kv.Key = key
			kv.Value = value
			err := db.Create(kv).Error
			if err != nil {
				log.Printf("create key error: %v\n", err)
				c.JSON(http.StatusOK, gin.H{
					"code": GET_ERROR,
					"msg":  "create key error",
				})
				return
			} else {
				log.Printf("create key success\n")
				c.JSON(http.StatusOK, gin.H{
					"code": 0,
					"msg":  "create key success",
				})
				return
			}
		} else {
			log.Printf("database query error: %v\n", err)
			c.JSON(http.StatusOK, gin.H{
				"code": GET_ERROR,
				"msg":  "database query error",
			})
			return
		}
	})

	r.Run()
}

type ErrorCode uint32

const (
	GET_ERROR ErrorCode = 10001
	SET_ERROR ErrorCode = 10002
)
