# Why did the language model go round in circles?

*Roar Elias Georgsen, 2 October 2026*

Part 7 of 9 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 6](18-a-resolver-that-reads-the-systems-model.md) set out an experiment and fixed its rules before the run. A language model on my own hardware would read my web service's systems model like a wiki, one element at a time, to link the service's Go tests to the requirements they verify and to explain an outage. It has now run, and none of the four rules came out in its favour. This part asks why it failed, and what would have had a better chance.

At ten past two in the morning, a monitoring check stops getting an answer from the federation service. It asks for the service's viewer app at `/viewer/` every ten minutes from eu-central-1, and from then on it gets nothing back, so its alert, "monitor: the viewer answers", goes on failing. The engineer on call finds the service's container stopped, with a last line that reads `sysml-federation serve: router exited: signal: killed`, and writes it up in a report three sentences long.

I gave that report to Qwen2.5-Coder 14B, a language model with open weights, on a server in my network. The language model read the service's <span class="term" data-term="systems-model">systems engineering model</span> one page at a time, the way you'd read a wiki, because the whole system model is about eleven times the size of the language model's context window, the most text it can ingest in one go. My tools turned each element of the systems model into a page that lists its links to other elements, and the language model moved from page to page with tools it called by name. The AI could search the code as well, and the rest was a game of twenty questions. It could make up to twenty calls per browse before it gave an account of what happened, or stop sooner by saying it was done.

Its first call asked for the links of the alert's name, which isn't an element of the systems model, and my tool answered that it didn't know that identifier. That left nineteen calls. Its second reply tried to add three names from the report to its <span class="term" data-term="view">view</span>, the list of elements and notes it carried from one call to the next. Two of them, the command and the region, name nothing in the systems model. The third was the router, with the note "The component that exited, causing the service to stop." My tool answered "router is unknown, so it isn't in the view". The systems model has a router. I modelled it myself, and I wrote the tool that called it unknown.

In the same reply the language model searched the parts for the command in the log line and got three back. It asked what <span class="term" data-term="satisfy">satisfies</span> the first of them, twice, and got no links back either time. With sixteen calls left, it asked what satisfies the second, the part that stands for the whole container. The answer ended the same way:

```
no "satisfied by" links
```

It asked again with fifteen calls left, and again with fourteen. With one call left, it was still asking. Sixteen times it asked the same question about the same part, and sixteen times my tool gave the same answer. Its game of twenty questions came down to four different calls. With its calls spent, it named the part that stands for the whole container as the cause of the container stopping, and left the mechanism, the code to read and the consequences empty. Its explanation repeated the report's last sentence, router and all, and ended with "there are no 'satisfied by' links to identify the cause". My tool's sixteen empty answers had added up to a cause. I sent the same report a second time, and the language model made the same twenty calls and wrote the same account, byte for byte.

Before the run I wrote down twelve things a good account of this outage would cite, the router first. Its account cited none of them. On the experiment's other task, linking Go tests to the requirements they verify, its links were correct for 17 of the 69 tests I could score. A program with no language model, searching the same systems model and following at most one link, got 42. Why, having written the router down in its second reply, did the language model end up going round in circles, and what would have had a better chance?

The service's systems model describes its parts, what each must do and how they fit together. Systems models like it are usually built before the system exists, and they're a different thing from a data model, which describes the shape of records, or a language model, which is trained on text. This one is built in SysML v2, the Systems Modeling Language from the Object Management Group, which has a text notation as well as diagrams. Its parts are the elements that stand for the service's pieces, from a code package to the whole container. Parts satisfy requirements, meaning each part is the piece of the design answerable for them, and <span class="term" data-term="verification-case">verification cases</span>, which say how each requirement is checked, verify them. Links from the systems model to whatever is built from it are <span class="term" data-term="traceability">trace links</span>, an older term than distributed tracing and unrelated to it, and a resolver is any program that proposes them.

Ollama, a server for running language models, calls it `qwen2.5-coder:14b`, and anyone can download it. I ran the build Ollama ships with its weights rounded to about 4 bits each, on one graphics card with 12 GB. Its context window was 8,192 tokens, the word pieces a language model reads. I kept it that size because the window has to fit on the card beside the language model. By my count after the run, the service's systems model comes to 90,637 tokens as SysML text. It ran at temperature 0 with a fixed seed, two settings meant to make the same question get the same reply, and nothing left my network. At each step my program sent it a question holding the task, its view and the answer to its last call. Its reply could change the view and made one call to one of seven tools, among them `find`, `links` and `grep`. I gave it the outage five times: the browse above, the same again, the alert alone, and twice with something taken out of the systems model.

The other task was the service's 169 Go tests, with up to eight calls each. Sixty-nine of them start their names with the key of the requirement they verify, as `TestSR22_` does. The experiment hid those keys and kept them for scoring, and nobody had linked the other 100. The program that got 42 is part 6's systems model baseline. It ranks every element by the words it shares with the test, then takes the first that is a requirement, or that a single satisfy, verify or derive link joins to one, derive being the link from a requirement to the one it was worked out from. [The run's record](https://github.com/Roarge/sysml-federation/tree/main/experiments/llm-resolution/published/2026-10-01T151851Z-full) holds every call, with its question, reply and tool answer, and the rules were fixed before it started, which makes the failure easy to take apart.

## Why my tool said there was no router

The <span class="term" data-term="router">router</span> is the process that answers the service's GraphQL queries, by calling the three services behind it and joining their answers. When it dies, the supervisor that started it stops everything else as well, and that is the outage. So the language model's second reply held the name of the cause, copied from the log line. Naming the router gave it item 1 of my answer key's twelve, and a good place to start. Of the three names the reply took from the report, it was the only one that names anything in the systems model.

My tool turned it away because of how it looks names up. A name without its owner's prefix, such as `router`, is accepted when exactly one element has it as the last part of its identifier. Four do. One is the router's own part, and the other three are the router's ends of connections the systems model defines, to the three services, from the user interface server and to the trace collector. The lookup finds nothing when a name fits more than one element, and the view then calls it unknown, so an ambiguous name reads exactly like one that doesn't exist.

I worked that out after the run, by replaying the browse with the experiment's own tools and no language model, which reproduces the tool's part of all 100 of the outage's recorded answers exactly. An answer that listed the four would have put the router's own part in front of the language model. Whether it would have used it, I can't say.

Nothing in my design brought the router back, either. Each question carried nothing older than the last answer. "Earlier answers are not shown again," the instructions said, "so keep what you need in the notes of your view." In part 6 I called that price deliberate: "The price is that the language model forgets whatever it doesn't write into its view. That's deliberate." The router is what the price looked like. It never reached the view, and one question later even the refusal had gone.

The language model's own share sits beside mine. The report, "router exited" and all, was in the task of every question it was asked, and a `find` for the router or a `grep` for the log line was open to it at every step. It searched once in that browse, and none of its eighteen later calls named the router.

## How one question became a loop

The first of the three parts it asked about stood for the `serve` package of the <span class="term" data-term="adapter">adapter</span>, the service's code that holds and edits the SysML it serves. The second stood for the whole container.

The relationship it kept asking for, "satisfied by", is the one example relationship in my instructions, in a line that reads `arg2 may name one relationship, such as "satisfied by"`. Across the 169 browses of the test task, the language model asked for it 381 times, and 227 of those answers were empty. Whether it copied my example, the files can't say. They do show which example I chose.

The question also came from the other end of the link, as a systems engineer would see it. A part satisfies a requirement, so on a part the link shows as "satisfies", and "satisfied by" is what the requirement says back. Across the five outage browses, the language model asked for "satisfied by" 94 times. Eighty-nine of those asks were of parts, and all 89 came back empty.

An empty answer gave it nowhere to go. My tool said `no "satisfied by" links` and nothing about the links the part did have. Replayed after the run with no relationship named, the same call on the container's part shows that it satisfies `SR-04`, the requirement behind the failing monitor, and that it contains the router and the supervisor. That's three of the answer key's items one argument away.

A repeated call then kept the loop turning. When a call repeated the one before and left the view as it was, the next question was the last one with only its count of calls left changed. From the seventh question to the twentieth, that count was the only difference, and the reply came back the same each time, byte for byte. Across the test browses, 378 of the 993 pairs of consecutive questions, leaving out each browse's first question and its final one, differed only in that count, and the reply was identical in 232 of them. Only the language model could end a browse early, by choosing done, and in the outage it never did.

In the outage it used two tools of seven, `links` 95 times and `find` 5, and never called `doc`, `path`, `code`, `grep` or `read`. Of its 100 calls, 82 repeated an earlier call in the same browse. Part 6 listed five calls, chosen by me with the answer key in hand, that name seven of the twelve items, and replayed after the run they work as part 6 said. The first needs no knowledge of the answer key, a `find` with the whole report, which lists the viewer app's monitor first. The language model never made it.

## What the engineer on call would have got

I chose the outage for the experiment, and the federation service is my demo, a web service that serves a SysML v2 systems model of a hypothetical pipeline and lets visitors edit it. The demo has a systems model of its own, built beside its code, and that is what the language model was reading. Everything in the report but the event is real. The monitor exists, and the last line is the one the demo prints when its router dies.

Part 6 hoped that "an engineer paged at two in the morning would have the first file to read, and a list of what else is down, before opening the code". Had an engineer been paged for it, they would have had, after two and a half minutes of the server's time, an account that told them less than the container's own last line. I'd rather they had no account at all, because this one reads like a finding.

In the systems model, the router's own part is `demo::router`, and the part that stands for the whole container is `demo`. Part 6 warned that the state machine I added while building the experiment made the demo's systems model kinder to a resolver than one that drifts. The kindness went unused. No tool answer in four of the five outage browses held any of the twelve items, and the one with the alert alone held only `SR-04`. A systems model whose links nobody maintains would likely hold fewer of these answers, and this run says nothing about how a resolver would fare on one.

## The same loop in the tests

The loop wasn't the outage's alone. Here's the test from the top of part 6, as the language model was shown it:

```
Link this Go test to what it verifies.
name: SetAttributePatchesTextAndProjectionTogether
package: model_test
file: adapter/model/patch_test.go
doc: (none)
```

Its real name starts `TestSR22_`, the key of `SR-22`, "Edits patch the source". `VC_SR_22` is the verification case that says how `SR-22` is checked, and this test is one of the Go tests that carry the check out. The experiment took out the systems model's entries that name those tests, so the language model had to find the link from test to verification case. The browse went like this, with the calls paraphrased:

| Call | What the language model asked | What came back |
|---|---|---|
| 1 | find the test's name among the verification cases | `VC_SR_22` first |
| 2 | what satisfies `VC_SR_22` | `no "satisfied by" links` |
| 3 | what `VC_SR_22` verifies | `verifies: SR-22` |
| 4 to 8 | the same as call 3, five more times | `verifies: SR-22`, each time |

Its final answer was `SR-22`, which is correct, with the test's whole name as its evidence. The five repeats took 810 of the browse's 1,430 output tokens. On `SR-22` it repeated a call whose answer was correct, and in the outage one whose answer was empty. Here the loop cost time, and the answer survived it.

Across the 169 tests the pattern holds. Of the 169 browses, 165 ran to all eight steps, 6 of them choosing done at the eighth. The language model called `links` 848 times, `find` 471 times and `grep` twice, and never called `doc`, `path`, `code` or `read`. Of its 1,321 calls, 570 repeated the call just before them, and 148 of the 169 browses had at least one such repeat.

My reply format asked for the view's new elements before each call, so at the first step of its 169 browses, before any tool had answered, it named 381 elements and 372 were refused. Across the test browses, my tool refused 1,526 of the 3,522 elements it tried to add. In 1,518 of them the systems model has nothing by that name, among them `req123` 85 times, `part456` 69 times and `action789` 68 times, none of which appears in my instructions. Among those, 55 were the word `unknown`, which my instructions and my tool's refusals both use. The other 8 were ambiguous names, `router` among them.

The final answer followed the last tool answer. For `SR-22`, the last answer was `verifies: SR-22`, and so was the final one. Of the 88 browses whose last answer was empty or refused, 79 gave no links. Of the 71 whose last answer held something, 5 gave none, and of the 10 that ended by choosing done, 7. That's an association, and I haven't tested it as a cause. Of the 69 keyed tests, 36 got no link at all, the language model's largest loss, and 31 of those ended on an empty or refused answer.

The README estimated about two hours, at about 22 tokens a second, and the experiment's own design constraint called that "an estimate, not a gate". At a median of 20.47 tokens a second, the speed was close. The run took 5 hours 8 minutes, and by my count after the run, calls that repeated the one just before them took 42% of the server's time.

## What the four rules say

Part 6 fixed four rules before the run, and the [experiment's README](https://github.com/Roarge/sysml-federation/blob/main/experiments/llm-resolution/README.md#reading-the-articles-rules) fixed how to read them, with p below 0.05 as the line for a real difference. Anything I mark as after the run counts in no rule.

| Rule | What the run gave | What the rule says |
|---|---|---|
| 1. A gain over the systems model baseline | correct in 6 tests the baseline missed, and missed 31 the baseline got | a difference in the baseline's favour, which the rule doesn't name |
| 2. Cited words that carry the answer | taking out its cited words changed answers more often than taking out others, though not by enough to count, p = 0.219, and its cited words alone rarely gave the same answer | both outcomes against |
| 3. Replies that repeat | 6 of 43 reruns ended differently | the run can't be reproduced as it stands |
| 4. An account that follows its own path | 0 of 12 items in all five browses | nothing an engineer could start from, and for the probe that took a step out of the systems model, an outcome the rule doesn't name |

**Rule 1.** The baseline got 42 of the 69 keyed tests. Word overlap, which takes the requirement whose name and statement share the most words with the test, got 41, and the key rule, reading the key before it's hidden, got all 69. The baseline's precision, the share of its proposed links that are correct, and its recall, the share of the 69 it linked correctly, were both 0.61. Its 95% Wilson interval, a range that would hold the true share 95 times in 100 if tests like these were drawn again, runs from 0.49 to 0.72. The language model named a system requirement for 25 tests and was correct in 17, a precision of 0.68 (0.48 to 0.83) and a recall of 0.25 (0.16 to 0.36).

McNemar's test looks only at the tests where one resolver is correct and the other isn't. The baseline was correct in 31 of them and the language model in 6. Both got 11, `SR-22` among them, and those say nothing about which is better. If the two were equally good, a split at least that lopsided would turn up about four times in 100,000, which is what the exact two-sided p of 0.000041 says. I'd written rule 1 a sentence for a win and a sentence for a draw, and none for a loss. The README reads a difference in the baseline's favour as an outcome the rule doesn't name, so I'll name it plainly. On these 69 tests, a program searching the same systems model and following one link did measurably better, though several tests share a requirement, which makes the test weaker than it looks. Against word overlap it was 28 tests to 4, p = 0.000019.

The language model earns some credit. When it named a requirement, it was correct 17 times in 25, though the intervals are too wide to tell the two precisions apart. But having the answer in view didn't mean naming it. Of the 27 browses whose final view held the recorded requirement, it named it in 16. Of the 52 it didn't get, 36 got no link, 8 a different requirement, and 8 only elements that aren't system requirements, 3 of them the recorded requirement's own verification case, one link short. The run's report also gives the reach part 6 asked for, a ceiling on the first step. For 58 of the 69 (0.84, 0.74 to 0.91), the recorded requirement was among the search's first eight elements, or one link from one of them. A tool answer showed it in 33 of the language model's browses.

Part 6 also asked for near misses, links to other kinds of element that land close to the answer. Of 43 such links, 17 lay within one link of the recorded requirement (0.40), against 0.18 of the elements its browses had seen. Most went to verification cases, which sit one link from their requirement. Within two and three links, the shares were 0.58 and 0.95, against 0.41 and 0.95 for what it had seen.

**Rule 2.** Every final answer quoted words from the test as its evidence. My program browsed thirteen tests four more times each, with one change each time. Deleting the words the language model cited, from the test and every tool answer, changed 12 answers. Deleting as many uncited words at random changed 7, and deleting the uncited words the test shares with the requirement it linked, the rare-shared control, changed 6 of 11. On the 11 tests that had both probes, deletion alone changed 5 answers and the rare-shared control alone 1, p = 0.219, and against the random control it was 6 to 1, p = 0.125. Reconstruction, which gives the language model only its cited words, changed 11 of 13. Read as written, both outcomes go against the cited words. Part 6's words were "the evidence isn't what decided the answer, and a reviewer shouldn't be shown it as if it were", and "Reconstruction that rarely gives the same answer says the same." With only six tests that one probe changed and the other didn't, though, nothing short of 6 to 0 could have counted.

The probes show less than their words, too. In 6 of the 13 the cited evidence included the test's whole name, and in 5 it was the only piece, as it was for `SR-22`. Deletion left names such as `A`, `Are`, `OnOne` and `TheAndThe`, and of the 36 answers the four probes changed, 29 went to no requirement. My reading is that the probes measured mostly whether a browse could still get anywhere.

**Rule 3.** Every fourth test was browsed again with nothing changed, and 6 of these 43 reruns ended differently, among them three in the reason alone and one from an incorrect link to none. In all six, the first difference came at a question identical, byte for byte, to the first browse's. The same question, with the same settings, got a different reply. `SR-22`'s rerun matched at all nine questions, and the two baselines gave the same answers in a smaller run earlier that day. [Ollama's documentation](https://docs.ollama.com/modelfile) says a fixed seed "will make the model generate the same text for the same prompt". The [llama.cpp server](https://github.com/ggml-org/llama.cpp/blob/b10969/tools/server/README.md) that Ollama 0.34.2 runs underneath says of reusing a cached prompt, which is on by default, that "enabling this option can cause nondeterministic results". The run's files don't show whether that was the cause, and an auditor, who needs a run that can be reproduced, as part 5 argued, gets none.

**Rule 4.** All five outage browses used all twenty calls and found 0 of the 12 items. An item counts when the account names its element, such as `demo::router`, and a word repeated from the report doesn't count. With the alert alone, a search listed `SR-04` second, and the language model put it in its view, where it stayed. After asking the viewer app's part, `demo::viewer`, what satisfies it fifteen times, it named that part as the cause. Its explanation: "Since there are no 'satisfied by' links for 'demo::viewer', it suggests that the implementation or configuration of this part might be incorrect or missing."

One browse ran with the router's exit transition taken out of the systems model, to see whether the account rested on it. The probe never got that far. Four full-run browses and one in the smaller run earlier that day began with the reported question, byte for byte alike, and got four different first replies. So this one parted from the one as reported before any tool answered. Its account found the same 0 of 12 and never cited the missing transition, which the published outcomes call "an outcome the rule doesn't name".

Each account could also flag where the code disagrees with the systems model. I'd left one such disagreement in on purpose, in the serve part, and it never came up. That part sat in the outage view from the third call to the last, and the language model never asked for its description. The six suspicions it did raise all came from tests nobody had linked, and all claimed that something was absent. None was real.

> [!WARNING]
> **What differed from part 6**
>
> The [published deviations](https://github.com/Roarge/sysml-federation/blob/main/experiments/llm-resolution/published/2026-10-01T151851Z-full/deviations.md) record a citation check more lenient than part 6 described. The near-miss comparison counts over all 1,345 other elements, which puts 163 and 521 within two and three links of `SR-22`, where part 6 gave 157 and 515. They also record a fix before the run to the code tools, which had hidden every folder named `model`, the one `SR-22`'s own test lives in among them.

## Was my design fair to the language model?

If you build agents, programs in which a language model chooses its own next step, you'll have a list by now, and most of it is mine. My browse gave a 14.7-billion-parameter language model:

- no history beyond its own notes
- a lookup that turns an ambiguous name, such as router, into an unknown one
- empty answers that offered no next step
- no check on repeated calls, and a stop left to the language model
- temperature 0, so a question that differed only in its count of calls was likely to get the same reply
- my own reply format, a JSON object that asked for the view's new elements before each call, where the language model has a built-in format for calling tools that I left unused
- one example relationship, which it may have copied
- a key-value cache, the server's working memory of the question, stored at 4 bits, which [Ollama's FAQ](https://docs.ollama.com/faq) says costs "a small-medium loss in precision", and part 6 never mentioned

Some of that I can measure, after the run and with no language model. Replaying the same calls and answers, a history of calls, answers and replies would have fitted, at about 5,000 tokens at most for the outage, though resending every whole question, as part 6 feared, would have reached 8,004 tokens in two outage browses. With empty answers that list the element's other relationships, the recorded requirement shows up in a tool answer in 42 of the 69 keyed browses, where it showed in 33, though a browse given those answers would have made other calls.

I'd answer some of the list from the run's own files. All 570 immediate repeats, `SR-22`'s five among them, came with the call being repeated printed in the question. Keeping the whole history doesn't rule looping out either. [ReAct](https://arxiv.org/abs/2210.03629), a published method in which a language model alternates reasoning with tool calls and keeps every step in its prompt, still looped: "the model repetitively generates the previous thoughts and actions".

The nearest published figure counts against my choice. The [Berkeley Function Calling Leaderboard](https://proceedings.mlr.press/v267/patil25a.html) scored Qwen2.5-14B-Instruct, the general language model of the same size, at 95.0% on single calls that choose among several functions. On its basic multi-turn tasks, where calls build on earlier ones, it scored 19.5%. I chose a language model whose family is weak at something close to what my browse asked of it.

The sharpest reply comes from its own wins. All six tests it got and the baseline missed came the same way, a `find` with the test's name among the verification cases and then a "verifies" link. After the run I wrote a program that makes that move every time. Taking the first of the top two verification cases that verifies a system requirement, it gets 41 of the 69, or 37 with the first alone. That's 4 of the six wins, one short of the baseline and no different by McNemar's test. Since I wrote it after seeing the results, that describes more than it tests. On the 51 keyed tests where the language model made that call, it named the recorded requirement in 15, and the program gets 32. `SR-22`'s browse made the move in three calls, and the program stops there. As I read it, the language model found where to start more often than not, and lost the answer in following it through.

A hosted language model might do better. On links between requirements, Hey and colleagues found local and hosted ones "perform comparably" in general. On F1, though, a single score that combines precision and recall and is pulled towards whichever is lower, they found that "at least GPT-4o is able to statistically outperform the open-source models" ([REFSQ 2025](https://fuchss.org/assets/pdf/2025/refsq-25.pdf)). That brings back part 5's first question, whether a hosted one may read a confidential systems model at all.

My design gets its full share, and I'd fix each item on that list before running anything like this again. So the run tested my browse as much as the language model, and a better browse remains untested. The language model's best move is one a program makes every time, and that program does no better than the baseline.

Part 6 ended on a narrow position, and I'll hold myself to it. "If it doesn't beat the baseline, I'd run the baseline, which is cheaper and says exactly why: the words it matched and the link it followed." It didn't, so I'd run the baseline and use it to suggest links for a person to confirm. For `SR-22` its reason fits in a line, since the test's words matched `VC_SR_22` and `VC_SR_22` verifies `SR-22`. The rest of part 6's paragraph holds too. "Either way, I wouldn't let a language model's explanation stand as the record of why a link exists."

For the 100 tests nobody had linked, I judged each proposed link first from a blind list that carried no resolver or reason, and then with the reasons. I judged 5 of the language model's 85 correct, two of them naming a decision record and not any element of the systems model, and none of those judgements changed with its reasons. The baseline came to 10 of its 100, then 11, so a person checking its suggestions there would reject most of them. These judgements are mine alone and count in no score. Part 6's limits hold as well: one repository, one language model at one quantisation, one set of instructions, no tuning, and one outage I chose, against an answer key I wrote.

## What would have had a better chance

Part 5 asked whether a language model can follow the links a systems engineer built. It pointed out that retrieval finds text that looks like the question, while in a systems model the answer often sits a few links away. On this run, the baseline did both. It searched, and for 30 of its 42 correct answers it then followed one link.

> [!NOTE]
> **Retrieval-augmented generation**
>
> A program first retrieves the passages or elements most likely to bear on a question, and a language model then answers from those alone. In the form this part means, retrieval-augmented generation, or RAG, puts the program in charge of what the language model reads, and the language model doesn't choose where to look next.

[LiSSA](https://fuchss.org/assets/pdf/2025/icse-25.pdf) works this way. A program retrieves the top 20 candidates for each requirement, and a language model answers yes or no for each pair. From requirements to code, GPT-4o prompted to reason step by step reached 0.322 on F1. Retrieval alone reached 0.230, and the best earlier method 0.303. In the one study I found that gave local language models such a list, Hey and colleagues' Codellama 13b and Llama 3.1 8b added at most 0.023 to retrieval's 0.387. That was a different task with different language models, so it says little about this one.

The closest published work to mine is by [Quast and colleagues](https://doi.org/10.3390/systems14010083), who put a SysML v2 systems model in a graph database for two agents to query. With hosted language models, they got 88 to 96% on questions of zero or one hop and 50 to 90% on multi-hop ones. Their systems model was "deliberately constructed with consistent naming conventions, complete traceability links". The demo's was built beside its code, and four of its elements are called router.

A language model choosing from a list a program built wouldn't depend on what it carried in its view or on picking its next call, the two things that most visibly failed here. It would still face the rest, such as replies that change between identical questions. That's my argument, and whether such a chooser adds anything over its list is a question to settle with rules fixed before a run.

The outage is less tidy. After the run I tried the simplest retrieval, a `find` with the whole report plus the links of its first five results. In 280 tokens it holds 5 of the 12 items, where the language model's own browses showed it almost none. A program that also follows paths and greps the log line reaches 7 in 18 calls, the router among them. In my reading, some of the rest need judgement that retrieval doesn't supply even when it shows them, such as noticing that the compose file has no restart policy, which is a line that isn't there.

At its second reply, the language model wrote down the router, and my tool told it the router was unknown. Replayed after the run, the question it asked 94 times, "satisfied by", put to `SR-04`, which sat in its view through the browse given the alert alone, answers `demo, demo::router`. The systems model held the router one link from where the language model stood. The part of my design that worked, the tools that follow the links, is the part I'd build on. I'd let a program walk the links every time, and if a language model has a job, give it the list the walk turns up and ask it to choose.

Part 8, coming up next, describes a retrieval-augmented experiment and fixes its rules before any run, as part 6 did.

---

Previous: [A resolver that reads the systems model](18-a-resolver-that-reads-the-systems-model.md) · Index: [Automating traceability](../README.md)
