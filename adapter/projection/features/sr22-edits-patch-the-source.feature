@SR-22
Feature: Edits patch the source
  As a visitor, I want the adapter to replace an edited literal at its source
  span, rebuild the projection and increment the version in one step, so that
  the text the viewer shows and the projection the document reads never
  disagree.

  The adapter serves the second fixture here, as every test of this package
  does, so that no word of the example appears in the package.

  @literalAndProjectionAgree
  Scenario: A mutation patches the literal and the projection together
    Given the adapter serves the second fixture at version 1
    When WH-A's rate is set to 45
    Then the served text carries "attribute :>> rate = 45;" and no longer "rate = 40;"
    And WH-A's rate is 45
    And the model version is 2

  @agreeUnderConcurrentReads
  Scenario: Reads during patches see the text and the projection agreeing
    Given the adapter serves the second fixture at version 1
    And four writers setting WH-R1's limit over and over
    When fifty reads each take the served text and WH-R1's limit together
    Then every read finds the limit it was given in the text it was given
    And the writers' patches landed while the reads ran
