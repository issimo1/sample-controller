package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/sample-controller/dao"
	"net/http"
)

func (h *handler) Create(c *gin.Context) {
	var req []Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, item := range req {
		svc := dao.Service{
			Name:  item.Name,
			Types: item.Type,
			Ip:    item.IP,
		}
		err := h.dao.Service.Create(context.Background(), &svc)
		if err != nil {
			err = fmt.Errorf("create service error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func (h *handler) UpdateAndInsert(c *gin.Context) {
	var req []Req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var svcs []*dao.Service
	for _, item := range req {
		svc := dao.Service{
			Name:  item.Name,
			Types: item.Type,
			Ip:    item.IP,
		}
		err := h.dao.Service.Update(context.TODO(), &svc)
		if err != nil {
			err = fmt.Errorf("update service error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		svcs = append(svcs, &svc)
	}
	//err := h.dao.Service.UpdateInBatchWithConflict(
	//	context.Background(),
	//	svcs,
	//	5,
	//	"name",
	//	[]string{"type", "ip"})
	//if err != nil {
	//	err = fmt.Errorf("update service error: %v", err)
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
