package kubeclient

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubehcl.sh/kubehcl/kube/kubeclient/storage"
)

// Lists all releases of a namespace
// Does that through listing the secrets matching all secrets matching the type
func (cfg *Config) List() ([]string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	client, err := cfg.Client.Factory.KubernetesClientSet()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Couldn't get secrets",
			Detail:   fmt.Sprintf("%s", err),
		})
	}

	if secretList, getSecretErr := client.CoreV1().Secrets(cfg.Settings.Namespace()).List(context.Background(), metav1.ListOptions{FieldSelector: "type=" + storage.SecretType}); apierrors.IsNotFound(getSecretErr) {
		return nil, diags
	} else {
		var secretNames []string
		for _, secret := range secretList.Items {
			secretNames = append(secretNames, secret.Name)
		}
		return secretNames, diags
	}
}
