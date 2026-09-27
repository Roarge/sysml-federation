@SR-05
Feature: Two platforms
  As a visitor, I want the release to publish the image for linux/amd64 and
  linux/arm64, so that Apple silicon and ARM servers run the demo as they are.

  @bothPlatformsInTheManifest @workflow
  Scenario: Both platforms in the manifest
    Given a published tag
    When the publish workflow reads the registry manifest
    Then both platforms are listed and the step fails if either is absent
