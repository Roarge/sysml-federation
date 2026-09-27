@SR-18
Feature: Refusal of unsupported syntax
  As a model owner, I want the adapter to refuse to start on a construct
  outside the supported subset, naming the file, line and column of the first
  one, so that a model that cannot be served faithfully is never served
  silently wrong.

  @refusalNamesThePlace
  Scenario: The adapter refuses a model outside the subset and names the place
    Given a model file whose third line declares an action, which the subset does not support
    When the adapter starts on it
    Then it refuses to start
    And the refusal names the file, line 3 and column 3

  @oneFixturePerRejection
  Scenario: Each rejection the parser distinguishes is refused at its own place
    Given the parser's rejection fixtures, one for each rejection it distinguishes
    When each fixture is read
    Then each is refused at the line and column its fixture gives, with the message it gives
