# A resolver that reads the systems model

*Roar Elias Georgsen, 28 September 2026*

Part 6 of 9 in [Automating traceability](../README.md).

> [!IMPORTANT]
> **The story so far**
>
> [Part 5](17-what-makes-a-resolver-worth-running.md) asked what decides whether a resolver, the program that proposes trace links, is worth running once it measures well. It ended on language models and three questions. May a hosted one read the systems model at all? Can one follow the links a systems engineer built? And are the explanations it writes the reasons it had? This part describes what I built to find out, and part 7 runs it.

Here's a test from my demo, exactly as a language model will be shown it:

```
Link this Go test to what it verifies.
name: SetAttributePatchesTextAndProjectionTogether
package: model_test
file: adapter/model/patch_test.go
doc: (none)
```

Its real name starts with `TestSR22_`, the key of the requirement it verifies, and that prefix is gone. Which of the demo's 56 requirements is it? If you've read the earlier parts, the body of the test will look familiar. The demo is a small web service that serves a SysML v2 <span class="term" data-term="systems-model">systems model</span>, the example pipeline's, and lets visitors edit it. This test raises the throughput of `PIPE-S2`, the pipeline's parse stage, from 1,200 queries per second to 1,700. Then it checks that the SysML text the demo serves changed in that one number and nowhere else. I know the answer is `SR-22`, "Edits patch the source", because I built the demo's own systems model and wrote the test. A language model starts from the box above, and from whatever the demo's systems model and code can tell it.

[Part 4](16-measuring-a-resolver.md) pointed out that a rule reading the key finds all 69 of the demo's keyed tests and proves nothing by it. Hiding the keys was the obvious next step, and I've now taken it. The language model is `qwen2.5-coder:14b`, an open-weights language model, which means anyone can download it and run it. It's small enough for a graphics card with 12 GB, and it runs under Ollama, a server for language models, on a machine in my own network. That sidesteps part 5's first question for this experiment, since nothing leaves the building. The other two, whether it can follow the links and whether its explanations are its reasons, take more machinery, and most of this part is about that machinery.

I want it to work, and there's reason to think it can. Language models have recovered trace links before. The authors of [LiSSA](https://fuchss.org/assets/pdf/2025/icse-25.pdf) had GPT-4o judge the candidate links a search turned up. They found that it "can significantly outperform the state-of-the-art on the code-related tasks", and add that it needs more work before it's fit for practice. Their requirements were plain documents. Mine sit in a systems model whose links a person already built, and in both of this part's examples the tools reach much of the answer before the language model has made a single choice. If you remember one thing from this part, make it this: what would show the language model following those links, or failing to, was written down before any answer existed.

## What the systems model had to learn first

A resolver that reads a systems model can only follow the links someone built into it. The demo's systems model had plenty when I started. It had 94 <span class="term" data-term="satisfy">satisfy</span> links saying which part meets which requirement, <span class="term" data-term="verification-case">verification cases</span> saying how each requirement is checked, <span class="term" data-term="derivation">derivations</span> from the stakeholders' needs to the requirements, and <span class="term" data-term="decision-record">decision records</span>. What it lacked was behaviour. Nothing in it said what the demo does when a part fails, or which part carries out each step of an edit.

That gap mattered most for the experiment's second task, explaining an outage, which comes further down. The code is plain about it. `supervisor.run` in `cmd/sysml-federation/serve.go` starts the three <span class="term" data-term="subgraph">subgraphs</span>, the services that each answer part of a query, then the <span class="term" data-term="router">router</span> that joins them, then opens the port. It returns the moment any of them fails. The supervisor is the container's first process (PID 1), so the container stops with it. None of that was in the systems model, whose description of the supervisor gave the start order and stopped there.

So before building the experiment, I modelled what the code already did, citing the file each piece came from:

- a state machine for the supervisor, `SupervisorStates`, with seven states from loading to stopped and a transition, `onRouterExit`, from serving to stopping
- 20 allocations, which say which part carries out each step of an edit, or of a request traced through the demo during a check
- a variation, SysML's way of saying one of several alternatives, for the connector that puts the demo on a public tunnel during a check session: a named tunnel or a quick one, never both
- two calculations for the capacity service, the maximum flow and the cut that finds the bottleneck

Both SysML v2 reference tools, the OMG's <span class="term" data-term="pilot-implementation">pilot implementation</span> and <span class="term" data-term="opensysml">OpenSysML</span>, accept all of it, and OpenSysML can run the state machine. Given a router exit while serving, it takes `SupervisorStates` to stopped, and `make model-state-check` checks that it still does.

I should own up to what this does to the experiment. The state machine carries the mechanism of the outage the language model will be asked to explain, and I added it while building the experiment. A systems model that has just been brought up to date with its code is kinder to a resolver than the drifting ones part 3 worried about. Part 7's results have to be read with that in mind.

## The systems model as a wiki

The language model can take in very little of the systems model at once. The experiment's parser reads the demo's SysML v2 text into 1,346 elements. Listed at one line each, with an identifier, a kind and a short description, they come to 31,777 tokens, the word pieces a language model reads. The language model runs with a context of 8,192 tokens, the most it can read in one go. That's a setting, and the card decides it. At Ollama's default settings, 31,777 tokens of context would take about 5.8 GB of the card's memory. The language model itself takes up 9 GB, and the card has 12.

So the language model never sees the whole systems model. It reads it the way a person reads a wiki, a page at a time, following links. Every element is a page, and every relationship the systems model states is a link that shows on both of its ends. The page for `SR-22` lists the part that satisfies it and the verification case that verifies it, and each of those pages lists `SR-22` back. A unit test holds the parser to the systems model. Every reference has to resolve, and every satisfy, allocate and bind statement has to give exactly one link.

The language model gets seven tools, which it calls by name, one at a time. Ordinary code does all of their work, so the language model spends its few tokens on choosing and judging:

| Tool | Answers with |
|---|---|
| `find` | the eight elements whose text shares the most words with the words it's given |
| `links` | an element's links, grouped by relationship, as identifiers |
| `doc` | its description, attributes, decisions and evidence |
| `path` | up to three shortest paths of at most three links between two elements |
| `code` | the files and lines the systems model names for the element |
| `grep` | lines of the demo's code that hold some text |
| `read` | thirty lines of a file around one line |

Take the test from the top. The obvious first call is `find` with the test's own text, and these are its first five answers, without their descriptions:

```
VC_SR_22::literalAndProjectionAgreeScenario (action)
VC_SR_22 (verification def)
VC_SR_22::agreeUnderConcurrentReadsScenario (action)
Adapter (part def)
SR-22 (requirement)
```

`links SR-22` would then show that the <span class="term" data-term="adapter">adapter</span> satisfies it, that `VC_SR_22` verifies it, and that it derives from four stakeholder stories, the user needs it came from.

Those answers make this test easier than it looks, and I'd rather say so than have you find it. `VC_SR_22` is the verification case for `SR-22`, and its name and its only verify link both give that away. For this task the tools hide two things. The verification register lists the evidence behind each verification case, and its entries that name Go tests are taken out of the systems model. And every requirement key in the code the tools show is masked. What stays is each verification case's description, which I wrote. All 32 that name Go tests open with their requirement's own title, and `VC_SR_22`'s goes on to say "six Go tests across the model and projection packages". A verification case is meant to say how its requirement is checked, and a resolver that reads one is doing its job. It does mean that plain word overlap finds the correct requirement for this test without any browsing at all.

The two scenario entries in that list came after I'd built the experiment. Every acceptance criterion of the demo's requirements, bar the seven design constraints, now also has a scenario in Gherkin. That's the plain-text given, when and then form of behaviour-driven development, and `go test` runs the 59 scenarios that can run there. The register names each of those as evidence, by its criterion and its feature file, and the tools keep these entries, which name no test. Their descriptions are the criteria themselves, so they share words with any test of the same behaviour. Their feature files are named after the requirement, as in `sr22-edits-patch-the-source.feature`, and give the key away just as the verification case's name does.

## One question per tool call

A browse is a series of steps, one per tool call. At each step the program sends the language model one message. It holds the task, the language model's viewpoint, a sentence saying what it's trying to find out, and its view so far, which is coming up next. It also holds the calls it has left and the answer to its last call, and nothing older. The language model replies in JSON under a fixed schema. The reply may sharpen the viewpoint, add up to three elements to the view with a note saying why each matters, take up to three out, and make the next call. A browse ends when the language model says it's done, or when its calls run out, at eight for a test and twenty for the outage. A last message then asks for the final answer.

Keeping the whole conversation would have been simpler, and it could have overflowed. The instructions alone come to 645 tokens. The tool answers I tried for the outage came to between 65 and 368 tokens each. Twenty answers like those, with the language model's replies between them, come close to 8,192 tokens, and longer answers would pass it. Because each message carries nothing older, a stopped run can also resume at any step. The price is that the language model forgets whatever it doesn't write into its view. That's deliberate. The view is its working memory, and its own summary of what it thought mattered. For the full record there's the results file, which keeps every message and every reply.

The view is a SysML v2 idea, not one I invented. In SysML v2 a viewpoint frames a concern, a question somebody needs answered, and a <span class="term" data-term="view">view</span> exposes the elements that answer it. At the end of each browse, the program writes the language model's view out as SysML v2. Here's the form, with two elements a browse of the test at the top might expose, as an example:

```
package TestView {
  viewpoint def WhichRequirementDoesThisTestVerify {
    frame concern question {
      doc /* Which requirement does this test verify? */
    }
  }
  view test {
    viewpoint answers : WhichRequirementDoesThisTestVerify;
    // edits patch the source, which the test checks
    expose Federation_SystemStories::SR_22_EditsPatchTheSource;
    // the test lives in the adapter's code
    expose Federation_LogicalArchitecture::demo::adapter;
  }
}
```

Both reference tools accept this one beside the systems model, and both refuse a view that exposes an element the systems model lacks. A person can open the view next to the systems model it came from, in any tool that reads SysML v2 text. It shows which elements the language model thought mattered, with its note on each.

## Two tasks

The first task is the one at the top, set for every test in the demo's own code. There are 169 of them, 23 more than part 4 counted, since running the scenarios took tests of its own. Of the 169, 69 carry a key. For each test, the language model proposes up to five elements that the test verifies or exercises, the requirement first. It backs the answer with up to five words or phrases copied from the test, and one sentence of reason. What gets scored is the first requirement it names, or none if it names none.

Four resolvers are scored on the 69 keyed tests. A key rule reads the key before it's hidden, so its links are correct by construction, and it's there to show what hiding the keys takes away. Word overlap compares the test's name, package, file and doc comment with each requirement's name and statement, weighting a word by how few requirements use it. It proposes the best one above a threshold I set before its first run. The systems model baseline makes the same comparison with every element the test task shows, the way `find` does. It takes the first ranked element that is a requirement, or that one satisfy, verify or derive link joins to one. For the test at the top it passes over the scenario entry ranked first, which no such link joins to a requirement, and follows the verify link of `VC_SR_22`, ranked second, to `SR-22`. Neither baseline needs a language model, so both have already run, and their figures wait for part 7. Last comes the language model. For the other 100 tests there's no recorded link to score against. Whatever the language model proposes for them goes on a list for a person to judge, first without its reasons and then with them, since a fluent reason can sway a judge.

The second task is an outage, and it's hypothetical. Nobody's router died at two in the morning. The demo has three monitors, which Checkly runs for the length of a check session, asking every ten minutes for the viewer app, the document and the root page. A fourth check, the session's heartbeat, waits for pings from the script that runs the session, in a container of its own. Here's the report from the engineer on call:

> The alert "monitor: the viewer answers" has failed since 02:10. It asks for /viewer/ every ten minutes from eu-central-1 and gets nothing back. The service's container has stopped, and its last line reads: sysml-federation serve: router exited: signal: killed

Everything in the report is real except the event. The monitor exists. That last line is the one the demo prints when its router dies, and a test with a stand-in router shows the supervisor shutting everything down when it does. From the report, the language model browses the whole systems model and the code. It gives the cause, the mechanism, the code to read first, the consequences, the path it followed, and two or three sentences for the engineer.

Before any run, I wrote down and committed an answer key: twelve things a good account names.

1. The router is the cause.
2. The supervisor's state machine is the mechanism.
3. So is its transition on the router's exit.
4. The supervisor's code, `serve.go`, is the first file to read.
5. So is the compose file that runs the demo, which sets no restart policy, so nothing brings the container back.
6. `SR-04`, which puts the demo's four paths on one port, goes unmet.
7. `SR-09` keeps the demo's state in memory, so every visitor's edit is lost with the container.
8. The document's monitor fails too.
9. So does the root page's.
10. A stakeholder story behind `SR-04` is affected.
11. So is the one behind `SR-09`.
12. The session's heartbeat keeps passing, and reports a healthy session while the demo is down.

An item counts as found when the account cites one of the identifiers the answer key accepts for it, and missed otherwise, so nobody judges a paraphrase. The answer key can't score one thing a good account would say, since it's neither an element nor a file. Files built into the demo's own web server make up the viewer app's page, and the router plays no part in sending it. So the failed monitor says nothing about the viewer app itself. What it shows is that the supervisor stops everything when any one part exits.

The answer key isn't out of reach. I ran the first calls a browse might make myself, with no language model involved. A `find` with the whole report lists `CHK_MonitorViewer`, the viewer app's monitor, first. The document's and the root page's monitors come third and fourth, and `SupervisorStates::onRouterExit` fifth. From the viewer app's monitor, `path` goes through `SR-04` to the router in two links, and from the router to `SupervisorStates` in three, by way of the supervisor's child process for the router. `code` on the state machine gives `serve.go`, and `grep` for "router exited" finds line 143 of that file. Five calls name seven of the twelve items: the cause, the mechanism and its transition, the first file to read, `SR-04` and the two other monitors. What's left for the language model is choosing which of those answers to follow, knowing when it has enough, and telling an engineer at two in the morning what it all means.

Two more checks apply to every account. Every element, link and file it cites has to be one the browse's tools could have shown. And each step of its path has to join the next, through a link in the systems model, a file the `code` tool gives for the element, or a file the browse read.

The run records its commit and hashes of the tests and the systems model. The answer key was committed on 25 September, and the systems model has grown since. Most recently it gained a second place the demo can run, on my own server, from the OpenTofu configuration part 4 took its gold set from. None of the answer key's items changed, and nothing in the report says which place stopped. The monitors only exist during a check session, which runs the demo in its compose stack, so the answer key stands, and a careful browse has one more place to rule out.

## Testing what it cites

This is where a sceptic should push, and the push is well founded. A language model's explanation is more text it generates, and there's good evidence that such text can be a plausible story and not the reason. Turpin and colleagues nudged language models towards particular answers, then read [426 explanations of answers that had followed the nudge](https://arxiv.org/abs/2305.04388). One of them mentioned it. For entity resolution in particular, Teofili and colleagues found language models' explanations of their own decisions "often unstable, weakly faithful, and poorly aligned with counterfactual evidence" ([PVLDB 19, 2026](https://arxiv.org/abs/2606.01210)).

Jacovi and Goldberg drew the distinction that makes this matter. Plausibility is how convincing an explanation is to people, and faithfulness is how accurately it reflects the reasoning behind the answer. They warn that ["a plausible but unfaithful interpretation may be the worst-case scenario"](https://aclanthology.org/2020.acl-main.386/). A traceability tool that explains its links convincingly and unfaithfully would be that worst case, because its reviewers would have every reason to believe it.

So the experiment doesn't take the explanations on trust. It changes the input and watches the answer. The tests it does this to are fixed by rule: the first test the language model links to a requirement, and every fourth such test after it, in the order the tests are read. Each is browsed again with one change, or with none for the repeat:

| Probe | What changes | What to expect if the cited words carry the answer |
|---|---|---|
| Deletion | The cited words are deleted from the test and from every answer the tools give. | The answer changes often. |
| Random control | As many of the test's uncited words are deleted, from the same places, chosen with a fixed seed. | The answer changes much less often than with deletion. |
| Rare-shared control | The uncited words the test shares with the requirement it linked are deleted, from the same places, rarest first, until as many are gone. | The answer still changes less often than with deletion. |
| Reconstruction | The test is replaced by the cited words alone, and its own code is hidden from the tools. | The same answer comes back. |
| Repeat | Nothing. | The same final reply, word for word. |

An answer counts as changed when its first requirement changes, or, for the repeat, when the reply differs at all.

The rare-shared control is the one I'd point a sceptic to. A random deletion is a weak comparison. The words a language model cites are usually the most telling ones in the test, so deleting them would change answers more often even if the citation were invented afterwards. The rare-shared control deletes the most telling words it can find that weren't cited. If the language model cited the rarest shared words itself, the ones left are a little less telling, which weakens this control, and I can't fix that without a larger test set. Deletion is close to the erasure measure Teofili and colleagues used, and reconstruction is one of the two tests [Atanasova and colleagues](https://aclanthology.org/2023.acl-short.25/) proposed for written explanations. The repeat is there because I found nobody who had tested whether Ollama gives identical output at temperature 0 with a fixed seed, the settings meant to take the chance out of its replies. [Its documentation](https://docs.ollama.com/modelfile) says that setting the seed to a specific number "will make the model generate the same text for the same prompt". The repeat checks.

One detail of the answer's shape matters here. A language model writes its reply one piece after another and can't go back. The final answer lists the links before the evidence, so the evidence is written after the answer is chosen. The probes test the evidence. Nothing tests the sentence of reason, or the notes in the view.

The outage is browsed five times. It goes as reported, then again unchanged, and then with the alert alone, without the container's last line. The fourth browse runs with `onRouterExit` taken out of the systems model, and the fifth with an unrelated allocation taken out instead, as a control. An account that rested on the transition should score differently when it's gone, or say that the systems model no longer explains what the code does. Citing a transition that no longer exists gets flagged, however good the sentence around it.

Every final answer also says where the language model suspects the system wasn't built as the systems model says. One mismatch is known, and I've left it in on purpose. According to the systems model, the adapter's `serve` package holds the store and its version counter. In the code they sit in `adapter/projection/store.go`, and the whole of `adapter/serve/server.go` is 33 lines. That part of the systems model came from a design article, and the code, written the same day as the article, put them somewhere else. Any suspicion the language model raises can be checked against that one, and against the code for whatever else it turns up.

## What will count for and against it

The scoring was fixed with the code, before any run, and this part fixes how I'll read it. On the 69 keyed tests each resolver gets precision and recall, each with the Wilson interval part 4 described. The language model's first requirement is compared with each baseline's, test by test, with McNemar's exact test, two-sided, and no answer counts as incorrect. McNemar's test looks only at the tests where the two disagree. A test both resolvers link correctly says nothing about which is better, and both baselines already link the one at the top correctly. The comparison that decides is the one with the systems model baseline. It reads the same systems engineering model through the same search, with no language model to choose where to go next, so a gain over it comes from the language model's choices. Word overlap's comparison is reported beside it. I'll call a difference real at p below 0.05, and at no other threshold. The probes are paired the same way. On the tests where deletion and a control were both asked, the report counts which of the two changed the answer, and applies the same test at the same threshold.

One more figure sets a ceiling on the first step. The report counts the keyed tests whose requirement is among the first eight elements the search ranks for the test's own words, or one trace link from one of them. For every other test, the answer is out of reach of that search and one link, and a browse has to search again or go further.

Links to elements that aren't requirements have nothing recorded to score them against. The report counts them by kind, and gives the shares that lie within one, two and three links of the recorded requirement, each distance apart. Part 4 argued for one or two links, and the numbers show why each share needs a comparison beside it. From `SR-22`, one link reaches 16 elements, the adapter among them. Two links reach 157, and three reach 515 of the 1,277 elements the test task shows, more than two fifths of the systems model. The other 69 of the 1,346 are the verification register's test entries, which the task hides. So beside each share the report gives the share of every element that lies as near. Even that comparison flatters the language model, since what it proposes comes from what it browsed, and what it browsed is near by construction. That's why the tools record every element their answers name, and the report also gives the share of those that lie as near. That's the share a near miss has to beat.

The experiment's limits need stating as plainly as its tests:

- It's one repository, one language model at one quantisation (the Q4_K_M build Ollama ships), one set of instructions and no tuning.
- There's one outage, and the person who chose it also built the systems model, wrote the answer key and designed the tools. I added the state machine that carries its mechanism while building the experiment.
- Every one of the 69 recorded links sat in a name, so they were the easy ones to write down. Several tests share a requirement, which makes the intervals narrower than they should be, and weakens McNemar's test the same way.
- I'll judge the 100 unkeyed tests alone.
- The outage's fifth browse is the only control for the fourth.
- Part 4's OpenTofu resources aren't in either task, since the experiment was built before them, and its code tools don't read OpenTofu files.

One thing in its favour: Qwen2.5-Coder was released in November 2024, and the demo's repository was started in August 2026, so the language model can't have seen these links in training.

These are the results that would count for the approach and against it, stated before any exist:

1. **A gain over the systems model baseline.** If McNemar's test finds the language model correct more often where the two disagree, its choices added something that following the links alone didn't. If it finds no difference, part 7 will say that on this data the language model gained nothing measurable over a program following the same links. On 69 tests that's weaker than showing it has none, and I'll say that too. A gain over word overlap alone doesn't change this, since it could come from the systems model and not the language model.
2. **Cited words that carry the answer.** Suppose the paired test finds that deleting the cited words changes answers more often than the rare-shared control does, and reconstruction mostly gives the same answer back. Then the words a reviewer is shown are the ones the answer rested on. If deletion changes answers no more often than the control, the evidence isn't what decided the answer, and a reviewer shouldn't be shown it as if it were. Reconstruction that rarely gives the same answer says the same.
3. **Replies that repeat.** If every repeat gives the same reply word for word, a run can be reproduced, which part 5 argued an auditor needs. If one differs, it can't be, as it stands.
4. **An account that follows its own path.** If the outage's account names the cause, the mechanism and the first file to read, an engineer could start from it. If it then changes when `onRouterExit` is taken out of the systems model, or says the systems model no longer explains the code, the path it reported is the one it used. If it stays the same, still citing the missing transition, it isn't.

My position before the run is narrow. If the language model beats the systems model baseline and its cited words hold up, I'd let a language model like this suggest links for a person to confirm, with its cited words beside each. The 100 tests nobody has linked would each get a proposal to accept or reject. If the outage's account holds up as well, an engineer paged at two in the morning would have the first file to read, and a list of what else is down, before opening the code. If it beats the baseline and its cited words don't hold up, it can still suggest links. Its reasons and notes go in the bin, and a reviewer sees only the link and the elements it looked at. If it doesn't beat the baseline, I'd run the baseline, which is cheaper and says exactly why: the words it matched and the link it followed. Either way, I wouldn't let a language model's explanation stand as the record of why a link exists.

I took the key off the test at the top so that an answer could be checked against it. The probes do the same to the explanations, taking away what the language model cited to see whether the answer goes with it.

[Part 7](19-why-did-the-language-model-go-round-in-circles.md) runs all of this on the server in my network and reports what happened, whichever way it went.

---

Previous: [What makes a resolver worth running?](17-what-makes-a-resolver-worth-running.md) · Index: [Automating traceability](../README.md) · Next: [Why did the language model go round in circles?](19-why-did-the-language-model-go-round-in-circles.md)
