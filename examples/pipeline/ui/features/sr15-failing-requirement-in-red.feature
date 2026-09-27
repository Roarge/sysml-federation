@SR-15
Feature: Failing requirement in red
  As a visitor, I want the viewer to show a failing requirement's block in red
  and every other block without it, so that the one colour accent in the
  viewer carries the whole message.

  @failingBlockIsRed @record
  Scenario: Failing block is red
    Given a requirement whose verdict is FAIL
    When the viewer shows it
    Then its block is red

  @otherVerdictsCarryNoRed @record
  Scenario: Other verdicts carry no red
    Given a requirement whose verdict is anything else
    When the viewer shows it
    Then its block carries no red
