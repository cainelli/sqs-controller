package controller

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Option defines the interface to set properties of Controller
type Option interface {
	setOption(opt *Controller)
}

type optionFunc func(c *Controller)

func (f optionFunc) setOption(c *Controller) {
	f(c)
}

// WithKubeClient configures the controller with the given kubernetes client
func WithKubeClient(kubecli client.Client) Option {
	return optionFunc(func(c *Controller) {
		c.Client = kubecli
	})
}

// WithSQSClient configures the controller with the given aw sqs client
func WithSQSClient(sqs *sqs.SQS) Option {
	return optionFunc(func(c *Controller) {
		c.sqscli = sqs
	})
}

// QueueURL configures the controller with the given aw sqs client
func QueueURL(url string) Option {
	return optionFunc(func(c *Controller) {
		c.queueURL = url
	})
}
