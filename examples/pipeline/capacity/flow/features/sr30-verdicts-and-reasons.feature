@SR-30
Feature: Verdicts and reasons
  As a visitor, I want the capacity service to return PASS, FAIL, INCONCLUSIVE
  or ERROR for each requirement by a stated precedence, with a reason built
  from a fixed template, so that a reader sees what the verdict is and why, in
  words the service is allowed to know.

  The shipped pipeline is the worked example of SR-29 as it ships, with
  capacity 1200 and parse its bottleneck.

  @oneCasePerVerdictKind
  Scenario Outline: Each verdict kind has a case that returns it
    Given a requirement on <quantity> <comparison> <limit> of <subject>, with verification case <case>
    When its verdict is returned
    Then it is <kind>, because "<reason>"

    Examples:
      | quantity | comparison | limit | subject                                   | case | kind         | reason                                       |
      | capacity | GE         | 1000  | the shipped pipeline                      | none | PASS         | capacity 1200 against 1000, limited by parse |
      | capacity | GE         | 1500  | the shipped pipeline                      | none | FAIL         | capacity 1200 against 1500, limited by parse |
      | latency  | LE         | 200   | the shipped pipeline                      | none | INCONCLUSIVE | no service computes latency                  |
      | capacity | GE         | 1500  | a pipeline whose indexA has no throughput | none | ERROR        | indexA has missing throughput                |

  @precedenceHolds
  Scenario Outline: When two conditions hold at once, the ordering decides
    Given a requirement on <quantity> <comparison> <limit> of <subject>, with verification case <case>
    When its verdict is returned
    Then it is <kind>, because "<reason>"

    Examples:
      | quantity | comparison | limit | subject                                   | case     | kind         | reason                                       |
      | none     | GE         | 1500  | the shipped pipeline                      | PIPE-VC1 | INCONCLUSIVE | no quantity to compute                       |
      | latency  | LE         | 200   | a pipeline whose indexA has no throughput | none     | INCONCLUSIVE | no service computes latency                  |
      | capacity | GE         | 1500  | the shipped pipeline with 9000 of its own | none     | FAIL         | capacity 1200 against 1500, limited by parse |

  @reasonsFollowTheTemplates
  Scenario Outline: Each reason is built from its template
    Given a requirement on <quantity> <comparison> <limit> of <subject>, with verification case <case>
    When its verdict is returned
    Then it is <kind>, because "<reason>"

    Examples:
      | quantity | comparison | limit | subject                                   | case     | kind         | reason                                              |
      | none     | GE         | 1500  | the shipped pipeline                      | none     | INCONCLUSIVE | no quantity to compute                              |
      | latency  | LE         | 200   | the shipped pipeline                      | PIPE-VC1 | INCONCLUSIVE | PIPE-VC1 is declared and no service runs it         |
      | latency  | LE         | 200   | the shipped pipeline                      | none     | INCONCLUSIVE | no service computes latency                         |
      | capacity | GE         | none  | the shipped pipeline                      | none     | INCONCLUSIVE | no limit to compare against                         |
      | capacity | none       | 1500  | the shipped pipeline                      | none     | INCONCLUSIVE | no comparison to apply                              |
      | capacity | GE         | 1500  | two groups with no path between them      | none     | FAIL         | capacity 0 against 1500, no path from entry to exit |
      | capacity | GE         | 1500  | the shipped pipeline                      | none     | FAIL         | capacity 1200 against 1500, limited by parse        |
      | capacity | GT         | 700   | a part whose throughput is 700            | none     | FAIL         | throughput 700 against 700                          |
      | capacity | LE         | 700   | a part whose throughput is 700            | none     | PASS         | throughput 700 against 700                          |
      | capacity | GE         | 1.5   | a part whose throughput is 2.25           | none     | PASS         | throughput 2.25 against 1.5                         |
      | capacity | GE         | 1500  | a pipeline whose indexA has no throughput | none     | ERROR        | indexA has missing throughput                       |
