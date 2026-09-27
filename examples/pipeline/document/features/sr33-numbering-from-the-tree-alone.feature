@SR-33
Feature: Numbering from the tree alone
  As a document owner, I want the document service to number heading and
  requirement nodes in dotted decimal from the ordered tree alone, skipping
  prose, so that the numbering is the document's own editorial decision and
  owes nothing to the model.

  @shippedTreeNumbersCorrectly
  Scenario: The shipped tree is numbered in dotted decimals, siblings from one
    Given the document service at version 1
    When the document is read
    Then the numbers are 1 PIPE-R1, 1.1 PIPE-R1.1, 1.2 PIPE-R1.2, 1.3 PIPE-R1.3, 1.4 PIPE-R1.4, 1.5 PIPE-R1.5, 2 PIPE-R2

  @proseIsSkippedAndUnnumbered
  Scenario: Prose is skipped by the numbering and carries no number
    Given the document service at version 1
    When prose reading "Allocated." is added as PIPE-R1's first child
    Then the document's two prose nodes carry no number
    And the numbers are 1 PIPE-R1, 1.1 PIPE-R1.1, 1.2 PIPE-R1.2, 1.3 PIPE-R1.3, 1.4 PIPE-R1.4, 1.5 PIPE-R1.5, 2 PIPE-R2

  @numberingFollowsEachEdit
  Scenario Outline: After each kind of edit the numbers follow the tree alone
    Given the document service at version 1
    When <edit>
    Then the numbers are <numbers>

    Examples:
      | edit                                                         | numbers                                                                                                           |
      | PIPE-R1.5 is moved under PIPE-R1 at position 1               | 1 PIPE-R1, 1.1 PIPE-R1.5, 1.2 PIPE-R1.1, 1.3 PIPE-R1.2, 1.4 PIPE-R1.3, 1.5 PIPE-R1.4, 2 PIPE-R2                   |
      | PIPE-R2 is moved under PIPE-R1 at position 6                 | 1 PIPE-R1, 1.1 PIPE-R1.1, 1.2 PIPE-R1.2, 1.3 PIPE-R1.3, 1.4 PIPE-R1.4, 1.5 PIPE-R1.5, 1.6 PIPE-R2                 |
      | a heading "Performance" is inserted above PIPE-R1            | 1 n1, 1.1 PIPE-R1, 1.1.1 PIPE-R1.1, 1.1.2 PIPE-R1.2, 1.1.3 PIPE-R1.3, 1.1.4 PIPE-R1.4, 1.1.5 PIPE-R1.5, 2 PIPE-R2 |
      | prose reading "Allocated." is added as PIPE-R1's first child | 1 PIPE-R1, 1.1 PIPE-R1.1, 1.2 PIPE-R1.2, 1.3 PIPE-R1.3, 1.4 PIPE-R1.4, 1.5 PIPE-R1.5, 2 PIPE-R2                   |
      | PIPE-R1.4 is excluded                                        | 1 PIPE-R1, 1.1 PIPE-R1.1, 1.2 PIPE-R1.2, 1.3 PIPE-R1.3, 1.4 PIPE-R1.5, 2 PIPE-R2                                  |
