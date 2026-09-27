@SR-91 @extra
Feature: What the agreement cannot read

  Scenario: No criterion
    Given a defined step

  @one @two
  Scenario: Two criteria
    Given a defined step

  @three
  Scenario Outline: Tagged examples
    Given the number <n> is noted

    @tagged
    Examples:
      | n |
      | 1 |

  Rule: A rule

    @four
    Scenario: Under a rule
      Given a defined step
