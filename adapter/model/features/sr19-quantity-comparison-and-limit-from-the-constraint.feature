@SR-19
Feature: Quantity comparison and limit from the constraint
  As a document owner, I want the adapter to read a requirement's constrained
  quantity, comparison and limit out of its constraint, and refuse a
  constraint of any other shape, so that one rule tells a throughput
  requirement from a latency one without either being declared twice.

  @quantityComparisonAndLimitRead
  Scenario Outline: Each requirement of the example carries its quantity, comparison and limit
    Given the example model
    When <requirement> is projected
    Then its quantity is <quantity>, its comparison is <comparison> and its limit is <limit>

    Examples:
      | requirement | quantity | comparison | limit  |
      | PIPE-R1     | capacity | GE         | 1500   |
      | PIPE-R1.1   | capacity | GE         | 1500   |
      | PIPE-R1.2   | capacity | GE         | 1500   |
      | PIPE-R1.3   | capacity | GE         | 750    |
      | PIPE-R1.4   | capacity | GE         | 750    |
      | PIPE-R1.5   | capacity | GE         | 1500   |
      | PIPE-R2     | latency  | LE         | 200 ms |

  @anyOtherShapeIsRefused
  Scenario: A constraint of any other shape is refused
    Given a fixture for each constraint shape the adapter does not read
    When each is read
    Then each is refused at the line and column its fixture gives, as unsupported syntax is
