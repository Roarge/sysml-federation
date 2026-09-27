@SR-49
Feature: The demo on the owners host when asked
  As a maintainer, I want the host configuration to start the demo on the
  owner's own host from the published image, public through the named tunnel,
  and stop it again, one command each, so that the demo can answer on its
  public hostname from the owner's own host for as long as the owner wants,
  with no hosted demo and no bill.

  @runsFromThePublishedImage @make-target
  Scenario: Runs from the published image
    Given a Proxmox VE host of version 9.1 or later
    When the configuration is applied with the demo switched on
    Then the host runs the demo from the published image

  @publicThroughTheTunnel @make-target
  Scenario: Public through the tunnel
    Given the demo running on the host
    When the public hostname is requested
    Then the tunnel hands the request to the demo on the host

  @nothingLeftWhenSwitchedOff @make-target
  Scenario: Nothing left when switched off
    Given the demo running on the host
    When the configuration is applied with the demo switched off
    Then nothing it created remains on the host and the route points at the compose session's demo again

  @theTunnelSurvives @inspection
  Scenario: The tunnel survives
    Given the tunnel and its DNS record under the configuration
    When the configuration is destroyed
    Then the destroy is refused and both remain
