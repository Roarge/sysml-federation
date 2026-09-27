@SR-26
Feature: Model version events
  As a document owner, I want the adapter to emit the new model version on a
  subscription whenever a mutation is accepted, so that each app learns of a
  change without polling and without a reload.

  The adapter serves the second fixture here, as every test of this package
  does, so that no word of the example appears in the package.

  @subscriberReceivesTheVersion
  Scenario: A subscriber receives each new model version
    Given the adapter serves the second fixture at version 1
    And a subscriber on the adapter's subscription
    When WH-B's rate is set to 35
    Then the subscriber receives model version 2
    When WH-R1's limit is set to 70
    Then the subscriber receives model version 3
