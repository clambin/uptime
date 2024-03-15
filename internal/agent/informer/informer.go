package informer

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"time"
)

type Informer struct {
	cache.SharedInformer
	cache.ResourceEventHandlerRegistration
	ch chan struct{}
}

func New(lw cache.ListerWatcher, interval time.Duration, example runtime.Object, handler cache.ResourceEventHandler) (*Informer, error) {
	i := Informer{
		SharedInformer: cache.NewSharedInformer(lw, example, interval),
		ch:             make(chan struct{}),
	}
	var err error
	i.ResourceEventHandlerRegistration, err = i.SharedInformer.AddEventHandler(handler)
	return &i, err

}

func (i *Informer) Run() {
	i.SharedInformer.Run(i.ch)
}

func (i *Informer) Cancel() {
	i.ch <- struct{}{}
	_ = i.SharedInformer.RemoveEventHandler(i.ResourceEventHandlerRegistration)
}
