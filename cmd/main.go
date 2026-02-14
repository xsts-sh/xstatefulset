/*
Copyright The XSTS-SH Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/xsts-sh/xstatefulset/pkg/webhook/cert"

	xstatefulsetclientset "github.com/xsts-sh/xstatefulset/client-go/clientset/versioned"
	"github.com/xsts-sh/xstatefulset/cmd/config"
	"github.com/xsts-sh/xstatefulset/pkg/controller"
	"github.com/xsts-sh/xstatefulset/pkg/controller/xstatefulset"
	"github.com/xsts-sh/xstatefulset/pkg/utils"
	"github.com/xsts-sh/xstatefulset/pkg/webhook"
	"golang.org/x/sync/errgroup"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	xappsv1 "github.com/xsts-sh/xstatefulset/api/apps/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	defaultLeaseDuration = 15 * time.Second
	defaultRenewDeadline = 10 * time.Second
	defaultRetryPeriod   = 2 * time.Second
	leaderElectionId     = "xstatefulset.controller-manager"
	leaseName            = "lease.xstatefulset.controller-manager"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = xappsv1.AddToScheme(scheme)
}

func main() {
	var enableWebhook bool
	var cc config.Config
	var wc config.WebhookConfiguration

	// Initialize klog flags first
	klog.InitFlags(nil)

	// Add controller flags
	pflag.StringVar(&cc.Kubeconfig, "kubeconfig", "", "kubeconfig file path")
	pflag.StringVar(&cc.MasterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	pflag.BoolVar(&cc.EnableLeaderElection, "leader-elect", false, "Enable leader election for controller. "+
		"Enabling this will ensure there is only one active controller. Default is false.")
	pflag.IntVar(&cc.Workers, "workers", 5, "number of workers to run. Default is 5")
	pflag.Float32Var(&cc.KubeAPIQPS, "kube-api-qps", 0, "QPS to use while talking with kubernetes apiserver. If 0, use default value.")
	pflag.IntVar(&cc.KubeAPIBurst, "kube-api-burst", 0, "Burst to use while talking with kubernetes apiserver. If 0, use default value.")

	// Webhook flags
	pflag.BoolVar(&enableWebhook, "enable-webhook", true, "Enable mutating admission webhook for defaulting. Default is true.")
	pflag.IntVar(&wc.Port, "webhook-port", 8443, "Port that the webhook server listens on")
	pflag.StringVar(&wc.ServiceName, "service-name", "xstatefulset-controller-manager-webhook", "Service name for the webhook server")
	pflag.StringVar(&wc.CertDir, "webhook-cert-dir", "/etc/tls", "Directory containing webhook TLS certificates")
	pflag.StringVar(&wc.TlsCert, "tls-cert-file", "/etc/tls/tls.crt", "File containing the x509 Certificate for HTTPS")
	pflag.StringVar(&wc.TlsPrivateKey, "tls-private-key-file", "/etc/tls/tls.key", "File containing the x509 private key to --tls-cert-file")
	pflag.StringVar(&wc.MutatingWebhookConfigurationName, "mutating-webhook-name", "xstatefulset-mutating-webhook", "Name of the mutating webhook configuration")
	pflag.StringVar(&wc.CertSecretName, "webhook-cert-secret", "xstatefulset-webhook-server-cert", "Name of the secret containing webhook certificates")

	// Add go flags (klog) to pflag, but skip flags that are already defined
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	pflag.CommandLine.VisitAll(func(f *pflag.Flag) {
		klog.Infof("Flag: %s, Value: %s", f.Name, f.Value.String())
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		cancel()
	}()

	// Build kubeconfig
	cfg, err := clientcmd.BuildConfigFromFlags(cc.MasterURL, cc.Kubeconfig)
	if err != nil {
		klog.Fatalf("build client config: %v", err)
	}

	// Set QPS and Burst if provided
	if cc.KubeAPIQPS > 0 {
		cfg.QPS = cc.KubeAPIQPS
	}
	if cc.KubeAPIBurst > 0 {
		cfg.Burst = cc.KubeAPIBurst
	}

	g, ctx := errgroup.WithContext(ctx)

	// Setup controller-runtime manager if webhook is enabled
	if enableWebhook {
		g.Go(func() error {
			klog.Info("Starting with webhook enabled using controller-runtime")
			return setupWebhookWithManager(ctx, cfg, cc, wc)
		})
	}

	g.Go(func() error {
		klog.Info("Starting xstatefulset controller manager")
		defer klog.Info("Shutting down xstatefulset controller manager")
		return setupController(ctx, cc)
	})

	if err := g.Wait(); err != nil {
		klog.Fatalf("Error running components: %v", err)
	}
}

func setupController(ctx context.Context, cc config.Config) error {
	cfg, err := clientcmd.BuildConfigFromFlags(cc.MasterURL, cc.Kubeconfig)
	if err != nil {
		return fmt.Errorf("build client config: %v", err)
	}
	// Set QPS and Burst if provided
	if cc.KubeAPIQPS > 0 {
		cfg.QPS = cc.KubeAPIQPS
	}
	if cc.KubeAPIBurst > 0 {
		cfg.Burst = cc.KubeAPIBurst
	}
	kubeClient := kubernetes.NewForConfigOrDie(cfg)

	xStatefulSetClient := xstatefulsetclientset.NewForConfigOrDie(cfg)

	startControllers := func(ctx context.Context) {
		controllerContext := controller.NewControllerContext(ctx, kubeClient, xStatefulSetClient)

		ssc := xstatefulset.NewStatefulSetController(
			ctx,
			controllerContext.KubeInformerFactory.Core().V1().Pods(),
			controllerContext.XStatefulsetInformerFactory.Apps().V1().XStatefulSets(),
			controllerContext.KubeInformerFactory.Core().V1().PersistentVolumeClaims(),
			controllerContext.KubeInformerFactory.Apps().V1().ControllerRevisions(),
			kubeClient,
			xStatefulSetClient)

		// Start the informers
		stopCh := ctx.Done()
		controllerContext.KubeInformerFactory.Start(stopCh)
		controllerContext.XStatefulsetInformerFactory.Start(stopCh)
		close(controllerContext.InformersStarted)

		go ssc.Run(ctx, cc.Workers)
		klog.Info("XStatefulSet controller started")
	}

	if cc.EnableLeaderElection {
		startedLeading := func(ctx context.Context) {
			startControllers(ctx)
			klog.Info("Start as leader")
		}
		leaderElector, err := initLeaderElector(kubeClient, startedLeading)
		if err != nil {
			return fmt.Errorf("init leader elector: %w", err)
		}
		leaderElector.Run(ctx)
	} else {
		startControllers(ctx)
		klog.Info("Started controllers without leader election")
	}
	<-ctx.Done()
	return nil
}

// initLeaderElector inits a leader elector for leader election
func initLeaderElector(kubeClient kubernetes.Interface, startedLeading func(ctx context.Context)) (*leaderelection.LeaderElector, error) {
	resourceLock, err := newResourceLock(kubeClient)
	if err != nil {
		return nil, err
	}
	leaderElector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          resourceLock,
		LeaseDuration: defaultLeaseDuration,
		RenewDeadline: defaultRenewDeadline,
		RetryPeriod:   defaultRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: startedLeading,
			OnStoppedLeading: func() {
				klog.Error("leader election lost")
			},
		},
		ReleaseOnCancel: false,
		Name:            leaderElectionId,
	})
	if err != nil {
		return nil, err
	}
	return leaderElector, nil
}

// newResourceLock returns a lease lock which is used to elect leader
func newResourceLock(client kubernetes.Interface) (*resourcelock.LeaseLock, error) {
	namespace, err := utils.GetInClusterNameSpace()
	if err != nil {
		return nil, err
	}
	// Leader id, should be unique
	id, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	id = id + "_" + string(uuid.NewUUID())
	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      leaseName,
			Namespace: namespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}, nil
}

// setupWebhookWithManager sets up the webhook using controller-runtime manager
func setupWebhookWithManager(ctx context.Context, cfg *rest.Config, cc config.Config, wc config.WebhookConfiguration) error {
	klog.Info("Setting up controller-runtime manager with webhook")

	// Set up controller-runtime logger to use klog
	ctrl.SetLogger(klog.NewKlogr())

	// Create Kubernetes client for certificate management
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Secret -> File -> Generate precedence for CA bundle selection
	namespace := getNamespace()
	var caBundle []byte

	if bundle, err := cert.LoadCertBundleFromSecret(ctx, kubeClient, namespace, wc.CertSecretName); err != nil {
		klog.Warningf("Error reading CA bundle from secret %s: %v", wc.CertSecretName, err)
	} else if bundle != nil {
		klog.Infof("Loaded CA bundle from secret %s", wc.CertSecretName)
		caBundle = bundle.CAPEM
	}

	if caBundle == nil {
		if !fileExists(wc.TlsPrivateKey) || !fileExists(wc.TlsCert) {
			bytes, err := ensureWebhookCertificate(ctx, kubeClient, wc)
			if err != nil {
				return fmt.Errorf("error ensuring webhook certificate: %w", err)
			}
			caBundle = bytes
		}
	}

	if caBundle != nil {
		if cert.UpdateMutatingWebhookCABundle(ctx, kubeClient, wc.MutatingWebhookConfigurationName, caBundle) != nil {
			return fmt.Errorf("Error updating mutating webhook certificate: %v", err)
		}
	}

	// Wait for both cert and key files to exist (in case they are mounted by Kubernetes)
	ok := waitForCertsReady(wc.TlsPrivateKey, wc.TlsCert)
	if !ok {
		return fmt.Errorf("TLS cert/key files not found, webhook server cannot start")
	}

	// Create manager with webhook server
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0", // Disable metrics server as we're not using it yet
		},
		WebhookServer: ctrlwebhook.NewServer(ctrlwebhook.Options{
			Port:    wc.Port,
			CertDir: wc.CertDir,
		}),
		HealthProbeBindAddress: ":9443",
		LeaderElection:         cc.EnableLeaderElection,
		LeaderElectionID:       leaderElectionId,
	})
	if err != nil {
		return fmt.Errorf("unable to create manager: %w", err)
	}

	// Setup webhook
	if err := (&webhook.XStatefulSetDefaulter{}).SetupWebhookWithManager(mgr); err != nil {
		return fmt.Errorf("unable to setup webhook: %w", err)
	}

	// Add health check endpoints
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	klog.Info("Webhook setup complete, starting manager")

	// Start the manager (this blocks)
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("manager error: %w", err)
	}

	return nil
}

// getNamespace returns the current pod namespace or "default".
func getNamespace() string {
	return os.Getenv("POD_NAMESPACE")
}

// fileExists returns true if the file exists.
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// ensureWebhookCertificate generates a certificate into the secret and returns the CA bundle.
func ensureWebhookCertificate(ctx context.Context, kubeClient kubernetes.Interface, wc config.WebhookConfiguration) ([]byte, error) {
	namespace := getNamespace()
	dnsNames := []string{
		fmt.Sprintf("%s.%s.svc", wc.ServiceName, namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", wc.ServiceName, namespace),
	}
	klog.Infof("Auto-generating certificate for webhook server (secret=%s service=%s)", wc.CertSecretName, wc.ServiceName)
	return cert.EnsureCertificate(ctx, kubeClient, namespace, wc.CertSecretName, dnsNames)
}

func waitForCertsReady(keyFile, CertFile string) bool {
	waitTimeout := 30 * time.Second
	waitInterval := 500 * time.Millisecond
	start := time.Now()
	for {
		if fileExists(CertFile) && fileExists(keyFile) {
			return true
		}
		if time.Since(start) > waitTimeout {
			klog.Warningf("timeout waiting for TLS cert/key files to appear at %s and %s", keyFile, CertFile)
			return false
		}
		time.Sleep(waitInterval)
	}
}
