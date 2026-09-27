@SR-36
Feature: Document operations leave the model unchanged
  As a document owner, I want the document service to leave the model version
  untouched by every editorial operation, so that shaping the document cannot
  reach the model.

  @documentReachesNoAdapterPackage
  Scenario: The document service reaches no adapter package
    Given the document service's sources, its tree package among them
    When their imports are followed
    Then none leads into the adapter

  @versionUnchangedThroughTheStack @record
  Scenario: Version unchanged through the stack
    Given the running stack where one store serves the adapter
    When each document operation is driven through the composed graph
    Then the model version read from the adapter's own endpoint is the same before and after
