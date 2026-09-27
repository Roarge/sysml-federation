@SR-11
Feature: Model text with the language visible
  As a visitor, I want the viewer to render the model source with keywords,
  names, numeric literals, strings and comments distinguished, so that the
  visitor sees the notation rather than a form over it.

  @notationIsVisible @record
  Scenario: Notation is visible
    Given the shipped model
    When the viewer opens
    Then the text pane distinguishes keywords, names, numeric literals, strings and comments

  @tokeniserMatchesTheReservedWords @inspection
  Scenario: Tokeniser matches the reserved words
    Given the tokeniser
    When it is read against the language's reserved-word list
    Then the words it distinguishes are those of the list
