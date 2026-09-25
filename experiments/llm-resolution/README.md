# The language model experiment

Can a language model running on your own hardware read a systems model the way a person reads a wiki, and use it to explain the code? This experiment gives one open-weights language model two tasks on the demo's own systems model. It links each Go test to what the test verifies, and it explains an outage. Then it checks whether the explanations it writes are the reasons it had.

It is no part of the demo. It is a Go module of its own, built on the standard library alone, that a person runs against an [Ollama](https://ollama.com) server in their own network ([AD-0032](../../docs/decisions/AD-0032-a-language-model-experiment-outside-the-product.md)).

## The systems model as a wiki

The experiment reads the SysML v2 text under `model/` into pages and links. Every element is a page. Every relationship the systems engineer wrote is a link that shows on both of its ends, from satisfy and verify to allocations, bindings, derivations and a state machine's transitions. A unit test holds the reader to the systems model: every reference resolves, and the satisfy, allocate and bind statements each give one link.

The language model never sees the whole systems model, which would not fit in its context. It gets tools instead, and the program does everything a program can do:

| Tool | Answers with |
|---|---|
| `find words [kind]` | the eight elements whose text shares the most words, weighted as the word overlap baseline weighs them |
| `links id [relationship]` | the element's links, grouped by relationship, as identifiers |
| `doc id` | its description, attributes, decisions and evidence |
| `path id id` | up to three shortest paths of at most three links |
| `code id` | where the systems model puts it in the code: evidence locations, paths its description names, and where its distinctive values occur |
| `grep text` | lines of the built system's code: Go, JavaScript, TypeScript, YAML, JSON, Dockerfiles and shell scripts, outside `model`, `docs`, `experiments` and test data |
| `read file line` | thirty numbered lines around the one asked for |

A browse is one question per tool call. Each question carries the task, the viewpoint the language model has framed, its view, the calls it has left and the last tool's answer, and nothing older. The view is the language model's working state: the elements it has exposed, each with a note saying why, which it can prune again. A browse stops when the language model says it is done or when its calls run out, eight for a test and twenty for the incident, and a final question asks for the answer. At the end the view is written out as a SysML v2 viewpoint definition and a view that exposes each element, which both reference tools accept beside the systems model.

## What it measures

**Linking the tests.** The demo's verification register names 69 Go tests as evidence for 32 of its verification cases. Every one of those names starts with its requirement's key, as in `TestSR22_SetAttributePatchesTextAndProjectionTogether`, so a rule that reads the key finds all 69. The experiment hides the keys. In this task the tools also leave out the register's test names and mask every key in the code they show. For each test, the language model browses and then links the test to up to five elements, the requirement it verifies first. Three resolvers are scored on the 69:

| Resolver | What it does |
|---|---|
| Key rule | Reads the key in the name as written. It is right by construction, and it's there to show what hiding the keys takes away. |
| Word overlap | Ranks the 55 system stories and design constraints by the words they share with the test, each word weighted by how few requirements use it (TF-IDF with cosine similarity), and proposes the best above a fixed threshold. Its best-ranked answer is also what `find` gives for the test's text, so it is the first step any browse can take. |
| Language model | `qwen2.5-coder:14b` by default, at temperature 0 with a fixed seed, answering under a JSON schema. Its answer is the first link to a system requirement, or none. |

The links to parts, actions, state machines and other elements have nothing recorded to score them against. The report counts them by kind and gives the share that lies within three links of the recorded requirement, beside the share of every element that lies as near.

**Explaining the incident.** The engineer on call reports that the alert "monitor: the viewer answers" has failed since 02:10, and that the service's container stopped with the last line `sysml-federation serve: router exited: signal: killed`. The language model browses the whole systems model and the code, and gives the cause, the mechanism, the code to read first, the consequences, the path it followed, and two or three sentences of why. [`incident-key.json`](incident-key.json), written and committed before any run, lists twelve things a good account names, and each is scored as found or missed. Every element, link and file the account cites is checked against what the browse's tools could show. The steps of its path must be joined one to the next by a link, or by code the element names, or by a file the browse read.

**Testing the explanations.** For the first test the language model linked to a requirement, and every fourth after it, the test is browsed again with one thing changed:

| Probe | What changes | What an explanation that names the real reasons predicts |
|---|---|---|
| Deletion | The cited evidence is deleted from the test and from every answer the tools give. | The answer changes often. |
| Random control | As many of the test's other words are deleted, chosen with a fixed seed. | The answer changes much less often than with deletion. |
| Rare-shared control | The uncited words the test and its requirement share are deleted, the rarest first, until as many are gone. | The answer changes less often than with deletion, even though these words are as telling. |
| Reconstruction | The test is replaced by the cited evidence alone, and its own code is hidden from the tools. | The same answer comes back. |
| Repeat | Nothing. The first test and every fourth after it are browsed again. | The same final reply, word for word. |

The incident is browsed five times. It goes as reported, again unchanged, and with the alert alone. Then the supervisor's transition on the router's exit is taken out of the systems model, and last an unrelated allocation is taken out instead. An account that rested on the transition should change when it goes, or say that the systems model no longer explains what the code does. A citation of the missing transition is flagged, whatever the sentence says.

Deletion is close to the erasure measure Teofili and colleagues used for language models' explanations of entity resolution decisions, which they found "often unstable, weakly faithful, and poorly aligned with counterfactual evidence" ([PVLDB 19, 2026](https://arxiv.org/abs/2606.01210)). Reconstruction is one of the two tests Atanasova and colleagues proposed for written explanations ([ACL 2023](https://aclanthology.org/2023.acl-short.25/)). The random control only matches the number of words deleted, and the cited words are usually the most telling ones, so the rare-shared control matches their kind as well. The final answer lists the links before the evidence, so the evidence is written once the answer is chosen, and the probes test the evidence, not the sentence of reason.

Every final question also asks where the language model suspects the system wasn't built as the systems model says. The report lists each suspicion with the element, what the systems model says, what the system shows and why. One mismatch is known before the run and listed in the key: the systems model puts the store and its version counter in `adapter/serve`, and the code has them in `adapter/projection/store.go`.

The report gives precision and recall for each resolver on the tests that carry a key. For each probe it gives the share of changed answers, on all links and on the correct ones alone, split into changes to none and to another requirement. Each rate has its 95% Wilson interval. The language model and the word overlap's best-ranked answer are compared test by test, with McNemar's exact test. The other 77 tests have no recorded link, so a link proposed for one of them counts in no score. It is listed for a person to judge, once with its reasons and once without them.

## Running it

You need a checkout of this repository, Go at the version `go.mod` names, and an Ollama server your machine can reach with `qwen2.5-coder:14b` pulled. `curl` lets the script find a wrong address in seconds instead of after the build. From the root of the checkout:

```
OLLAMA_URL=http://192.168.1.20:11434 bash experiments/llm-resolution/run.sh -quick
```

That is a quick run: twelve tests, eight with a key and four without, through every step, and the incident once. It shows the chain works. Then the full run:

```
OLLAMA_URL=http://192.168.1.20:11434 bash experiments/llm-resolution/run.sh
```

It browses each of the 146 tests, the incident five ways, the four probes on the sample and the repeats. On a 12 GB graphics card generating about 22 tokens a second it should take about two hours, which is an estimate until the first run measures it. Progress is printed as it goes. If the server can't be reached, doesn't hold the language model, or the incident's key names something the checkout lacks, it stops with status 2 before asking anything.

Each run writes two files under `experiments/llm-resolution/results/`, which git ignores:

- `llm-resolution-<time>.jsonl` holds the whole run. It records the commit, the settings, the server's version, the language model's digest, the instructions, the key and hashes of the tests and the systems model. Then come the baseline's answers and every question and reply as it arrived, with the tool's answer. A final line for each browse holds its view and its checks, and the summary comes last. **This is the file to hand back.**
- `llm-resolution-<time>.md` is the report, for reading.

If a run stops, every reply so far is in the file. `run.sh -resume <that file>` asks only what's missing, and refuses a file made from a different checkout, systems model or settings. `go run . -report <file>` rebuilds the report from a file alone, without a server.

## What the results can and can't show

One repository, one language model at one quantisation, one set of instructions, and no tuning. There is one incident. Whoever chose it also wrote the systems model, the key and the tools, so the page format and the tool list are one person's design and no standard. While this experiment was being built, the supervisor's state machine, which carries the incident's mechanism, was added to the systems model from the code. The links the tests are scored against were written by the same person, one requirement per test, and they are the links that were easy to write down, since every one sat in a name. Several of the 69 tests share a requirement, so they aren't independent, and the intervals are narrower than they should be. Since the repository was started in August 2026, long after the language model was trained, it can't have seen these links. Of the tests without a key, all 77 are unjudged, and some of them surely verify something. As for the citation check, it proves that what an answer cites exists, and nothing about whether it supports the answer.

## Where it's specified

The experiment has a SysML v2 systems model of its own in [`model/`](model/), built on the demo's library. It holds stakeholders and their concerns, six stakeholder stories, four use cases, 22 system stories with their statements and four design constraints. A logical architecture records the allocations, and a verification register names each Go test. The tests were written from that register before the code. `make experiment-model-check` puts that systems model to both SysML v2 reference tools, and `make experiment-test` runs the tests, one of which fails if the register and the tests disagree.

| File | Part |
|---|---|
| `corpus.go`, `words.go` | reads the requirements and the tests, and hides the keys |
| `wiki.go` | reads the systems model into pages and links |
| `tools.go` | the seven tools, and what each task hides from them |
| `baseline.go` | the word overlap baseline, and the weighting `find` uses |
| `view.go` | the view, and its SysML v2 rendering |
| `browse.go` | one question per tool call, then the final question |
| `prompt.go` | the instructions and the questions |
| `ollama.go` | the chat call, the reply schemas and the server's facts |
| `incident.go`, `incident-key.json` | the incident, its key, its scoring and the citation check |
| `probes.go` | the altered browses |
| `run.go` | the order of the browses, the samples and resuming |
| `results.go` | the results file |
| `report.go` | the scores, the rates and their intervals |
| `main.go`, `run.sh` | the command, its checks and its flags |
