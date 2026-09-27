@SR-38
Feature: The documents editable set
  As a document owner, I want the document app to offer the same editable set
  as the viewer, reached from the document, so that a number is corrected from
  where the document owner already works.

  @editableSetIsExact @record
  Scenario: Editable set is exact
    Given the document is open
    When its controls are read
    Then every editable control maps to a derived requirement's subject server throughput or to the limit of the global throughput requirement, and no other value has one
