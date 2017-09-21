package operator

import (
	"errors"
	"reflect"

	"github.com/appscode/go/log"
	acrt "github.com/appscode/go/runtime"
	"github.com/appscode/kubed/pkg/util"
	kutil "github.com/appscode/kutil/prometheus/v1"
	prom "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

// Blocks caller. Intended to be called as a Go routine.
func (op *Operator) WatchServiceMonitorV1() {
	if !util.IsPreferredAPIResource(op.KubeClient, prom.Group+"/"+prom.Version, prom.ServiceMonitorsKind) {
		log.Warningf("Skipping watching non-preferred GroupVersion:%s Kind:%s", prom.Group+"/"+prom.Version, prom.ServiceMonitorsKind)
		return
	}

	defer acrt.HandleCrash()

	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return op.PromClient.ServiceMonitors(apiv1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return op.PromClient.ServiceMonitors(apiv1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&prom.ServiceMonitor{},
		op.Opt.ResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if res, ok := obj.(*prom.ServiceMonitor); ok {
					log.Infof("ServiceMonitor %s@%s added", res.Name, res.Namespace)
					kutil.AssignTypeKind(res)

					si := op.SearchIndex()
					if si != nil {
						if err := si.HandleAdd(obj); err != nil {
							log.Errorln(err)
						}
					}

					ri := op.ReverseIndex()
					if ri != nil {
						if err := ri.ServiceMonitor.Add(res); err != nil {
							log.Errorln(err)
						}
						if ri.Prometheus != nil {
							proms, err := op.PromClient.Prometheuses(apiv1.NamespaceAll).List(metav1.ListOptions{})
							if err != nil {
								log.Errorln(err)
								return
							}
							if promList, ok := proms.(*prom.PrometheusList); ok {
								ri.Prometheus.AddServiceMonitor(res, promList.Items)
							}
						}
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if res, ok := obj.(*prom.ServiceMonitor); ok {
					log.Infof("ServiceMonitor %s@%s deleted", res.Name, res.Namespace)
					kutil.AssignTypeKind(res)

					si := op.SearchIndex()
					if si != nil {
						if err := si.HandleDelete(obj); err != nil {
							log.Errorln(err)
						}
					}
					tc := op.TrashCan()
					if tc != nil {
						tc.Delete(res.TypeMeta, res.ObjectMeta, obj)
					}

					ri := op.ReverseIndex()
					if ri != nil {
						if err := ri.ServiceMonitor.Delete(res); err != nil {
							log.Errorln(err)
						}
						if ri.Prometheus != nil {
							ri.Prometheus.DeleteServiceMonitor(res)
						}
					}
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldRes, ok := old.(*prom.ServiceMonitor)
				if !ok {
					log.Errorln(errors.New("Invalid ServiceMonitor object"))
					return
				}
				newRes, ok := new.(*prom.ServiceMonitor)
				if !ok {
					log.Errorln(errors.New("Invalid ServiceMonitor object"))
					return
				}
				kutil.AssignTypeKind(oldRes)
				kutil.AssignTypeKind(newRes)

				si := op.SearchIndex()
				if si != nil {
					si.HandleUpdate(old, new)
				}
				tc := op.TrashCan()
				if tc != nil && op.Config.RecycleBin.HandleUpdates {
					if !reflect.DeepEqual(oldRes.Labels, newRes.Labels) ||
						!reflect.DeepEqual(oldRes.Annotations, newRes.Annotations) ||
						!reflect.DeepEqual(oldRes.Spec, newRes.Spec) {
						tc.Update(newRes.TypeMeta, newRes.ObjectMeta, old, new)
					}
				}

				ri := op.ReverseIndex()
				if ri != nil {
					if err := ri.ServiceMonitor.Update(oldRes, newRes); err != nil {
						log.Errorln(err)
					}
				}
			},
		},
	)
	ctrl.Run(wait.NeverStop)
}
