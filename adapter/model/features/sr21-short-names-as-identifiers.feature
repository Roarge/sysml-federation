@SR-21
Feature: Short names as identifiers
  As a visitor, I want the adapter to identify every projected element by its
  declared short name, or by its qualified name where none is declared, so
  that the three services join on a stable identifier the model's author
  controls.

  @shortNameIsTheIdentifier
  Scenario: Every element of the example is identified by its short name
    Given the example model
    When it is projected
    Then its parts are identified as PIPE-P1, PIPE-S1, PIPE-S2, PIPE-S3, PIPE-S4 and PIPE-S5
    And its requirements as PIPE-R1, PIPE-R1.1, PIPE-R1.2, PIPE-R1.3, PIPE-R1.4, PIPE-R1.5 and PIPE-R2
    And its verification case as PIPE-VC1

  @qualifiedNameWhereNoneIsDeclared
  Scenario: An element with no short name is identified by its qualified name
    Given a package Plant whose part line owns a part a with the short name A and a part b with none
    When it is projected
    Then line is identified as Plant::line
    And a is identified as A
    And b is identified as Plant::line::b
