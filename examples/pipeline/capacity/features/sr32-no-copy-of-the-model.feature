@SR-32
Feature: No copy of the model
  As a document owner, I want the capacity service to hold no model data
  between requests, so that the view is live rather than exported.

  @restartChangesNothing
  Scenario: The same request before and after a restart gets the same answer
    Given the capacity service, asked for the worked example with parse at 1700
    When the service is started again and asked the same
    Then the two answers are the same, byte for byte

  @theServiceHasNoStore
  Scenario: The capacity service's source declares no store
    Given the capacity service's hand-written source, the generated files left out
    When it is read
    Then its resolver holds the configured names and nothing else
    And no hand-written file of the service declares a variable at package level
    And the flow package declares only its errors at package level
