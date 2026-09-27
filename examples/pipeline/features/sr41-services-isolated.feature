@SR-41
Feature: Services isolated
  As a document owner, I want the demo to keep every service from importing,
  calling or reading the data of another, so that no side channel undoes the
  argument the demo is making.

  @noServiceReachesAnother
  Scenario: No service imports, calls or holds the address of another
    Given the three services behind the router
    When their imports and hand-written sources are read
    Then none imports a package of another
    And none holds an address, is handed one by a flag, reads the environment or opens a connection
