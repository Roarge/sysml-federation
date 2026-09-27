@SR-07
Feature: Licence carried in the image
  As a maintainer, I want the image to carry the router's Apache-2.0 text
  beside the binary copied into the image, so that the image is published and
  copied without a licence obligation being missed.

  @licenceBesideTheBinary @inspection
  Scenario: Licence beside the binary
    Given the published image
    When its contents are inspected
    Then the router's Apache-2.0 text sits beside the copied binary
