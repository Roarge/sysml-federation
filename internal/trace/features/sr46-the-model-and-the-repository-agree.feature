@SR-46
Feature: The model and the repository agree
  As a maintainer, I want the repository to keep the model and the repository
  in agreement by a test that fails on the first identifier, test name,
  scenario, check file or image that one names and the other lacks, and on a
  criterion without its one scenario, so that the record cannot drift from
  what ships without a check failing.

  Each scenario runs one check of the agreement test against this repository,
  and the check fails the scenario on the first name one side carries and the
  other lacks.

  @identifiers
  Scenario: The identifiers and the decision records agree
    Given the model and the decisions folder
    When the agreement test reads them
    Then every identifier the model writes is declared exactly once in it
    And every decision record the model names exists under the decisions folder, and every record there is named in the model

  @requirementsAffected
  Scenario: Every requirement a decision record names is declared
    Given the decision records
    When the agreement test reads the requirements each names as affected
    Then every one is a short name the model declares

  @goTests
  Scenario: The Go tests and the register agree
    Given the Go tests of the requirement scheme and the verification register
    When the agreement test compares them
    Then every Go test the model names exists in the file it names, and every Go test of the scheme is named by the model

  @checkFiles
  Scenario: The check files and the case registers agree
    Given the check project and the case registers
    When the agreement test compares them
    Then every check file a case names exists in the check project, and every check file of that project is named by a case

  @images
  Scenario: The published images and the views agree
    Given the published images and the views
    When the agreement test compares them
    Then every published image is named by exactly one view, and every image a view names is published

  @coverage
  Scenario: Every story that is done is verified, satisfied and derived
    Given the stories, the cases and the logical architecture
    When the agreement test reads them
    Then every story that is done is verified by a case, satisfied by something in the logical architecture, and a derived end of a derivation connection

  @checkInventory
  Scenario: The check register and the manifest agree
    Given the check register and the manifest the check project reads
    When the agreement test compares them
    Then both carry the same checks with the same fields

  @sessionParts
  Scenario: The session's services and its composite agree
    Given the compose file of the check session and the session composite
    When the agreement test compares them
    Then the services the compose file starts are the parts the composite carries, and nothing else

  @scenarios
  Scenario: The feature files, the stories and the register agree
    Given the feature files, the system stories and the verification register
    When the agreement test reads them
    Then every feature file is named after the one system story its tag names, and every scenario names one acceptance criterion of that story
    And every criterion of a story that is done has exactly one scenario
    And every kind of evidence a scenario names in place of running is offered by its story's case
    And every scenario go test runs is named by that case and has a runner in its package
