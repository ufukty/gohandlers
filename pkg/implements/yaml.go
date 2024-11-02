package implements

import (
	"fmt"
	"gohandlers/pkg/inspects"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlHandler struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

func YamlFile(dst string, infoss map[inspects.Receiver]map[string]inspects.Info) error {
	hs := map[string]YamlHandler{}
	for _, infos := range infoss {
		for n, h := range infos {
			hs[n] = YamlHandler{
				Method: h.Method,
				Path:   h.Path,
			}
		}
	}
	o, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating yaml file: %w", err)
	}
	defer o.Close()
	e := yaml.NewEncoder(o)
	e.SetIndent(2)
	err = e.Encode(hs)
	if err != nil {
		return fmt.Errorf("writing yaml file: %w", err)
	}
	return nil
}
