@SR-47
Feature: The demos model accepted by the reference tools
  As a maintainer, I want the repository to validate the demo's own model with
  both reference tools on every change, so that a form one tool refuses never
  lands on main.

  @bothToolsAccept @validator
  Scenario: Both tools accept
    Given the whole model tree
    When the OMG pilot implementation release 2026-07 and OpenSysML v0.6.0 read it
    Then neither reports an error or a warning

  @runsInContinuousIntegration @workflow
  Scenario: Runs in continuous integration
    Given a pull request or a push to main that changes a file under model/
    When the workflow runs
    Then it puts the model to both tools
