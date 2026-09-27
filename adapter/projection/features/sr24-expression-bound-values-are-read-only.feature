@SR-24
Feature: Expression bound values are read only
  As a model owner, I want the adapter to reject a mutation that targets a
  value bound by an expression, so that an allocation cannot be broken by an
  edit.

  The adapter serves the second fixture here, as every test of this package
  does, so that no word of the example appears in the package.

  @boundValueIsNotEdited
  Scenario: A limit bound by an expression is refused and keeps its value
    Given the adapter serves the second fixture at version 1
    And WH-R1.1's limit is bound by an expression
    When WH-R1.1's limit is set to 1
    Then the mutation is refused as not editable
    And WH-R1.1's limit is still 40
    And the model version is still 1
