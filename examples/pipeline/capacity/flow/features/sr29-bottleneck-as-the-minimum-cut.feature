@SR-29
Feature: Bottleneck as the minimum cut
  As a visitor, I want the capacity service to report the child parts of the
  source-side minimum cut as the bottleneck, so that the reported set is the
  same for every maximum flow, ties included.

  The worked example is the pipeline ingest, parse, the index pair indexA and
  indexB, and serve, wired ingest>parse, parse>indexA, parse>indexB,
  indexA>serve and indexB>serve.

  @cutIsTheBottleneck
  Scenario Outline: The bottleneck of each worked state is the source-side minimum cut
    Given the worked example with throughputs <throughputs>
    When the bottleneck is reported
    Then the capacity is <capacity>, limited by <bottleneck>

    Examples:
      | throughputs                                                 | capacity | bottleneck     |
      | ingest 2000, parse 1200, indexA 700, indexB 700, serve 1800 | 1200     | parse          |
      | ingest 2000, parse 1700, indexA 700, indexB 700, serve 1800 | 1400     | indexA, indexB |
      | ingest 2000, parse 1700, indexA 900, indexB 700, serve 1800 | 1600     | indexA, indexB |

  @aTieHasOneAnswer
  Scenario: A tie between minimum cuts reports the source side, whatever the order
    Given the worked example with throughputs ingest 2000, parse 1600, indexA 900, indexB 700, serve 1800
    When the bottleneck is reported for fifty orderings of its children and connections
    Then every report is 1600, limited by parse
