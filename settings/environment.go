/*
This file was inspired from https://github.com/helm/helm
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: Apache-2.0
Licesne: https://www.apache.org/licenses/LICENSE-2.0
*/
/*
Copyright The Helm Authors.

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

/*
Package cli describes the operating environment for the Helm CLI.

Helm's environment encapsulates all of the service dependencies Helm has.
These dependencies are expressed as interfaces so that alternate implementations
(mocks, etc.) can be easily generated.
*/
package settings

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	"helm.sh/helm/v4/pkg/kube"
)

// defaultMaxHistory sets the maximum number of releases to 0: unlimited
const defaultMaxHistory = 10

// defaultBurstLimit sets the default client-side throttling limit
const defaultBurstLimit = 100

const defaultTimeout = 100

// defaultQPS sets the default QPS value to 0 to use library defaults unless specified
const defaultQPS = float32(0)

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	namespace string
	config    *genericclioptions.ConfigFlags

	// KubeConfig is the path to the kubeconfig file
	KubeConfig string
	// KubeContext is the name of the kubeconfig context.
	KubeContext string
	// Bearer KubeToken used for authentication
	KubeToken string
	// Username to impersonate for the operation
	KubeAsUser string
	// Groups to impersonate for the operation, multiple groups parsed from a comma delimited list
	KubeAsGroups []string
	// Kubernetes API Server Endpoint for authentication
	KubeAPIServer string
	// Custom certificate authority file.
	KubeCaFile string
	// KubeInsecureSkipTLSVerify indicates if server's certificate will not be checked for validity.
	// This makes the HTTPS connections insecure
	KubeInsecureSkipTLSVerify bool
	// KubeTLSServerName overrides the name to use for server certificate validation.
	// If it is not provided, the hostname used to contact the server is used
	KubeTLSServerName string
	// Debug indicates whether or not Helm is running in Debug mode.
	Debug bool

	MaxHistory int
	// BurstLimit is the default client-side throttling limit.
	BurstLimit int

	// Timeout for the operation
	Timeout int
	// QPS is queries per second which may be used to avoid throttling.
	QPS float32
}

func NewSettings() *EnvSettings {
	env := &EnvSettings{
		namespace:                 os.Getenv("KUBEHCL_NAMESPACE"),
		MaxHistory:                envIntOr("KUBEHCL_MAX_HISTORY", defaultMaxHistory),
		KubeContext:               os.Getenv("KUBEHCL_KUBECONTEXT"),
		KubeToken:                 os.Getenv("KUBEHCL_KUBETOKEN"),
		KubeAsUser:                os.Getenv("KUBEHCL_KUBEASUSER"),
		KubeAsGroups:              envCSV("KUBEHCL_KUBEASGROUPS"),
		KubeAPIServer:             os.Getenv("KUBEHCL_KUBEAPISERVER"),
		KubeCaFile:                os.Getenv("KUBEHCL_KUBECAFILE"),
		KubeTLSServerName:         os.Getenv("KUBEHCL_KUBETLS_SERVER_NAME"),
		KubeInsecureSkipTLSVerify: envBoolOr("KUBEHCL_KUBEINSECURE_SKIP_TLS_VERIFY", false),
		BurstLimit:                envIntOr("KUBEHCL_BURST_LIMIT", defaultBurstLimit),
		Timeout:                   envIntOr("KUBEHCL_TIMEOUT", defaultTimeout),
		QPS:                       envFloat32Or("KUBEHCL_QPS", defaultQPS),
	}
	env.Debug, _ = strconv.ParseBool(os.Getenv("KUBEHCL_DEBUG"))

	// bind to kubernetes config flags
	config := &genericclioptions.ConfigFlags{
		Namespace:        &env.namespace,
		Context:          &env.KubeContext,
		BearerToken:      &env.KubeToken,
		APIServer:        &env.KubeAPIServer,
		CAFile:           &env.KubeCaFile,
		KubeConfig:       &env.KubeConfig,
		Impersonate:      &env.KubeAsUser,
		Insecure:         &env.KubeInsecureSkipTLSVerify,
		TLSServerName:    &env.KubeTLSServerName,
		ImpersonateGroup: &env.KubeAsGroups,
		WrapConfigFn: func(config *rest.Config) *rest.Config {
			config.Burst = env.BurstLimit
			config.QPS = env.QPS
			config.Wrap(func(rt http.RoundTripper) http.RoundTripper {
				return &kube.RetryingRoundTripper{Wrapped: rt}
			})
			config.UserAgent = "kubehclv1"
			return config
		},
	}
	if env.BurstLimit != defaultBurstLimit {
		config = config.WithDiscoveryBurst(env.BurstLimit)
	}
	env.config = config

	return env
}

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&s.namespace, "namespace", "n", s.namespace, "namespace scope for this request")
	fs.StringVar(&s.KubeConfig, "kubeconfig", "", "path to the kubeconfig file")
	fs.StringVar(&s.KubeContext, "kube-context", s.KubeContext, "name of the kubeconfig context to use")
	fs.StringVar(&s.KubeToken, "kube-token", s.KubeToken, "bearer token used for authentication")
	fs.StringVar(&s.KubeAsUser, "kube-as-user", s.KubeAsUser, "username to impersonate for the operation")
	fs.StringArrayVar(&s.KubeAsGroups, "kube-as-group", s.KubeAsGroups, "group to impersonate for the operation, this flag can be repeated to specify multiple groups.")
	fs.StringVar(&s.KubeAPIServer, "kube-apiserver", s.KubeAPIServer, "the address and the port for the Kubernetes API server")
	fs.StringVar(&s.KubeCaFile, "kube-ca-file", s.KubeCaFile, "the certificate authority file for the Kubernetes API server connection")
	fs.StringVar(&s.KubeTLSServerName, "kube-tls-server-name", s.KubeTLSServerName, "server name to use for Kubernetes API server certificate validation. If it is not provided, the hostname used to contact the server is used")
	fs.BoolVar(&s.KubeInsecureSkipTLSVerify, "kube-insecure-skip-tls-verify", s.KubeInsecureSkipTLSVerify, "if true, the Kubernetes API server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	fs.BoolVar(&s.Debug, "debug", s.Debug, "enable verbose output")
	fs.IntVar(&s.BurstLimit, "burst-limit", s.BurstLimit, "client-side default throttling limit")
	fs.Float32Var(&s.QPS, "qps", s.QPS, "queries per second used when communicating with the Kubernetes API, not including bursting")
	fs.IntVar(&s.Timeout, "timeout", s.Timeout, "Timeout for resource creation")

}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}

func envBoolOr(name string, def bool) bool {
	if name == "" {
		return def
	}
	envVal := envOr(name, strconv.FormatBool(def))
	ret, err := strconv.ParseBool(envVal)
	if err != nil {
		return def
	}
	return ret
}

func envIntOr(name string, def int) int {
	if name == "" {
		return def
	}
	envVal := envOr(name, strconv.Itoa(def))
	ret, err := strconv.Atoi(envVal)
	if err != nil {
		return def
	}
	return ret
}

func envFloat32Or(name string, def float32) float32 {
	if name == "" {
		return def
	}
	envVal := envOr(name, strconv.FormatFloat(float64(def), 'f', 2, 32))
	ret, err := strconv.ParseFloat(envVal, 32)
	if err != nil {
		return def
	}
	return float32(ret)
}

func envCSV(name string) (ls []string) {
	trimmed := strings.Trim(os.Getenv(name), ", ")
	if trimmed != "" {
		ls = strings.Split(trimmed, ",")
	}
	return
}

func (s *EnvSettings) EnvVars() map[string]string {
	envvars := map[string]string{
		"KUBEHCL_BIN":         os.Args[0],
		"KUBEHCL_DEBUG":       fmt.Sprint(s.Debug),
		"KUBEHCL_NAMESPACE":   s.Namespace(),
		"KUBEHCL_MAX_HISTORY": strconv.Itoa(s.MaxHistory),
		"KUBEHCL_BURST_LIMIT": strconv.Itoa(s.BurstLimit),
		"KUBEHCL_QPS":         strconv.FormatFloat(float64(s.QPS), 'f', 2, 32),

		// broken, these are populated from KUBEHCL flags and not kubeconfig.
		"KUBEHCL_KUBECONTEXT":                  s.KubeContext,
		"KUBEHCL_KUBETOKEN":                    s.KubeToken,
		"KUBEHCL_KUBEASUSER":                   s.KubeAsUser,
		"KUBEHCL_KUBEASGROUPS":                 strings.Join(s.KubeAsGroups, ","),
		"KUBEHCL_KUBEAPISERVER":                s.KubeAPIServer,
		"KUBEHCL_KUBECAFILE":                   s.KubeCaFile,
		"KUBEHCL_KUBEINSECURE_SKIP_TLS_VERIFY": strconv.FormatBool(s.KubeInsecureSkipTLSVerify),
		"KUBEHCL_KUBETLS_SERVER_NAME":          s.KubeTLSServerName,
	}
	if s.KubeConfig != "" {
		envvars["KUBECONFIG"] = s.KubeConfig
	}
	return envvars
}

// Namespace gets the namespace from the configuration
func (s *EnvSettings) Namespace() string {
	if ns, _, err := s.config.ToRawKubeConfigLoader().Namespace(); err == nil {
		return ns
	}
	if s.namespace != "" {
		return s.namespace
	}
	return "default"
}

// SetNamespace sets the namespace in the configuration
func (s *EnvSettings) SetNamespace(namespace string) {
	s.namespace = namespace
}

// RESTClientGetter gets the kubeconfig from EnvSettings
func (s *EnvSettings) RESTClientGetter() genericclioptions.RESTClientGetter {
	return s.config
}
