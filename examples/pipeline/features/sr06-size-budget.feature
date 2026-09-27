@SR-06
Feature: Size budget
  As a visitor, I want the release to keep each platform's compressed layers
  within 80 MB, so that the pull finishes while the visitor is still watching.

  @layersWithinTheBudget @workflow
  Scenario: Layers within the budget
    Given a published tag
    When the same workflow step reads each platform's layer sizes
    Then the compressed total is at most 80 MB
