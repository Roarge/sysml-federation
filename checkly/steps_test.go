package checkly

import (
	"regexp"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Roarge/sysml-federation/internal/assert"
)

// steps binds the scenarios of this package to a world of their own, one per
// scenario, reported through the criterion's subtest.
func steps(t *testing.T, sc *godog.ScenarioContext) {
	w := &world{t: t}
	sc.Given(`^the check session's compose file, with no credential and no profile chosen$`, w.theComposeFile)
	sc.When(`^the services it starts are read$`, w.theDefaultServicesAreRead)
	sc.Then(`^the demo is the only one$`, w.theDemoIsTheOnlyOne)
	sc.Then(`^it runs the published image on port 8080, as the docker run line does$`, w.itRunsThePublishedImage)
	sc.Then(`^the two variables it passes on stay empty until the session sets them$`, w.theVariablesStayEmpty)
}

// world is what one scenario of this package knows.
type world struct {
	t        *testing.T
	file     composeFile
	starting []string
}

func (w *world) theComposeFile() { w.file = loadCompose(w.t) }

// theDefaultServicesAreRead takes the services that belong to no profile,
// which are the ones docker compose up starts when none is chosen.
func (w *world) theDefaultServicesAreRead() {
	for name, service := range w.file.Services {
		if len(service.Profiles) == 0 {
			w.starting = append(w.starting, name)
		}
	}
}

func (w *world) theDemoIsTheOnlyOne() { assert.SliceEqual(w.t, w.starting, []string{"demo"}) }

// defaultRE reads the value a compose substitution falls back to:
// ${DEMO_IMAGE:-ghcr.io/roarge/sysml-federation}.
var defaultRE = regexp.MustCompile(`^\$\{\w+:-([^}]*)\}$`)

func (w *world) itRunsThePublishedImage() {
	demo := w.file.Services["demo"]
	image := defaultRE.FindStringSubmatch(demo.Image)
	assert.True(w.t, image != nil, "the image falls back to a default: "+demo.Image)
	if image != nil {
		assert.Equal(w.t, image[1], "ghcr.io/roarge/sysml-federation")
	}
	assert.True(w.t, publishes(demo, "8080"), "the demo publishes host port 8080")
}

// emptyRE matches a variable passed on with an empty default: NAME=${VAR:-}.
var emptyRE = regexp.MustCompile(`^\w+=\$\{\w+:-\}$`)

func (w *world) theVariablesStayEmpty() {
	demo := w.file.Services["demo"]
	assert.Len(w.t, demo.Environment, 2)
	for _, entry := range demo.Environment {
		assert.True(w.t, emptyRE.MatchString(entry), "the variable is empty unless the session sets it: "+entry)
	}
}
