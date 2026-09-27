@SR-02
Feature: Ready within ten seconds
  As a visitor, I want the demo to serve both apps and the router endpoint
  within ten seconds of the container starting, so that the start never
  outlasts the visitor's attention.

  The whole process tree runs inside the test, with a stand-in for the router
  that answers what the UI server passes to it. What is under test is the
  supervisor and the published port, not the federation behind them.

  @readyWithinTenSeconds
  Scenario: Everything answers within ten seconds of the start
    Given a stand-in for the router
    When the demo starts
    Then the viewer answers within ten seconds of the start
    And the three subgraphs, the router and the document all answer
