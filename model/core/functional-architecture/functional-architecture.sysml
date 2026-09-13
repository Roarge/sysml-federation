package Federation_FunctionalArchitecture {
    doc /* What the demo does, in four pieces. The first three correspond to a
           published board or sheet and are named against it. The fourth is
           the optional check session, which appears on no board. The
           successions carry the order and nothing else: no condition and no
           calculation is modelled here. */
    private import Federation_Domain::*;
    private import ScalarValues::*;

    // ------------------------------------------------------ the V3 runtime --

    package EditPropagation {
        doc /* The V3 runtime board, docs/img/v3-runtime.png, and the companion
               board that shows the same seven steps changing nothing,
               docs/img/v3-nothing-moves.png. */

        action def PropagateAnEdit {
            doc /* An edit in the viewer: parse raised from 1200 to 1700, and
                   the seven steps until both apps show it. Bounded by SR-39:
                   two seconds from the third step to the seventh. */
            action mutate { doc /* setAttribute(PIPE-S2, throughput, 1700) from the viewer to the one published port */ }
            action proxyAndPlan { doc /* the UI server proxies to the router, which plans the mutation to the adapter */ }
            action validatePatchRebuild { doc /* the adapter validates the value (SR-25), checks it is a literal (SR-24), patches the literal at its span, rebuilds the projection as version 2, and emits modelChanged: 2 */ }
            action pushVersionEvent { doc /* the router fans the event out over SSE to every open subscription, both apps */ }
            action refetch { doc /* each app refetches its query: parts and requirements, documentNumber and included by key, representations for capacity and verdict */ }
            action recomputeVerdict { doc /* the capacity service builds the flow network, finds the maximum flow 1400 and the cut indexA and indexB, and answers FAIL with the reason capacity 1400 against 1500, limited by indexA, indexB */ }
            action renderTheView { doc /* the viewer draws the red outline on indexA and indexB, and the document shows PIPE-R1.2 PASS and PIPE-R1's new reason */ }
            first mutate then proxyAndPlan;
            first proxyAndPlan then validatePatchRebuild;
            first validatePatchRebuild then pushVersionEvent;
            first pushVersionEvent then refetch;
            first refetch then recomputeVerdict;
            first recomputeVerdict then renderTheView;
        }

        action def NothingMoves {
            doc /* ingest raised to 3000. Steps 1 to 5 are identical. In step 6
                   the cut is still parse and every number and verdict is
                   unchanged. In step 7 only the number on ingest changes. The
                   event still fires and the apps still refetch. Nothing
                   detects a no-op. */
        }

        action def ADocumentEdit {
            doc /* A document edit. PIPE-R2 moved under PIPE-R1: moveNode(id,
                   parentId, index) reaches the document service. It reorders,
                   renumbers (SR-33), increments its version, emits
                   documentChanged. The viewer, not subscribed, does nothing.
                   The model version is unchanged (SR-36). */
        }

        action def Reset {
            doc /* Reset. One mutation from either app carries both root
                   fields, resetModel and resetDocument. Each service restores
                   its shipped state, increments its version and emits its
                   event, and both apps refetch (SR-44). If one field fails
                   the other is still applied. Both are sent because neither
                   service knows about the other (SR-41). */
        }
    }

    // ------------------------------------------ the L2b sheet, model side --

    package VerdictComputation {
        doc /* The functional view of the L2b sheet,
               docs/img/a3-l2b-model-side.png.

               The arithmetic the sheet tabulates, four rows. As shipped: the
               cut is parse, the capacity 1200, PIPE-R1 FAIL, and the derived
               verdicts PASS FAIL FAIL FAIL PASS. With ingest at 3000: the cut
               is parse, the capacity 1200, PIPE-R1 FAIL, and the derived
               verdicts PASS FAIL FAIL FAIL PASS. With parse at 1700: the cut
               is indexA and indexB, the capacity 1400, PIPE-R1 FAIL, and the
               derived verdicts PASS PASS FAIL FAIL PASS. With indexA at 900:
               the cut is indexA and indexB, the capacity 1600, PIPE-R1 PASS,
               and the derived verdicts PASS PASS PASS FAIL PASS. The derived
               verdicts are given in the order ingest, parse, indexA, indexB,
               serve. The limits are 1500 on the pipeline and on ingest, parse
               and serve, and 750 on each index server.

               No calculation is modelled here. The arithmetic is the capacity
               service's, and the rollup is a maximum flow (AD-0007). */

        action def ComputeCapacityAndVerdicts {
            doc /* How the number is made, and why raising one server changes
                   nothing and raising another changes everything. */
            action readServersAndWiring { doc /* the projection, via the router: servers and their wiring */ }
            action buildFlowNetwork { doc /* one edge per server, sized by its throughput */ }
            action findMaximumFlow { doc /* entry to exit servers: the capacity */ }
            action findSourceSideCut { doc /* saturated servers nearest the entry: the bottleneck */ }
            action compareWithEachLimit { doc /* PASS if capacity >= limit, else FAIL: verdicts and reasons */ }
            action raiseOutsideTheCut { doc /* ingest 2000 to 3000: nothing moves, the same number, the same cut */ }
            action raiseInsideTheCut { doc /* parse 1200 to 1700: the cut migrates to indexA and indexB. Then indexA 700 to 900: capacity 1600, PIPE-R1 PASS, and PIPE-R1.4 on indexB still fails, 700 against its allocated 750 */ }
            first readServersAndWiring then buildFlowNetwork;
            first buildFlowNetwork then findMaximumFlow;
            first findMaximumFlow then findSourceSideCut;
            first findSourceSideCut then compareWithEachLimit;
            first compareWithEachLimit then raiseOutsideTheCut;
            first raiseOutsideTheCut then raiseInsideTheCut;
        }
    }

    // -------------------------------------------- the L0 sheet, model side --

    package TheArgument {
        doc /* The functional view of the L0 overview sheet,
               docs/img/a3-l0-model-side.png. */

        action def TheArgument {
            doc /* The argument in eight steps, as the overview sheet reads
                   it. */
            action authorTheModel { doc /* as text, in Git */ }
            action publishAProjection { doc /* plain types, no metamodel */ }
            action attachAnAnalysis { doc /* that knows nothing about SysML */ }
            action attachEditorialStructure { doc /* that knows nothing about SysML */ }
            action composeIntoOneGraph { doc /* checked before deployment */ }
            action serveTwoApps { doc /* from the one graph */ }
            action changeANumber { doc /* in either app */ }
            action watchBothFollow { doc /* the verdict and the document */ }
            first authorTheModel then publishAProjection;
            first publishAProjection then attachAnAnalysis;
            first attachAnAnalysis then attachEditorialStructure;
            first attachEditorialStructure then composeIntoOneGraph;
            first composeIntoOneGraph then serveTwoApps;
            first serveTwoApps then changeANumber;
            first changeANumber then watchBothFollow;
        }
    }

    // --------------------------------------------------- the check session --

    package Session {
        doc /* The behaviour of the optional check session, written as the
               design and read against the session script when it exists. */

        action def RunASession {
            doc /* The session runner's twelve steps. With no credentials
                   nothing runs, unless the session is asked for without an
                   account, when it stops after the probe. With a tunnel token
                   the named tunnel carries the public hostname, and without it
                   a quick tunnel is opened and the subscriptions are probed.
                   The last step returns the test session's exit code. */
            in credentialsPresent : Boolean;
            in tunnelToken : Boolean;
            action installDependencies { doc /* npm ci in the runner, with the HTTP client the runner image lacks */ }
            action resolveTheDemoUrl { doc /* the public hostname when a tunnel token is set, else the quick tunnel's hostname read from its metrics endpoint */ }
            action waitForTheDemo { doc /* GET /viewer/ through the tunnel answers 200, up to 120 seconds */ }
            action probeSubscriptions { doc /* quick tunnel only: a POST subscription and a parallel mutation. An event frame within ten seconds means streams arrive. Without an account the runner stops here and keeps the stack up */ }
            action readTheAccountPlan { doc /* whoami and the account plan to the log */ }
            action publishTheDemoUrl { doc /* the demo URL becomes the account variable the checks read */ }
            action recordATestSession { doc /* checkly test, recorded, every check once against this instance */ }
            action recordASuiteSession { doc /* the browser suite, run against the same instance and recorded as a session beside the first */ }
            action deployTheProject { doc /* deploy, and read the heartbeat ping address back from the output, or from the account when the output lacks it */ }
            action triggerOnce { doc /* every deployed check runs once so the dashboard and status page fill */ }
            action keepTheHeartbeat { doc /* ping every five minutes while the session lives */ }
            action destroyOnStop { doc /* on stop, destroy the project and remove the account variable, then exit with the test session's code */ }
            first installDependencies then resolveTheDemoUrl;
            first resolveTheDemoUrl then waitForTheDemo;
            first waitForTheDemo then probeSubscriptions;
            first probeSubscriptions then readTheAccountPlan;
            first readTheAccountPlan then publishTheDemoUrl;
            first publishTheDemoUrl then recordATestSession;
            first recordATestSession then recordASuiteSession;
            first recordASuiteSession then deployTheProject;
            first deployTheProject then triggerOnce;
            first triggerOnce then keepTheHeartbeat;
            first keepTheHeartbeat then destroyOnStop;
        }

        action def TraceOneCheckRequest {
            doc /* The path of one traced request, from the monitoring
                   service's runner to the trace beside the check result. */
            action injectTraceContext { doc /* the runner adds traceparent and tracestate checkly=true to the request */ }
            action crossTheTunnel { doc /* the edge terminates the public hostname and hands the request to the connector beside the demo */ }
            action proxyToTheRouter { doc /* the UI server proxies /graphql to the router, headers intact */ }
            action openTheRouterSpan { doc /* the router opens a span under the request's trace id, parent-based sampling keeping it */ }
            action fetchThreeSubgraphs { doc /* one child span per subgraph fetch: model, capacity, document */ }
            action exportToTheCollector { doc /* the router exports the spans over OTLP to the collector on the compose network */ }
            action keepCheckMarkedSpans { doc /* the collector's filter keeps spans whose trace state carries checkly=true for the ingest pipeline */ }
            action showInTheViewer { doc /* every span reaches the local viewer */ }
            action ingestBesideTheCheckResult { doc /* the marked spans reach the monitoring service, which shows the trace beside the check result */ }
            first injectTraceContext then crossTheTunnel;
            first crossTheTunnel then proxyToTheRouter;
            first proxyToTheRouter then openTheRouterSpan;
            first openTheRouterSpan then fetchThreeSubgraphs;
            first fetchThreeSubgraphs then exportToTheCollector;
            first exportToTheCollector then keepCheckMarkedSpans;
            first keepCheckMarkedSpans then showInTheViewer;
            first showInTheViewer then ingestBesideTheCheckResult;
        }

        action def ExerciseAStory {
            doc /* The shape every browser check follows. */
            action openThePages { doc /* the viewer, the document, or both in one browser context */ }
            action readTheShippedState { doc /* the caption capacity 1200, bottleneck parse and the shipped verdicts */ }
            action actOnTheStory { doc /* the story's edit, query or reset */ }
            action expectWithinTwoSeconds { doc /* the other page reflects the change within two seconds, no reload */ }
            action resetBoth { doc /* resetModel and resetDocument */ }
            action expectTheShippedCaption { doc /* the caption reads capacity 1200, bottleneck parse again */ }
            first openThePages then readTheShippedState;
            first readTheShippedState then actOnTheStory;
            first actOnTheStory then expectWithinTwoSeconds;
            first expectWithinTwoSeconds then resetBoth;
            first resetBoth then expectTheShippedCaption;
        }
    }
}
