package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/coreos/pkg/flagutil"
	"github.com/golang/glog"

	"github.com/coreos/container-linux-update-operator/pkg/analytics"
	"github.com/coreos/container-linux-update-operator/pkg/k8sutil"
	"github.com/coreos/container-linux-update-operator/pkg/operator"
	"github.com/coreos/container-linux-update-operator/pkg/version"
)

var (
	kubeconfig       = flag.String("kubeconfig", "", "Path to a kubeconfig file. Default to the in-cluster config if not provided.")
	analyticsEnabled = flag.Bool("analytics", true, "Send analytics to Google Analytics")
	printVersion     = flag.Bool("version", false, "Print version and exit")
	// deprecated
	manageAgent    = flag.Bool("manage-agent", false, "Manage the associated update-agent")
	agentImageRepo = flag.String("agent-image-repo", "quay.io/coreos/container-linux-update-operator", "The image to use for the managed agent, without version tag")
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	if err := flagutil.SetFlagsFromEnv(flag.CommandLine, "UPDATE_OPERATOR"); err != nil {
		glog.Fatalf("Failed to parse environment variables: %v", err)
	}
	// respect KUBECONFIG without the prefix as well
	if *kubeconfig == "" {
		*kubeconfig = os.Getenv("KUBECONFIG")
	}

	if *printVersion {
		fmt.Println(version.Format())
		os.Exit(0)
	}

	if *analyticsEnabled {
		analytics.Enable()
	}

	if *manageAgent {
		glog.Warning("Use of -manage-agent=true is deprecated and will be removed in the future")
	}

	// create Kubernetes client (clientset)
	client, err := k8sutil.GetClient(*kubeconfig)
	if err != nil {
		glog.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// update-operator
	o, err := operator.New(operator.Config{
		Client:         client,
		ManageAgent:    *manageAgent,
		AgentImageRepo: *agentImageRepo,
	})
	if err != nil {
		glog.Fatalf("Failed to initialize %s: %v", os.Args[0], err)
	}

	glog.Infof("%s running", os.Args[0])

	analytics.ControllerStarted()

	// Run operator until the stop channel is closed
	stop := make(chan struct{})
	defer close(stop)

	if err := o.Run(stop); err != nil {
		glog.Fatalf("Error while running %s: %v", os.Args[0], err)
	}
}
