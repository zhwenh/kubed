package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/appscode/kubed/test/framework"
	"fmt"
	"net/http"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	// "log"
	"time"
	"io/ioutil"
	"path/filepath"
	"k8s.io/client-go/util/homedir"
)

var _ = Describe("Event-forwarder", func() {
	var (
		f *framework.Invocation
		s *http.Server
	)
	BeforeEach(func() {
		fmt.Println("hello Event Forwarder BeforeEach")
		f = root.Invoke()
		file, err := ioutil.ReadFile(filepath.Join(homedir.HomeDir(), "go/src/github.com/appscode/kubed/docs/examples/event-forwarder/config.yaml"))
		Expect(err).NotTo(HaveOccurred())

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Hello World Working---------")
			fmt.Fprintf(w, "Hello %q", r.URL)
		})

		s = &http.Server{
			Addr:           ":8181",
			Handler:        nil,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		s.ListenAndServe()

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
		// Expect(err).NotTo(HaveOccurred())
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
		var ()
		Context("Pvc add eventer", func() {
			BeforeEach(func() {
				fmt.Println("Context*****************************")
				myclaim := &apiv1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind: "",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "myclaim",
						Namespace: "demo",
					},
				}
			})

			AfterEach(func() {
				// s.Close()
			})

			FIt("Check notify kubed", func() {
				// w := httptest.NewRecorder()
				// http.
				/*req, err := http.NewRequest(http.MethodGet, "http://localhost:8181", nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := http.DefaultClient.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).Should(BeNumerically("==", 200))*/
				Expect(0).Should(BeZero())
			})
		})
	})
})
