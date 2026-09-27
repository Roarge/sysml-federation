@SR-20
Feature: Connection direction
  As a visitor, I want the adapter to take a connection's direction from the
  order of its ends, and refuse a wiring whose ports face the wrong way, so
  that a wiring error is a model error rather than a capacity that is quietly
  wrong.

  @directionFollowsTheEnds
  Scenario: Each connection of the example runs from its first end to its second
    Given the example model
    When it is projected
    Then its connections are:
      | from    | out    | to      | in    |
      | PIPE-S1 | output | PIPE-S2 | input |
      | PIPE-S2 | output | PIPE-S3 | input |
      | PIPE-S2 | output | PIPE-S4 | input |
      | PIPE-S3 | output | PIPE-S5 | input |
      | PIPE-S4 | output | PIPE-S5 | input |

  @reversedPortsAreRefused
  Scenario Outline: A connection whose ends face the wrong way is refused
    Given parts a and b, each with an in port i, an out port o and an inout port b
    When a connection is declared from <first> to <second>
    Then the adapter refuses it: <refusal>

    Examples:
      | first | second | refusal                            |
      | b.i   | a.o    | first end "b.i" is not an out port |
      | a.o   | b.o    | second end "b.o" is not an in port |
      | a.b   | b.i    | first end "a.b" is not an out port |
