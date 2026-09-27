@SR-25
Feature: Rejection of invalid values
  As a visitor, I want the adapter to reject a value that is not a finite,
  non-negative number and leave the previous one in place, so that the first
  nonsense a visitor types does not stop the demo.

  The adapter serves the second fixture here, as every test of this package
  does, so that no word of the example appears in the package.

  @invalidValuesAreRejected
  Scenario Outline: An invalid value is rejected and the previous one stands
    Given the adapter serves the second fixture at version 1
    And WH-A's rate is 40
    When <input> is submitted as WH-A's rate
    Then the mutation is rejected
    And WH-A's rate is still 40
    And the model version is still 1

    Examples:
      | input          |
      | an empty value |
      | the text abc   |
      | -1             |
      | infinity       |
      | not a number   |
