@SR-90
Feature: A fixture story
  The scenarios the package's own tests read and run.

  @passes
  Scenario: A step that is defined passes
    Given a defined step

  @undefined
  Scenario: A step with no definition fails
    Given a step nobody defined

  @recordedElsewhere @record
  Scenario: A criterion verified by a recorded run
    Given a defined step

  @eachExample
  Scenario Outline: Each example starts from a fresh world
    Given the number <n> is noted
    Then one number has been noted

    Examples:
      | n |
      | 1 |
      | 2 |
