package informers

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"log"
)

var CacheIndex map[string]cache.SharedIndexInformer

type ServiceController struct {
	informer cache.SharedIndexInformer
}

func NewServiceController(cfg *rest.Config) (*ServiceController, error) {
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	lw := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(),
		"services", metav1.NamespaceAll, fields.Everything())
	informers := cache.NewSharedIndexInformer(lw, &v1.Service{},
		0,
		cache.Indexers{"annotation": annotationFunc})
	informers.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//svc := obj.(*v1.Service)
			index, err := informers.GetIndexer().ByIndex("annotation", "slb")
			if err != nil {
				return
			}
			for _, item := range index {
				fmt.Printf("Adding service %s to service controller\n", item.(*v1.Service).Name)
			}
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			oldsvc := oldObj.(*v1.Service)
			newsvc := newObj.(*v1.Service)
			log.Default().Println(newsvc.Name)
			log.Default().Println(oldsvc.Name)
		},
	})
	CacheIndex = make(map[string]cache.SharedIndexInformer)
	CacheIndex["services"] = informers
	return &ServiceController{
		informer: informers,
	}, nil
}

func annotationFunc(obj interface{}) ([]string, error) {
	if service, ok := obj.(*v1.Service); ok {
		if service.Annotations == nil {
			return []string{}, nil
		}
		if _, ok := service.ObjectMeta.Annotations["service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id"]; ok {
			return []string{"slb"}, nil
		}
	}
	return []string{}, nil
}

func (s *ServiceController) Run(stopCh <-chan struct{}, logger klog.Logger) {
	logger.Info("starting wait for cache sync")
	go s.informer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, s.informer.HasSynced) {
		logger.Error(errors.New("timed out waiting for cache to sync"), "service controller")
		return
	}
	logger.Info("cache sync complete, starting informers")
	go s.informer.Run(stopCh)
}
