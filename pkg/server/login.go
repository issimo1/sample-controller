package server

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"k8s.io/sample-controller/utils"
	"net/http"
)

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *handler) login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.validate(context.Background(), req.Username, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, _ := utils.GenerateToken(req.Username, req.Password)
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"token":  token,
	})
}

func (h *handler) validate(ctx context.Context, user, pwd string) error {
	pwds, err := h.redis.Get(context.Background(), user).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return errors.New("user not found")
		}
		return errors.New("redis get user error")
	}
	if pwd == "" || pwds != pwd {
		if pwds != pwd {
			return errors.New("username or password is wrong")
		}
		return errors.New("user maybe not registry")
	}
	return nil
}
