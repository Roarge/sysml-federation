package main

import (
	"testing"

	"github.com/cucumber/godog"
)

// steps binds the scenarios of this package. None is defined yet, so every
// scenario that runs fails on its first step.
func steps(*testing.T, *godog.ScenarioContext) {}
