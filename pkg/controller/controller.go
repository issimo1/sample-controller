package controller

import (
	"context"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"log"
	"time"
)

type Controller struct {
	clientSet *kubernetes.Clientset
	reflector *cache.Reflector
	deltaFifo *cache.DeltaFIFO
}

var CacheIndexer map[string]cache.Indexer

func NewServiceController(cfg *rest.Config) (*Controller, error) {
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	listAndWatch := cache.NewListWatchFromClient(client.CoreV1().RESTClient(),
		"services",
		metav1.NamespaceAll,
		fields.Everything())
	store := cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{})
	reflector := cache.NewNamedReflector("services", listAndWatch, &corev1.Service{}, store, 30*time.Second)
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{
		"test": func(obj interface{}) ([]string, error) {
			if svc, ok := obj.(*corev1.Service); ok {
				if svc.Annotations["lwcontroller"] == "true" {
					return []string{"success"}, nil
				}
			}
			return []string{}, nil
		},
	})
	CacheIndexer = make(map[string]cache.Indexer)
	CacheIndexer["services"] = indexer
	return &Controller{
		clientSet: client,
		reflector: reflector,
		deltaFifo: store,
	}, nil
}

func (c *Controller) Run(ctx context.Context) {
	stopCh := make(chan struct{})
	go c.reflector.Run(stopCh)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("context cancelled")
			return
		case <-stopCh:
			return
		default:
			_, err := c.deltaFifo.Pop(func(obj interface{}, isInInitialList bool) error {
				switch v := obj.(type) {
				case cache.Deltas:
					vv := v.Newest()
					svc, ok := vv.Object.(*corev1.Service)
					if !ok {
						return errors.New("object is not a service")
					}
					switch vv.Type {
					case cache.Sync, cache.Added, cache.Updated:
						if _, exists, err := CacheIndexer["services"].Get(svc); err == nil && exists {
							if err := CacheIndexer["services"].Update(svc); err != nil {
								return err
							}
						} else {
							if err := CacheIndexer["services"].Add(svc); err != nil {
								return err
							}
						}
						return updateService(svc)
					}
				}
				return nil
			})
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func updateService(svc *corev1.Service) error {
	if _, ok := svc.Annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id"]; !ok {
		log.Printf("service：%v does not contain annotation 'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id'", svc.Name)
	}
	return nil
}
