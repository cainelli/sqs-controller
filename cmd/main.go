package main

import (
	"flag"
	"os"

	sqscontroller "sqs-controller/pkg/controller"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	ctx := signals.SetupSignalHandler()

	var queueURL = flag.String("queue-url", "", "name of SQS queue to consume")
	flag.Parse()

	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		NameKey:    "name",
		MessageKey: "msg",
		TimeKey:    "timestamp",
		EncodeTime: zapcore.ISO8601TimeEncoder,
	})
	logger := zap.New(
		zap.Encoder(jsonEncoder),
		zap.Level(zapcore.InfoLevel),
	)

	log.SetLogger(logger)
	entryLog := log.Log.WithName("entrypoint")

	k8sConfig, err := config.GetConfig()
	if err != nil {
		entryLog.Error(err, "could not initialize kubeconfig")
		os.Exit(1)
	}

	mgr, err := manager.New(k8sConfig, manager.Options{})
	if err != nil {
		entryLog.Error(err, "could not start a new manager")
		os.Exit(1)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	sqscli := sqs.New(sess)
	sqsctrl, err := sqscontroller.New(
		sqscontroller.WithSQSClient(sqscli),
		sqscontroller.QueueURL(*queueURL),
	)
	if err != nil {
		entryLog.Error(err, "could not initialize sqs controlelr")
		os.Exit(1)
	}

	if err := sqsctrl.SetupWithManager(mgr); err != nil {
		os.Exit(1)
	}
	go sqsctrl.SQSWorker(ctx) // watch sqs queue in the background

	if err := mgr.Start(ctx); err != nil {
		entryLog.Error(err, "could not start manager")
		os.Exit(1)
	}
}
