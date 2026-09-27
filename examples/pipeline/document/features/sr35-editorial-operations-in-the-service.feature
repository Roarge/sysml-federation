@SR-35
Feature: Editorial operations in the service
  As a document owner, I want the document service to apply an accepted
  operation to the tree, restore an excluded requirement as the last child of
  its former parent, and renumber, so that the service owns the tree and its
  numbering, so the app never has to.

  @eachOperationIsApplied
  Scenario Outline: Each operation is applied to the tree, which is numbered again
    Given the document service at version 1
    When <operation>
    Then the document version is 2
    And <outcome>

    Examples:
      | operation                                                    | outcome                                             |
      | PIPE-R1.5 is moved under PIPE-R1 at position 1               | PIPE-R1.5 is numbered 1.1                           |
      | PIPE-R2 is moved under PIPE-R1 at position 6                 | PIPE-R2 is numbered 1.6                             |
      | a heading "Performance" is inserted above PIPE-R1            | PIPE-R1 is numbered 1.1                             |
      | prose reading "Allocated." is added as PIPE-R1's first child | PIPE-R1's first child is prose reading "Allocated." |
      | the opening prose is rewritten as "Rewritten."               | the opening prose reads "Rewritten."                |
      | PIPE-R1.4 is excluded                                        | PIPE-R1.4 is out of the document                    |

  @restoreReturnsToTheFormerParent
  Scenario: A restored requirement returns as the last child of its former parent
    Given the document service at version 1
    And PIPE-R1.4 has been excluded
    When PIPE-R1.4 is restored
    Then PIPE-R1.4 is PIPE-R1's last child, numbered 1.5
