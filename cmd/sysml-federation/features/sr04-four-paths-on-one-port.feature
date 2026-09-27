@SR-04
Feature: Four paths on one port
  As a visitor, I want the demo to serve both apps, the GraphQL endpoint and
  the playground on one published port, with the root redirecting to the
  viewer, so that one origin keeps the launch line to a single flag and the
  browser clear of cross-origin rules.

  The whole process tree runs inside the test, with a stand-in for the router
  that answers what the UI server passes to it. What is under test is the
  supervisor and the published port, not the federation behind them.

  @fourPathsAnswer
  Scenario: The four paths answer on the published port
    Given the demo is running with a stand-in for the router
    When the four paths are requested on the published port
    Then each is served by the app or the endpoint it names:
      | path        | served by                 |
      | /viewer/    | the Model viewer page     |
      | /document/  | the Requirements document |
      | /graphql    | the router                |
      | /playground | the router                |

  @rootRedirectsToTheViewer
  Scenario: The root redirects to the viewer
    Given the demo is running with a stand-in for the router
    When the root path is requested
    Then it redirects to /viewer/
