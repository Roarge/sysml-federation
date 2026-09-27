@SR-39
Feature: Live reflection in the other app
  As a document owner, I want the demo to reflect a model or document version
  change in each app within two seconds and without a reload, so that both
  apps are seen to read one graph rather than each other.

  @changeReachesTheOtherApp @record
  Scenario: Change reaches the other app
    Given both apps open in a local browser session
    When the model version or the document version changes
    Then the other app reflects the change within two seconds, timed, and without a page reload
