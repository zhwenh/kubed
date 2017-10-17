package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/appscode/kubed/test/framework"
	"fmt"
	"net/http"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"time"
	"io/ioutil"
	"path/filepath"
	"k8s.io/client-go/util/homedir"
	"net/http/httptest"
	"strings"
)

var _ = Describe("Event-forwarder", func() {
	var (
		f       *framework.Invocation
		reqs    []*http.Request
		s       *http.Server
		handler = func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%q", r.URL)
		}
	)
	BeforeEach(func() {
		f = root.Invoke()
		file, err := ioutil.ReadFile(filepath.Join(homedir.HomeDir(), "go/src/github.com/appscode/kubed/docs/examples/event-forwarder/config.yaml"))
		Expect(err).NotTo(HaveOccurred())
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			reqs = append(reqs, r)
			fmt.Fprintf(w, "%q", r.URL)
		})

		s = &http.Server{
			Addr:           ":8181",
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		go s.ListenAndServe()
		notifierConfig := &apiv1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "notifier-config",
				Namespace: "kube-system",
			},
			Data: map[string][]byte{
				"WEBHOOK_URL": []byte("http://localhost:8181"),
			},
		}

		_, err = f.KubeClient.CoreV1().Secrets("kube-system").Get("notifier-config", metav1.GetOptions{})
		if err != nil {
			_, err = f.KubeClient.CoreV1().Secrets("kube-system").Create(notifierConfig)
		} else {
			_, err = f.KubeClient.CoreV1().Secrets("kube-system").Update(notifierConfig)
		}
		Expect(err).NotTo(HaveOccurred())

		kubedConfig := &apiv1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubed-config",
				Namespace: "kube-system",
				Labels: map[string]string{
					"app": "kubed",
				},
			},
			Data: map[string][]byte{
				"config.yaml": file,
			},
		}

		_, err = f.KubeClient.CoreV1().Secrets("kube-system").Get("kubed-config", metav1.GetOptions{})
		if err != nil {
			_, err = f.KubeClient.CoreV1().Secrets("kube-system").Create(kubedConfig)
		} else {
			_, err = f.KubeClient.CoreV1().Secrets("kube-system").Update(kubedConfig)
		}
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Checkout event-forward", func() {
		Context("Pvc add eventer", func() {
			BeforeEach(func() {
				myclaim := &apiv1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "PersistentVolumeClaim",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myclaim",
						Namespace: f.Namespace(),
					},
					Spec: apiv1.PersistentVolumeClaimSpec{
						AccessModes: append([]apiv1.PersistentVolumeAccessMode{}, "ReadWriteOnce"),
						Resources: apiv1.ResourceRequirements{
							Requests: apiv1.ResourceList{
								"storage": resource.Quantity{

								},
							},
						},
					},
				}
				_, err := f.KubeClient.CoreV1().PersistentVolumeClaims(f.Namespace()).Get("myclaim", metav1.GetOptions{})
				if err == nil {
					err = f.KubeClient.CoreV1().PersistentVolumeClaims(f.Namespace()).Delete("myclaim", &metav1.DeleteOptions{})
				}
				_, err = f.KubeClient.CoreV1().PersistentVolumeClaims(f.Namespace()).Create(myclaim)
				Expect(err).NotTo(HaveOccurred())

			})

			FIt("Check notify kubed", func() {
				Eventually(func() bool {
					for _, val := range reqs {
						wr := httptest.NewRecorder()
						handler(wr, val)
						result := wr.Result()
						bit, err := ioutil.ReadAll(result.Body)
						Expect(err).NotTo(HaveOccurred())
						respStr := string(bit)
						if strings.Contains(respStr, "PersistentVolumeClaim") && result.StatusCode == 200 {
							return true
						}
					}
					return false
				}).Should(BeTrue())
				Expect(0).Should(BeZero())
			})
		})
		Context("Warning Event", func() {
			BeforeEach(func() {
				wPod := &apiv1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "busybox",
						Namespace: "kube-system",
					},
					Spec: apiv1.PodSpec{
						RestartPolicy: "Never",
						Containers: append([]apiv1.Container{}, apiv1.Container{
							Name:            "busybox",
							Image:           "busybox",
							ImagePullPolicy: "IfNotPresent",
							Command:         []string{"bad", "3600"},
						}),
					},
				}
				_, err := f.KubeClient.CoreV1().Pods("kube-system").Create(wPod)
				Expect(err).NotTo(HaveOccurred())
			})
			AfterEach(func() {
				err := f.KubeClient.CoreV1().Pods("kube-system").Delete("busybox", &metav1.DeleteOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			FIt("Check warning event", func() {
				Eventually(func() bool {
					for _, val := range reqs {
						wr := httptest.NewRecorder()
						handler(wr, val)
						resp := wr.Result()
						byt, err := ioutil.ReadAll(resp.Body)
						Expect(err).NotTo(HaveOccurred())
						respStr := string(byt)
						if resp.StatusCode == 200 && strings.Contains(respStr, "busybox"){
							return true
						}
					}
					return false
				}).Should(BeTrue())
			})
		})
	})
})
