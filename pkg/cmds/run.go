package cmds

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/runtime"
	"github.com/appscode/kubed/pkg/operator"
	srch_cs "github.com/appscode/searchlight/client/clientset"
	scs "github.com/appscode/stash/client/clientset"
	vcs "github.com/appscode/voyager/client/clientset"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	"github.com/fsnotify/fsnotify"
	kcs "github.com/k8sdb/apimachinery/client/clientset"
	"github.com/spf13/cobra"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type Event struct {
	Name string
	Op   Op
}

type Op uint32

// runtime.GOPath() + "/src/github.com/appscode/kubed/hack/config/clusterconfig.yaml"
func NewCmdRun() *cobra.Command {
	opt := operator.Options{
		ConfigPath:        "/srv/kubed/config.yaml",
		APIAddress:        ":8080",
		WebAddress:        ":56790",
		ScratchDir:        "/tmp",
		OperatorNamespace: namespace(),
		ResyncPeriod:      5 * time.Minute,
	}
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run daemon",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			log.Infoln("Starting kubed...")

			Run(opt)
		},
	}
	fmt.Println("Hello World!!!!")
	fmt.Println("read config file")
	cmd.Flags().StringVar(&opt.KubeConfig, "kubeconfig", opt.KubeConfig, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	cmd.Flags().StringVar(&opt.Master, "master", opt.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig)")
	cmd.Flags().StringVar(&opt.ConfigPath, "clusterconfig", opt.ConfigPath, "Path to cluster config file")
	cmd.Flags().StringVar(&opt.ScratchDir, "scratch-dir", opt.ScratchDir, "Directory used to store temporary files. Use an `emptyDir` in Kubernetes.")
	cmd.Flags().StringVar(&opt.APIAddress, "api.address", opt.APIAddress, "The address of the Kubed API Server (overrides any value in clusterconfig)")
	cmd.Flags().StringVar(&opt.WebAddress, "web.address", opt.WebAddress, "Address to listen on for web interface and telemetry.")
	cmd.Flags().DurationVar(&opt.ResyncPeriod, "resync-period", opt.ResyncPeriod, "If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out.")

	fmt.Println("___________________________ Hello Test")
	go fileWatchTest()
	return cmd
}

func Run(opt operator.Options) {
	log.Infoln("configurations provided for kubed", opt)
	defer runtime.HandleCrash()

	c, err := clientcmd.BuildConfigFromFlags(opt.Master, opt.KubeConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	op := &operator.Operator{
		KubeClient:        clientset.NewForConfigOrDie(c),
		VoyagerClient:     vcs.NewForConfigOrDie(c),
		SearchlightClient: srch_cs.NewForConfigOrDie(c),
		StashClient:       scs.NewForConfigOrDie(c),
		KubeDBClient:      kcs.NewForConfigOrDie(c),
		Opt:               opt,
	}
	op.PromClient, err = pcm.NewForConfig(c)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = op.Setup()
	if err != nil {
		log.Fatalln(err)
	}

	log.Infoln("Running kubed watcher")
	op.RunAndHold()
}

func namespace() string {
	if ns := os.Getenv("OPERATOR_NAMESPACE"); ns != "" {
		return ns
	}
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return apiv1.NamespaceDefault
}

func fileWatchTest() {
	fmt.Println("-----------------File Watch Test Began----------------")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error Occured ***************")
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Infoln("Event:-------------------------------------------------------", event, reflect.TypeOf(event), event.String())
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Infoln("modified file:", event.Name)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Infoln("Removed file:---------------", event.Name)

					if filepath.Clean(event.Name) == "/srv/kubed/config.yaml" {
						err = printMD5("/srv/kubed/config.yaml")
						if err != nil {
							log.Errorln("fffffffffffffffffff Error", err)
						}
						err = watcher.Add("/srv/kubed/config.yaml")
						if err != nil {
							log.Errorln("wwwwwwwwwwwwwwwwwww Error", err)
						}
					}
				}
			case err := <-watcher.Errors:
				log.Infoln("error:", err)
			}
		}
	}()

	err = watcher.Add("/srv/kubed/config.yaml")
	if err != nil {
		log.Fatalln("1st Error", err)
	}
	err = watcher.Add("/srv/kubed/")
	if err != nil {
		log.Fatalln("2nd Error", err)
	}
	<-done
}

func printMD5(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	fmt.Printf("%x\n", h.Sum(nil))
	return nil
}
