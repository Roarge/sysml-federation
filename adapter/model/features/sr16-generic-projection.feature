@SR-16
Feature: Generic projection
  As a visitor, I want the adapter to project every part, requirement,
  relationship and verification case of a model in the supported subset,
  together with the model's text and version, so that every other service
  reads what it needs from one projection and never from the source.

  The second fixture is the model package's warehouse, whose names, wiring and
  constraints differ from the example's on purpose.

  @everyFieldProjected
  Scenario: The second fixture projects every field
    Given the second fixture, which shares no name with the example
    When it is projected
    Then its parts are:
      | id    | name  | definition | within | attributes                      | ports             |
      | WH-L1 | line  | Line       |        | capacity, cycleTime, shifts = 2 |                   |
      | WH-A  | pick  | Robot      | WH-L1  | capacity, rate = 40             | take in, give out |
      | WH-B  | pack  | Robot      | WH-L1  | capacity, rate = 30             | take in, give out |
      | WH-C  | label | Robot      | WH-L1  | capacity, rate = 25             | take in, give out |
      | WH-D  | ship  | Robot      | WH-L1  | capacity, rate = 50             | take in, give out |
    And its connections are:
      | from | out  | to   | in   |
      | WH-A | give | WH-B | take |
      | WH-B | give | WH-D | take |
      | WH-C | give | WH-D | take |
    And its requirements are:
      | id                  | subject | quantity  | comparison | limit | editable | derived from | satisfied by | verified by |
      | WH-R1               | WH-L1   | capacity  | GE         | 60    | yes      |              | WH-L1        |             |
      | WH-R1.1             | WH-D    | capacity  | GE         | 40    | no       | WH-R1        | WH-D         |             |
      | Warehouse::packRate | WH-B    | capacity  | GE         | 30    | no       | WH-R1        | WH-B         |             |
      | WH-R2               | WH-L1   | cycleTime | LE         | 30 s  | yes      |              | WH-L1        | WH-VC1      |
    And WH-R1 reads "The line shall handle the required parcel rate"
    And its verification case WH-VC1, named cycleTest, verifies WH-R2
