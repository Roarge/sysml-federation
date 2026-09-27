@SR-14
Feature: Verdict and reason in the viewer
  As a visitor, I want the viewer to show each requirement's verdict and
  reason as read through the router, so that the visitor sees what the
  analysis concluded and the reason it gives.

  @verdictAndReasonShown @record
  Scenario: Verdict and reason shown
    Given each of the three states of the worked example
    When the viewer is open
    Then every requirement shows its verdict and its reason as read through the router
