package controller

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Controller reconciles a RoleBinding objects
type Controller struct {
	client.Client
	queueURL string
	ch       chan event.GenericEvent
	sqscli   *sqs.SQS
}

// New creates a new dev auth controller with options
func New(opts ...Option) (*Controller, error) {
	ctrl := &Controller{}
	for _, f := range opts {
		f.setOption(ctrl)
	}

	if ctrl.queueURL == "" {
		return nil, ErrMissingQueueURL
	}
	return ctrl, nil
}

// Reconcile namespaces of Services that exist in the Organisation upserting Roles required in that namespace.
func (c *Controller) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := logr.FromContextOrDiscard(ctx)

	var ns corev1.Namespace
	err := c.Get(ctx, request.NamespacedName, &ns)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("reconciling")

	return reconcile.Result{}, nil
}

// SetupWithManager adds the controller to the manager
func (c *Controller) SetupWithManager(mgr manager.Manager) error {
	if c.Client == nil {
		c.Client = mgr.GetClient()
	}
	if c.ch == nil {
		c.ch = make(chan event.GenericEvent)
	}
	sourceChan := &source.Channel{Source: c.ch}

	return ctrl.NewControllerManagedBy(mgr).
		Named("sqs").
		For(&corev1.Namespace{}).
		Owns(&rbacv1.Role{}).
		Watches(sourceChan, handler.EnqueueRequestsFromMapFunc(mapToNamespaceRequest)).
		Complete(c)
}
