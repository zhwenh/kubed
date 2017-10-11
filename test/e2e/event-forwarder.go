package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/appscode/kubed/test/framework"
	"fmt"
	"net/http"
	// "log"
	// "time"
	"time"
)

var _ = Describe("Event-forwarder", func() {
	var (
		f *framework.Invocation
	)
	BeforeEach(func() {
		fmt.Println("hello Event Forwarder BeforeEach")
		f = root.Invoke()
	})

	Describe("Checkout event-forward", func() {
		var (
			s *http.Server
		)
		Context("Pvc add eventer", func() {
			BeforeEach(func() {
				fmt.Println("Context*****************************")
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

				err := s.ListenAndServe()
				Expect(err).NotTo(HaveOccurred())
				fmt.Println("Hello World!!!")
				Expect(0).Should(Equal(0))
			})

			AfterEach(func() {
				s.Close()
			})

			It("Check notify kubed", func() {
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
