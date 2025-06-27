package server

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/sample-controller/cfg"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func init() {
	prometheus.MustRegister(walSegment, walSize)
}

var (
	// Metrics
	walSegment = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "prometheus_wal_segment_count",
			Help: "The number of WAL segments currently in use",
		},
	)
	walSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "prometheus_wal_size_bytes",
			Help: "The size of the WAL in bytes",
		},
	)
)

func RegisterPrometheusHandler(c *ClientManager, r *gin.RouterGroup, cfg *cfg.Config) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	go func() {
		for {
			client := c.Get("mt")
			promPod, err := client.CoreV1().Pods("monitor").List(context.Background(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=prometheus"})
			if err != nil {
				continue
			}
			var promName string
			for _, pod := range promPod.Items {
				if pod.Status.Phase == "Running" {
					promName = pod.Name
				}
			}
			req := client.CoreV1().RESTClient().Post().
				Resource("pods").
				Name(promName).
				Namespace("monitor").
				SubResource("exec").
				Param("container", "prometheus").
				VersionedParams(&corev1.PodExecOptions{
					Command: []string{"sh", "-c", `ls -lh /prometheus/wal/*`},
					Stdin:   false,
					Stdout:  true,
					Stderr:  true,
					TTY:     false,
				}, scheme.ParameterCodec)

			exec, err := remotecommand.NewSPDYExecutor(cfg.Kubeconfig, "POST", req.URL())
			if err != nil {
				continue
			}
			stdout := bytes.Buffer{}
			stderr := bytes.Buffer{}
			err = exec.Stream(remotecommand.StreamOptions{
				Stdin:  nil,
				Stdout: &stdout,
				Stderr: &stderr,
				Tty:    false,
			})
			if err != nil {
				log.Fatalln(err)

			}
			re := regexp.MustCompile(`^\d{8}$`)
			log.Println(stdout.String())
			for _, line := range strings.Split(stdout.String(), "\n") {
				parts := strings.Fields(line)
				if len(parts) < 8 {
					continue
				}
				if strings.Contains(line, "total") {
					continue
				}
				filename := parts[len(parts)-1]
				file := filepath.Base(filename)
				if re.MatchString(file) {
					log.Printf("文件名:%s", file)
					walSegment.Add(float64(1))
				}
			}
			walSegment.Set(float64(0))
			time.Sleep(time.Duration(cfg.PromMetricsIntervals) * time.Second)
		}
	}()
}
