package checkly

import (
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// composePath is the compose file of the check session, beside this test.
const composePath = "compose.yml"

// The parts of the compose file the test reads. The decoder ignores every
// other key, so the structs name only what is checked. The image is kept as
// the text the file carries, so that the default the README names is what is
// compared, not what the environment would substitute.
type composeFile struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Image       string   `yaml:"image"`
	Profiles    []string `yaml:"profiles"`
	Environment []string `yaml:"environment"`
	Ports       []string `yaml:"ports"`
	Volumes     []string `yaml:"volumes"`
}

// sessionProfiles are the profiles a service of the session may carry, one
// each: the shared services under checkly, one tunnel under named and the
// other under quick, and the private-location route under checkly-private.
var sessionProfiles = []string{"checkly", "named", "quick", "checkly-private"}

func loadCompose(t *testing.T) composeFile {
	t.Helper()
	var file composeFile
	data, err := os.ReadFile(composePath)
	raw := assert.Must(t, data, err)
	assert.NoError(t, yaml.Unmarshal(raw, &file))
	return file
}

// hostPort returns the host side of a published port in the short syntax,
// HOST:CONTAINER or IP:HOST:CONTAINER, and the empty string for an entry that
// publishes no host port.
func hostPort(entry string) string {
	fields := strings.Split(entry, ":")
	if len(fields) < 2 {
		return ""
	}
	return fields[len(fields)-2]
}

func publishes(service composeService, port string) bool {
	for _, entry := range service.Ports {
		if hostPort(entry) == port {
			return true
		}
	}
	return false
}

// TestSR48_TheCheckProfileIsOptional: the demo service carries no profile, so
// docker compose up demo runs it alone as the README's docker run line does,
// with the published image by default and the two variables that stay empty
// until the session sets them. Every other service belongs to exactly one
// profile, so nothing of the session starts unless a profile is asked for.
func TestSR48_TheCheckProfileIsOptional(t *testing.T) {
	file := loadCompose(t)
	demo, present := file.Services["demo"]
	assert.True(t, present, "the compose file declares a demo service")

	assert.Len(t, demo.Profiles, 0)
	assert.Equal(t, demo.Image, "${DEMO_IMAGE:-ghcr.io/roarge/sysml-federation}")

	assert.Len(t, demo.Environment, 2)
	for _, prefix := range []string{"SYSML_FEDERATION_ROUTER_CONFIG_PATH=", "LOG_LEVEL="} {
		found := false
		for _, entry := range demo.Environment {
			found = found || strings.HasPrefix(entry, prefix)
		}
		assert.True(t, found, "the demo's environment carries "+prefix)
	}

	assert.Len(t, demo.Volumes, 1)
	for _, entry := range demo.Volumes {
		assert.True(t, strings.HasSuffix(entry, ":/otel/router.yaml:ro"), "the router configuration is mounted read-only: "+entry)
	}

	assert.True(t, publishes(demo, "8080"), "the demo publishes host port 8080")
	for name, service := range file.Services {
		if name == "demo" {
			continue
		}
		assert.True(t, !publishes(service, "8080"), name+" does not publish host port 8080")
		assert.Len(t, service.Profiles, 1)
		for _, profile := range service.Profiles {
			assert.Contains(t, sessionProfiles, profile)
		}
	}

	assert.SliceEqual(t, file.Services["tunnel-named"].Profiles, []string{"named"})
	assert.SliceEqual(t, file.Services["tunnel-quick"].Profiles, []string{"quick"})
}
