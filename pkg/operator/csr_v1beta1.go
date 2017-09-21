package operator

import (
	"errors"
	"reflect"

	"github.com/appscode/go/log"
	acrt "github.com/appscode/go/runtime"
	"github.com/appscode/kubed/pkg/util"
	kutil "github.com/appscode/kutil/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	certificates "k8s.io/client-go/pkg/apis/certificates/v1beta1"
	"k8s.io/client-go/tools/cache"
)

// Blocks caller. Intended to be called as a Go routine.
func (op *Operator) WatchCertificateSigningRequests() {
	if !util.IsPreferredAPIResource(op.KubeClient, certificates.SchemeGroupVersion.String(), "CertificateSigningRequest") {
		log.Warningf("Skipping watching non-preferred GroupVersion:%s Kind:%s", certificates.SchemeGroupVersion.String(), "CertificateSigningRequest")
		return
	}

	defer acrt.HandleCrash()

	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return op.KubeClient.CertificatesV1beta1().CertificateSigningRequests().List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return op.KubeClient.CertificatesV1beta1().CertificateSigningRequests().Watch(metav1.ListOptions{})
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&certificates.CertificateSigningRequest{},
		op.Opt.ResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if res, ok := obj.(*certificates.CertificateSigningRequest); ok {
					log.Infof("CertificateSigningRequest %s@%s added", res.Name, res.Namespace)
					kutil.AssignTypeKind(res)

					messenger := op.Messenger()
					if messenger != nil &&
						messenger.Spec.CSREvents.Handle &&
						util.IsRecent(res.ObjectMeta.CreationTimestamp) {
						err := op.messenger.Forward(res.TypeMeta, res.ObjectMeta, "added", obj)
						if err != nil {
							log.Errorln(err)
						}
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if res, ok := obj.(*certificates.CertificateSigningRequest); ok {
					log.Infof("CertificateSigningRequest %s@%s deleted", res.Name, res.Namespace)
					kutil.AssignTypeKind(res)

					tc := op.TrashCan()
					if tc != nil {
						tc.Delete(res.TypeMeta, res.ObjectMeta, obj)
					}
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldRes, ok := old.(*certificates.CertificateSigningRequest)
				if !ok {
					log.Errorln(errors.New("Invalid CertificateSigningRequest object"))
					return
				}
				newRes, ok := new.(*certificates.CertificateSigningRequest)
				if !ok {
					log.Errorln(errors.New("Invalid CertificateSigningRequest object"))
					return
				}
				kutil.AssignTypeKind(oldRes)
				kutil.AssignTypeKind(newRes)

				messenger := op.Messenger()
				if messenger != nil &&
					messenger.Spec.CSREvents.Handle {
					for _, cond := range newRes.Status.Conditions {
						if util.IsRecent(cond.LastUpdateTime) {
							err := op.messenger.Forward(newRes.TypeMeta, newRes.ObjectMeta, string(cond.Type), newRes)
							if err != nil {
								log.Errorln(err)
							}
						}
					}
				}

				tc := op.TrashCan()
				if tc != nil && tc.Spec.HandleUpdates {
					if !reflect.DeepEqual(oldRes.Labels, newRes.Labels) ||
						!reflect.DeepEqual(oldRes.Annotations, newRes.Annotations) ||
						!reflect.DeepEqual(oldRes.Spec, newRes.Spec) {
						tc.Update(newRes.TypeMeta, newRes.ObjectMeta, old, new)
					}
				}
			},
		},
	)
	ctrl.Run(wait.NeverStop)
}
