package server

import (
	"github.com/gin-gonic/gin"
	"k8s.io/sample-controller/utils"
	"net/http"
	"strings"
)

func (h *handler) auth(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Not authorized",
			"error":   "basic author error",
		})
	}
	// token= Bearer/Basic xx
	auth := strings.Split(token, " ")
	if len(auth) != 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Not authorized",
			"error":   "invalid token",
		})
	}
	switch auth[0] {
	case "Basic":
		user, pwd, ok := c.Request.BasicAuth()
		if !ok || user != "admin" || pwd != "admin" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": http.StatusUnauthorized,
			})
		}
	case "Bearer":
		claims, err := utils.ParseToken(auth[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Not authorized",
				"error":   "parse token error",
			})
			return
		}
		c.Set("userid", claims.Userid)
		c.Set("role", claims.Role)
	}
	c.Next()
}

func (h *handler) cache(c *gin.Context) {
	name := c.Query("name")
	namespace := c.Query("namespace")
	if name == "" {
		c.Next()
		return
	}
	val, err := h.GetService(namespace, name)
	if err != nil {
		c.Next()
		return
	}
	if svc, ok := val.(string); ok {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   svc,
		})
		return
	}
	c.Next()
}
