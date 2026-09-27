@SR-23
Feature: Expression bound values are evaluated
  As a document owner, I want the adapter to evaluate a value bound by an
  expression over literals, feature references, the four arithmetic operators
  and parentheses, so that an allocated limit follows the limit it is derived
  from, as the model says it does.

  @derivedLimitsFollowTheGlobalLimit
  Scenario: The derived limits follow the global limit
    Given the example model
    When PIPE-R1's limit is set to 2000
    Then the limits are:
      | requirement | limit |
      | PIPE-R1     | 2000  |
      | PIPE-R1.1   | 2000  |
      | PIPE-R1.2   | 2000  |
      | PIPE-R1.3   | 1000  |
      | PIPE-R1.4   | 1000  |
      | PIPE-R1.5   | 2000  |

  @everyOperatorEvaluates
  Scenario Outline: Each operator evaluates to the value its expression gives
    Given a part u whose attribute b is 100, beside a part other whose attribute a is 10 ms
    When u's attribute x is bound to <expression>
    Then x reads <value>

    Examples:
      | expression  | value |
      | 1 + 2       | 3     |
      | 5 - 7       | -2    |
      | 3 * 4       | 12    |
      | 9 / 2       | 4.5   |
      | 1 + 2 * 3   | 7     |
      | (1 + 2) * 3 | 9     |
      | other.a / 2 | 5 ms  |
      | b + 1       | 101   |
      | 2 * other.a | 20 ms |
      | 3[s] + 1    | 4 s   |
