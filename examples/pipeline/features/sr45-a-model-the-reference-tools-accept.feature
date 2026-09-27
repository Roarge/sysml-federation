@SR-45
Feature: A model the reference tools accept
  As a model owner, I want the example model to be accepted without error by
  the OMG pilot implementation release 2026-07 and by the OpenSysML command
  line, so that a reader who knows the language loads the file in a real tool
  and trusts what the demo says about it.

  @bothToolsAcceptTheExample @validator
  Scenario: Both tools accept the example
    Given the example model
    When the OMG pilot implementation and the OpenSysML command line read it
    Then neither reports an error

  @recordedInTheExampleReadme
  Scenario: The example's verification record names the releases of both tools
    Given the example's README
    When its verification record is read
    Then its rows for the OMG pilot implementation name the release and the kernel
    And its rows for OpenSysML name the version
