@SR-42
Feature: Composition that cannot drift
  As an adopting organisation, I want the repository to embed in the committed
  router configuration subgraph schemas identical to the repository's schema
  files, so that the contract between producer and consumer is checked
  mechanically before deployment rather than discovered in production.

  @driftFailsTheTest
  Scenario: A schema that differs from the committed configuration is caught
    Given the committed configuration and the three schema files
    When each schema file is compared with the schema the configuration embeds for it
    Then all three match
    And a schema changed by one character is reported as drift

  @compositionProducesTheCommittedFile @record
  Scenario: Composition produces the committed file
    Given the maintainer's compose step
    When it is run
    Then its output is the committed configuration
