package Federation_ValidationCases {
    doc /* The validation case register. One case per stakeholder story,
           VAL_US_nn, naming the demo as its subject and reading that story's
           acceptance in its objective. For the storyboard stories the actions
           are the recorded runs that exercised the story: a row of the example's
           verification record appears under every story whose requirements it
           bears on, so one row can stand under several cases. The stories added
           with the model were exercised by the system stories' own cases rather
           than by a run of the demo, and their doc names those cases instead. */
    private import Federation_StakeholderStories::*;
    private import Federation_Context::Demo;
    private import Federation_Library::Evidence;

    // ------------------------------------------- the storyboard, US-01 to US-12 --

    verification def VAL_US_01 {
        doc /* Launch with one command: the one command, the start, the paths,
               the platforms, the size and the run with no route out. */
        subject demo : Demo;
        objective { verify US_01_LaunchWithOneCommand.acceptance; }
        action <'record: 2026-08-27 SR-04'> pathsOnTheRunningStack {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The running stack answered the redirect and the three app paths on the one port. */
        }
        action <'record: 2026-08-28 SR-10'> requestListOnReload {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A reload of the viewer made every one of its eight requests to the one published port. */
        }
        action <'record: 2026-08-28 SR-04'> pathsOnTheContainer {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The container answered the redirect, the three paths, a query and a held subscription on port 8080. */
        }
        action <'record: 2026-08-28 SR-02'> containerStartMeasured {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The viewer answered after 0.75 s and a query through the router after 0.77 s. */
        }
        action <'record: 2026-08-28 SR-06'> layersMeasured {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Both platforms came in under the size budget, so the image the one command pulls is the small one. */
        }
        action <'record: 2026-08-28 cross-build router binary'> arm64BinaryResolves {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The arm64 build gave ARM binaries, so the command works on the other architecture as well. */
        }
        action <'record: 2026-08-29 no outbound connection'> noConnectionWithTheRouteWithheld {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* With no network interface the container still reached healthy and named no address but the loopback. */
        }
        action <'record: 2026-08-29 SR-03'> fourVariablesInTheChild {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The router child carried the four variables that switch its outbound paths off. */
        }
        action <'record: 2026-08-29 SR-01'> oneCommandUnauthenticated {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The one docker run line pulled the published image from a shell holding no credential and served the demo. */
        }
        action <'us01-launch.spec.ts'> bothAppsRenderAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us01-launch.spec.ts"; }
            doc /* The viewer and the document rendered from the one published port and every request either page made went to it, checked against a running instance. */
        }
    }

    verification def VAL_US_02 {
        doc /* Read the model: the text, the sketch and the verdict as the
               viewer serves them. */
        subject demo : Demo;
        objective { verify US_02_ReadTheModel.acceptance; }
        action <'record: 2026-08-28 SR-11'> notationDrawn {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The text pane drew the served file with keywords, numbers, quoted names and doc bodies distinguished. */
        }
        action <'record: 2026-08-28 SR-12'> sketchAndSecondWiring {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The sketch showed the five servers left to right with their throughputs, the capacity and the bottleneck marked. */
        }
        action <'record: 2026-08-28 SR-14, SR-15'> verdictsAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The requirements column carried the limit, the verdict and the reason for every requirement. */
        }
        action <'record: 2026-08-28 SR-12, SR-14, SR-22'> viewerUnderEdits {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The same three readings held while values were edited in the viewer's panel. */
        }
        action <'us02-read-the-model.spec.ts'> textSketchAndVerdictAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us02-read-the-model.spec.ts"; }
            doc /* The text pane, the sketch with its capacity and bottleneck, and the verdict of PIPE-R1 with its reason, read from a running instance. */
        }
    }

    verification def VAL_US_03 {
        doc /* Raise a server that is not the bottleneck: nothing moves, a bad
               value is refused, and zero stops the chain. */
        subject demo : Demo;
        objective { verify US_03_RaiseAServerThatIsNotTheBottleneck.acceptance; }
        action <'record: 2026-08-27 SR-25'> negativeThroughTheRouter {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A negative throughput sent through the router was refused and the previous value stood. */
        }
        action <'record: 2026-08-28 SR-12'> sketchAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The sketch showed the bottleneck before anything was raised. */
        }
        action <'record: 2026-08-28 SR-14, SR-15'> verdictsAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The verdicts and reasons before the edit, which the raise is read against. */
        }
        action <'record: 2026-08-28 SR-13, SR-38'> controlsCounted {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The throughput of each server was editable and nothing else was. */
        }
        action <'record: 2026-08-28 SR-12, SR-14, SR-22'> raisingIngestChangedNothing {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Raising ingest from 2000 to 3000 left the capacity at 1200 and the bottleneck where it was, and setting the bottleneck to zero stopped the chain. */
        }
        action <'record: 2026-08-28 SR-25'> badValueInEitherApp {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Letters, a negative number and an emptied field were refused in both apps and the served value came back. */
        }
        action <'record: 2026-08-28 Refusal text'> refusalTextReachesTheClient {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A refusal reached the client as the service's own message, so what a visitor sees is what the service said. */
        }
        action <'us03-nothing-moves.spec.ts'> nothingMovesAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us03-nothing-moves.spec.ts"; }
            doc /* Ingest raised to 3000 in the viewer of a running instance left the capacity, the bottleneck and the verdict of PIPE-R1 where they were, with PIPE-R1.1 passing on the new value. */
        }
    }

    verification def VAL_US_04 {
        doc /* Raise the bottleneck: the capacity rises, the bottleneck moves
               and the requirement turns from red. */
        subject demo : Demo;
        objective { verify US_04_RaiseTheBottleneck.acceptance; }
        action <'record: 2026-08-28 SR-12'> sketchAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The sketch showed the bottleneck stroked red before the raise. */
        }
        action <'record: 2026-08-28 SR-14, SR-15'> verdictsAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The failing requirement was red and the others were not. */
        }
        action <'record: 2026-08-28 SR-13, SR-38'> controlsCounted {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The bottleneck's throughput was one of the six editable values. */
        }
        action <'record: 2026-08-28 SR-12, SR-14, SR-22'> bottleneckMoved {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Raising the bottleneck to 1700 gave capacity 1400 with the index pair red, and raising one of the pair to 900 gave 1600 and a pass. */
        }
        action <'us04-bottleneck-moves.spec.ts'> bottleneckMovesAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us04-bottleneck-moves.spec.ts"; }
            doc /* Parse raised to 1700 in the viewer of a running instance moved the bottleneck to the index pair at capacity 1400, and indexA raised to 900 gave 1600 and a pass. */
        }
    }

    verification def VAL_US_05 {
        doc /* Tighten the limit: the verdict follows the limit and the edited
               number appears in the served text. */
        subject demo : Demo;
        objective { verify US_05_TightenTheLimit.acceptance; }
        action <'record: 2026-08-28 SR-14, SR-15'> verdictsAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The verdict and the reason as served, which the change of limit is read against. */
        }
        action <'record: 2026-08-28 SR-13, SR-38'> controlsCounted {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The limit of the global requirement was editable and no derived limit was. */
        }
        action <'record: 2026-08-28 SR-12, SR-14, SR-22'> limitChangedBothWays {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The limit at 1000 gave a pass and at 2500 a red failure naming the bottleneck, with the new number in the served text where the old literal was. */
        }
        action <'us05-tighten-limit.spec.ts'> limitChangedBothWaysAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us05-tighten-limit.spec.ts"; }
            doc /* The limit of PIPE-R1 at 1000 gave a pass and at 2500 a red failure naming parse, with the new number in the served text where the old literal was, in the viewer of a running instance. */
        }
    }

    verification def VAL_US_06 {
        doc /* Read the document: the numbering, the rows and the depth the app
               shows. */
        subject demo : Demo;
        objective { verify US_06_ReadTheDocument.acceptance; }
        action <'record: 2026-08-28 SR-33, SR-37'> documentAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A freshly started stack showed the sections numbered from the tree and each row carrying everything the model and the analysis know about it. */
        }
        action <'record: 2026-08-28 SR-33, SR-34'> numberingAfterAHeading {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The numbering stayed the document's own through a heading, a prose node, an exclusion and a restore. */
        }
        action <'record: 2026-08-28 Depth limit'> depthLimit {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* At the sixth level the app says how deep the document is shown and offers only what still applies. */
        }
        action <'us06-read-document.spec.ts'> documentAsServedAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us06-read-document.spec.ts"; }
            doc /* The numbering, the row of PIPE-R1.4 with everything the model knows about it, and PIPE-R2 INCONCLUSIVE, read from the document of a running instance. */
        }
    }

    verification def VAL_US_07 {
        doc /* Reorder and nest: the order changes, the numbering follows and
               the model stays as it was. */
        subject demo : Demo;
        objective { verify US_07_ReorderAndNest.acceptance; }
        action <'record: 2026-08-27 SR-36'> versionsThroughTheRouter {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Moves driven through the router raised the document version and left the model version where it was. */
        }
        action <'record: 2026-08-27 SR-41'> importSweep {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The document service reaches no other service, which is why a move cannot touch the model. */
        }
        action <'record: 2026-08-28 SR-33, SR-37'> documentAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The shipped order and numbering the moves were read against. */
        }
        action <'record: 2026-08-28 SR-34'> dragsAndNesting {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A requirement dragged above another took its number, and one dragged under another was renumbered as its child while keeping its relationships. */
        }
        action <'record: 2026-08-28 SR-33, SR-34'> headingProseAndExclusion {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The same tree took a heading, a prose node, an exclusion and a restore, renumbering each time. */
        }
        action <'record: 2026-08-28 SR-36'> countersInTheHeader {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Through every move the document counter rose and the model counter did not. */
        }
        action <'record: 2026-08-28 Depth limit'> depthLimit {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Nesting stops at the sixth level, and the app says so rather than refusing silently. */
        }
        action <'us07-reorder-and-nest.spec.ts'> reorderAndNestAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us07-reorder-and-nest.spec.ts"; }
            doc /* PIPE-R1.5 moved above PIPE-R1.1 took 1.1 and the others shifted, PIPE-R2 moved under PIPE-R1 took 1.6 and kept its relations, and the model version stood, on a running instance. */
        }
    }

    verification def VAL_US_08 {
        doc /* Shape the document: headings, prose, an exclusion and a restore,
               with the model untouched. */
        subject demo : Demo;
        objective { verify US_08_ShapeTheDocument.acceptance; }
        action <'record: 2026-08-27 SR-36'> versionsThroughTheRouter {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A heading and an exclusion driven through the router raised the document version alone. */
        }
        action <'record: 2026-08-27 SR-41'> importSweep {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The document service holds requirement keys and structure and reaches nothing else. */
        }
        action <'record: 2026-08-28 SR-33, SR-37'> documentAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The shipped structure the editorial changes were read against. */
        }
        action <'record: 2026-08-28 SR-34'> dragsAndNesting {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The moves that put a requirement where the heading was to be inserted above it. */
        }
        action <'record: 2026-08-28 SR-33, SR-34'> headingProseAndExclusion {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A heading took the number, its children renumbered, a paragraph appeared unnumbered, an excluded requirement left the document and stayed in the model, and a restore returned it as the last child. */
        }
        action <'record: 2026-08-28 SR-36'> countersInTheHeader {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The model counter did not move through any of it. */
        }
        action <'record: 2026-08-28 Depth limit'> depthLimit {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Adding headings stops at the sixth level, with the reason on the row. */
        }
        action <'us08-shape-the-document.spec.ts'> shapeTheDocumentAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us08-shape-the-document.spec.ts"; }
            doc /* A heading took number 1 and pushed PIPE-R1 to 1.1, a paragraph appeared unnumbered, PIPE-R1.4 left for the tray and came back last on restore, and the model version stood, in the document of a running instance. */
        }
    }

    verification def VAL_US_09 {
        doc /* Change a value from the document: the correction is made where
               the document owner works and the viewer follows. */
        subject demo : Demo;
        objective { verify US_09_ChangeAValueFromTheDocument.acceptance; }
        action <'record: 2026-08-27 SR-26'> versionAfterEachMutation {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Each accepted mutation put the new model version on the subscription within a second. */
        }
        action <'record: 2026-08-28 SR-12'> sketchAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The viewer's sketch before the correction, which the change is read against. */
        }
        action <'record: 2026-08-28 SR-13, SR-38'> controlsCounted {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The document offered the same six editable values as the viewer. */
        }
        action <'record: 2026-08-28 SR-33, SR-37'> documentAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The row the correction was made in, carrying the value as a control beside the verdict. */
        }
        action <'record: 2026-08-28 SR-12, SR-14, SR-22'> editedNumberInTheText {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* An edited number appeared in the served text exactly where the original literal was. */
        }
        action <'record: 2026-08-28 SR-39, SR-44'> bothAppsFollow {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A change made in one app was picked up by the other within two seconds. */
        }
        action <'record: 2026-08-28 SR-39 reconnection'> reconnectionAfterAStop {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* After the stack came back the next change reached the open page with no reload. */
        }
        action <'record: 2026-08-28 SR-39 background tab'> recoveryFromAFrozenTab {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A page the browser had frozen redrew the served state when it became visible again. */
        }
        action <'record: 2026-08-28 SR-39 edits in both apps'> correctionInTheDocument {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A throughput and a limit corrected in the document moved the verdicts there and reached the viewer's text and sketch without a reload. */
        }
        action <'us09-edit-from-document.spec.ts'> viewerFollowsTheDocumentAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us09-edit-from-document.spec.ts"; }
            doc /* The limit of PIPE-R1 and the throughput of parse corrected in the document of a running instance reached the viewer on a second page within two seconds, with no reload. Skipped where the session carries no stream. */
        }
    }

    verification def VAL_US_10 {
        doc /* Change from the viewer, watch the document: the other app moves
               by itself, which is what shows both read one graph. */
        subject demo : Demo;
        objective { verify US_10_ChangeFromTheViewerWatchTheDocument.acceptance; }
        action <'record: 2026-08-27 SR-26'> versionAfterEachMutation {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The version event each mutation raises is what the other app acts on. */
        }
        action <'record: 2026-08-27 SR-41'> importSweep {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* No service reads another, so the document can only be following the graph. */
        }
        action <'record: 2026-08-28 SR-33, SR-37'> documentAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The document rows before the edit, which the change is read against. */
        }
        action <'record: 2026-08-28 SR-39, SR-44'> bothAppsFollow {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A change made in one app was picked up by the other within two seconds. */
        }
        action <'record: 2026-08-28 SR-39 reconnection'> reconnectionAfterAStop {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* An edit made elsewhere reached the open viewer after the stack came back, with a focused control keeping focus. */
        }
        action <'record: 2026-08-28 SR-39 background tab'> recoveryFromAFrozenTab {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A page two changes out of date redrew the served version, values, wiring and verdicts when it became visible. */
        }
        action <'record: 2026-08-28 SR-39 edits in both apps'> editsFromTheViewer {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Two values raised in the viewer moved the document's verdicts and reasons with no reload. */
        }
        action <'us10-viewer-to-document.spec.ts'> documentFollowsTheViewerAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us10-viewer-to-document.spec.ts"; }
            doc /* IndexA raised to 900 in the viewer of a running instance, with parse at 1700, moved the document's verdicts on a second page within two seconds, with no reload. Skipped where the session carries no stream. */
        }
    }

    verification def VAL_US_11 {
        doc /* Query the graph: one query answered from all three services, at
               an endpoint on the one published port. */
        subject demo : Demo;
        objective { verify US_11_QueryTheGraph.acceptance; }
        action <'record: 2026-08-27 SR-43'> oneQueryThreeServices {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* One query for a requirement's text, verdict and document number was answered in one response, and introspection listed types from each subgraph. */
        }
        action <'record: 2026-08-27 SR-04'> playgroundOnTheRunningStack {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The playground answered on the same port as the two apps. */
        }
        action <'record: 2026-08-28 SR-04'> playgroundOnTheContainer {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The same held for the container, where a query on the endpoint answered beside the three app paths. */
        }
        action <'us11-query-the-graph.spec.ts'> oneQueryThreeServicesAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us11-query-the-graph.spec.ts"; }
            doc /* One query for the text, the verdict and the document number of PIPE-R1 answered in one response by a running instance, its schema carrying a type from each subgraph, and the playground loading beside it. */
        }
    }

    verification def VAL_US_12 {
        doc /* Reset: the shipped state comes back in both apps, and a restart
               returns the demo to where it started. */
        subject demo : Demo;
        objective { verify US_12_Reset.acceptance; }
        action <'record: 2026-08-27 SR-09'> restartAfterEdits {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A restart after the recorded edits came back on the shipped values and the shipped structure. */
        }
        action <'record: 2026-08-27 SR-44'> resetThroughTheRouter {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Both reset mutations through the router brought the shipped verdict back. */
        }
        action <'record: 2026-08-28 SR-39, SR-44'> resetFromEitherApp {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Reset pressed in each app in turn showed the shipped values in both within two seconds. */
        }
        action <'record: 2026-08-28 SR-39 reconnection'> reconnectionAfterAStop {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* After a stop and a restart the open page picked the demo up again without a reload. */
        }
        action <'record: 2026-08-28 SR-39 background tab'> recoveryFromAFrozenTab {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A frozen page redrew the state it had missed rather than needing a reload before the next showing. */
        }
        action <'record: 2026-08-28 SR-39 edits in both apps'> editsBeforeTheReset {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The edits in both apps that the reset was pressed on top of. */
        }
        action <'us12-reset.spec.ts'> resetFromTheViewerAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/us12-reset.spec.ts"; }
            doc /* After an edit in each app, Reset in the viewer of a running instance brought both apps back to the shipped state within two seconds. Skipped where the session carries no stream. */
        }
    }

    // ----------------------------- added with the model, US-13 to US-19 --

    verification def VAL_US_13 {
        doc /* Licences travel with what is copied. Exercised by the system
               stories' cases rather than by a run of the demo: VC_SR_07 for the
               text beside the copied binary and VC_SR_08 for the notice and the
               text beside the vendored file. */
        subject demo : Demo;
        objective { verify US_13_LicencesTravelWithWhatIsCopied.acceptance; }
    }

    verification def VAL_US_14 {
        doc /* Nothing of the example in the adapter. Exercised by VC_SR_17,
               which searches the adapter source and projects a second fixture
               that shares no name with the example. */
        subject demo : Demo;
        objective { verify US_14_NothingOfTheExampleInTheAdapter.acceptance; }
    }

    verification def VAL_US_15 {
        doc /* A served model that is never silently wrong. Exercised by
               VC_SR_18 for the refusal that names file, line and column, by
               VC_SR_24 for the value bound by an expression that an edit cannot
               reach, and against a running instance by the refusals check. */
        subject demo : Demo;
        objective { verify US_15_AServedModelThatIsNeverSilentlyWrong.acceptance; }
        action <'refusals.multistep.spec.ts'> boundValuesStayBoundThroughTheRouter {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/refusals.multistep.spec.ts"; }
            doc /* An edit aimed at a limit bound by an expression is refused through the router, and the model version is the same before and after. */
        }
    }

    verification def VAL_US_16 {
        doc /* The contract checked before deployment. Exercised by VC_SR_42,
               which compares the committed configuration with the schema files
               it was composed from. */
        subject demo : Demo;
        objective { verify US_16_TheContractCheckedBeforeDeployment.acceptance; }
    }

    verification def VAL_US_17 {
        doc /* The example accepted by the reference tools. Exercised by
               VC_SR_45, which puts the example model and the second fixture to
               both tools. */
        subject demo : Demo;
        objective { verify US_17_TheExampleAcceptedByTheReferenceTools.acceptance; }
    }

    verification def VAL_US_18 {
        doc /* The demo's own model as the record. Exercised by VC_SR_46 for the
               agreement between the model and the repository, and by VC_SR_47
               for the two reference tools over the whole tree on every
               change. */
        subject demo : Demo;
        objective { verify US_18_TheDemosOwnModelAsTheRecord.acceptance; }
    }

    verification def VAL_US_19 {
        doc /* The stories exercised against a running instance. Exercised by
               VC_SR_48, whose evidence is the session record that a later change
               adds. */
        subject demo : Demo;
        objective { verify US_19_TheStoriesExercisedAgainstARunningInstance.acceptance; }
    }
}
