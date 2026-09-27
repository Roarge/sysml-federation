@SR-40
Feature: Every request through the router
  As a visitor, I want the apps to send every query, mutation and subscription
  to the router, so that the router stays the only place where the three
  services meet.

  @everyCallGoesToTheRouter @inspection
  Scenario: Every call goes to the router
    Given the apps
    When their network calls are inspected
    Then every query, mutation and subscription is addressed to the router

  @appsFailWithoutTheRouter
  Scenario: With no router, the apps' one endpoint fails
    Given no router answers behind the UI server
    When each app's page and the GraphQL endpoint are requested
    Then both pages are served
    And every request to /graphql fails with 502 Bad Gateway
    And neither app's files name the address of a subgraph
