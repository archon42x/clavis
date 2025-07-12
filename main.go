package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	token  string
	mu     sync.Mutex
	config map[string]any = make(map[string]any)
)

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

func init() {
	token = os.Getenv("CLAVIS_TOKEN")
	if token == "" {
		log.Fatal("CLAVIS_TOKEN is empty")
	}
	log.Printf("load token success\n")
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

		val, exists := config[key]
		if !exists {
			log.Printf("key not exists\n")
			c.JSON(http.StatusOK, gin.H{
				"code": GET_ERROR,
				"msg":  "key not exists",
			})
			return
		}

		log.Printf("get key success\n")
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "get key success",
			"data": val,
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

		mu.Lock()
		defer mu.Unlock()
		config[key] = req.Value

		log.Printf("set key success\n")
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "set key success",
		})
	})

	r.Run()
}

type ErrorCode uint32

const (
	GET_ERROR ErrorCode = 10001
	SET_ERROR ErrorCode = 10002
)
