package openshiftinstall

import (
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/version"
)

var (
	// ConfigPath is the relative path of openshift-install within the asset
	// directory.
	ConfigPath = filepath.Join("openshift", "openshift-install.yaml")
)

// Config generates the openshift-install ConfigMap.
type Config struct {
	File *asset.File
}

var _ asset.WritableAsset = (*Config)(nil)

// Name returns a human friendly name for the asset.
func (*Config) Name() string {
	return "OpenShift Install"
}

// Dependencies returns all of the dependencies directly needed to generate
// the asset.
func (*Config) Dependencies() []asset.Asset {
	return []asset.Asset{}
}

// Generate generates the openshift-install ConfigMap.
func (i *Config) Generate(dependencies asset.Parents) error {
	cm, err := CreateInstallConfig("")
	if err != nil {
		return err
	}

	if cm != "" {
		i.File = &asset.File{
			Filename: ConfigPath,
			Data:     []byte(cm),
		}
	}

	return nil
}

// Files returns the files generated by the asset.
func (i *Config) Files() []*asset.File {
	if i.File != nil {
		return []*asset.File{i.File}
	}
	return []*asset.File{}
}

// Load loads the already-rendered files back from disk.
func (i *Config) Load(f asset.FileFetcher) (bool, error) {
	file, err := f.FetchByName(ConfigPath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	i.File = file
	return true, nil
}

// CreateInstallConfig creates the openshift-install ConfigMap from the
// OPENSHIFT_INSTALL_INVOKER environment variable, and if not present, from the
// provided default invoker. If both the environment variable and the default
// are the empty string, this returns an empty string (indicting that no
// ConfigMap should be created. This returns an error if the marshalling to
// YAML fails.
func CreateInstallConfig(defaultInvoker string) (string, error) {
	var invoker string
	if env := os.Getenv("OPENSHIFT_INSTALL_INVOKER"); env != "" {
		invoker = env
	} else if defaultInvoker != "" {
		invoker = defaultInvoker
	} else {
		return "", nil
	}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "openshift-config",
			Name:      "openshift-install",
		},
		Data: map[string]string{
			"version": version.Raw,
			"invoker": invoker,
		},
	}

	cmData, err := yaml.Marshal(cm)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create install-config ConfigMap")
	}

	return string(cmData), nil
}
