@SR-10
Feature: No resource from another origin
  As a visitor, I want the apps to reference no font, script or image at any
  origin other than the published port, so that the apps render fully on a
  machine with no route to the internet.

  @noAbsoluteUrlInTheAssets
  Scenario: No embedded file names another origin
    Given the files the demo embeds for both apps
    When they are scanned for absolute and protocol-relative URLs
    Then none is found
    And the shared module, the vendored library and the viewer's page were among the files scanned

  @appsRenderWithTheHostOffline @record
  Scenario: Apps render with the host offline
    Given the host is offline
    When either app is opened
    Then it renders fully
