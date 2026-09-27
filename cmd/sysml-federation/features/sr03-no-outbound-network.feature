@SR-03
Feature: No outbound network
  As a visitor, I want the demo to run with no connection to any address
  outside the container, the router's telemetry switched off, and a
  configuration file handed over only when the operator names one, so that the
  claim that the demo suits an air-gapped machine holds, and telemetry is
  something the operator asks for rather than something they inherit.

  The whole process tree runs inside the test, with a stand-in for the router
  that answers what the UI server passes to it. What is under test is the
  supervisor and the published port, not the federation behind them.

  @environmentCarriesTheFourVariables
  Scenario: The router's environment keeps it off the network
    Given a stand-in for the router
    When the demo starts
    Then the router's environment sets:
      | variable                 | value |
      | DO_NOT_TRACK             | 1     |
      | COSMO_TELEMETRY_DISABLED | true  |
      | TRACING_ENABLED          | false |
      | METRICS_OTLP_ENABLED     | false |
    And every address the composed configuration names is on loopback

  @outboundPathsAnalysed @analysis
  Scenario: Outbound paths analysed
    Given the router's code paths that could reach an address outside the container
    When they are analysed
    Then none is reached with those variables set

  @noConnectionAttemptWithTheRouteWithheld @record
  Scenario: No connection attempt with the route withheld
    Given the container run with the route to the internet withheld and the log level set to debug
    When it starts
    Then it reaches readiness and logs no connection attempt

  @configurationFileIsAnOptIn
  Scenario Outline: A configuration file reaches the router only when one is named
    Given a stand-in for the router
    And the router's configuration file is <named>
    When the demo starts
    Then the router's environment carries the nine variables the supervisor sets
    And its CONFIG_PATH is <config path>

    Examples:
      | named                      | config path       |
      | not named                  | absent            |
      | named as /otel/router.yaml | /otel/router.yaml |
