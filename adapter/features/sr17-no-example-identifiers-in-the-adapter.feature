@SR-17
Feature: No example identifiers in the adapter
  As an adopting organisation, I want the adapter to contain no identifier of
  the example model, so that an adopting reader judges what they would keep by
  reading the adapter alone.

  @noExampleNameInTheSource
  Scenario: The adapter's source names nothing of the example
    Given the adapter's Go and schema files, its tests left out
    When they are searched for the example's names
    Then none is found
    And at least twenty-two files were searched

  @secondFixtureProjectsWithoutThem
  Scenario: The second fixture projects with none of the example's names
    Given the second fixture of SR-16
    When it is projected
    Then it projects five parts, four requirements and one verification case
    And its text carries no name of the example but capacity, which both models declare
