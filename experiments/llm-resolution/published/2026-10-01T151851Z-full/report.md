# Language model experiment: report

Run started 2026-10-01T15:18:51Z on commit `6346d12ac2f2db325f2c7e6601780dfb0ab494cc`. Language model `qwen2.5-coder:14b` (digest `9ec8897f747e246e970bc5cfdda85d22f1123dc2e3d34978a010a75968716849`, 14.8B, Q4_K_M) on Ollama 0.34.2.
Temperature 0, seed 42, context 8192 tokens. 169 tests, 69 of them carrying a key, 56 requirements and 1346 elements of the systems model. At most 8 tool calls a test and 20 for the incident.

To judge the links proposed for tests with no key without being swayed by the reasons given, start with the last list, which has none, before reading the rest.

## The incident

| Browse | Key items found | Citations checked | Citations missing | Steps | Seconds |
|---|---|---|---|---|---|
| as reported | 0 of 12 | 2 | 0 | 20 | 151 |
| again | 0 of 12 | 2 | 0 | 20 | 152 |
| alert only | 0 of 12 | 4 | 0 | 20 | 174 |
| transition removed | 0 of 12 | 12 | 4 | 20 | 181 |
| control link removed | 0 of 12 | 12 | 4 | 20 | 181 |

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

### again

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

### alert only

> The alert 'monitor: the viewer answers' failing since 02:10 is likely due to the 'demo::viewer' part not functioning as expected. This part is responsible for satisfying SR-14, which is related to the viewer's verdict and reason. Since there are no 'satisfied by' links for 'demo::viewer', it suggests that the implementation or configuration of this part might be incorrect or missing.

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
    viewpoint def WhatIsTheSourceOfTheAlertMonitorTheViewerAnswersFailingSince {
        frame concern question {
            doc /* What is the source of the alert 'monitor: the viewer answers' failing since 02:10? */
        }
    }
    view incident {
        viewpoint answers : WhatIsTheSourceOfTheAlertMonitorTheViewerAnswersFailingSince;
        // Requirement related to the viewer's verdict and reason.
        expose Federation_SystemStories::SR_14_VerdictAndReasonInTheViewer;
        // Requirement related to paths on one port.
        expose Federation_SystemStories::SR_04_FourPathsOnOnePort;
        // Requirement related to failing requirements.
        expose Federation_SystemStories::SR_15_FailingRequirementInRed;
        // The part that satisfies SR-14, related to the viewer's verdict and reason.
        expose Federation_LogicalArchitecture::demo::viewer;
    }
}
```

### transition removed

> The service container stopping and the alert failure are likely caused by issues with the viewer components. The 'demo::viewer' and 'session::viewer' parts, along with the 'CheckSession::session' check, are involved in the service's operation and might have failed, leading to the container being killed and the alert not receiving a response.

Mechanism cited: demo::viewer, session::viewer, CheckSession::session.

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
- cited in consequences but missing: `monitor: the viewer answers`, no such element in what the tools could show
- cited in code but missing: `sysml-federation serve: router exited: signal: killed`, no such element in what the tools could show
- cited in path but missing: `session::viewer`, no link joins demo::viewer and session::viewer
- cited in path but missing: `sysml-federation serve: router exited: signal: killed`, no such element in what the tools could show

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
        // This part is related to the viewer and might be involved in the service container stopping.
        expose Federation_LogicalArchitecture::demo::viewer;
        // This part is also related to the viewer and might be involved in the service container stopping.
        expose Federation_LogicalArchitecture::CheckSession::session::viewer;
        // This part is related to the session and might be involved in the service container stopping.
        expose Federation_LogicalArchitecture::CheckSession::session;
    }
}
```

### control link removed

> The service container stopping and the alert failure are likely caused by issues with the viewer components. The 'demo::viewer' and 'session::viewer' parts, along with the 'CheckSession::session' check, are involved in the service's operation and might have failed, leading to the container being killed and the alert not receiving a response.

Mechanism cited: demo::viewer, session::viewer, CheckSession::session.

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
- cited in consequences but missing: `monitor: the viewer answers`, no such element in what the tools could show
- cited in code but missing: `sysml-federation serve: router exited: signal: killed`, no such element in what the tools could show
- cited in path but missing: `session::viewer`, no link joins demo::viewer and session::viewer
- cited in path but missing: `sysml-federation serve: router exited: signal: killed`, no such element in what the tools could show

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
        // This part is related to the viewer and might be involved in the service container stopping.
        expose Federation_LogicalArchitecture::demo::viewer;
        // This part is also related to the viewer and might be involved in the service container stopping.
        expose Federation_LogicalArchitecture::CheckSession::session::viewer;
        // This part is related to the session and might be involved in the service container stopping.
        expose Federation_LogicalArchitecture::CheckSession::session;
    }
}
```

## Resolvers on the tests that carry a key

The keys were hidden from every resolver but the key rule, which reads the names as written.

| Resolver | Proposed | Correct | Precision (95% interval) | Recall (95% interval) |
|---|---|---|---|---|
| key rule | 69 | 69 | 1.00 (0.95 to 1.00) | 1.00 (0.95 to 1.00) |
| word overlap | 68 | 41 | 0.60 (0.48 to 0.71) | 0.59 (0.48 to 0.70) |
| word overlap, best-ranked | 69 | 41 | 0.59 (0.48 to 0.70) | 0.59 (0.48 to 0.70) |
| systems model baseline | 69 | 42 | 0.61 (0.49 to 0.72) | 0.61 (0.49 to 0.72) |
| language model | 25 | 17 | 0.68 (0.48 to 0.83) | 0.25 (0.16 to 0.36) |

For 58 of the 69 tests that carry a key, 0.84 (0.74 to 0.91), the recorded requirement is among the first eight elements the search ranks for the test's text, or one trace link from one of them.

Test by test, the language model against the word overlap's best-ranked answer:

| Both correct | Only the language model | Only the word overlap | Neither |
|---|---|---|---|
| 13 | 4 | 28 | 24 |

McNemar's exact test on the 32 tests only one answered correctly: p = 0.000.

Test by test, the language model against the systems model baseline:

| Both correct | Only the language model | Only the baseline | Neither |
|---|---|---|---|
| 11 | 6 | 31 | 21 |

McNemar's exact test on the 37 tests only one answered correctly: p = 0.000.

## Links to elements of other kinds

| Kind | Links | Within three links of the recorded requirement |
|---|---|---|
| verification def | 31 | 31 |
| part | 6 | 6 |
| action | 3 | 1 |
| requirement | 2 | 2 |
| part def | 1 | 1 |

41 of 43 such links lie within three links of the recorded requirement, 0.95 (0.85 to 0.99). Of every element, the share that lies as near, taken over the same links, is 0.35.

By distance, beside the share of every element and of the visited elements, the ones the browse's tool answers named, that lie as near. Each share is taken over the same links.

| Within | Links (95% interval) | Every element | Visited elements |
|---|---|---|---|
| one link | 17 of 43, 0.40 (0.26 to 0.54) | 0.01 | 0.18 |
| two links | 25 of 43, 0.58 (0.43 to 0.72) | 0.11 | 0.41 |
| three links | 41 of 43, 0.95 (0.85 to 0.99) | 0.35 | 0.95 |

## The explanation tests

Each probe browses again for a test the language model linked, with one thing changed, and counts how often the answer changed.

| Probe | What changed | Asked | Answer changed (95% interval) | to none | to another | Skipped | Errors |
|---|---|---|---|---|---|---|---|
| deletion | the cited words deleted from the test and from every tool answer | 13 | 0.92 (0.67 to 0.99) | 11 | 1 | 0 | 0 |
| control | as many other words of the test deleted at random | 13 | 0.54 (0.29 to 0.77) | 5 | 2 | 0 | 0 |
| rare-shared | uncited words the test and its requirement share deleted, rarest first | 11 | 0.55 (0.28 to 0.79) | 4 | 2 | 2 | 0 |
| reconstruction | the test replaced by the cited words alone, its code hidden | 13 | 0.85 (0.58 to 0.96) | 9 | 2 | 0 | 0 |
| repeat | the same browse again (changed means the final reply's text differs) | 43 | 0.14 (0.07 to 0.27) | 0 | 0 | 0 | 0 |
| deletion, correct links only | the cited words deleted from the test and from every tool answer | 3 | 1.00 (0.44 to 1.00) | 3 | 0 | 0 | 0 |
| control, correct links only | as many other words of the test deleted at random | 3 | 0.33 (0.06 to 0.79) | 1 | 0 | 0 | 0 |
| rare-shared, correct links only | uncited words the test and its requirement share deleted, rarest first | 3 | 0.33 (0.06 to 0.79) | 1 | 0 | 0 | 0 |
| reconstruction, correct links only | the test replaced by the cited words alone, its code hidden | 3 | 0.67 (0.21 to 0.94) | 1 | 1 | 0 | 0 |

Deletion against each control, test by test, on the tests where both were asked and answered:

| Deletion against | Asked | Both changed | Only deletion | Only the control | Neither | McNemar's exact p |
|---|---|---|---|---|---|---|
| control | 13 | 6 | 6 | 1 | 0 | 0.125 |
| rare-shared | 11 | 5 | 5 | 1 | 0 | 0.219 |

Evidence found as written, apart from case, in the test's fields: 61 of 91 pieces, 0.67 (0.57 to 0.76).

## Suspected mismatches

| Raised in | Element | The systems model says | The system shows | Why |
|---|---|---|---|---|
| adapter/serve/scenarios_test.go#TestScenarios | `demo::adapter::serve::TestScenarios` | unknown | unknown | The test Scenarios is not found in the systems model. |
| adapter/serve/scenarios_test.go#TestScenarios | `AD-0034` | unknown | unknown | The requirement AD-0034 is not found in the systems model. |
| adapter/serve/scenarios_test.go#TestScenarios | `demo::adapter::serve::adapter` | unknown | unknown | The adapter element is not found in the systems model. |
| adapter/syntax/scenarios_test.go#TestScenarios | `unknown` | TestScenarios is not in the systems model | TestScenarios is in the code | The systems model does not include the test 'TestScenarios', which is present in the code. |
| examples/pipeline/document/scenarios_test.go#TestScenarios | `AD-0034` | unknown | not found in the model | The test references an acceptance criterion (AD-0034) that is not present in the systems model. |
| examples/pipeline/ui/ui_test.go#TestPureModulesUnderNode | `unknown` | tokeniser, layout, shared client's frame parser | unknown | These modules are not present in the systems model, making it impossible to verify or exercise them through the test. |

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
| `adapter/model/model_test.go#TestAttributeShapes` | word overlap | SR-26 | model, adapter |
| `adapter/model/model_test.go#TestAttributeShapes` | systems model baseline | SR-45 | through demo::exampleModel |
| `adapter/model/model_test.go#TestAttributeShapes` | language model | CHK_Us08ShapeTheDocument | The test 'AttributeShapes' is linked to the verification definition CHK_Us08ShapeTheDocument, which verifies the requirement US-08. The validation VAL_US_08 is also related and exercises the requirement. |
| `adapter/model/model_test.go#TestAttributeShapes` | language model | VAL_US_08 | The test 'AttributeShapes' is linked to the verification definition CHK_Us08ShapeTheDocument, which verifies the requirement US-08. The validation VAL_US_08 is also related and exercises the requirement. |
| `adapter/model/model_test.go#TestAttributeShapes` | language model | US-08 | The test 'AttributeShapes' is linked to the verification definition CHK_Us08ShapeTheDocument, which verifies the requirement US-08. The validation VAL_US_08 is also related and exercises the requirement. |
| `adapter/model/model_test.go#TestBuildRefusals` | word overlap | SR-26 | model, adapter |
| `adapter/model/model_test.go#TestBuildRefusals` | systems model baseline | SR-45 | through demo::exampleModel |
| `adapter/model/model_test.go#TestBuildRefusals` | language model | CHK_Refusals | The test 'BuildRefusals' is linked to the verification definition 'CHK_Refusals' which specifically mentions refusal of unsupported syntax. |
| `adapter/model/model_test.go#TestLinkRefusals` | word overlap | SR-26 | model, adapter |
| `adapter/model/model_test.go#TestLinkRefusals` | systems model baseline | SR-45 | through demo::exampleModel |
| `adapter/model/model_test.go#TestLinkRefusals` | language model | CHK_Refusals | The test 'LinkRefusals' is linked to the verification definition 'CHK_Refusals' as it seems to match the test name and description. Additionally, it exercises the verification definitions 'VAL_US_15' and 'VC_SR_18' as they are related to served models and refusal of unsupported syntax, respectively. |
| `adapter/model/model_test.go#TestLinkRefusals` | language model | VAL_US_15 | The test 'LinkRefusals' is linked to the verification definition 'CHK_Refusals' as it seems to match the test name and description. Additionally, it exercises the verification definitions 'VAL_US_15' and 'VC_SR_18' as they are related to served models and refusal of unsupported syntax, respectively. |
| `adapter/model/model_test.go#TestLinkRefusals` | language model | VC_SR_18 | The test 'LinkRefusals' is linked to the verification definition 'CHK_Refusals' as it seems to match the test name and description. Additionally, it exercises the verification definitions 'VAL_US_15' and 'VC_SR_18' as they are related to served models and refusal of unsupported syntax, respectively. |
| `adapter/model/model_test.go#TestPartsTreeAttributesAndPorts` | word overlap | SR-16 | model, part, port, attribute, adapter |
| `adapter/model/model_test.go#TestPartsTreeAttributesAndPorts` | systems model baseline | SR-45 | through demo::exampleModel |
| `adapter/model/model_test.go#TestRelationships` | word overlap | SR-26 | model, adapter |
| `adapter/model/model_test.go#TestRelationships` | systems model baseline | SR-45 | through demo::exampleModel |
| `adapter/model/patch_internal_test.go#TestPatchGuardsItsInputs` | word overlap | SR-22 | patch, literal, model, set, value |
| `adapter/model/patch_internal_test.go#TestPatchGuardsItsInputs` | systems model baseline | SR-22 | through VC_SR_22 |
| `adapter/model/patch_internal_test.go#TestPatchGuardsItsInputs` | language model | SR-25 | The test PatchGuardsItsInputs exercises the internal write, which is likely related to the system requirement SR-25 that mentions rejection of invalid values, including negative values. |
| `adapter/model/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, model, acceptance, criterion, run |
| `adapter/model/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `adapter/projection/projection_test.go#TestAnUntypedPartPublishesNoDefinition` | word overlap | SR-16 | part, projection, model, case, adapter |
| `adapter/projection/projection_test.go#TestAnUntypedPartPublishesNoDefinition` | systems model baseline | SR-16 |  |
| `adapter/projection/projection_test.go#TestEntitiesResolveTheThreeKeyedTypes` | word overlap | SR-22 | projection, adapter |
| `adapter/projection/projection_test.go#TestEntitiesResolveTheThreeKeyedTypes` | systems model baseline | SR-22 |  |
| `adapter/projection/projection_test.go#TestEntitiesResolveTheThreeKeyedTypes` | language model | SR-43 | The test 'EntitiesResolveTheThreeKeyedTypes' is linked to the verification definition VC_SR_43, which mentions 'One query, three services'. This indicates that the test verifies the system requirement SR-43. |
| `adapter/projection/projection_test.go#TestQueriesServeEveryFieldOfTheProjection` | word overlap | SR-22 | projection, query, adapter |
| `adapter/projection/projection_test.go#TestQueriesServeEveryFieldOfTheProjection` | systems model baseline | SR-22 |  |
| `adapter/projection/projection_test.go#TestQueriesServeEveryFieldOfTheProjection` | language model | SR-43 | The test 'QueriesServeEveryFieldOfTheProjection' is linked to the verification case VC_SR_43, which verifies the system requirement SR-43. |
| `adapter/projection/projection_test.go#TestUnknownIDsAreErrors` | word overlap | SR-22 | projection, adapter |
| `adapter/projection/projection_test.go#TestUnknownIDsAreErrors` | systems model baseline | SR-22 |  |
| `adapter/projection/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `adapter/projection/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `adapter/projection/scenarios_test.go#TestScenarios` | language model | SC-04 | The test Scenarios is linked to the requirement SC-04, which is about being tracked by an allowlist. |
| `adapter/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `adapter/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `adapter/serve/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `adapter/serve/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `adapter/serve/server_test.go#TestHandlerServesPostAndHealth` | word overlap | SR-02 | serve |
| `adapter/serve/server_test.go#TestHandlerServesPostAndHealth` | systems model baseline | SR-01 | through VC_SR_01 |
| `adapter/syntax/lexer_test.go#TestLexErrorsCarryFileLineAndColumn` | word overlap | SR-18 | syntax, column, line, adapter |
| `adapter/syntax/lexer_test.go#TestLexErrorsCarryFileLineAndColumn` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/lexer_test.go#TestLexKindsAndText` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/lexer_test.go#TestLexKindsAndText` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/lexer_test.go#TestLexKindsAndText` | language model | SR-22 | The test 'LexKindsAndText' is linked to the verification definition VC_SR_22, which in turn verifies the system requirement SR-22. |
| `adapter/syntax/lexer_test.go#TestLexSpansAreByteOffsets` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/lexer_test.go#TestLexSpansAreByteOffsets` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/lexer_test.go#TestPosition` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/lexer_test.go#TestPosition` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/lexer_test.go#TestReservedWordsAreKeywords` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/lexer_test.go#TestReservedWordsAreKeywords` | systems model baseline | SR-11 | through VC_SR_11 |
| `adapter/syntax/lexer_test.go#TestReservedWordsAreKeywords` | language model | SR-11 | The test 'ReservedWordsAreKeywords' is linked to the verification definition VC_SR_11, which mentions the tokeniser and reserved words. This verification definition is associated with the requirement SR-11. |
| `adapter/syntax/parser_test.go#TestExampleModelParses` | word overlap | SR-17 | example, adapter, model |
| `adapter/syntax/parser_test.go#TestExampleModelParses` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | language model | VC_SR_28::aLeafReadsItsOwnAttributeScenario | The test 'ParseAttributeWithBody' exercises actions related to parsing, such as reading an attribute and raising the bottleneck for parsing. It does not directly verify a specific requirement but exercises scenarios involving parsing. |
| `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | language model | UC_04_RaiseTheBottleneck::raiseParse | The test 'ParseAttributeWithBody' exercises actions related to parsing, such as reading an attribute and raising the bottleneck for parsing. It does not directly verify a specific requirement but exercises scenarios involving parsing. |
| `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | language model | CHK_RouterPlayground::expectAnHtmlPage | The test 'ParseAttributeWithBody' exercises actions related to parsing, such as reading an attribute and raising the bottleneck for parsing. It does not directly verify a specific requirement but exercises scenarios involving parsing. |
| `adapter/syntax/parser_test.go#TestParseDefinitionsAndParts` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/parser_test.go#TestParseDefinitionsAndParts` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParseDerivationAndVerification` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/parser_test.go#TestParseDerivationAndVerification` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | language model | SR-25 | The test 'ParseExpressionPrecedenceAndSpans' is named after the requirement 'SR-25' about expression precedence and spans, and it exercises the verification definition 'VC_SR_25'. |
| `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | language model | VC_SR_25 | The test 'ParseExpressionPrecedenceAndSpans' is named after the requirement 'SR-25' about expression precedence and spans, and it exercises the verification definition 'VC_SR_25'. |
| `adapter/syntax/parser_test.go#TestParseNegativeLiteralSpanIncludesTheSign` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/parser_test.go#TestParseNegativeLiteralSpanIncludesTheSign` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | language model | VC_SR_36 | The test 'ParsePackageHeaderAndImports' verifies the system requirement SR-36 and exercises actions related to reading the package layout and import sweep. |
| `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | language model | VC_SR_41 | The test 'ParsePackageHeaderAndImports' verifies the system requirement SR-36 and exercises actions related to reading the package layout and import sweep. |
| `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | language model | VC_SC_03 | The test 'ParsePackageHeaderAndImports' verifies the system requirement SR-36 and exercises actions related to reading the package layout and import sweep. |
| `adapter/syntax/parser_test.go#TestParsePortsConnectAndSatisfy` | word overlap | SR-20 | port, connect, adapter |
| `adapter/syntax/parser_test.go#TestParsePortsConnectAndSatisfy` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParsePortsConnectAndSatisfy` | language model | US-11 | The test 'ParsePortsConnectAndSatisfy' is named in a way that suggests it verifies the requirement 'US-11'. The verification definition 'CHK_Us11QueryTheGraph' is linked to 'US-11', further supporting this connection. |
| `adapter/syntax/parser_test.go#TestParseRequirements` | word overlap | SR-18 | syntax, adapter |
| `adapter/syntax/parser_test.go#TestParseRequirements` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/parser_test.go#TestParseRequirements` | language model | VC_SR_37 | The test 'ParseRequirements' is linked to the verification definitions VC_SR_37 and VC_SR_38, which are related to parsing requirements. |
| `adapter/syntax/parser_test.go#TestParseRequirements` | language model | VC_SR_38 | The test 'ParseRequirements' is linked to the verification definitions VC_SR_37 and VC_SR_38, which are related to parsing requirements. |
| `adapter/syntax/parser_test.go#TestParseUnitIsAQualifiedName` | word overlap | SR-21 | name, qualifi, adapter |
| `adapter/syntax/parser_test.go#TestParseUnitIsAQualifiedName` | systems model baseline | SR-18 | through VC_SR_18 |
| `adapter/syntax/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `adapter/syntax/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `checkly/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `checkly/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `cmd/sysml-federation/main_test.go#TestAddressesMatchTheComposedConfiguration` | word overlap | SR-03 | address, router, configuration, federation, sysml |
| `cmd/sysml-federation/main_test.go#TestAddressesMatchTheComposedConfiguration` | systems model baseline | SR-03 |  |
| `cmd/sysml-federation/main_test.go#TestHealthcheckFailsWhenEitherProbeFails` | word overlap | SR-15 | fail |
| `cmd/sysml-federation/main_test.go#TestHealthcheckFailsWhenEitherProbeFails` | systems model baseline | SR-47 |  |
| `cmd/sysml-federation/main_test.go#TestHealthcheckProbesThePublishedPort` | word overlap | SR-10 | port, publish, other, app |
| `cmd/sysml-federation/main_test.go#TestHealthcheckProbesThePublishedPort` | systems model baseline | SR-04 | through demo |
| `cmd/sysml-federation/main_test.go#TestHelperRouter` | word overlap | SR-03 | router, process, federation, sysml |
| `cmd/sysml-federation/main_test.go#TestHelperRouter` | systems model baseline | SR-03 | through demo::router |
| `cmd/sysml-federation/main_test.go#TestRouterFromEnvReadsTheThreePaths` | word overlap | SR-03 | router, path, federation, sysml, set |
| `cmd/sysml-federation/main_test.go#TestRouterFromEnvReadsTheThreePaths` | systems model baseline | SR-03 | through demo::router |
| `cmd/sysml-federation/main_test.go#TestRouterFromEnvReadsTheThreePaths` | language model | CHK_RouterJoin | The test 'RouterFromEnvReadsTheThreePaths' is linked to the verification definition 'CHK_RouterJoin', which verifies the requirement 'SR-43'. The test's description mentions covering each variable set and unset, which aligns with the verification definition's focus on the join query through the router. |
| `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | word overlap | SR-03 | process, router, federation, sysml |
| `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | systems model baseline | SR-03 | through demo::router |
| `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | language model | VC_SR_02 | The test 'RouterRunsAsAChildProcess' exercises the Router part and verifies the readiness and restart scenarios. |
| `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | language model | VC_SR_09 | The test 'RouterRunsAsAChildProcess' exercises the Router part and verifies the readiness and restart scenarios. |
| `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | language model | Router | The test 'RouterRunsAsAChildProcess' exercises the Router part and verifies the readiness and restart scenarios. |
| `cmd/sysml-federation/main_test.go#TestRunDispatchesSubcommandsAndExitCodes` | word overlap | SR-01 | federation, sysml, run |
| `cmd/sysml-federation/main_test.go#TestRunDispatchesSubcommandsAndExitCodes` | systems model baseline | SR-47 |  |
| `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | word overlap | SR-03 | router, federation, sysml |
| `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | systems model baseline | SR-47 |  |
| `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | language model | VC_SR_15 | The test name and the requirement text suggest that the test exercises the behavior of the Router when it dies, and verifies a requirement related to failing. |
| `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | language model | SR-15 | The test name and the requirement text suggest that the test exercises the behavior of the Router when it dies, and verifies a requirement related to failing. |
| `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | language model | Router | The test name and the requirement text suggest that the test exercises the behavior of the Router when it dies, and verifies a requirement related to failing. |
| `cmd/sysml-federation/main_test.go#TestServeHandsTheRouterItsConfigurationFile` | word overlap | SR-02 | ten, serve, within, second, router |
| `cmd/sysml-federation/main_test.go#TestServeHandsTheRouterItsConfigurationFile` | systems model baseline | SR-03 | through demo::router |
| `cmd/sysml-federation/main_test.go#TestServeRefusesAModelItCannotRead` | word overlap | SR-47 | main, model |
| `cmd/sysml-federation/main_test.go#TestServeRefusesAModelItCannotRead` | systems model baseline | SR-47 |  |
| `cmd/sysml-federation/main_test.go#TestServerDrainIsBoundedByItsTimeout` | word overlap | SR-24 | bound, leave, read |
| `cmd/sysml-federation/main_test.go#TestServerDrainIsBoundedByItsTimeout` | systems model baseline | SR-47 |  |
| `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | word overlap | SR-06 | budget |
| `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | systems model baseline | SR-06 | through demo |
| `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | language model | CHK_SessionHeartbeat | The test StopBudgetFitsAContainerGrace verifies the sequential stop behavior of the system, which is linked to requirement SR-48. It also exercises verification definitions related to 'Size budget' and 'Licence carried in the image'. |
| `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | language model | VC_SR_06 | The test StopBudgetFitsAContainerGrace verifies the sequential stop behavior of the system, which is linked to requirement SR-48. It also exercises verification definitions related to 'Size budget' and 'Licence carried in the image'. |
| `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | language model | VC_SR_07 | The test StopBudgetFitsAContainerGrace verifies the sequential stop behavior of the system, which is linked to requirement SR-48. It also exercises verification definitions related to 'Size budget' and 'Licence carried in the image'. |
| `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | language model | SR-48 | The test StopBudgetFitsAContainerGrace verifies the sequential stop behavior of the system, which is linked to requirement SR-48. It also exercises verification definitions related to 'Size budget' and 'Licence carried in the image'. |
| `cmd/sysml-federation/main_test.go#TestStopRouterKillsAChildThatIgnoresSIGTERM` | word overlap | SR-03 | router, federation, sysml |
| `cmd/sysml-federation/main_test.go#TestStopRouterKillsAChildThatIgnoresSIGTERM` | systems model baseline | SR-03 | through demo::router |
| `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | word overlap | SR-01 | federation, sysml |
| `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | systems model baseline | SR-47 |  |
| `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | language model | VAL_US_02 | The test 'SubcommandsServeOneComponent' appears to verify the system requirements and verification definitions related to reading the model and executing a command with a published image. |
| `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | language model | VC_SR_01 | The test 'SubcommandsServeOneComponent' appears to verify the system requirements and verification definitions related to reading the model and executing a command with a published image. |
| `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | language model | US-02 | The test 'SubcommandsServeOneComponent' appears to verify the system requirements and verification definitions related to reading the model and executing a command with a published image. |
| `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | language model | SR-01 | The test 'SubcommandsServeOneComponent' appears to verify the system requirements and verification definitions related to reading the model and executing a command with a published image. |
| `cmd/sysml-federation/main_test.go#TestUIRefusesADirectoryWithNoPage` | word overlap | SC-04 | directory, rule, html, path, nam |
| `cmd/sysml-federation/main_test.go#TestUIRefusesADirectoryWithNoPage` | systems model baseline | SC-04 |  |
| `cmd/sysml-federation/main_test.go#TestUIRefusesToListTheSharedDirectory` | word overlap | SR-44 | app, shipp, both |
| `cmd/sysml-federation/main_test.go#TestUIRefusesToListTheSharedDirectory` | systems model baseline | SC-01 | through VC_SC_01 |
| `cmd/sysml-federation/main_test.go#TestUIRefusesToListTheSharedDirectory` | language model | CHK_Refusals | The Go test UIRefusesToListTheSharedDirectory verifies the CHK_Refusals verification definition, which involves checks for refusals and a guard against shipped files, aligning with the test's description of refusing to list the shared directory and ensuring other resources are still served. |
| `cmd/sysml-federation/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `cmd/sysml-federation/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `cmd/sysml-federation/scenarios_test.go#TestScenarios` | language model | req-AD-0034 | The test 'TestScenarios' is explicitly linked to the requirement AD-0034 in its documentation. |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | word overlap | SR-28 | capacity |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | systems model baseline | SR-28 | through demo::capacity |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | language model | CHK_RouterHealthNotProxied | The 'Health' test likely verifies the router health check as indicated by the 'CHK_RouterHealthNotProxied' action, and exercises other related actions like 'VC_SR_48' and 'VC_SR_04'. |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | language model | VC_SR_48 | The 'Health' test likely verifies the router health check as indicated by the 'CHK_RouterHealthNotProxied' action, and exercises other related actions like 'VC_SR_48' and 'VC_SR_04'. |
| `examples/pipeline/capacity/capacity_test.go#TestHealth` | language model | VC_SR_04 | The 'Health' test likely verifies the router health check as indicated by the 'CHK_RouterHealthNotProxied' action, and exercises other related actions like 'VC_SR_48' and 'VC_SR_04'. |
| `examples/pipeline/capacity/capacity_test.go#TestLeafCapacityIsItsOwnAttribute` | word overlap | SR-28 | capacity, attribute, own |
| `examples/pipeline/capacity/capacity_test.go#TestLeafCapacityIsItsOwnAttribute` | systems model baseline | SR-28 | through demo::capacity |
| `examples/pipeline/capacity/capacity_test.go#TestLeafCapacityIsItsOwnAttribute` | language model | SR-28 | The test name 'LeafCapacityIsItsOwnAttribute' suggests it verifies a requirement related to a leaf reading its own attribute. The action VC_SR_28 is linked to this requirement, indicating it is verified by the test. |
| `examples/pipeline/capacity/flow/flow_test.go#TestRollupNamesTheFaultyChild` | word overlap | SR-28 | flow, capacity, child, rollup |
| `examples/pipeline/capacity/flow/flow_test.go#TestRollupNamesTheFaultyChild` | systems model baseline | SR-28 |  |
| `examples/pipeline/capacity/flow/flow_test.go#TestRollupNamesTheFaultyChild` | language model | SR-28 | The test 'RollupNamesTheFaultyChild' is linked to the verification definition 'VC_SR_28', which is related to 'Rollup by maximum flow'. This indicates that the test verifies the system requirement 'SR-28'. |
| `examples/pipeline/capacity/flow/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `examples/pipeline/capacity/flow/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `examples/pipeline/capacity/flow/scenarios_test.go#TestScenarios` | language model | req_AD-0034 | The test Scenarios is linked to the requirement AD-0034, which is related to tracking by allowlist. |
| `examples/pipeline/capacity/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `examples/pipeline/capacity/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `examples/pipeline/capacity/scenarios_test.go#TestScenarios` | language model | SR-28 | The test Scenarios is likely verifying the requirement SR-28, which is related to maximum flow, and exercising the part demo::capacity, which is likely related to the capacity aspect of the system. |
| `examples/pipeline/capacity/scenarios_test.go#TestScenarios` | language model | demo::capacity | The test Scenarios is likely verifying the requirement SR-28, which is related to maximum flow, and exercising the part demo::capacity, which is likely related to the capacity aspect of the system. |
| `examples/pipeline/document/document_test.go#TestARefusedMutationIsAnErrorAndLeavesTheVersion` | word overlap | SR-27 | document, version |
| `examples/pipeline/document/document_test.go#TestARefusedMutationIsAnErrorAndLeavesTheVersion` | systems model baseline | SR-27 | through demo::document |
| `examples/pipeline/document/document_test.go#TestARefusedMutationIsAnErrorAndLeavesTheVersion` | language model | SR-18 | The test name 'ARefusedMutationIsAnErrorAndLeavesTheVersion' indicates it verifies the refusal of unsupported syntax (SR-18) and the rejection of invalid values (SR-25). |
| `examples/pipeline/document/document_test.go#TestARefusedMutationIsAnErrorAndLeavesTheVersion` | language model | SR-25 | The test name 'ARefusedMutationIsAnErrorAndLeavesTheVersion' indicates it verifies the refusal of unsupported syntax (SR-18) and the rejection of invalid values (SR-25). |
| `examples/pipeline/document/document_test.go#TestEntityAnswersForUnknownIDs` | word overlap | SR-27 | document |
| `examples/pipeline/document/document_test.go#TestEntityAnswersForUnknownIDs` | systems model baseline | SR-27 | through demo::document |
| `examples/pipeline/document/document_test.go#TestEntityAnswersForUnknownIDs` | language model | VAL_US_11 | The test 'EntityAnswersForUnknownIDs' is related to querying the graph, which is verified by the verification definitions VAL_US_11 and CHK_Us11QueryTheGraph. |
| `examples/pipeline/document/document_test.go#TestEntityAnswersForUnknownIDs` | language model | CHK_Us11QueryTheGraph | The test 'EntityAnswersForUnknownIDs' is related to querying the graph, which is verified by the verification definitions VAL_US_11 and CHK_Us11QueryTheGraph. |
| `examples/pipeline/document/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `examples/pipeline/document/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `examples/pipeline/document/scenarios_test.go#TestScenarios` | language model | SC-06 | The test 'TestScenarios' is linked to the requirements 'SC-06' and 'SC-04' as they are the only requirements in the view, and the test's doc mentions an acceptance criterion (AD-0034) that is not found in the model. |
| `examples/pipeline/document/scenarios_test.go#TestScenarios` | language model | SC-04 | The test 'TestScenarios' is linked to the requirements 'SC-06' and 'SC-04' as they are the only requirements in the view, and the test's doc mentions an acceptance criterion (AD-0034) that is not found in the model. |
| `examples/pipeline/document/tree/tree_test.go#TestEditText` | word overlap | SR-33 | tree, document |
| `examples/pipeline/document/tree/tree_test.go#TestEditText` | systems model baseline | SR-33 | through VC_SR_33 |
| `examples/pipeline/document/tree/tree_test.go#TestLoadRefusesDuplicateIDs` | word overlap | SR-33 | tree, document |
| `examples/pipeline/document/tree/tree_test.go#TestLoadRefusesDuplicateIDs` | systems model baseline | SR-33 | through VC_SR_33 |
| `examples/pipeline/document/tree/tree_test.go#TestLoadRefusesDuplicateIDs` | language model | VC_SR_19 | The test 'LoadRefusesDuplicateIDs' is linked to the verification definition 'VC_SR_19', which verifies the system requirement 'SR-19'. |
| `examples/pipeline/document/tree/tree_test.go#TestRefusals` | word overlap | SR-33 | tree, document |
| `examples/pipeline/document/tree/tree_test.go#TestRefusals` | systems model baseline | SR-33 | through VC_SR_33 |
| `examples/pipeline/document/tree/tree_test.go#TestRefusals` | language model | CHK_Refusals | The test 'Refusals' in the 'tree' package is linked to the verification definition 'CHK_Refusals', which verifies system requirements SR-24, SR-25, and user story US-15. |
| `examples/pipeline/document/tree/tree_test.go#TestRefusals` | language model | SR-24 | The test 'Refusals' in the 'tree' package is linked to the verification definition 'CHK_Refusals', which verifies system requirements SR-24, SR-25, and user story US-15. |
| `examples/pipeline/document/tree/tree_test.go#TestRefusals` | language model | SR-25 | The test 'Refusals' in the 'tree' package is linked to the verification definition 'CHK_Refusals', which verifies system requirements SR-24, SR-25, and user story US-15. |
| `examples/pipeline/document/tree/tree_test.go#TestRefusals` | language model | US-15 | The test 'Refusals' in the 'tree' package is linked to the verification definition 'CHK_Refusals', which verifies system requirements SR-24, SR-25, and user story US-15. |
| `examples/pipeline/document/tree/tree_test.go#TestRequirementLookup` | word overlap | SR-33 | tree, requirement, document |
| `examples/pipeline/document/tree/tree_test.go#TestRequirementLookup` | systems model baseline | SR-33 | through VC_SR_33 |
| `examples/pipeline/document/tree/tree_test.go#TestRequirementLookup` | language model | VC_SR_37 | The test 'RequirementLookup' is linked to the verification definition 'VC_SR_37', which verifies the system requirement 'SR-37'. |
| `examples/pipeline/isolation_test.go#TestEveryEditorialOperationIsAcceptedThroughTheGraph` | word overlap | SR-35 | operation, editorial, accept, renumber, last |
| `examples/pipeline/isolation_test.go#TestEveryEditorialOperationIsAcceptedThroughTheGraph` | systems model baseline | SR-35 | through VC_SR_35 |
| `examples/pipeline/isolation_test.go#TestEveryEditorialOperationIsAcceptedThroughTheGraph` | language model | SR-35 | The test verifies that editorial operations are accepted through the graph, which aligns with the requirement SR-35. |
| `examples/pipeline/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `examples/pipeline/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `examples/pipeline/scenarios_test.go#TestScenarios` | language model | SC-04 | The test Scenarios is linked to the requirement SC-04, which is related to tracking by allowlist. |
| `examples/pipeline/ui/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `examples/pipeline/ui/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |
| `examples/pipeline/ui/scenarios_test.go#TestScenarios` | language model | SC-06 | The test 'TestScenarios' is linked to the requirement 'SC-06' through the mention of 'AD-0034' in the test's documentation. |
| `examples/pipeline/ui/ui_test.go#TestANodeListKeepsRoomPastItsLastRow` | systems model baseline | SR-37 | through VC_SR_37 |
| `examples/pipeline/ui/ui_test.go#TestBothAppsRefetchWhenTheTabIsSeenAgain` | systems model baseline | SR-39 |  |
| `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | word overlap | SC-06 | about, module, ui, stay, gate |
| `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | systems model baseline | SC-06 |  |
| `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | language model | SR-18 | The Go test 'EveryModuleParses' verifies the system requirement SR-18 by checking for syntax errors using 'node --check'. The verification definitions VC_SR_18 and VC_SC_04 are also relevant as they mention gates that catch unsupported syntax. |
| `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | language model | VC_SR_18 | The Go test 'EveryModuleParses' verifies the system requirement SR-18 by checking for syntax errors using 'node --check'. The verification definitions VC_SR_18 and VC_SC_04 are also relevant as they mention gates that catch unsupported syntax. |
| `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | language model | VC_SC_04 | The Go test 'EveryModuleParses' verifies the system requirement SR-18 by checking for syntax errors using 'node --check'. The verification definitions VC_SR_18 and VC_SC_04 are also relevant as they mention gates that catch unsupported syntax. |
| `examples/pipeline/ui/ui_test.go#TestPureModulesUnderNode` | word overlap | SC-05 | module, node, app, them, es |
| `examples/pipeline/ui/ui_test.go#TestPureModulesUnderNode` | systems model baseline | SC-05 |  |
| `internal/assert/assert_test.go#TestContains` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestContains` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestContainsAcceptsANamedSliceType` | word overlap | SC-04 | assert, internal, nam |
| `internal/assert/assert_test.go#TestContainsAcceptsANamedSliceType` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestDeepEqual` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestDeepEqual` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnErrors` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnErrors` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnErrors` | language model | CHK_TraceCorrelation | The test 'DeepEqualDoesNotPanicOnErrors' is linked to the verification definitions 'CHK_TraceCorrelation' and 'CHK_Us07ReorderAndNest' based on the evidence found in the system model. |
| `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnErrors` | language model | CHK_Us07ReorderAndNest | The test 'DeepEqualDoesNotPanicOnErrors' is linked to the verification definitions 'CHK_TraceCorrelation' and 'CHK_Us07ReorderAndNest' based on the evidence found in the system model. |
| `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnUnexportedFields` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnUnexportedFields` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestEqualFails` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestEqualFails` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestEqualPasses` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestEqualPasses` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestError` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestError` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestErrorAs` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestErrorAs` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestErrorIs` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestErrorIs` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestLen` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestLen` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestMapEqual` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestMapEqual` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestMust` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestMust` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestNoError` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestNoError` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestNotEqual` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestNotEqual` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestNotEqual` | language model | SR-04 | The test 'NotEqual' is likely verifying that the system does not redirect to a specific URL, as indicated by the requirement SR-04. |
| `internal/assert/assert_test.go#TestSliceEqual` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestSliceEqual` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestSliceEqual` | language model | CHK_RouterRootRedirect | The Go test 'SliceEqual' is likely verifying the 'CHK_RouterRootRedirect' action, as it involves a redirect which could be verified by a slice comparison. |
| `internal/assert/assert_test.go#TestTrue` | word overlap | SC-04 | assert, internal |
| `internal/assert/assert_test.go#TestTrue` | systems model baseline | SC-04 |  |
| `internal/assert/assert_test.go#TestTrue` | language model | CHK_Us09EditFromDocument | The test 'True' is linked to the verification definition CHK_Us09EditFromDocument, which verifies the requirement US-09. |
| `internal/scenario/scenario_test.go#TestElsewhereKindsAreNotGoTestsOrScenarios` | word overlap | SR-46 | scenario |
| `internal/scenario/scenario_test.go#TestElsewhereKindsAreNotGoTestsOrScenarios` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestElsewhereKindsAreNotGoTestsOrScenarios` | language model | SR-30 | The test 'ElsewhereKindsAreNotGoTestsOrScenarios' mentions 'verdicts and reasons', which aligns with the verification definition SR-30. |
| `internal/scenario/scenario_test.go#TestReadFindsEachScenariosCriterionAndItsEvidence` | word overlap | SR-46 | scenario, criterion |
| `internal/scenario/scenario_test.go#TestReadFindsEachScenariosCriterionAndItsEvidence` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestReadNamesAFileGodogCannotParse` | word overlap | SR-46 | scenario, name |
| `internal/scenario/scenario_test.go#TestReadNamesAFileGodogCannotParse` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestReadNamesAFileGodogCannotParse` | language model | CHK_Us02ReadTheModel | The test name 'ReadNamesAFileGodogCannotParse' suggests it is related to reading a model, which is verified by CHK_Us02ReadTheModel. |
| `internal/scenario/scenario_test.go#TestReadReportsAFileWithNothingToRun` | word overlap | SR-46 | scenario, run |
| `internal/scenario/scenario_test.go#TestReadReportsAFileWithNothingToRun` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestReadReportsWhatTheAgreementCannotRead` | word overlap | SR-46 | scenario |
| `internal/scenario/scenario_test.go#TestReadReportsWhatTheAgreementCannotRead` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestRunRunsEachCriterionInASubtestOfItsStory` | word overlap | SR-46 | scenario, story, criterion, run |
| `internal/scenario/scenario_test.go#TestRunRunsEachCriterionInASubtestOfItsStory` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestSuiteFailsAStepWithNoDefinition` | word overlap | SR-46 | scenario, fail |
| `internal/scenario/scenario_test.go#TestSuiteFailsAStepWithNoDefinition` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestSuiteGivesEachExampleAFreshWorld` | word overlap | SR-46 | scenario |
| `internal/scenario/scenario_test.go#TestSuiteGivesEachExampleAFreshWorld` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestSuiteRunsNothingForATagNoScenarioCarries` | word overlap | SR-46 | scenario, carry, run |
| `internal/scenario/scenario_test.go#TestSuiteRunsNothingForATagNoScenarioCarries` | systems model baseline | SR-46 |  |
| `internal/scenario/scenario_test.go#TestSuiteRunsNothingForATagNoScenarioCarries` | language model | SR-32 | The test name 'SuiteRunsNothingForATagNoScenarioCarries' directly corresponds to the verification definition 'VC_SR_32', which verifies the system requirement 'SR-32'. |
| `internal/scenario/scenario_test.go#TestSuiteRunsTheTaggedScenarioWithItsSteps` | word overlap | SR-46 | scenario, run |
| `internal/scenario/scenario_test.go#TestSuiteRunsTheTaggedScenarioWithItsSteps` | systems model baseline | SR-46 |  |
| `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | word overlap | SC-04 | tabletest, internal |
| `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | systems model baseline | SC-04 |  |
| `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | language model | SR-35 | The test 'RunErrExercisesEveryCase' is linked to the verification case VC_SR_35, which verifies the requirement SR-35. The test also exercises other verification cases and validation cases. |
| `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | language model | VC_SR_29 | The test 'RunErrExercisesEveryCase' is linked to the verification case VC_SR_35, which verifies the requirement SR-35. The test also exercises other verification cases and validation cases. |
| `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | language model | VC_SR_35 | The test 'RunErrExercisesEveryCase' is linked to the verification case VC_SR_35, which verifies the requirement SR-35. The test also exercises other verification cases and validation cases. |
| `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | language model | VAL_US_13 | The test 'RunErrExercisesEveryCase' is linked to the verification case VC_SR_35, which verifies the requirement SR-35. The test also exercises other verification cases and validation cases. |
| `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | language model | US-13 | The test 'RunErrExercisesEveryCase' is linked to the verification case VC_SR_35, which verifies the requirement SR-35. The test also exercises other verification cases and validation cases. |
| `internal/tabletest/tabletest_test.go#TestRunErrMatchesTheNamedSentinel` | word overlap | SC-04 | tabletest, internal, nam |
| `internal/tabletest/tabletest_test.go#TestRunErrMatchesTheNamedSentinel` | systems model baseline | SC-04 |  |
| `internal/tabletest/tabletest_test.go#TestRunExercisesEveryCase` | word overlap | SC-04 | tabletest, internal |
| `internal/tabletest/tabletest_test.go#TestRunExercisesEveryCase` | systems model baseline | SC-04 |  |
| `internal/trace/scenarios_test.go#TestScenarios` | word overlap | SR-46 | scenario, acceptance, criterion, run |
| `internal/trace/scenarios_test.go#TestScenarios` | systems model baseline | SR-46 |  |

## For blind judgement

The same proposals again, each once, in a scrambled order and with nothing to say where they came from. Judge these first, then compare with the reasons further up.

| Number | Test | Proposed | Correct? |
|---|---|---|---|
| 1 | `examples/pipeline/document/tree/tree_test.go#TestRequirementLookup` | SR-33 | |
| 2 | `checkly/scenarios_test.go#TestScenarios` | SR-46 | |
| 3 | `adapter/model/model_test.go#TestAttributeShapes` | US-08 | |
| 4 | `adapter/model/model_test.go#TestBuildRefusals` | SR-45 | |
| 5 | `adapter/model/model_test.go#TestBuildRefusals` | SR-26 | |
| 6 | `examples/pipeline/document/scenarios_test.go#TestScenarios` | SR-46 | |
| 7 | `adapter/model/scenarios_test.go#TestScenarios` | SR-46 | |
| 8 | `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | VC_SR_15 | |
| 9 | `examples/pipeline/capacity/capacity_test.go#TestHealth` | SR-28 | |
| 10 | `internal/tabletest/tabletest_test.go#TestRunErrMatchesTheNamedSentinel` | SC-04 | |
| 11 | `internal/scenario/scenario_test.go#TestSuiteRunsTheTaggedScenarioWithItsSteps` | SR-46 | |
| 12 | `internal/scenario/scenario_test.go#TestSuiteGivesEachExampleAFreshWorld` | SR-46 | |
| 13 | `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | VC_SR_18 | |
| 14 | `adapter/serve/server_test.go#TestHandlerServesPostAndHealth` | SR-01 | |
| 15 | `adapter/serve/server_test.go#TestHandlerServesPostAndHealth` | SR-02 | |
| 16 | `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | UC_04_RaiseTheBottleneck::raiseParse | |
| 17 | `adapter/model/model_test.go#TestLinkRefusals` | CHK_Refusals | |
| 18 | `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | SR-18 | |
| 19 | `internal/assert/assert_test.go#TestNotEqual` | SR-04 | |
| 20 | `adapter/syntax/parser_test.go#TestParseUnitIsAQualifiedName` | SR-18 | |
| 21 | `adapter/syntax/parser_test.go#TestParseUnitIsAQualifiedName` | SR-21 | |
| 22 | `adapter/syntax/scenarios_test.go#TestScenarios` | SR-46 | |
| 23 | `adapter/syntax/lexer_test.go#TestLexSpansAreByteOffsets` | SR-18 | |
| 24 | `internal/scenario/scenario_test.go#TestSuiteRunsNothingForATagNoScenarioCarries` | SR-46 | |
| 25 | `internal/scenario/scenario_test.go#TestSuiteRunsNothingForATagNoScenarioCarries` | SR-32 | |
| 26 | `examples/pipeline/document/tree/tree_test.go#TestLoadRefusesDuplicateIDs` | VC_SR_19 | |
| 27 | `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | Router | |
| 28 | `adapter/model/model_test.go#TestLinkRefusals` | SR-45 | |
| 29 | `adapter/model/model_test.go#TestLinkRefusals` | SR-26 | |
| 30 | `cmd/sysml-federation/main_test.go#TestServerDrainIsBoundedByItsTimeout` | SR-47 | |
| 31 | `cmd/sysml-federation/main_test.go#TestServerDrainIsBoundedByItsTimeout` | SR-24 | |
| 32 | `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | US-02 | |
| 33 | `adapter/syntax/lexer_test.go#TestLexKindsAndText` | SR-18 | |
| 34 | `adapter/syntax/lexer_test.go#TestLexKindsAndText` | SR-22 | |
| 35 | `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | VC_SR_29 | |
| 36 | `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | VC_SR_35 | |
| 37 | `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | SR-18 | |
| 38 | `internal/assert/assert_test.go#TestTrue` | SC-04 | |
| 39 | `adapter/syntax/lexer_test.go#TestLexErrorsCarryFileLineAndColumn` | SR-18 | |
| 40 | `internal/assert/assert_test.go#TestContains` | SC-04 | |
| 41 | `internal/tabletest/tabletest_test.go#TestRunExercisesEveryCase` | SC-04 | |
| 42 | `examples/pipeline/document/tree/tree_test.go#TestRefusals` | SR-24 | |
| 43 | `examples/pipeline/document/tree/tree_test.go#TestRefusals` | SR-25 | |
| 44 | `examples/pipeline/document/tree/tree_test.go#TestRefusals` | SR-33 | |
| 45 | `examples/pipeline/ui/ui_test.go#TestBothAppsRefetchWhenTheTabIsSeenAgain` | SR-39 | |
| 46 | `adapter/model/model_test.go#TestBuildRefusals` | CHK_Refusals | |
| 47 | `examples/pipeline/capacity/capacity_test.go#TestHealth` | VC_SR_48 | |
| 48 | `examples/pipeline/capacity/capacity_test.go#TestHealth` | VC_SR_04 | |
| 49 | `internal/scenario/scenario_test.go#TestElsewhereKindsAreNotGoTestsOrScenarios` | SR-30 | |
| 50 | `internal/scenario/scenario_test.go#TestElsewhereKindsAreNotGoTestsOrScenarios` | SR-46 | |
| 51 | `adapter/syntax/parser_test.go#TestExampleModelParses` | SR-17 | |
| 52 | `adapter/syntax/parser_test.go#TestExampleModelParses` | SR-18 | |
| 53 | `adapter/model/model_test.go#TestAttributeShapes` | CHK_Us08ShapeTheDocument | |
| 54 | `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | SR-03 | |
| 55 | `examples/pipeline/ui/scenarios_test.go#TestScenarios` | SC-06 | |
| 56 | `examples/pipeline/ui/ui_test.go#TestANodeListKeepsRoomPastItsLastRow` | SR-37 | |
| 57 | `cmd/sysml-federation/main_test.go#TestRouterFromEnvReadsTheThreePaths` | CHK_RouterJoin | |
| 58 | `adapter/projection/projection_test.go#TestAnUntypedPartPublishesNoDefinition` | SR-16 | |
| 59 | `examples/pipeline/document/document_test.go#TestEntityAnswersForUnknownIDs` | VAL_US_11 | |
| 60 | `examples/pipeline/document/document_test.go#TestEntityAnswersForUnknownIDs` | CHK_Us11QueryTheGraph | |
| 61 | `examples/pipeline/document/tree/tree_test.go#TestLoadRefusesDuplicateIDs` | SR-33 | |
| 62 | `examples/pipeline/document/tree/tree_test.go#TestRefusals` | US-15 | |
| 63 | `adapter/syntax/lexer_test.go#TestReservedWordsAreKeywords` | SR-18 | |
| 64 | `adapter/syntax/lexer_test.go#TestReservedWordsAreKeywords` | SR-11 | |
| 65 | `adapter/projection/projection_test.go#TestQueriesServeEveryFieldOfTheProjection` | SR-43 | |
| 66 | `adapter/projection/projection_test.go#TestQueriesServeEveryFieldOfTheProjection` | SR-22 | |
| 67 | `internal/assert/assert_test.go#TestMapEqual` | SC-04 | |
| 68 | `internal/assert/assert_test.go#TestErrorAs` | SC-04 | |
| 69 | `examples/pipeline/document/tree/tree_test.go#TestRequirementLookup` | VC_SR_37 | |
| 70 | `adapter/projection/scenarios_test.go#TestScenarios` | SR-46 | |
| 71 | `adapter/model/eval_test.go#TestEvalRefusals` | VAL_US_15 | |
| 72 | `adapter/model/model_test.go#TestAttributeShapes` | VAL_US_08 | |
| 73 | `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | VC_SC_03 | |
| 74 | `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | VC_SR_01 | |
| 75 | `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | CHK_RouterPlayground::expectAnHtmlPage | |
| 76 | `examples/pipeline/document/tree/tree_test.go#TestRefusals` | CHK_Refusals | |
| 77 | `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | VC_SR_07 | |
| 78 | `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | VC_SR_06 | |
| 79 | `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | VAL_US_13 | |
| 80 | `adapter/model/model_test.go#TestRelationships` | SR-45 | |
| 81 | `adapter/model/model_test.go#TestRelationships` | SR-26 | |
| 82 | `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | SR-35 | |
| 83 | `examples/pipeline/isolation_test.go#TestEveryEditorialOperationIsAcceptedThroughTheGraph` | SR-35 | |
| 84 | `internal/scenario/scenario_test.go#TestReadNamesAFileGodogCannotParse` | CHK_Us02ReadTheModel | |
| 85 | `internal/scenario/scenario_test.go#TestRunRunsEachCriterionInASubtestOfItsStory` | SR-46 | |
| 86 | `examples/pipeline/capacity/flow/flow_test.go#TestRollupNamesTheFaultyChild` | SR-28 | |
| 87 | `adapter/model/model_test.go#TestPartsTreeAttributesAndPorts` | SR-16 | |
| 88 | `adapter/model/model_test.go#TestPartsTreeAttributesAndPorts` | SR-45 | |
| 89 | `internal/scenario/scenario_test.go#TestReadReportsAFileWithNothingToRun` | SR-46 | |
| 90 | `adapter/syntax/parser_test.go#TestParseRequirements` | VC_SR_38 | |
| 91 | `adapter/syntax/parser_test.go#TestParseRequirements` | VC_SR_37 | |
| 92 | `internal/assert/assert_test.go#TestNotEqual` | SC-04 | |
| 93 | `cmd/sysml-federation/main_test.go#TestUIRefusesADirectoryWithNoPage` | SC-04 | |
| 94 | `adapter/syntax/parser_test.go#TestParsePortsConnectAndSatisfy` | US-11 | |
| 95 | `cmd/sysml-federation/main_test.go#TestRunDispatchesSubcommandsAndExitCodes` | SR-47 | |
| 96 | `cmd/sysml-federation/main_test.go#TestRunDispatchesSubcommandsAndExitCodes` | SR-01 | |
| 97 | `cmd/sysml-federation/main_test.go#TestUIRefusesToListTheSharedDirectory` | CHK_Refusals | |
| 98 | `adapter/model/model_test.go#TestLinkRefusals` | VAL_US_15 | |
| 99 | `adapter/model/eval_test.go#TestEvalRefusals` | VC_SR_18 | |
| 100 | `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | SR-15 | |
| 101 | `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | SR-03 | |
| 102 | `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | SR-47 | |
| 103 | `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | CHK_SessionHeartbeat | |
| 104 | `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | VAL_US_02 | |
| 105 | `adapter/syntax/parser_test.go#TestParseDerivationAndVerification` | SR-18 | |
| 106 | `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | SR-25 | |
| 107 | `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | SR-18 | |
| 108 | `adapter/syntax/parser_test.go#TestParseExpressionPrecedenceAndSpans` | VC_SR_25 | |
| 109 | `examples/pipeline/document/document_test.go#TestEntityAnswersForUnknownIDs` | SR-27 | |
| 110 | `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnErrors` | CHK_TraceCorrelation | |
| 111 | `adapter/serve/scenarios_test.go#TestScenarios` | SR-46 | |
| 112 | `examples/pipeline/document/scenarios_test.go#TestScenarios` | SC-06 | |
| 113 | `examples/pipeline/document/scenarios_test.go#TestScenarios` | SC-04 | |
| 114 | `adapter/model/patch_internal_test.go#TestPatchGuardsItsInputs` | SR-22 | |
| 115 | `adapter/model/patch_internal_test.go#TestPatchGuardsItsInputs` | SR-25 | |
| 116 | `examples/pipeline/capacity/flow/scenarios_test.go#TestScenarios` | SR-46 | |
| 117 | `internal/assert/assert_test.go#TestErrorIs` | SC-04 | |
| 118 | `examples/pipeline/scenarios_test.go#TestScenarios` | SR-46 | |
| 119 | `adapter/model/eval_test.go#TestEvalRefusals` | SR-26 | |
| 120 | `adapter/model/eval_test.go#TestEvalRefusals` | SR-24 | |
| 121 | `adapter/model/eval_test.go#TestEvalRefusals` | SR-25 | |
| 122 | `cmd/sysml-federation/scenarios_test.go#TestScenarios` | req-AD-0034 | |
| 123 | `examples/pipeline/capacity/scenarios_test.go#TestScenarios` | demo::capacity | |
| 124 | `internal/assert/assert_test.go#TestSliceEqual` | CHK_RouterRootRedirect | |
| 125 | `internal/assert/assert_test.go#TestTrue` | CHK_Us09EditFromDocument | |
| 126 | `adapter/model/model_test.go#TestAttributeShapes` | SR-45 | |
| 127 | `adapter/model/model_test.go#TestAttributeShapes` | SR-26 | |
| 128 | `adapter/projection/projection_test.go#TestUnknownIDsAreErrors` | SR-22 | |
| 129 | `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | VC_SC_04 | |
| 130 | `adapter/syntax/parser_test.go#TestParseRequirements` | SR-18 | |
| 131 | `cmd/sysml-federation/main_test.go#TestServeFailsWhenTheRouterDies` | Router | |
| 132 | `examples/pipeline/capacity/capacity_test.go#TestLeafCapacityIsItsOwnAttribute` | SR-28 | |
| 133 | `cmd/sysml-federation/main_test.go#TestServeHandsTheRouterItsConfigurationFile` | SR-03 | |
| 134 | `cmd/sysml-federation/main_test.go#TestServeHandsTheRouterItsConfigurationFile` | SR-02 | |
| 135 | `internal/scenario/scenario_test.go#TestSuiteFailsAStepWithNoDefinition` | SR-46 | |
| 136 | `cmd/sysml-federation/main_test.go#TestHelperRouter` | SR-03 | |
| 137 | `internal/scenario/scenario_test.go#TestReadFindsEachScenariosCriterionAndItsEvidence` | SR-46 | |
| 138 | `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | SC-06 | |
| 139 | `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | SR-48 | |
| 140 | `cmd/sysml-federation/main_test.go#TestStopBudgetFitsAContainerGrace` | SR-06 | |
| 141 | `examples/pipeline/ui/ui_test.go#TestPureModulesUnderNode` | SC-05 | |
| 142 | `internal/assert/assert_test.go#TestDeepEqual` | SC-04 | |
| 143 | `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | US-13 | |
| 144 | `internal/scenario/scenario_test.go#TestReadNamesAFileGodogCannotParse` | SR-46 | |
| 145 | `cmd/sysml-federation/main_test.go#TestUIRefusesToListTheSharedDirectory` | SR-44 | |
| 146 | `internal/assert/assert_test.go#TestNoError` | SC-04 | |
| 147 | `internal/scenario/scenario_test.go#TestReadReportsWhatTheAgreementCannotRead` | SR-46 | |
| 148 | `cmd/sysml-federation/main_test.go#TestAddressesMatchTheComposedConfiguration` | SR-03 | |
| 149 | `adapter/projection/scenarios_test.go#TestScenarios` | SC-04 | |
| 150 | `internal/assert/assert_test.go#TestSliceEqual` | SC-04 | |
| 151 | `internal/assert/assert_test.go#TestMust` | SC-04 | |
| 152 | `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | VC_SR_02 | |
| 153 | `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess` | VC_SR_09 | |
| 154 | `cmd/sysml-federation/main_test.go#TestHealthcheckProbesThePublishedPort` | SR-10 | |
| 155 | `cmd/sysml-federation/main_test.go#TestHealthcheckProbesThePublishedPort` | SR-04 | |
| 156 | `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnUnexportedFields` | SC-04 | |
| 157 | `adapter/syntax/parser_test.go#TestParseDefinitionsAndParts` | SR-18 | |
| 158 | `internal/assert/assert_test.go#TestLen` | SC-04 | |
| 159 | `cmd/sysml-federation/scenarios_test.go#TestScenarios` | SR-46 | |
| 160 | `internal/assert/assert_test.go#TestEqualFails` | SC-04 | |
| 161 | `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnErrors` | CHK_Us07ReorderAndNest | |
| 162 | `adapter/model/model_test.go#TestLinkRefusals` | VC_SR_18 | |
| 163 | `examples/pipeline/document/document_test.go#TestARefusedMutationIsAnErrorAndLeavesTheVersion` | SR-18 | |
| 164 | `examples/pipeline/document/document_test.go#TestARefusedMutationIsAnErrorAndLeavesTheVersion` | SR-27 | |
| 165 | `examples/pipeline/document/document_test.go#TestARefusedMutationIsAnErrorAndLeavesTheVersion` | SR-25 | |
| 166 | `cmd/sysml-federation/main_test.go#TestServeRefusesAModelItCannotRead` | SR-47 | |
| 167 | `internal/trace/scenarios_test.go#TestScenarios` | SR-46 | |
| 168 | `examples/pipeline/capacity/capacity_test.go#TestHealth` | CHK_RouterHealthNotProxied | |
| 169 | `internal/tabletest/tabletest_test.go#TestRunErrExercisesEveryCase` | SC-04 | |
| 170 | `adapter/syntax/parser_test.go#TestParseNegativeLiteralSpanIncludesTheSign` | SR-18 | |
| 171 | `examples/pipeline/scenarios_test.go#TestScenarios` | SC-04 | |
| 172 | `adapter/scenarios_test.go#TestScenarios` | SR-46 | |
| 173 | `cmd/sysml-federation/main_test.go#TestStopRouterKillsAChildThatIgnoresSIGTERM` | SR-03 | |
| 174 | `examples/pipeline/ui/scenarios_test.go#TestScenarios` | SR-46 | |
| 175 | `adapter/model/eval_test.go#TestEvalRefusals` | CHK_Refusals | |
| 176 | `internal/assert/assert_test.go#TestContainsAcceptsANamedSliceType` | SC-04 | |
| 177 | `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnErrors` | SC-04 | |
| 178 | `examples/pipeline/document/tree/tree_test.go#TestEditText` | SR-33 | |
| 179 | `adapter/syntax/parser_test.go#TestParseAttributeWithBody` | VC_SR_28::aLeafReadsItsOwnAttributeScenario | |
| 180 | `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | VC_SR_41 | |
| 181 | `adapter/syntax/parser_test.go#TestParsePackageHeaderAndImports` | VC_SR_36 | |
| 182 | `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | SR-47 | |
| 183 | `cmd/sysml-federation/main_test.go#TestSubcommandsServeOneComponent` | SR-01 | |
| 184 | `adapter/projection/projection_test.go#TestEntitiesResolveTheThreeKeyedTypes` | SR-43 | |
| 185 | `adapter/projection/projection_test.go#TestEntitiesResolveTheThreeKeyedTypes` | SR-22 | |
| 186 | `examples/pipeline/ui/ui_test.go#TestEveryModuleParses` | SR-18 | |
| 187 | `cmd/sysml-federation/main_test.go#TestRouterFromEnvReadsTheThreePaths` | SR-03 | |
| 188 | `adapter/syntax/lexer_test.go#TestPosition` | SR-18 | |
| 189 | `examples/pipeline/capacity/scenarios_test.go#TestScenarios` | SR-28 | |
| 190 | `examples/pipeline/capacity/scenarios_test.go#TestScenarios` | SR-46 | |
| 191 | `internal/assert/assert_test.go#TestError` | SC-04 | |
| 192 | `cmd/sysml-federation/main_test.go#TestHealthcheckFailsWhenEitherProbeFails` | SR-47 | |
| 193 | `cmd/sysml-federation/main_test.go#TestHealthcheckFailsWhenEitherProbeFails` | SR-15 | |
| 194 | `internal/assert/assert_test.go#TestEqualPasses` | SC-04 | |
| 195 | `adapter/syntax/parser_test.go#TestParsePortsConnectAndSatisfy` | SR-20 | |
| 196 | `adapter/syntax/parser_test.go#TestParsePortsConnectAndSatisfy` | SR-18 | |
| 197 | `cmd/sysml-federation/main_test.go#TestUIRefusesToListTheSharedDirectory` | SC-01 | |
| 198 | `examples/pipeline/capacity/flow/scenarios_test.go#TestScenarios` | req_AD-0034 | |

## Time

| Probe | Questions | Mean seconds | Mean prompt tokens | Largest prompt |
|---|---|---|---|---|
| base | 1500 | 7.8 | 834 | 1149 |
| incident | 105 | 8.0 | 827 | 914 |
| reconstruction | 117 | 7.5 | 806 | 1053 |
| deletion | 111 | 7.1 | 815 | 1055 |
| control | 117 | 7.8 | 833 | 1134 |
| rare-shared | 99 | 8.1 | 850 | 1078 |
| repeat | 384 | 7.7 | 834 | 1142 |

Total time spent on questions: 315 minutes.
