@SR-34
Feature: Editorial operations in the app
  As a document owner, I want the document app to offer the document owner
  move, nest, insert a heading, add prose, edit text, exclude and restore, so
  that the document reads the way its readers expect, from the app the owner
  already has open.

  @everyOperationIsOffered @record
  Scenario: Every operation is offered
    Given the shipped tree
    When each operation is taken in turn
    Then the app offers all seven of them
