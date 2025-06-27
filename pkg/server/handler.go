package server

import (
	"context"
	"fmt"
	"io"
	"k8s.io/sample-controller/cfg"
	"k8s.io/sample-controller/dao"
	"net/http"
	"os"

	"k8s.io/sample-controller/pkg/controller"
	"k8s.io/sample-controller/pkg/informers"
	"k8s.io/sample-controller/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	v1 "k8s.io/api/core/v1"
	v2 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
)

type handler struct {
	redis         *redis.Client
	clientManager *ClientManager
	dao           *dao.Dao
}

func InstallHandler(group *gin.RouterGroup, cfg map[string]*kubernetes.Clientset, c *cfg.Config) {
	redisClient, err := utils.NewRedisClient(context.Background(), utils.RedisConfig{Addr: os.Getenv("REDIS_ADDR")})
	if err != nil {
		//panic(err)
	}
	h := &handler{
		clientManager: NewClientManager(cfg),
		redis:         redisClient,
		dao:           dao.NewDao(),
	}
	group.POST("/login", h.login)
	router := group.Group("/v1")
	router.Use(h.auth)

	service := router.Group("/service")
	service.Use(h.cache)
	service.GET("cache", h.testcase)
	service.GET("get", h.getService)
	service.GET("update", h.updateService)
	service.GET("testCacheIndexer", h.subSvc)
	service.GET("testPool", h.getResource)
	service.POST("create", h.Create)
	service.POST("update", h.UpdateAndInsert)

	test := group.Group("/v2")
	test.GET("testDelay", h.delayServer)

	prom := group.Group("/v1")
	RegisterPrometheusHandler(h.clientManager, prom, c)
}

// SetService auto handler redis set key,value,
// key = service-namespace-name value = service
func (h *handler) SetService(svc *v1.Service) error {
	name := fmt.Sprintf("service-%s-%s", svc.Namespace, svc.Name)
	return h.redis.Set(context.Background(), name, svc.Annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id"], 0).Err()
}

func (h *handler) GetService(namespace, name string) (interface{}, error) {
	name = fmt.Sprintf("service-%s-%s", namespace, name)
	res, err := h.redis.Get(context.Background(), name).Result()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *handler) testcase(c *gin.Context) {
	if informers.CacheIndex == nil {
		c.JSON(400, gin.H{
			"message": "no informers available",
		})
		return
	}
	var errors field.ErrorList
	if informer, ok := informers.CacheIndex["services"]; ok {
		index, err := informer.GetIndexer().ByIndex("annotation", "slb")
		if err != nil {
			c.JSON(400, gin.H{
				"message": err.Error(),
			})
			return
		}
		for _, item := range index {
			if svc, ok := item.(*v1.Service); ok {
				err := h.SetService(svc)
				if err != nil {
					errors = append(errors, field.Invalid(field.NewPath("metadata", "annotation", "slb", "service"), svc.Name, err.Error()))
					continue
				}
			}
		}
	}
	if len(errors) > 0 {
		c.JSON(400, gin.H{
			"message": errors.ToAggregate().Error(),
		})
		return
	}
	c.JSON(200, nil)
}

func (h *handler) getService(c *gin.Context) {
	val, err := h.GetService(c.Query("namespace"), c.Query("name"))
	if err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	if svc, ok := val.(*v1.Service); ok {
		c.JSON(200, gin.H{
			"status": "success",
			"data":   svc,
		})
		return
	}
	c.JSON(200, nil)
}

func (h *handler) updateService(c *gin.Context) {

}

func (h *handler) subSvc(c *gin.Context) {
	if controller.CacheIndexer == nil {
		c.JSON(400, gin.H{
			"message": "no informers available",
		})
	}
	var errors field.ErrorList
	for _, indexer := range controller.CacheIndexer {
		svcs, err := indexer.ByIndex("test", "success")
		if err != nil {
			errors = append(errors, field.Invalid(field.NewPath("metadata", "annotation", "slb", "service"), svcs, err.Error()))
			continue
		}
		for _, svc := range svcs {
			if svc, ok := svc.(*v1.Service); ok {
				c.JSON(200, gin.H{
					"service": svc,
					"status":  "success",
				})
				return
			}
		}
	}
	if len(errors) > 0 {
		c.JSON(400, gin.H{
			"message": errors.ToAggregate().Error(),
		})
		return
	}
	c.JSON(200, nil)
}

func (h *handler) getResource(c *gin.Context) {
	var req ReqResource
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	var errors field.ErrorList
	type resp struct {
		Name string `json:"name"`
	}
	respMap := make(map[string]map[string]resp)
	for namespace, clusters := range req.Data {
		tmp := make(map[string]resp)
		for _, cluster := range clusters {
			client := h.clientManager.Get(cluster)
			if client == nil {
				errors = append(errors, field.Invalid(field.NewPath("poll", "cluster"), namespace, cluster))
				continue
			}
			services, err := client.CoreV1().Services(namespace).List(context.Background(), v2.ListOptions{})
			if err != nil {
				errors = append(errors, field.Invalid(field.NewPath("list", "service"), namespace, cluster))
				continue
			}
			for _, service := range services.Items {
				tmp[cluster] = resp{
					Name: service.Name,
				}
			}
			h.clientManager.Put(cluster, client)
		}
		respMap[namespace] = tmp
	}
	c.JSON(200, respMap)
}

func (h *handler) delayServer(c *gin.Context) {

}

func repeat(depth int) (interface{}, error) {
	if depth <= 0 {
		return nil, nil
	}
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:8980/v1alpha/getLogs?time=%d", depth))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return repeat(depth / 2)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, nil
}
