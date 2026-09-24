# Use cases

The storyboard is one PDF, [use-cases.pdf](use-cases.pdf): an overview board, then one board per use case in the order below. This page is its text form, with the criteria each use case is judged by. The story around it is in [Twelve use cases and one moving bottleneck](../articles/04-twelve-use-cases-and-one-moving-bottleneck.md).

Values follow the design brief: ingest 2000, parse 1200, indexA 700, indexB 700, serve 1800, and a limit of 1500 on PIPE-R1. The derived limits are 1500 for ingest, parse and serve and 750 for each index server.

## Personas

- **The visitor** is a developer or architect at an engineering organisation of fewer than 25 engineers, evaluating whether federation could connect the tools they already have. "They have Docker, a browser and fifteen minutes, and they will not read a manual."
- **The model owner** is a systems engineer who writes the SysML v2 model and wants it to stay the source of truth.
- **The document owner** is a requirements engineer who owns the ordering, numbering and inclusion decisions of a requirements document, and "has never seen SysML, and must never need to".

## What can be edited

Exactly six numbers are editable in the apps: the five server throughputs and the limit of PIPE-R1. Everything else is read-only in both apps. The adapter's mutation accepts any literal in the source, including the 200 ms of PIPE-R2, which the playground can reach. Edits land in the served model text and its version counter, never on disk ([editing as scaffolding](../decisions/AD-0004-editing-as-scaffolding.md)).

A throughput that isn't a finite number, or is negative, is refused and the previous value stands.

## The twelve use cases

Each is written persona first. Its criteria say what is observed, not which control is used.

**1. Launch with one command (visitor).** With the image already pulled, `docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation` renders the viewer at `http://localhost:8080/viewer/` and the document at `http://localhost:8080/document/` within ten seconds. The pull is outside the ten seconds. With no route to the internet, either app renders fully and every request it makes goes to the one published port.

**2. Read the model (visitor).** The viewer's text pane shows the model file with the servers, their throughputs, the connections, the requirements and their short names. The sketch shows the five servers left to right as wired, each with its throughput, the capacity of 1200 and parse marked as the bottleneck. PIPE-R1 shows its limit of 1500, FAIL, and a reason naming parse at 1200.

**3. Raise a server that is not the bottleneck (visitor).** Ingest goes from 2000 to 3000. Capacity stays 1200, PIPE-R1 stays FAIL and the bottleneck stays parse. The only other change is PIPE-R1.1 passing with its new value. Ingest is used rather than an index server because raising an index server could flip that server's own derived requirement, which is use case 10's lesson. Parse set to 0 gives capacity 0, PIPE-R1 failing, and a reason naming parse at 0.

**4. Raise the bottleneck (visitor).** Parse goes from 1200 to 1700. Capacity becomes 1400, PIPE-R1 stays FAIL, and the bottleneck moves to indexA and indexB. With parse at 1700, indexA goes from 700 to 900. Capacity becomes 1600, PIPE-R1 becomes PASS with a reason naming the index pair at 1600 against 1500, and its block is no longer red.

**5. Tighten the limit (model owner).** The limit of PIPE-R1 set to 1000 gives PASS. Set to 2500 it gives FAIL, the block turns red, and the reason names parse at 1200. Whatever value is edited in the viewer, the model text read back shows the edited number exactly where the original literal was.

**6. Read the document (document owner).** The shipped structure is PIPE-R1 as section 1, its derived requirements as 1.1 to 1.5, and PIPE-R2 as section 2, numbers that belong to the document and not to the model. Each requirement shows its short name, text and limit, and the requirement it derives from or those derived from it. It also shows the part that satisfies it, the verification case that verifies it, the current value of its subject, and its verdict with a reason. PIPE-R2 is INCONCLUSIVE, with the reason "PIPE-VC1 is declared and no service runs it".

**7. Reorder and nest (document owner).** PIPE-R1.5 moved above PIPE-R1.1 becomes 1.1, and the others shift to 1.2 to 1.5 in their former order with nothing else renumbered. PIPE-R2 moved under PIPE-R1 becomes 1.6, and its relationships in the model are unchanged and still shown. The viewer's text and sketch are exactly as they were.

**8. Shape the document (document owner).** A heading "Performance" inserted as the parent of PIPE-R1 takes number 1, PIPE-R1 becomes 1.1 and its derived requirements 1.1.1 to 1.1.5. A paragraph added under the heading appears in place and carries no number. Excluding PIPE-R1.4 removes it from the document, PIPE-R1.5 becomes 1.1.4, and the model still lists PIPE-R1.4. Restoring it returns it as the last child of PIPE-R1, numbered 1.1.5.

**9. Change a value from the document (document owner).** With the viewer open in another tab, the throughput of parse is changed from 1200 to 1700 in the row of PIPE-R1.2. PIPE-R1.2 becomes PASS, and PIPE-R1 stays FAIL with a reason naming indexA and indexB at 1400. Within two seconds, and without a reload, the viewer shows 1700 in its text and capacity 1400 with the index pair marked. A change to the limit of PIPE-R1 in its row reaches the viewer's requirement text the same way.

**10. Change from the viewer and watch the document (visitor).** Starting from parse at 1700, indexA is raised from 700 to 900 in the viewer. Within two seconds and without a reload, the document shows PIPE-R1 as PASS with a reason naming the index pair at 1600 against 1500. PIPE-R1.3 shows PASS, and PIPE-R1.4 still FAIL with 700 against its limit of 750. The document's unnumbered first paragraph explains why.

**11. One query in the playground (visitor).** One query for the text, verdict and document number of PIPE-R1 returns all three in one response, and the served schema contains types contributed by all three subgraphs.

**12. Reset (visitor).** The reset control in either app returns both apps to the shipped values and document structure within two seconds.

---

Index: [Federating a systems model](../README.md) · Repository: https://github.com/Roarge/sysml-federation
