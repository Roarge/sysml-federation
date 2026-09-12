package Federation_UseCases {
    doc /* The use case register. One use case per storyboard story, US-01 to
           US-12, and one for the check session, US-19. Each names the whole
           demo as its subject, the story's persona as its actor, the story it
           realises as its objective, and the storyboard's given, when and then
           bullets as owned actions in reading order. The successions carry the
           order and nothing else: no condition is modelled here, and the
           arithmetic behind every number is the capacity service's. */
    private import Federation_StakeholderStories::*;
    private import Federation_Stakeholders::*;
    private import Federation_Context::Demo;

    // ------------------------------------------- the storyboard, UC-01 to UC-12 --

    use case def UC_01_LaunchWithOneCommand {
        doc /* The visitor runs the one command from the README and both apps
               render, then runs the container with no route to the internet
               and watches every request stay on the one published port. */
        subject demo : Demo;
        actor visitor : Visitor;
        objective realises :> US_01_LaunchWithOneCommand { subject :>> system = demo; }
        action runTheOneCommand { doc /* the one docker run line from the README, the image already pulled */ }
        action seeBothAppsRender { doc /* within ten seconds the model viewer and the requirements document render */ }
        action withholdTheRoute { doc /* the container running with no route to the internet */ }
        action seeEveryRequestStayLocal { doc /* each app renders fully and every request it makes goes to the one published port */ }
        first runTheOneCommand then seeBothAppsRender;
        first seeBothAppsRender then withholdTheRoute;
        first withholdTheRoute then seeEveryRequestStayLocal;
    }

    use case def UC_02_ReadTheModel {
        doc /* The visitor opens the viewer and reads the model three ways: as
               text, as a sketch of the wiring, and as a requirement carrying
               its verdict. */
        subject demo : Demo;
        actor visitor : Visitor;
        objective realises :> US_02_ReadTheModel { subject :>> system = demo; }
        action readTheModelText { doc /* the text pane shows the servers, their throughput values, the connections, the requirements and their short names */ }
        action readTheWiring { doc /* the sketch shows the five servers left to right as wired, each with its throughput, the rolled-up capacity of 1200, and parse marked as the bottleneck */ }
        action readTheVerdict { doc /* PIPE-R1 shows its limit of 1500, a verdict of FAIL and a reason naming parse at 1200 */ }
        first readTheModelText then readTheWiring;
        first readTheWiring then readTheVerdict;
    }

    use case def UC_03_RaiseAServerThatIsNotTheBottleneck {
        doc /* The visitor raises a server that is not the bottleneck and
               nothing moves, then sees a bad value refused and zero stop the
               chain. */
        subject demo : Demo;
        actor visitor : Visitor;
        objective realises :> US_03_RaiseAServerThatIsNotTheBottleneck { subject :>> system = demo; }
        action raiseIngest { doc /* ingest from 2000 to 3000: capacity still 1200, PIPE-R1 still FAIL, bottleneck still parse, and PIPE-R1.1 passing with its new value */ }
        action offerABadValue { doc /* a throughput set to a value that is not a number or is negative is rejected and the previous value stands */ }
        action setParseToZero { doc /* parse set to 0: the capacity becomes 0, PIPE-R1 stays FAIL and the reason names parse at 0 */ }
        first raiseIngest then offerABadValue;
        first offerABadValue then setParseToZero;
    }

    use case def UC_04_RaiseTheBottleneck {
        doc /* The visitor raises the bottleneck and watches the cut move, then raises one of the pair and watches the requirement pass. */
        subject demo : Demo;
        actor visitor : Visitor;
        objective realises :> US_04_RaiseTheBottleneck { subject :>> system = demo; }
        action raiseParse { doc /* parse from 1200 to 1700 in the viewer's panel */ }
        action watchTheCut { doc /* capacity 1400, the pair indexA and indexB marked, PIPE-R1 still FAIL */ }
        action raiseIndexA { doc /* indexA from 700 to 900 */ }
        action watchItPass { doc /* capacity 1600, PIPE-R1 PASS, the pair still marked */ }
        first raiseParse then watchTheCut;
        first watchTheCut then raiseIndexA;
        first raiseIndexA then watchItPass;
    }

    use case def UC_05_TightenTheLimit {
        doc /* The model owner lowers the limit and the verdict passes, raises
               it and the verdict fails, then reads the edit back in the model
               text. */
        subject demo : Demo;
        actor modelOwner : ModelOwner;
        objective realises :> US_05_TightenTheLimit { subject :>> system = demo; }
        action lowerTheLimit { doc /* the limit of PIPE-R1 from 1500 to 1000: the verdict becomes PASS */ }
        action raiseTheLimit { doc /* the limit of PIPE-R1 from 1500 to 2500: the verdict becomes FAIL, the requirement block turns red, and the reason names parse as the bottleneck at 1200 */ }
        action readTheEditInTheText { doc /* the edited number appears in the model text exactly where the original literal was */ }
        first lowerTheLimit then raiseTheLimit;
        first raiseTheLimit then readTheEditInTheText;
    }

    use case def UC_06_ReadTheDocument {
        doc /* The document owner opens the requirements document, reads what
               one row carries, and finds the case nobody runs
               inconclusive. */
        subject demo : Demo;
        actor documentOwner : DocumentOwner;
        objective realises :> US_06_ReadTheDocument { subject :>> system = demo; }
        action openTheDocument { doc /* PIPE-R1 as section 1 with its derived requirements as 1.1 to 1.5 and PIPE-R2 as section 2, numbers that are the document's own and not the model's short names */ }
        action readOneRow { doc /* short name, text, limit, what it is derived from or what is derived from it, the part that satisfies it, the case that verifies it, the current value of its subject, and its verdict with a reason */ }
        action readTheUnrunCase { doc /* PIPE-R2 is INCONCLUSIVE with the reason "PIPE-VC1 is declared and no service runs it" */ }
        first openTheDocument then readOneRow;
        first readOneRow then readTheUnrunCase;
    }

    use case def UC_07_ReorderAndNest {
        doc /* The document owner moves one requirement above another, nests a
               second under a first, and reads the model back unchanged. */
        subject demo : Demo;
        actor documentOwner : DocumentOwner;
        objective realises :> US_07_ReorderAndNest { subject :>> system = demo; }
        action moveOneUp { doc /* PIPE-R1.5 above PIPE-R1.1: it becomes 1.1, the others shift to 1.2 to 1.5 in their former order, and nothing else changes number */ }
        action nestUnderAnother { doc /* PIPE-R2 under PIPE-R1: it is numbered 1.6, and its relationships in the model are unchanged and still shown */ }
        action readTheModelBack { doc /* the model text and the sketch are unchanged */ }
        first moveOneUp then nestUnderAnother;
        first nestUnderAnother then readTheModelBack;
    }

    use case def UC_08_ShapeTheDocument {
        doc /* The document owner adds a heading and a paragraph of prose,
               hides a requirement, and restores it. */
        subject demo : Demo;
        actor documentOwner : DocumentOwner;
        objective realises :> US_08_ShapeTheDocument { subject :>> system = demo; }
        action insertAHeading { doc /* a heading "Performance" as the parent of PIPE-R1: the heading takes number 1, PIPE-R1 becomes 1.1 and its derived requirements 1.1.1 to 1.1.5 */ }
        action addProse { doc /* a prose paragraph under the heading appears in place and carries no number */ }
        action excludeARequirement { doc /* PIPE-R1.4 excluded: it disappears from the document, PIPE-R1.5 becomes 1.4, and the model still lists PIPE-R1.4 */ }
        action restoreIt { doc /* PIPE-R1.4 restored: it returns as the last child of PIPE-R1 and is numbered 1.5 */ }
        first insertAHeading then addProse;
        first addProse then excludeARequirement;
        first excludeARequirement then restoreIt;
    }

    use case def UC_09_ChangeAValueFromTheDocument {
        doc /* The document owner corrects a number in the document and the
               viewer follows, then changes the limit in its row and reads it
               in the viewer's requirement text. */
        subject demo : Demo;
        actor documentOwner : DocumentOwner;
        objective realises :> US_09_ChangeAValueFromTheDocument { subject :>> system = demo; }
        action editTheValueInItsRow { doc /* the throughput of parse from 1200 to 1700 in the row of PIPE-R1.2, the viewer open in another tab */ }
        action watchTheViewerFollow { doc /* PIPE-R1.2 PASS, PIPE-R1 still FAIL with a reason now naming indexA and indexB at 1400, and within two seconds the viewer's text shows 1700 and its sketch capacity 1400 with the pair marked, without a reload */ }
        action editTheLimitInItsRow { doc /* the limit of PIPE-R1 changed in its row */ }
        action seeTheNewLimitInTheViewer { doc /* the viewer shows the new limit in the requirement text */ }
        first editTheValueInItsRow then watchTheViewerFollow;
        first watchTheViewerFollow then editTheLimitInItsRow;
        first editTheLimitInItsRow then seeTheNewLimitInTheViewer;
    }

    use case def UC_10_ChangeFromTheViewerWatchTheDocument {
        doc /* The visitor edits a value in the viewer with the document open
               beside it, and the document updates by itself. */
        subject demo : Demo;
        actor visitor : Visitor;
        objective realises :> US_10_ChangeFromTheViewerWatchTheDocument { subject :>> system = demo; }
        action raiseIndexAInTheViewer { doc /* both apps open and parse at 1700, indexA from 700 to 900 in the viewer */ }
        action watchTheDocumentFollow { doc /* within two seconds and without a reload the document shows PIPE-R1 PASS with a reason naming the index pair at 1600 against 1500, PIPE-R1.3 PASS, and PIPE-R1.4 still FAIL at 700 against its limit of 750 */ }
        first raiseIndexAInTheViewer then watchTheDocumentFollow;
    }

    use case def UC_11_QueryTheGraph {
        doc /* The visitor asks one query for fields from all three services
               and reads them in one response. */
        subject demo : Demo;
        actor visitor : Visitor;
        objective realises :> US_11_QueryTheGraph { subject :>> system = demo; }
        action askForAllThree { doc /* one query in the playground for PIPE-R1's text, verdict and document number */ }
        action readOneResponse { doc /* one response carries all three, and the served schema contains types contributed by all three subgraphs */ }
        first askForAllThree then readOneResponse;
    }

    use case def UC_12_Reset {
        doc /* The visitor returns the demo to its shipped state from either
               app. */
        subject demo : Demo;
        actor visitor : Visitor;
        objective realises :> US_12_Reset { subject :>> system = demo; }
        action useTheResetControl { doc /* values and the document structure edited, the reset control used in either app */ }
        action seeTheShippedState { doc /* both apps show the shipped values and the shipped document structure within two seconds */ }
        first useTheResetControl then seeTheShippedState;
    }

    // ------------------------------------------------ the check session, UC-19 --

    use case def UC_19_RunACheckSession {
        doc /* The maintainer starts the demo with the monitoring credentials
               and the tunnel token, and watches the suite run against the
               running instance, the project deploy for the duration, a trace
               arrive beside a check result, and everything go on stop. */
        subject demo : Demo;
        actor maintainer : Maintainer;
        objective realises :> US_19_TheStoriesExercisedAgainstARunningInstance { subject :>> system = demo; }
        action startWithCredentials { doc /* start the stack with the monitoring credentials and the tunnel token */ }
        action watchTheSuiteRun { doc /* the suite runs against the instance and records a session */ }
        action watchTheProjectDeploy { doc /* the project deploys for the duration */ }
        action watchATraceArrive { doc /* a trace arrives beside a check result */ }
        action stopAndWatchItGo { doc /* stop the stack and the project is destroyed */ }
        first startWithCredentials then watchTheSuiteRun;
        first watchTheSuiteRun then watchTheProjectDeploy;
        first watchTheProjectDeploy then watchATraceArrive;
        first watchATraceArrive then stopAndWatchItGo;
    }
}
