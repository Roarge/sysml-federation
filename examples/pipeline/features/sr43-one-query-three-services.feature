@SR-43
Feature: One query three services
  As a visitor, I want the router to answer a requirement's text, verdict and
  document number in one response, from a schema all three subgraphs
  contribute to, so that the visitor sees the federation rather than taking it
  on trust.

  @oneResponseCarriesAllThree @record @checkly
  Scenario: One response carries all three
    Given the running stack
    When one query asks for a requirement's text, verdict and document number
    Then one response carries all three

  @schemaCarriesAllThreeSubgraphs
  Scenario: The schema the router serves carries types from all three subgraphs
    Given the composed configuration the router is started with
    When the schema it serves is read from it
    Then it carries types that only the model, only the capacity service and only the document service declare
