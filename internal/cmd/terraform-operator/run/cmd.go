// Copyright 2020 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package run

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	zapf "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	tfconfig "github.com/operator-framework/operator-sdk/internal/terraform/config"
	"github.com/operator-framework/operator-sdk/internal/terraform/controller"
	"github.com/operator-framework/operator-sdk/internal/terraform/flags"
	"github.com/operator-framework/operator-sdk/internal/terraform/watches"
	"github.com/operator-framework/operator-sdk/internal/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/internal/version"
)

var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info("Version",
		"Go Version", runtime.Version(),
		"GOOS", runtime.GOOS,
		"GOARCH", runtime.GOARCH,
		"terraform-operator", sdkVersion.Version,
		"commit", sdkVersion.GitCommit)
}

func NewCmd() *cobra.Command {
	f := &flags.Flags{}
	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := &zapf.Options{}
	opts.BindFlags(zapfs)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the operator",
		Run: func(cmd *cobra.Command, _ []string) {
			logf.SetLogger(zapf.New(zapf.UseFlagOptions(opts)))
			run(cmd, f)
		},
	}

	f.AddTo(cmd.Flags())
	cmd.Flags().AddGoFlagSet(zapfs)
	return cmd
}

func run(cmd *cobra.Command, f *flags.Flags) {
	printVersion()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "Failed to get config.")
		os.Exit(1)
	}

	// Set default manager options
	options := manager.Options{
		MetricsBindAddress:      f.MetricsAddress,
		LeaderElection:          f.EnableLeaderElection,
		LeaderElectionID:        f.LeaderElectionID,
		LeaderElectionNamespace: f.LeaderElectionNamespace,
		ClientBuilder:           manager.NewClientBuilder(),
	}

	namespace, found := os.LookupEnv(k8sutil.WatchNamespaceEnvVar)
	log = log.WithValues("Namespace", namespace)
	if found {
		if namespace == metav1.NamespaceAll {
			log.Info("Watching all namespaces.")
			options.Namespace = metav1.NamespaceAll
		} else {
			if strings.Contains(namespace, ",") {
				log.Info("Watching multiple namespaces.")
				options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(namespace, ","))
			} else {
				log.Info("Watching single namespace.")
				options.Namespace = namespace
			}
		}
	} else {
		log.Info(fmt.Sprintf("%v environment variable not set. Watching all namespaces.",
			k8sutil.WatchNamespaceEnvVar))
		options.Namespace = metav1.NamespaceAll
	}

	mgr, err := manager.New(cfg, options)
	if err != nil {
		log.Error(err, "Failed to create a new manager.")
		os.Exit(1)
	}

	ws, err := watches.Load(f.WatchesFile)
	if err != nil {
		log.Error(err, "Failed to create new manager factories.")
		os.Exit(1)
	}
	for _, w := range ws {
		// Register the controller with the factory.
		err := controller.Add(mgr, controller.Options{
			Namespace:               namespace,
			GVK:                     w.GroupVersionKind,
			ManagerFactory:          tfconfig.NewManagerFactory(mgr, w.TemplateDir),
			ReconcilePeriod:         f.ReconcilePeriod,
			WatchDependentResources: *w.WatchDependentResources,
			OverrideValues:          w.OverrideValues,
			MaxConcurrentReconciles: f.MaxConcurrentReconciles,
		})
		if err != nil {
			log.Error(err, "Failed to add manager factory to controller.")
			os.Exit(1)
		}
	}

	// Start the Cmd
	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero.")
		os.Exit(1)
	}
}
