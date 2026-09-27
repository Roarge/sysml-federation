@SR-27
Feature: Document version events
  As a document owner, I want the document service to emit the new document
  version on a subscription whenever a document change is accepted, so that
  each app learns of a document change without polling and without a reload.

  @subscriberReceivesTheDocumentVersion
  Scenario: A subscriber receives the new document version
    Given the document service at version 1
    And a subscriber on the document service's subscription
    When prose reading "Allocated." is added at the top
    Then the subscriber receives document version 2
