// +build experimental

package mustache

import (
	"github.com/cbroglie/mustache"
	"github.com/docker/app/internal/renderer"
	"github.com/pkg/errors"
)

func init() {
	renderer.Register("mustache", &Driver{})
}

// Driver is the mustache implementation of rendered drivers.
type Driver struct{}

// Apply applies the settings to the string
func (d *Driver) Apply(s string, settings map[string]interface{}) (string, error) {
	data, err := mustache.Render(s, settings)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute mustache template")
	}
	return data, nil
}
