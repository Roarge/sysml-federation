# Language model experiment: report

Run started 2026-10-01T12:10:33Z on commit `6346d12ac2f2db325f2c7e6601780dfb0ab494cc`. Language model `qwen2.5-coder:14b` (digest `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`, 14.8B, Q4_K_M) on Ollama 0.34.2.
Temperature 0, seed 42, context 8192 tokens. 12 tests, 8 of them carrying a key, 56 requirements and 1346 elements of the systems model. At most 8 tool calls a test and 20 for the incident. A quick run.

To judge the links proposed for tests with no key without being swayed by the reasons given, start with the last list, which has none, before reading the rest.

## The incident

| Browse | Key items found | Citations checked | Citations missing | Steps | Seconds |
|---|---|---|---|---|---|
| as reported | 0 of 12 | 2 | 0 | 20 | 154 |

### as reported

> The service's container has stopped, and its last line reads: sysml-federation serve: router exited: signal: killed. The 'demo' part represents the container and its process tree, but there are no 'satisfied by' links to identify the cause.

- the router is the cause: missed
- the supervisor's state machine is the mechanism: missed
- the transition on the router's exit: missed
- the supervisor's code: missed
- the compose file, which restarts nothing: missed
- SR-04, four paths on one port, is unmet: missed
- SR-09, every edit lost with the memory it lived in: missed
- the document's monitor fails too: missed
- the root's monitor fails too: missed
- a stakeholder story behind SR-04: missed
- the stakeholder story behind SR-09: missed
- the session's heartbeat, which keeps passing: missed

The view it ended with:

```
package IncidentView {
    viewpoint def IdentifyTheCauseOfTheServiceContainerStoppingAndTheAlertFail {
        frame concern question {
            doc /* Identify the cause of the service container stopping and the alert failure. */
        }
    }
    view incident {
        viewpoint answers : IdentifyTheCauseOfTheServiceContainerStoppingAndTheAlertFail;
        // This part is related to serving the viewer.
        expose Federation_LogicalArchitecture::Adapter::serve;
        // This part represents the container and its process tree.
        expose Federation_LogicalArchitecture::demo;
        // This part represents the repository where the service is built and published.
        expose Federation_LogicalArchitecture::project::repository;
    }
}
```

## Resolvers on the tests that carry a key

The keys were hidden from every resolver but the key rule, which reads the names as written.

| Resolver | Proposed | Correct | Precision (95% interval) | Recall (95% interval) |
|---|---|---|---|---|
| key rule | 8 | 8 | 1.00 (0.68 to 1.00) | 1.00 (0.68 to 1.00) |
| word overlap | 8 | 5 | 0.62 (0.31 to 0.86) | 0.62 (0.31 to 0.86) |
| word overlap, best-ranked | 8 | 5 | 0.62 (0.31 to 0.86) | 0.62 (0.31 to 0.86) |
| systems model baseline | 8 | 6 | 0.75 (0.41 to 0.93) | 0.75 (0.41 to 0.93) |
| language model | 2 | 2 | 1.00 (0.34 to 1.00) | 0.25 (0.07 to 0.59) |

For 8 of the 8 tests that carry a key, 1.00 (0.68 to 1.00), the recorded requirement is among the first eight elements the search ranks for the test's text, or one trace link from one of them.

Test by test, the language model against the word overlap's best-ranked answer:

| Both correct | Only the language model | Only the word overlap | Neither |
|---|---|---|---|
| 1 | 1 | 4 | 2 |

McNemar's exact test on the 5 tests only one answered correctly: p = 0.375.

Test by test, the language model against the systems model baseline:

| Both correct | Only the language model | Only the baseline | Neither |
|---|---|---|---|
| 1 | 1 | 5 | 1 |

McNemar's exact test on the 6 tests only one answered correctly: p = 0.219.

## Links to elements of other kinds

| Kind | Links | Within three links of the recorded requirement |
|---|---|---|
| verification def | 3 | 3 |

3 of 3 such links lie within three links of the recorded requirement, 1.00 (0.44 to 1.00). Of every element, the share that lies as near, taken over the same links, is 0.41.

By distance, beside the share of every element and of the visited elements, the ones the browse's tool answers named, that lie as near. Each share is taken over the same links.

| Within | Links (95% interval) | Every element | Visited elements |
|---|---|---|---|
| one link | 1 of 3, 0.33 (0.06 to 0.79) | 0.01 | 0.12 |
| two links | 1 of 3, 0.33 (0.06 to 0.79) | 0.13 | 0.25 |
| three links | 3 of 3, 1.00 (0.44 to 1.00) | 0.41 | 1.00 |

## The explanation tests

Each probe browses again for a test the language model linked, with one thing changed, and counts how often the answer changed.

| Probe | What changed | Asked | Answer changed (95% interval) | to none | to another | Skipped | Errors |
|---|---|---|---|---|---|---|---|
| deletion | the cited words deleted from the test and from every tool answer | 1 | 1.00 (0.21 to 1.00) | 1 | 0 | 0 | 0 |
| control | as many other words of the test deleted at random | 0 | no cases | 0 | 0 | 1 | 0 |
| rare-shared | uncited words the test and its requirement share deleted, rarest first | 0 | no cases | 0 | 0 | 1 | 0 |
| reconstruction | the test replaced by the cited words alone, its code hidden | 1 | 1.00 (0.21 to 1.00) | 1 | 0 | 0 | 0 |
| repeat | the same browse again (changed means the final reply's text differs) | 3 | 0.33 (0.06 to 0.79) | 0 | 0 | 0 | 0 |
| deletion, correct links only | the cited words deleted from the test and from every tool answer | 1 | 1.00 (0.21 to 1.00) | 1 | 0 | 0 | 0 |
| control, correct links only | as many other words of the test deleted at random | 0 | no cases | 0 | 0 | 1 | 0 |
| rare-shared, correct links only | uncited words the test and its requirement share deleted, rarest first | 0 | no cases | 0 | 0 | 1 | 0 |
| reconstruction, correct links only | the test replaced by the cited words alone, its code hidden | 1 | 1.00 (0.21 to 1.00) | 1 | 0 | 0 | 0 |

Deletion against each control, test by test, on the tests where both were asked and answered:

| Deletion against | Asked | Both changed | Only deletion | Only the control | Neither | McNemar's exact p |
|---|---|---|---|---|---|---|
| control | 0 | 0 | 0 | 0 | 0 | 1.000 |
| rare-shared | 0 | 0 | 0 | 0 | 0 | 1.000 |

Evidence found as written, apart from case, in the test's fields: 2 of 10 pieces, 0.20 (0.06 to 0.51).

## Suspected mismatches

None.

Known before the run:

- `Federation_LogicalArchitecture::Adapter::serve`: The systems model puts the store and the version counter in adapter/serve. The code has them in adapter/projection/store.go.

## Links proposed for tests with no key

Nobody recorded an answer for these tests, so these are for a person to judge and count in no score.

| Test | Resolver | Proposed | Reason given |
|---|---|---|---|
| `adapter/model/eval_test.go#TestEvalRefusals` | word overlap | SR-26 | model, adapter |
| `adapter/model/eval_test.go#TestEvalRefusals` | systems model baseline | SR-26 | through demo::adapter |
| `adapter/model/eval_test.go#TestEvalRefusals` | language model | CHK_Refusals | The test 'EvalRefusals' appears to verify multiple elements in the systems model, including verification definitions and system requirements related to refusal of unsupported syntax and the behavior of the served model. |
| `adapter/model/eval_test.go#TestEvalRefusals` | language model | VAL_US_15 | The test 'EvalRefusals' appears to verify multiple elements in the systems model, including verification definitions and system requirements related to refusal of unsupported syntax and the behavior of the served model. |
| `adapter/model/eval_test.go#TestEvalRefusals` | language model | VC_SR_18 | The test 'EvalRefusals' appears to verify multiple elements in the systems model, including verification definitions and system requirements related to refusal of unsupported syntax and the behavior of the served model. |
| `adapter/model/eval_test.go#TestEvalRefusals` | language model | SR-24 | The test 'EvalRefusals' appears to verify multiple elements in the systems model, including verification definitions and system requirements related to refusal of unsupported syntax and the behavior of the served model. |
| `adapter/model/eval_test.go#TestEvalRefusals` | language model | SR-25 | The test 'EvalRefusals' appears to verify multiple elements in the systems model, including verification definitions and system requirements related to refusal of unsupported syntax and the behavior of the served model. |
| `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | systems model baseline | SR-18 | through VC_SR_18 |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | word overlap | SR-28 | capacity |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | systems model baseline | SR-28 | through demo::capacity |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | language model | CHK_RouterHealthNotProxied | The 'Health' test likely verifies the router health check as indicated by the 'CHK_RouterHealthNotProxied' action, and exercises other related actions like 'VC_SR_48' and 'VC_SR_04'. |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | language model | VC_SR_48 | The 'Health' test likely verifies the router health check as indicated by the 'CHK_RouterHealthNotProxied' action, and exercises other related actions like 'VC_SR_48' and 'VC_SR_04'. |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | language model | VC_SR_04 | The 'Health' test likely verifies the router health check as indicated by the 'CHK_RouterHealthNotProxied' action, and exercises other related actions like 'VC_SR_48' and 'VC_SR_04'. |
| `internal/assert/assert_test.go#TestEqualPasses` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestEqualPasses` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestEqualPasses` | language model | CHK_RouterRootRedirect | The test 'EqualPasses' appears to verify several actions and verification definitions related to redirects, specific pipe passes, and document watching. The evidence from the test supports these verifications. |
| `internal/assert/assert_test.go#TestEqualPasses` | language model | CHK_Us03NothingMoves | The test 'EqualPasses' appears to verify several actions and verification definitions related to redirects, specific pipe passes, and document watching. The evidence from the test supports these verifications. |
| `internal/assert/assert_test.go#TestEqualPasses` | language model | UC_10_ChangeFromTheViewerWatchTheDocument | The test 'EqualPasses' appears to verify several actions and verification definitions related to redirects, specific pipe passes, and document watching. The evidence from the test supports these verifications. |
| `internal/assert/assert_test.go#TestEqualPasses` | language model | CHK_RouterHealthNotProxied | The test 'EqualPasses' appears to verify several actions and verification definitions related to redirects, specific pipe passes, and document watching. The evidence from the test supports these verifications. |
| `internal/assert/assert_test.go#TestEqualPasses` | language model | CHK_Us04BottleneckMoves | The test 'EqualPasses' appears to verify several actions and verification definitions related to redirects, specific pipe passes, and document watching. The evidence from the test supports these verifications. |

## For blind judgement

The same proposals again, each once, in a scrambled order and with nothing to say where they came from. Judge these first, then compare with the reasons further up.

| Number | Test | Proposed | Correct? |
|---|---|---|---|
| 1 | `examples/pipeline/capacity/capacity_test.go#TestHealth` | SR-28 | |
| 2 | `internal/assert/assert_test.go#TestEqualPasses` | CHK_RouterHealthNotProxied | |
| 3 | `examples/pipeline/capacity/capacity_test.go#TestHealth` | VC_SR_48 | |
| 4 | `examples/pipeline/capacity/capacity_test.go#TestHealth` | VC_SR_04 | |
| 5 | `internal/assert/assert_test.go#TestEqualPasses` | UC_10_ChangeFromTheViewerWatchTheDocument | |
| 6 | `adapter/model/eval_test.go#TestEvalRefusals` | VAL_US_15 | |
| 7 | `internal/assert/assert_test.go#TestEqualPasses` | CHK_Us03NothingMoves | |
| 8 | `adapter/model/eval_test.go#TestEvalRefusals` | VC_SR_18 | |
| 9 | `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | SR-18 | |
| 10 | `adapter/model/eval_test.go#TestEvalRefusals` | SR-26 | |
| 11 | `adapter/model/eval_test.go#TestEvalRefusals` | SR-24 | |
| 12 | `adapter/model/eval_test.go#TestEvalRefusals` | SR-25 | |
| 13 | `internal/assert/assert_test.go#TestEqualPasses` | CHK_RouterRootRedirect | |
| 14 | `internal/assert/assert_test.go#TestEqualPasses` | CHK_Us04BottleneckMoves | |
| 15 | `examples/pipeline/capacity/capacity_test.go#TestHealth` | CHK_RouterHealthNotProxied | |
| 16 | `adapter/model/eval_test.go#TestEvalRefusals` | CHK_Refusals | |
| 17 | `internal/assert/assert_test.go#TestEqualPasses` | SC-04 | |

## Time

| Probe | Questions | Mean seconds | Mean prompt tokens | Largest prompt |
|---|---|---|---|---|
| base | 104 | 8.4 | 842 | 1031 |
| incident | 21 | 7.3 | 814 | 850 |
| reconstruction | 9 | 6.7 | 786 | 829 |
| deletion | 9 | 4.6 | 727 | 737 |
| repeat | 27 | 8.9 | 846 | 1020 |

Total time spent on questions: 23 minutes.
