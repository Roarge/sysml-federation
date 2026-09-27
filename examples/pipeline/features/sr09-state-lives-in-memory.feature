@SR-09
Feature: State lives in memory
  As a visitor, I want the demo to hold every edited value and every document
  change in memory alone, so that a restart is the whole of the reset.

  The adapter and the document service run in the test process, each from what
  the image ships, and are asked through their own graphs.

  @restartReturnsTheShippedState
  Scenario: After a restart both services show the shipped state
    Given the adapter and the document service as they ship
    And parse's throughput has been set to 1700 and PIPE-R2 excluded from the document
    When both services are started again
    Then parse's throughput is 1200 and PIPE-R2 is numbered 2
    And both versions are 1
