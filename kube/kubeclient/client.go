/*
This file was inspired from https://github.com/helm/helm
This file has been modified from the original version
Changes made to fit kubehcl purposes
This file retains its' original license
// SPDX-License-Identifier: Apache-2.0
Licesne: https://www.apache.org/licenses/LICENSE-2.0
*/
package kubeclient

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"helm.sh/helm/v4/pkg/kube"

	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubehcl.sh/kubehcl/kube/kubeclient/storage"
	"kubehcl.sh/kubehcl/settings"
)

type Config struct {
	Settings *settings.EnvSettings
	Client   *kube.Client
	Storage  storage.Storage
	Name     string
	Timeout  time.Duration
	// WaitStrategy kube.WaitStrategy
	Version string
}

// Applies the settings and creates a config to create,destroy and  validate all configuration files
func New(name string, conf *settings.EnvSettings) (*Config, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	cfg := &Config{}
	cfg.Settings = conf
	cfg.Client = kube.New(cfg.Settings.RESTClientGetter())
	cfg.Client.SetWaiter(kube.StatusWatcherStrategy)

	cfg.Storage = storage.New(cfg.Client, name, conf.Namespace())
	cfg.Name = name
	if conf.Timeout < 0 {
		cfg.Timeout = 100 * time.Second
	} else {
		cfg.Timeout = time.Duration(conf.Timeout) * time.Second
	}
	diags = append(diags, cfg.IsReachable()...)
	if !diags.HasErrors() {
		client, err := cfg.Client.Factory.KubernetesClientSet()
		if err != nil {
			panic("Couldn't get client")
		}
		version, err := client.ServerVersion()
		if err != nil {
			panic("Couldn't get version")
		}
		cfg.Version = version.Major + "." + version.Minor

	}

	return cfg, diags
}

func (cfg *Config) validateNamespace() hcl.Diagnostics {
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		panic("Couldn't get client")
	}

	var diags hcl.Diagnostics
	if _, err := client.CoreV1().Namespaces().Get(context.Background(), cfg.Settings.Namespace(), metav1.GetOptions{}); apierrors.IsNotFound(err) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Namespace \"%s\" does not exist", cfg.Settings.Namespace()),
			Detail:   fmt.Sprintf("%s, in order to create it add --create-namespace", err.Error()),
		})
	}
	return diags
}

// Build resource build the resource from cty.value type into a json
func (cfg *Config) buildResource(key string, value cty.Value, rg *hcl.Range) (kube.ResourceList, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	data, err := ctyjson.Marshal(value, value.Type())

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't convert resource config to json",
			Detail:   fmt.Sprintf("%s", err),
			Subject:  rg,
		})
	}

	cfg.Storage.Add(key, data)
	reader := bytes.NewReader(data)
	kubeResourceList, buildErr := cfg.Client.Build(reader, true)
	if buildErr != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't build resource",
			Detail:   fmt.Sprintf("%s", buildErr),
			Subject:  rg,
		})
	}
	return kubeResourceList, diags
}

// Format for diagnostic error
func formatErr(err error) string {
	errStr := strings.Join(strings.Split(err.Error(), ", "), "\n")
	errStr = strings.ReplaceAll(errStr, "[", "")
	return strings.ReplaceAll(errStr, "]", "")
}

/*
Checks if the client is reachable
*/
func (cfg *Config) IsReachable() hcl.Diagnostics {
	var diags hcl.Diagnostics
	if err := cfg.Client.IsReachable(); err != nil {
		return append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Client is not reachable",
			Detail:   fmt.Sprintf("Error: %s", err.Error()),
		})
	}
	return diags
}
