# Outcomes, 2026-10-01T151851Z-full

This file sets the figures of this run beside the four results that [the article that set the experiment out](../../../../docs/articles/18-a-resolver-that-reads-the-systems-model.md#what-will-count-for-and-against-it) said would count for the approach and against it, read as [the README reads the article's rules](../../README.md#reading-the-articles-rules). A path such as `paired_systems_model.only_baseline` names a field of the summary line, the last line of [run.jsonl](run.jsonl), with array entries counted from 0, and a figure the summary line lacks comes from the call lines or from another file of the run, computed as stated where it is given.

## Checks on the totals

Unless a value is given as recorded, shares and interval bounds are rounded to three decimal places. Every p is given in full, as the summary line records it, and is compared with 0.05 at that precision. In a rate, the share after k of n is computed from k and n, and the interval is the one the summary line records.

Each figure taken from the summary line was also computed again from the baseline lines and the call lines of run.jsonl, following the code at the run's commit, and the two agree to the precision of floating-point arithmetic. The checks the run recorded on its call lines, such as the key items found, the citations and each link's distance from the recorded requirement, are taken as recorded. A first browse is a browse of the probe `base`. The totals agree with each other.

| Check | Sum | Against | Holds |
|---|---|---|---|
| `paired_systems_model`: both_right + only_language_model + only_baseline + neither_right | 11 + 6 + 31 + 21 = 69 | `keyed_tests` = 69 | yes |
| `paired`: both_right + only_language_model + only_word_overlap + neither_right | 13 + 4 + 28 + 24 = 69 | `keyed_tests` = 69 | yes |
| `probe_pairs[0]`, the random control: both_changed + only_deletion + only_control + neither | 6 + 6 + 1 + 0 = 13 | `probe_pairs[0].asked` = 13 | yes |
| `probe_pairs[1]`, the rare-shared control: both_changed + only_deletion + only_control + neither | 5 + 5 + 1 + 0 = 11 | `probe_pairs[1].asked` = 11 | yes |
| language model correct: both_right + only_language_model, in `paired_systems_model` and in `paired` | 11 + 6 = 17, and 13 + 4 = 17 | `resolvers[4].correct` = 17 | yes |
| systems model baseline correct: `paired_systems_model` both_right + only_baseline | 11 + 31 = 42 | `resolvers[3].correct` = 42 | yes |
| word overlap's best-ranked answer correct: `paired` both_right + only_word_overlap | 13 + 28 = 41 | `resolvers[2].correct` = 41 | yes |
| deletion changed, within `probe_pairs[0]`: both_changed + only_deletion | 6 + 6 = 12, of 13 asked | `probes[0].changed` = 12 of 13 | yes |
| random control changed, within `probe_pairs[0]`: both_changed + only_control | 6 + 1 = 7, of 13 asked | `probes[1].changed` = 7 of 13 | yes |
| rare-shared control changed, within `probe_pairs[1]`: both_changed + only_control | 5 + 1 = 6, of 11 asked | `probes[2].changed` = 6 of 11 | yes |
| deletion changed, within `probe_pairs[1]`: both_changed + only_deletion | 5 + 5 = 10, of 11 asked | no more than `probes[0].changed.k` = 12 | yes |
| to_none + to_another, `probes[0]` to `probes[3]` | 11 + 1 = 12, 5 + 2 = 7, 4 + 2 = 6, 9 + 2 = 11 | changed.k: 12, 7, 6, 11 | yes |
| to_none + to_another, `probes[5]` to `probes[8]`, correct links only | 3 + 0 = 3, 1 + 0 = 1, 1 + 0 = 1, 1 + 1 = 2 | changed.k: 3, 1, 1, 2 | yes |
| changed.n + skipped + errors, `probes[0]` to `probes[4]` | 13 + 0 + 0 = 13, 13 + 0 + 0 = 13, 11 + 2 + 0 = 13, 13 + 0 + 0 = 13, 43 + 0 + 0 = 43 | final call lines of each probe, counted: 13, 13, 13, 13, 43 | yes |
| `other_kinds`, links summed | 31 + 6 + 3 + 2 + 1 = 43 | `other_near.n` = 43, and each `near_by_distance[i].near.n` = 43 | yes |
| `other_kinds`, near summed | 31 + 6 + 1 + 2 + 1 = 41 | `other_near.k` = 41, and `near_by_distance[2].near.k` = 41 | yes |
| `base_rate` | 0.35056626610184155 | `near_by_distance[2].every_element` = 0.35056626610184155 | yes |
| items marked found in `incident[i].items`, for each browse | 0, 0, 0, 0, 0 | `incident[i].found`: 0, 0, 0, 0, 0 | yes |
| `tests` and `keyed_tests` | 169 and 69 | the header's `tests` and `keyed_tests`, 169 and 69, and the baseline lines and the first browses' final call lines, 169 and 169 | yes |
| entries of `proposals`, by resolver: word overlap, systems model baseline, language model | 98 + 100 + 85 = 283 | entries of `proposals` = 283 | yes |
| entries of `blind` | 198 | rows of judgements.csv = 198, each with the number, test and proposed element of one entry | yes |

## 1. A gain over the systems model baseline

The article's first rule, from [What will count for and against it](../../../../docs/articles/18-a-resolver-that-reads-the-systems-model.md#what-will-count-for-and-against-it), reads as follows.

> **A gain over the systems model baseline.** If McNemar's test finds the language model correct more often where the two disagree, its choices added something that following the links alone didn't. If it finds no difference, part 7 will say that on this data the language model gained nothing measurable over a program following the same links. On 69 tests that's weaker than showing it has none, and I'll say that too. A gain over word overlap alone doesn't change this, since it could come from the systems model and not the language model.

| Figure | Path | Value |
|---|---|---|
| Tests that carry a key | `keyed_tests` | 69 |
| First browses that ended without a final answer | `base_errors` | 0 |
| Both correct | `paired_systems_model.both_right` | 11 |
| Only the language model correct, b | `paired_systems_model.only_language_model` | 6 |
| Only the systems model baseline correct, c | `paired_systems_model.only_baseline` | 31 |
| Neither correct | `paired_systems_model.neither_right` | 21 |
| McNemar's exact p | `paired_systems_model.mcnemar_exact_p` | 0.000041257590055465746 |

| Resolver | Path | Proposed | Correct | Precision | Recall |
|---|---|---|---|---|---|
| language model | `resolvers[4]` | 25 | 17 | 17 of 25, 0.680 (computed), 95% interval 0.484 to 0.828 | 17 of 69, 0.246 (computed), 95% interval 0.160 to 0.360 |
| systems model baseline | `resolvers[3]` | 69 | 42 | 42 of 69, 0.609 (computed), 95% interval 0.491 to 0.715 | 42 of 69, 0.609 (computed), 95% interval 0.491 to 0.715 |

Beside them stands the comparison with the word overlap's best-ranked answer, `paired`. Both were correct on 13 tests (`paired.both_right`), only the language model on 4 (`paired.only_language_model`), only the word overlap's best-ranked answer on 28 (`paired.only_word_overlap`) and neither on 24 (`paired.neither_right`), with p = 0.000019301194697618183 (`paired.mcnemar_exact_p`). The best-ranked answer itself, `resolvers[2]`, proposed 69 and was correct on 41, a precision and recall of 41 of 69, 0.594 (computed), 95% interval 0.476 to 0.702.

In the README's reading, p = 0.000041257590055465746 is below 0.05 and c = 31 is greater than b = 6, so the difference lies in the baseline's favour. The figures meet neither of the rule's stated outcomes. A difference in the baseline's favour is an outcome the rule doesn't name.

## 2. Cited words that carry the answer

The article's second rule, from [What will count for and against it](../../../../docs/articles/18-a-resolver-that-reads-the-systems-model.md#what-will-count-for-and-against-it), reads as follows.

> **Cited words that carry the answer.** Suppose the paired test finds that deleting the cited words changes answers more often than the rare-shared control does, and reconstruction mostly gives the same answer back. Then the words a reviewer is shown are the ones the answer rested on. If deletion changes answers no more often than the control, the evidence isn't what decided the answer, and a reviewer shouldn't be shown it as if it were. Reconstruction that rarely gives the same answer says the same.

| Figure | Path | Value |
|---|---|---|
| Tests where deletion and the rare-shared control were both asked and answered | `probe_pairs[1].asked` | 11 |
| Both changed the answer | `probe_pairs[1].both_changed` | 5 |
| Only deletion changed it, b | `probe_pairs[1].only_deletion` | 5 |
| Only the rare-shared control changed it, c | `probe_pairs[1].only_control` | 1 |
| Neither changed it | `probe_pairs[1].neither` | 0 |
| McNemar's exact p | `probe_pairs[1].mcnemar_exact_p` | 0.21875000000000017 |

Beside it stands the random control, `probe_pairs[0]`, asked and answered with deletion on 13 tests. Both changed the answer on 6, only deletion on 6, only the control on 1 and neither on 0, with p = 0.12499999999999985.

| Probe | Path | Changed | k/n, computed | 95% interval | To none | To another | Skipped | Errors |
|---|---|---|---|---|---|---|---|---|
| deletion | `probes[0]` | 12 of 13 | 0.923 | 0.667 to 0.986 | 11 | 1 | 0 | 0 |
| control | `probes[1]` | 7 of 13 | 0.538 | 0.291 to 0.768 | 5 | 2 | 0 | 0 |
| rare-shared | `probes[2]` | 6 of 11 | 0.545 | 0.280 to 0.787 | 4 | 2 | 2 | 0 |
| reconstruction | `probes[3]` | 11 of 13 | 0.846 | 0.578 to 0.957 | 9 | 2 | 0 | 0 |
| deletion, correct links only | `probes[5]` | 3 of 3 | 1.000 | 0.439 to 1.000 | 3 | 0 | 0 | 0 |
| control, correct links only | `probes[6]` | 1 of 3 | 0.333 | 0.061 to 0.792 | 1 | 0 | 0 | 0 |
| rare-shared, correct links only | `probes[7]` | 1 of 3 | 0.333 | 0.061 to 0.792 | 1 | 0 | 0 | 0 |
| reconstruction, correct links only | `probes[8]` | 2 of 3 | 0.667 | 0.208 to 0.939 | 1 | 1 | 0 | 0 |

The 2 skipped rare-shared probes asked nothing. The final call line of each carries the note "the test and the linked requirement share no uncited word". The control named `control` in the summary line is the random one.

Beside these, `evidence_grounded`, the pieces of evidence found as written, apart from case, in the test's fields on the first browses that linked a requirement, whether the test carries a key or not, is 61 of 91, 0.670 (computed), 95% interval 0.569 to 0.758.

In the README's reading, p = 0.21875000000000017 for deletion against the rare-shared control is not below 0.05, so deletion changes answers no more often than the control. The 95% interval of reconstruction's share of changed answers over all its links, `probes[3].changed`, runs from 0.5776536898051746 to 0.9567418216419189 and lies wholly above one half, so reconstruction rarely gives the same answer back. The figures meet two of the rule's stated outcomes. For deletion the rule says "If deletion changes answers no more often than the control, the evidence isn't what decided the answer, and a reviewer shouldn't be shown it as if it were." For reconstruction it says "Reconstruction that rarely gives the same answer says the same."

## 3. Replies that repeat

The article's third rule, from [What will count for and against it](../../../../docs/articles/18-a-resolver-that-reads-the-systems-model.md#what-will-count-for-and-against-it), reads as follows.

> **Replies that repeat.** If every repeat gives the same reply word for word, a run can be reproduced, which part 5 argued an auditor needs. If one differs, it can't be, as it stands.

| Figure | Path or source | Value |
|---|---|---|
| Repeats whose final reply differs from the first browse's | `probes[4].changed` | 6 of 43, 0.140 (computed), 95% interval 0.066 to 0.273 |
| Repeats skipped | `probes[4].skipped` | 0 |
| Repeats whose final reply was missing or did not parse | `probes[4].errors` | 0 |
| Final call lines of the repeat probe | the call lines with probe `repeat` and final true, counted | 43 |

The 43 final call lines are the number expected, one in four of the 169 tests rounded up, and the tests they repeat are the first test and every fourth after it in the order the tests are read. The summary line counts a repeat as changed when the raw text of its final reply differs from that of the first browse's final reply. Compared again from the call lines in the same way, the 6 that differ are these.

- `adapter/model/fixture_test.go#TestSR16_WarehouseTextVersionAndIdentifiers`
- `adapter/syntax/lexer_test.go#TestLexSpansAreByteOffsets`
- `cmd/sysml-federation/main_test.go#TestRouterRunsAsAChildProcess`
- `cmd/sysml-federation/main_test.go#TestServeHandsTheRouterItsConfigurationFile`
- `internal/assert/assert_test.go#TestDeepEqualDoesNotPanicOnErrors`
- `internal/scenario/scenario_test.go#TestReadNamesAFileGodogCannotParse`

In the README's reading, 6 of the 43 repeats give a final reply that differs, and none ended in an error. One or more differ, so the figures meet the rule's second outcome: "If one differs, it can't be, as it stands."

The incident's "again" browse is no part of this rule. From the call lines, the raw text of its final reply is identical to that of the "as reported" browse.

## 4. An account that follows its own path

The article's fourth rule, from [What will count for and against it](../../../../docs/articles/18-a-resolver-that-reads-the-systems-model.md#what-will-count-for-and-against-it), reads as follows.

> **An account that follows its own path.** If the outage's account names the cause, the mechanism and the first file to read, an engineer could start from it. If it then changes when `onRouterExit` is taken out of the systems model, or says the systems model no longer explains the code, the path it reported is the one it used. If it stays the same, still citing the missing transition, it isn't.

### The five accounts

| Browse | Path | Items found | Cited | Missing | Mechanism | Error |
|---|---|---|---|---|---|---|
| as reported | `incident[0]` | 0 of 12 | 2 | 0 | empty | none |
| again | `incident[1]` | 0 of 12 | 2 | 0 | empty | none |
| alert only | `incident[2]` | 0 of 12 | 4 | 0 | empty | none |
| transition removed | `incident[3]` | 0 of 12 | 12 | 4 | `demo::viewer`, `session::viewer`, `CheckSession::session` | none |
| control link removed | `incident[4]` | 0 of 12 | 12 | 4 | `demo::viewer`, `session::viewer`, `CheckSession::session` | none |

Each of the five final replies was the JSON asked for, so no account counts as one that did not parse.

The missing entries, `incident[3].missing` and `incident[4].missing`, are the same 4 in both browses. The browses "as reported", "again" and "alert only" have none.

| Field | Ref | Why |
|---|---|---|
| consequences | `monitor: the viewer answers` | no such element in what the tools could show |
| code | `sysml-federation serve: router exited: signal: killed` | no such element in what the tools could show |
| path | `session::viewer` | no link joins demo::viewer and session::viewer |
| path | `sysml-federation serve: router exited: signal: killed` | no such element in what the tools could show |

The why of each account, `incident[i].why`, verbatim:

- "as reported" and "again", the same text in each

  > The service's container has stopped, and its last line reads: sysml-federation serve: router exited: signal: killed. The 'demo' part represents the container and its process tree, but there are no 'satisfied by' links to identify the cause.

- "alert only"

  > The alert 'monitor: the viewer answers' failing since 02:10 is likely due to the 'demo::viewer' part not functioning as expected. This part is responsible for satisfying SR-14, which is related to the viewer's verdict and reason. Since there are no 'satisfied by' links for 'demo::viewer', it suggests that the implementation or configuration of this part might be incorrect or missing.

- "transition removed" and "control link removed", the same text in each

  > The service container stopping and the alert failure are likely caused by issues with the viewer components. The 'demo::viewer' and 'session::viewer' parts, along with the 'CheckSession::session' check, are involved in the service's operation and might have failed, leading to the container being killed and the alert not receiving a response.

No entry of `mismatches` has the where "the incident, transition removed", and none comes from any browse of the incident. In the final call lines, all five accounts carry an empty list of suspicions.

### The incident key, item by item

The twelve items are the header's `incident_key.items`, and each cell is `incident[i].items[j].found` for the browse and the item. Each text is that of `incident_key.items[j].item`, with backticks added around the requirement keys `SR-04` and `SR-09`.

| Item | Text | as reported | again | alert only | transition removed | control link removed |
|---|---|---|---|---|---|---|
| 1 | the router is the cause | missed | missed | missed | missed | missed |
| 2 | the supervisor's state machine is the mechanism | missed | missed | missed | missed | missed |
| 3 | the transition on the router's exit | missed | missed | missed | missed | missed |
| 4 | the supervisor's code | missed | missed | missed | missed | missed |
| 5 | the compose file, which restarts nothing | missed | missed | missed | missed | missed |
| 6 | `SR-04`, four paths on one port, is unmet | missed | missed | missed | missed | missed |
| 7 | `SR-09`, every edit lost with the memory it lived in | missed | missed | missed | missed | missed |
| 8 | the document's monitor fails too | missed | missed | missed | missed | missed |
| 9 | the root's monitor fails too | missed | missed | missed | missed | missed |
| 10 | a stakeholder story behind `SR-04` | missed | missed | missed | missed | missed |
| 11 | the stakeholder story behind `SR-09` | missed | missed | missed | missed | missed |
| 12 | the session's heartbeat, which keeps passing | missed | missed | missed | missed | missed |

### Whether an engineer could start from it

On "as reported", item 1, the cause, is missed. Items 2 and 3, the mechanism, are both missed, and so are items 4 and 5, the first file to read. The account names neither the cause, nor the mechanism, nor the first file to read, so the figures do not meet the condition of the rule's first sentence.

### Whether it follows its own path

The found-or-missed pattern of "transition removed" is the same as that of "as reported" in all twelve items, every one missed, so the account does not change. The browse raised no suspicion, so none names `onRouterExit`, `SupervisorStates` or `serve.go`, and the account does not say that the systems model no longer explains the code. Of its 4 missing entries, 3 have the why "no such element in what the tools could show", and their refs, `monitor: the viewer answers` in 1 entry and `sysml-federation serve: router exited: signal: killed` in 2, do not end in `onRouterExit`. The account does not still cite the missing transition.

The account stays the same without citing the missing transition. That meets neither "the path it reported is the one it used" nor "it isn't", and is an outcome the rule doesn't name.

Beside it, compared with "as reported" in the same way and without a reading:

| Browse | Pattern differs from "as reported" | A suspicion names `onRouterExit`, `SupervisorStates` or `serve.go` | A missing entry ends in `onRouterExit` |
|---|---|---|---|
| again | no | no, none raised | no, none missing |
| transition removed | no | no, none raised | no |
| control link removed | no | no, none raised | no |

From the call lines, the raw text of the final reply of "control link removed" is identical to that of "transition removed".

## Other figures

### Reach

The reach, `reach`, is 58 of 69, 0.841 (computed), 95% interval 0.737 to 0.909. It counts the tests that carry a key whose recorded requirement is among the first eight elements the search ranks for the test's own words, or one satisfy, verify or derive link from one of them. It is measured on the test's own words and is no bound on either resolver.

### Near misses

The links to elements other than system requirements, on the first browses of the tests that carry a key, by distance from the recorded requirement:

| Within | Path | Links as near | k/n, computed | 95% interval | Every element | Visited elements |
|---|---|---|---|---|---|---|
| one link | `near_by_distance[0]` | 17 of 43 | 0.395 | 0.264 to 0.544 | 0.009 | 0.178 |
| two links | `near_by_distance[1]` | 25 of 43 | 0.581 | 0.433 to 0.716 | 0.108 | 0.409 |
| three links | `near_by_distance[2]` | 41 of 43 | 0.953 | 0.845 to 0.987 | 0.351 | 0.952 |

Within three links, `other_near` gives the same 41 of 43, 0.953 (computed), 95% interval 0.845 to 0.987, and `base_rate`, the every-element share within three links, is 0.35056626610184155.

| Kind | Path | Links | Within three links |
|---|---|---|---|
| verification def | `other_kinds[0]` | 31 | 31 |
| part | `other_kinds[1]` | 6 | 6 |
| action | `other_kinds[2]` | 3 | 1 |
| requirement | `other_kinds[3]` | 2 | 2 |
| part def | `other_kinds[4]` | 1 | 1 |

The every-element share is defined in [run.go](../../run.go) and averaged in [report.go](../../report.go). For the first browse of a test that carries a key, the program counts the elements of the whole systems model, of every kind and packages among them, that lie within one, two and three links of the recorded requirement by paths that pass through no package, and divides each count by the number of elements less one, 1,345 in this run. The whole systems model holds the verification register's test entries that the test task hides, so those entries count in the divisor, count as near when they lie within the distance, and carry paths to other elements. The visited share is the same count taken over the elements the browse's tool answers named, other than the requirement, and divided by their number. The summary line's shares are the means of these per-browse shares over the links to other kinds, each link counting once, so a browse with three such links counts three times. A link whose identifier names no element gets the kind unknown and no distance: it counts among the links, adds its browse's share to the every-element figure, and is never near. None of the 43 links has the kind unknown. A browse whose tool answers named no element still adds its every-element share for each of its links, and adds 0 to the visited share while counting in its divisor. None of the 43 links in this run comes from such a browse.

On the article's example, the final call line of the first browse of `adapter/model/patch_test.go#TestSR22_SetAttributePatchesTextAndProjectionTogether` records `base_rates` of 0.011895910780669145, 0.12118959107806691 and 0.387360594795539. The header gives 1,346 elements, and each rate multiplied by 1,345 gives the counts below (computed).

| Within | `base_rates` | Multiplied by 1,345 | The article |
|---|---|---|---|
| one link | 0.011895910780669145 | 16 | 16 |
| two links | 0.12118959107806691 | 163 | 157 |
| three links | 0.387360594795539 | 521 | 515 |

In [What will count for and against it](../../../../docs/articles/18-a-resolver-that-reads-the-systems-model.md#what-will-count-for-and-against-it), the article gives its three counts as parts of "the 1,277 elements the test task shows".

### Citation failures

The incident's missing entries, grouped by why, by browse (`incident[i].missing`):

| Why | as reported | again | alert only | transition removed | control link removed |
|---|---|---|---|---|---|
| no such element in what the tools could show | 0 | 0 | 0 | 3 | 3 |
| no link joins demo::viewer and session::viewer | 0 | 0 | 0 | 1 | 1 |
| missing, of all cited | 0 of 2 | 0 of 2 | 0 of 4 | 4 of 12 | 4 of 12 |

The test answers' citations are not in the summary line. Computed from the citations on every final call line that holds a test answer, one citation for each link of the answer, the failures are these.

| Browses | Links cited | Not found |
|---|---|---|
| first browses | 160 | 2 |
| deletion | 8 | 0 |
| random control | 17 | 1 |
| rare-shared control | 23 | 0 |
| reconstruction | 14 | 0 |
| repeat | 50 | 1 |
| all | 272 | 4 |

All 4 have the why "no such element in what the tools could show".

| Browse | Test | Ref |
|---|---|---|
| first browses | `cmd/sysml-federation/scenarios_test.go#TestScenarios` | `req-AD-0034` |
| first browses | `examples/pipeline/capacity/flow/scenarios_test.go#TestScenarios` | `req_AD-0034` |
| random control | `examples/pipeline/isolation_test.go#TestEveryEditorialOperationIsAcceptedThroughTheGraph` | `VC_SR_48::Operations` |
| repeat | `cmd/sysml-federation/scenarios_test.go#TestScenarios` | `req-AD-0034` |

The check is `CheckCitations` and `cite` in [incident.go](../../incident.go), and `CheckTestCitations` for a test answer. A reference with a slash in it is taken as a file and any other as an element. A file passes when a file or folder of that path exists anywhere in the checkout, once a trailing line number or line range and a leading `./` are taken off. An element passes when the systems model as the browse's tools saw it holds it, found by the lookup the tools use, which drops surrounding spaces, quotes and backticks, reads a dot as `::` and accepts a suffix of a qualified name that only one element has. In an account, each step of the path after the first that passes on its own must also join the step before it. Two elements join when they are the same element or when any link joins them, in either direction. An element and a file join, in either order, when the file is among the places the `code` tool gives for the element. Two files join when the second was read, or shown by `grep` or `code`, in the same browse. A test answer's check applies the element and file test to each of its links, and nothing more.

In [Two tasks](../../../../docs/articles/18-a-resolver-that-reads-the-systems-model.md#two-tasks), the article asks that every element, link and file an account cites be "one the browse's tools could have shown", and that each step of the path join the next "through a link in the systems model, a file the `code` tool gives for the element, or a file the browse read". For an element, the check asks the same. It is more lenient than the article in these places.

- A file need only exist in the checkout. `grep` and `read` show only the built system's code and `code` names only the places the systems model gives, so a file or folder that no tool could have shown passes too. A line number past the end of the file passes, and the path is not held inside the checkout, so one that climbs out of it with `..` is looked up outside it.
- The same element twice in a row passes as a join, with no link between.
- For an element and a file, the file may be any of the places `code` gives for the element, including those cut from the tool's answer, which shows eight lines at most.
- For two files, a file shown in an answer of `grep` or `code` counts as read.

An account has no field for links, so the only links checked are the joins between the steps of its path.

### Suspicions

The summary line's `mismatches` lists the suspicions raised in the first browses of the tests and in the incident's accounts. It holds 6 entries, all from first browses of tests.

| Path | Where | Id | The systems model says | The system shows | Why |
|---|---|---|---|---|---|
| `mismatches[0]` | `adapter/serve/scenarios_test.go#TestScenarios` | `demo::adapter::serve::TestScenarios` | unknown | unknown | The test Scenarios is not found in the systems model. |
| `mismatches[1]` | `adapter/serve/scenarios_test.go#TestScenarios` | `AD-0034` | unknown | unknown | The requirement AD-0034 is not found in the systems model. |
| `mismatches[2]` | `adapter/serve/scenarios_test.go#TestScenarios` | `demo::adapter::serve::adapter` | unknown | unknown | The adapter element is not found in the systems model. |
| `mismatches[3]` | `adapter/syntax/scenarios_test.go#TestScenarios` | `unknown` | TestScenarios is not in the systems model | TestScenarios is in the code | The systems model does not include the test 'TestScenarios', which is present in the code. |
| `mismatches[4]` | `examples/pipeline/document/scenarios_test.go#TestScenarios` | `AD-0034` | unknown | not found in the model | The test references an acceptance criterion (AD-0034) that is not present in the systems model. |
| `mismatches[5]` | `examples/pipeline/ui/ui_test.go#TestPureModulesUnderNode` | `unknown` | tokeniser, layout, shared client's frame parser | unknown | These modules are not present in the systems model, making it impossible to verify or exercise them through the test. |

The summary line does not list the suspicions raised in the probe and repeat browses. Computed from their final call lines, the 50 probe browses that gave a final answer raised no suspicion, and the 43 repeats raised 1, in the repeat of `adapter/syntax/scenarios_test.go#TestScenarios`, the same in every field as `mismatches[3]`.

The header's incident key names one known mismatch, `Federation_LogicalArchitecture::Adapter::serve`: "The systems model puts the store and the version counter in adapter/serve. The code has them in adapter/projection/store.go." [The README](../../README.md#reading-the-articles-rules) says it "counts when a suspicion names `Adapter::serve` as its element and `adapter/projection/store.go` in its text", and gives no further detail. In this file, a suspicion counts as that mismatch when its id, with spaces, backticks and quotes trimmed and a dot read as `::`, is `Federation_LogicalArchitecture::Adapter::serve` or ends in `Adapter::serve`, and its id, model_says, system_shows or why names `adapter/projection/store.go`. None of the 7 suspicions meets either condition, so there is none to list with one of the two alone. The ids of `mismatches[0]` and `mismatches[2]` hold `adapter::serve` with a further segment after it, so they neither are nor end in `Adapter::serve`. The known mismatch was not among the suspicions.

### Duration and throughput

The run took 18,495 seconds of wall-clock time, 5 h 08 min 15 s, from 15:18:51 to 20:27:06 UTC on 1 October 2026 (`wall_clock_seconds`, `started_utc` and `finished_utc` in [environment.json](environment.json)). Over its 2,433 replies, none of them reused, the language model generated 366,028 tokens, at a median of 20.47 tokens a second per reply, a mean of 20.54 and a pooled rate of 20.51 (`throughput` in environment.json). Each reply's rate is its output tokens over its output time, both from the timings the server returned with it, and the pooled rate is all output tokens over all output time. The time the server spent loading is left out. The same figures follow from the timings of the replies in run.jsonl.

[`EXP-SC-03`](../../model/design-constraints.sysml) states: "A full run finishes in about two hours on the operator's server, a 14B model at 4-bit quantisation on a 12 GB graphics card that generates about 22 tokens a second. It is an estimate, not a gate." The acceptance criterion of [`EXP-US-01`](../../model/stakeholder-stories.sysml) on time, `withinTwoHours`, reads: "On the operator's server the full run finishes in about two hours."

The program's own seconds, on each call line and in the summary line's `timing`, come from the clock that [deviations.md](deviations.md) describes.

### Judgement of the links proposed for tests with no key

These judgements count in no score of the experiment. Each of the 198 rows of [judgements.csv](judgements.csv) is one entry of the summary line's `blind`, with the same number, test and proposed element, judged first blind and then with the reasons (`judgement_blind` and `judgement_with_reasons`). A row counts for every resolver with an entry in `proposals` that carries its test and element, so the rows of the three resolvers add up to more than 198.

| Resolver | Judged | Blind, correct | Blind, incorrect | Blind, unsure | With the reasons, correct | With the reasons, incorrect | With the reasons, unsure |
|---|---|---|---|---|---|---|---|
| word overlap | 98 | 9 | 89 | 0 | 10 | 88 | 0 |
| systems model baseline | 100 | 10 | 90 | 0 | 11 | 89 | 0 |
| language model | 85 | 5 | 80 | 0 | 5 | 80 | 0 |
| all items | 198 | 15 | 183 | 0 | 16 | 182 | 0 |

No row is marked unsure.
