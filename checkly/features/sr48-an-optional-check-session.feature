@SR-48
Feature: An optional check session
  As a maintainer, I want the check session to run the whole check suite
  against a running instance through a tunnel, deploy it for the session,
  trace every request, and destroy it on stop, only when asked, so that the
  demo's claims are checked whenever it runs with credentials, and the demo is
  untouched otherwise.

  @reachableThroughATunnel @record
  Scenario: Reachable through a tunnel
    Given credentials supplied to the compose profile
    When the session starts
    Then the demo answers on a public hostname through the tunnel

  @suiteRunsAndRecords @checkly
  Scenario: Suite runs and records
    Given the demo is reachable
    When the check suite runs against it
    Then every check runs and the run is recorded as a session

  @deployedForTheDuration @record
  Scenario: Deployed for the duration
    Given the session is running
    When the project is read at the monitoring service
    Then its checks are deployed for as long as the session lasts

  @destroyedOnStop @record
  Scenario: Destroyed on stop
    Given a running session
    When the compose profile stops
    Then the deployed project is destroyed

  @tracesReachTheCollector @record
  Scenario: Traces reach the collector
    Given the session is running
    When a request passes through the router
    Then its trace reaches the collector beside the demo

  @unchangedWithoutCredentials
  Scenario: With no profile chosen, the compose file starts the demo as docker run does
    Given the check session's compose file, with no credential and no profile chosen
    When the services it starts are read
    Then the demo is the only one
    And it runs the published image on port 8080, as the docker run line does
    And the two variables it passes on stay empty until the session sets them
