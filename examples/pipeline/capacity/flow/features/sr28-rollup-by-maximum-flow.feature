@SR-28
Feature: Rollup by maximum flow
  As a visitor, I want the capacity service to compute a part's capacity as
  the maximum flow through its children's wiring, and a childless part's as
  its own configured attribute, so that a chain and its parallel branches are
  answered by one rule that stays defined where the simple rules are not.

  A wiring is written as its children with their throughputs, and its
  connections as from>to. A child whose throughput is ? has none.

  @tableCasesForEachWiring
  Scenario Outline: The capacity is the maximum flow through each kind of wiring
    Given the children <children>
    And the connections <connections>
    When the capacity is rolled up
    Then the rollup gives <result>

    Examples: a chain, fan-out and fan-in, nested branches, a cycle and a disconnected child
      | children                                                             | connections                                                          | result                          |
      | a 2000, b 1200, c 1800                                               | a>b, b>c                                                             | 1200, limited by b              |
      | ingest 2000, parse 1200, indexA 700, indexB 700, serve 1800          | ingest>parse, parse>indexA, parse>indexB, indexA>serve, indexB>serve | 1200, limited by parse          |
      | ingest 2000, parse 1700, indexA 700, indexB 700, serve 1800          | ingest>parse, parse>indexA, parse>indexB, indexA>serve, indexB>serve | 1400, limited by indexA, indexB |
      | a 5000, b 600, c 500, d 300, e 400, f 900, g 2000                    | a>b, a>c, b>d, b>e, c>f, d>g, e>g, f>g                               | 1100, limited by b, c           |
      | a 100, b 50, c 80, d 200                                             | a>b, b>c, c>b, c>d                                                   | 50, limited by b                |
      | ingest 2000, parse 1200, indexA 700, indexB 700, serve 1800, spare 1 | ingest>parse, parse>indexA, parse>indexB, indexA>serve, indexB>serve | 1200, limited by parse          |

    Examples: zero and missing values
      | children  | connections | result                        |
      | a 0, b 5  | a>b         | 0, limited by a               |
      | a ?, b 1  | a>b         | refused, a has no throughput  |
      | a -1, b 1 | a>b         | refused, a has a negative one |

  @aLeafReadsItsOwnAttribute
  Scenario: A part with no children reads its own attribute
    Given a part ingest with no children, whose throughput is 2000
    When the capacity is analysed
    Then the capacity is 2000, with no bottleneck

  @differentialAgainstMinimumAndSum
  Scenario: The rollup agrees with the minimum and the sum on series-parallel wirings
    Given 300 series-parallel wirings, generated from a fixed seed
    When each is rolled up
    Then each capacity is the minimum over series and the sum over parallel
    And each bottleneck's throughputs add up to its capacity
