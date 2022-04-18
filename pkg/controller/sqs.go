package controller

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	runtimeLog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var ErrMissingSQSClient error = errors.New("missing sqs client")
var ErrMissingQueueURL error = errors.New("missing queue url")

type Message struct {
	Namespace string `json:"namespace"`
}

func (c *Controller) SQSWorker(ctx context.Context) error {
	if c.sqscli == nil {
		return ErrMissingSQSClient
	}
	ticker := time.NewTicker(time.Second)
	log := runtimeLog.Log.WithName("sqs-worker")
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			output, err := c.sqscli.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
				MaxNumberOfMessages: aws.Int64(10),
				WaitTimeSeconds:     aws.Int64(20),
				QueueUrl:            aws.String(c.queueURL),
			})
			if err != nil {
				log.Error(err, "could not receive message")
				continue
			}
			for _, m := range output.Messages {
				var message Message
				if err := json.Unmarshal([]byte(*m.Body), &message); err != nil {
					log.Error(err, "could not serialize message body %q", *m.Body)
					continue
				}
				c.ch <- event.GenericEvent{Object: &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: message.Namespace,
					},
				}}
			}
		}
	}
}

func mapToNamespaceRequest(obj client.Object) []reconcile.Request {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return []reconcile.Request{}
	}
	return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: ns.Name}}}
}
