@SR-01
Feature: One command
  As a visitor, I want the demo to start the whole demo from one command, with
  no other file, argument, volume or network configuration, so that a curious
  visitor reaches the running demo before anything has a chance to stop them.

  @startsFromOneCommand @record @workflow
  Scenario: Starts from one command
    Given a host with Docker and no authentication to the registry
    When the one command from the README is run
    Then the demo starts with no other file, argument, volume or network configuration

  @recordedOnThreePlatforms @record
  Scenario: Recorded on three platforms
    Given the same run on Linux, macOS and Windows
    When each is recorded in the example README
    Then the record names the hosts used
