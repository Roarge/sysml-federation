@SR-13
Feature: The viewers editable set
  As a visitor, I want the viewer to offer an editable control for each
  server's throughput and for the limit of the global throughput requirement,
  and for nothing else, so that the demo stays explainable and an allocated
  limit cannot be edited out of its allocation.

  @editableSetIsExact @record
  Scenario: Editable set is exact
    Given the viewer is open
    When its controls are read
    Then every editable control maps to a server's throughput or to the limit of the global throughput requirement, and no other value has one
