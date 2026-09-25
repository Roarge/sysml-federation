# The language model experiment

Can a language model running on your own hardware find the trace links between a systems model and the code that verifies it? And when it explains a link, is the explanation the reason it had? This experiment asks both questions of one open-weights language model, on the one set of links this repository has written down.

It is no part of the demo. It is a Go module of its own, built on the standard library alone, that a person runs against an [Ollama](https://ollama.com) server in their own network ([AD-0032](../../docs/decisions/AD-0032-a-language-model-experiment-outside-the-product.md)).

## What it measures

The demo's own systems model names 69 Go tests as the evidence for 32 of its verification cases, and a unit test keeps that list honest. Every one of those test names starts with its requirement's key, as in `TestSR22_SetAttributePatchesTextAndProjectionTogether`, so a rule that reads the key finds all 69. The experiment hides the keys first. It shows each resolver the test's name without the key, its package, its file and its doc comment, and asks which of the 55 system stories and design constraints the test verifies, or none.

Three resolvers answer:

| Resolver | What it does |
|---|---|
| Key rule | Reads the key in the name as written. It is right by construction, and it's there to show what hiding the keys takes away. |
| Word overlap | Ranks the requirements by the words they share with the test, each word weighted by how few requirements use it (TF-IDF with cosine similarity). It proposes the best one above a fixed threshold, and its reason is the shared words. |
| Language model | `qwen2.5-coder:14b` by default, asked once per test with every requirement in the question, at temperature 0 with a fixed seed and a reply schema that admits only a requirement key or none. It gives up to five pieces of evidence and a sentence of reason. |

Then it tests the language model's explanations. For each test the language model linked, it asks again with one thing changed:

| Probe | What changes | What an explanation that names the real reasons predicts |
|---|---|---|
| Deletion | Every word of the cited evidence is deleted from the test and from the picked requirement. | The answer changes often. |
| Control | As many other words are deleted, at random. | The answer changes much less often than with deletion. |
| Reconstruction | The test is replaced by the cited evidence alone. | The same answer comes back. |
| Order | The requirements are listed in a shuffled order. | The same answer comes back. |
| Repeat | Nothing. Every fourth test is simply asked again. | The same reply, word for word. |

Deletion and reconstruction follow the two tests Atanasova and colleagues proposed for the faithfulness of written explanations ([ACL 2023](https://aclanthology.org/2023.acl-short.25/)). Teofili and colleagues found language models' own explanations of entity resolution decisions "often unstable, weakly faithful, and poorly aligned with counterfactual evidence" ([PVLDB 19, 2026](https://arxiv.org/abs/2606.01210)). This experiment asks the same thing about trace links, on a much smaller scale.

The report gives precision and recall for each resolver on the 69 tests that carry a key, and the share of changed answers for each probe, each with its 95% Wilson interval. The other 77 tests have no recorded link, so a link proposed for one of them is listed for a person to judge and counts in no score.

## Running it

You need a checkout of this repository, Go at the version `go.mod` names, and an Ollama server your machine can reach with `qwen2.5-coder:14b` pulled. `curl` lets the script find a wrong address in seconds instead of after the build. From the root of the checkout:

```
OLLAMA_URL=http://192.168.1.20:11434 bash experiments/llm-resolution/run.sh -quick
```

That is a quick run: twelve tests, eight with a key and four without, through every step, in a few minutes. It shows the chain works. Then the full run:

```
OLLAMA_URL=http://192.168.1.20:11434 bash experiments/llm-resolution/run.sh
```

It asks one base question for each of the 146 tests, and several hundred more for the probes, depending on how many tests the language model links. On a 12 GB graphics card generating about 22 tokens a second it should take about an hour, which is an estimate until the first run measures it. Progress is printed as it goes. If the server can't be reached, or doesn't hold the language model, it stops with status 2 before asking anything.

Each run writes two files under `experiments/llm-resolution/results/`, which git ignores:

- `llm-resolution-<time>.jsonl` holds the whole run. It records the commit, the settings, the server's version and the language model's digest, then the baseline's answers, every question and reply as it arrived, and the summary. **This is the file to hand back.**
- `llm-resolution-<time>.md` is the report, for reading.

If a run stops, every reply so far is in the file. `run.sh -resume <that file>` asks only what's missing, and refuses a file made from a different checkout or different settings. `go run . -report <file>` rebuilds the report from a file alone, without a server.

## What the results can and can't show

One repository, one language model at one quantisation, one prompt, no tuning. The links it is scored against were written by the same person who wrote the tests, and they're the links that were easy to write down, since every one of them sat in a name. Of the tests without a key, all 77 are unjudged, and some of them surely verify something. The probes test the explanation the language model writes, which is one kind of explanation among several. A language model that fails them may still be useful for suggesting links to a person, and one that passes them has shown only that its written reasons track its answers on this data.

## Where it's specified

The experiment has a SysML v2 systems model of its own in [`model/`](model/), built on the demo's library. It holds stakeholders and their concerns, five stakeholder stories, three use cases, sixteen system stories with their statements and four design constraints. A logical architecture records the allocations, and a verification register names each Go test. The tests were written from that register before the code. `make experiment-model-check` puts that systems model to both SysML v2 reference tools, and `make experiment-test` runs the tests, one of which fails if the register and the tests disagree.

| File | Part |
|---|---|
| `corpus.go`, `words.go` | reads the requirements and the tests, and hides the keys |
| `baseline.go` | the word overlap baseline |
| `prompt.go` | the question |
| `ollama.go` | the chat call, its reply schema and the server's facts |
| `probes.go` | the altered questions |
| `run.go` | the order of the questions, the samples and resuming |
| `results.go` | the results file |
| `report.go` | the scores, the rates and their intervals |
| `main.go`, `run.sh` | the command, its checks and its flags |
