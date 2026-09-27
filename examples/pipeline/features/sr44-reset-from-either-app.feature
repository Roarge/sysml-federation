@SR-44
Feature: Reset from either app
  As a visitor, I want the demo to return both apps to the shipped model
  values and the shipped document structure on a request from either, so that
  the demo is shown again from the beginning.

  The adapter and the document service run in the test process, each from what
  the image ships, and are asked through their own graphs.

  @resetMutationsReturnTheShippedState
  Scenario: The two reset mutations return both services to their shipped state
    Given the adapter and the document service as they ship
    And parse's throughput has been set to 1700 and PIPE-R2 excluded from the document
    When the adapter's resetModel and the document service's resetDocument are called
    Then parse's throughput is 1200 and PIPE-R2 is numbered 2
    And both versions are 3, since a reset moves a version on and never back

  @resetFromEitherApp @record
  Scenario: Reset from either app
    Given both apps
    When the reset control is used in either
    Then both show the shipped model values and the shipped document structure
