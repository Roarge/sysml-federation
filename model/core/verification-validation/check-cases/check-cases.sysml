package Federation_CheckCases {
    doc /* The check register. One case per check and monitor of the check
           project, the check itself as a model element: its attributes equal
           the manifest's entry for it, and its actions are the check's own
           steps, a request and what is asserted of the answer. Each case reads
           the acceptance of the story the check exercises, and its evidence is
           the file that declares or drives it. */
    private import Federation_StakeholderStories::*;
    private import Federation_SystemStories::*;
    private import Federation_Context::*;
    private import Federation_LogicalArchitecture::*;
    private import Federation_Library::*;
    private import ScalarValues::*;

    // ------------------------------------------------ the router, hourly --

    verification def CHK_RouterVersion {
        doc /* The model version through the router: the smallest query the composed graph answers. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_43_OneQueryThreeServices.acceptance; }
        attribute logicalId : String = "router-version";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action postTheVersionQuery { doc /* POST the query for the model's version. */ }
        action expectAVersionAboveZero { doc /* Status 200 and a version greater than 0. */ }
    }

    verification def CHK_RouterJoin {
        doc /* The join query through the router: one requirement with fields from all three subgraphs. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_43_OneQueryThreeServices.acceptance; }
        attribute logicalId : String = "router-join";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action postTheJoinQuery { doc /* POST the requirement query for PIPE-R1 asking for text, verdict, verdictReason and documentNumber. */ }
        action expectTheThreeSubgraphsAnswer { doc /* Status 200, the text equals the shipped sentence, the verdict is not empty and documentNumber is 1. */ }
    }

    verification def CHK_RouterCapacityBottleneck {
        doc /* The capacity and the bottleneck of the pipeline through the router: the minimum cut, computed on the way out. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_29_BottleneckAsTheMinimumCut.acceptance; }
        attribute logicalId : String = "router-capacity-bottleneck";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action postTheCapacityQuery { doc /* POST the part query for PIPE-P1 asking for capacity and the bottleneck's ids. */ }
        action expectACapacityAndABottleneck { doc /* Status 200, a capacity that is not null and a bottleneck that is not empty. */ }
    }

    verification def CHK_RouterIntrospection {
        doc /* Introspection through the router: one schema that carries a type from each subgraph. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_42_CompositionThatCannotDrift.acceptance; }
        attribute logicalId : String = "router-introspection";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action postTheIntrospectionQuery { doc /* POST the schema query asking for the name of every type. */ }
        action expectTypesFromEachSubgraph { doc /* Status 200 and a body that contains Verdict, Document, Node and Model. */ }
    }

    verification def CHK_RouterPlayground {
        doc /* The playground on the one port: an HTML page for a browser that asks for one. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "router-playground";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/playground";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action getThePlayground { doc /* GET the playground path with Accept set to text/html. */ }
        action expectAnHtmlPage { doc /* Status 200 and a body that contains an html tag. */ }
    }

    verification def CHK_RouterRootRedirect {
        doc /* The root on the one port: a redirect to the viewer, read without following it. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "router-root-redirect";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action getTheRootWithoutFollowing { doc /* GET the root with redirects not followed. */ }
        action expectARedirectToTheViewer { doc /* Status 302 and a location header equal to /viewer/. */ }
    }

    verification def CHK_RouterHealthNotProxied {
        doc /* The router's health path is not one of the four: a 404 on the one port is the pass. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "router-health-not-proxied";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/health/ready";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action getTheHealthPath { doc /* GET the router's readiness path, with the check set to pass on an error status. */ }
        action expectNotFound { doc /* Status 404. */ }
    }

    // ------------------------------------------------ the refusals walk --

    verification def CHK_Refusals {
        doc /* Two edits the router refuses, with the model version read before and after to show that nothing moved. Four steps, one request each, and nothing to put back. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_15_AServedModelThatIsNeverSilentlyWrong.acceptance.boundValuesStayBound;
            verify SR_24_ExpressionBoundValuesAreReadOnly.acceptance.boundValueIsNotEdited;
            verify SR_25_RejectionOfInvalidValues.acceptance.invalidValuesAreRejected;
        }
        attribute logicalId : String = "refusals";
        attribute kind : CheckKind = CheckKind::MultiStepCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/refusals.multistep.spec.ts"; }
        action readTheVersion { doc /* Read the model version and keep it. */ }
        action refuseADerivedLimit { doc /* Send setLimit for PIPE-R1.3 at 900 and expect the first error to say the value is not a literal in the source, with no data. */ }
        action refuseANegativeThroughput { doc /* Send setAttribute for PIPE-S1's throughput at -5 and expect the first error to say the value must be a finite, non-negative number, with no data. */ }
        action readTheVersionAgain { doc /* Read the model version again and expect the one kept. */ }
    }

    // ------------------------------------------ the session's monitors --

    verification def CHK_MonitorViewer {
        doc /* The viewer answers, every ten minutes. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "monitor-viewer";
        attribute kind : CheckKind = CheckKind::UrlMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action getTheViewer { doc /* GET the viewer's path and expect status 200. */ }
    }

    verification def CHK_MonitorDocument {
        doc /* The document answers, every ten minutes. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "monitor-document";
        attribute kind : CheckKind = CheckKind::UrlMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/document/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action getTheDocument { doc /* GET the document's path and expect status 200. */ }
    }

    verification def CHK_MonitorRoot {
        doc /* The root answers through its redirect, every ten minutes. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "monitor-root";
        attribute kind : CheckKind = CheckKind::UrlMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action getTheRootFollowingTheRedirect { doc /* GET the root with redirects followed and expect status 200. */ }
    }

    verification def CHK_SessionHeartbeat {
        doc /* The session script is still running: a ping every ten minutes, five minutes of grace, and then a missed ping is a failed session. A heartbeat runs on no schedule and from no location, so the case carries neither. */
        subject runner : Federation_LogicalArchitecture::CheckSession::SessionRunner;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        attribute logicalId : String = "session-heartbeat";
        attribute kind : CheckKind = CheckKind::HeartbeatMonitor;
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "the ping URL";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action expectAPingEveryTenMinutes { doc /* Expect a ping on the monitor's URL within each ten-minute period plus five minutes of grace. */ }
    }

    // ------------------------------- the public hostname, when one is set --

    verification def CHK_DemoCertificate {
        doc /* The certificate the edge presents for the public hostname is valid, with an alert two weeks before it expires. Deployed only when a hostname is set. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        attribute logicalId : String = "demo-certificate";
        attribute kind : CheckKind = CheckKind::SslMonitor;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : String = "whenHostnameSet";
        attribute mutating : Boolean = false;
        attribute target : String = "DEMO_HOSTNAME";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action readTheCertificate { doc /* Complete a TLS handshake with the hostname on port 443 and expect a valid certificate more than 14 days from expiry. */ }
    }

    verification def CHK_DemoDns {
        doc /* The public hostname resolves. The tunnel's CNAME is flattened at the edge, so an A query is what answers. Deployed only when a hostname is set. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        attribute logicalId : String = "demo-dns";
        attribute kind : CheckKind = CheckKind::DnsMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : String = "whenHostnameSet";
        attribute mutating : Boolean = false;
        attribute target : String = "DEMO_HOSTNAME";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action queryTheARecord { doc /* Query the A record of the hostname and expect an answer that is not empty. */ }
    }

    verification def CHK_DemoTcp {
        doc /* Port 443 of the public hostname accepts a connection over IPv4. Deployed only when a hostname is set. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        attribute logicalId : String = "demo-tcp";
        attribute kind : CheckKind = CheckKind::TcpMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : String = "whenHostnameSet";
        attribute mutating : Boolean = false;
        attribute target : String = "DEMO_HOSTNAME:443";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action connectToPort443 { doc /* Open a TCP connection to the hostname on port 443 over IPv4 and expect it to be accepted. */ }
    }

    // ------------------------------------------ the stories, US-01 to US-12 --

    verification def CHK_Us01Launch {
        doc /* Launch with one command: both apps render from the one published port, and every request either page makes goes to that port. Nothing is edited. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_01_LaunchWithOneCommand.acceptance.bothAppsRender;
            verify US_01_LaunchWithOneCommand.acceptance.noRouteToTheInternet;
        }
        attribute logicalId : String = "us01-launch";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us01-launch.spec.ts"; }
        action openTheViewer { doc /* Open the viewer and expect it rendered: the status line and the version mark present, and the caption capacity 1200, bottleneck parse. */ }
        action openTheDocument { doc /* Open the document and expect the tree with PIPE-R1 rendered as section 1 under the versions line. */ }
        action expectEveryRequestOnTheOnePort { doc /* Expect every request the two pages made to have gone to the one published port. */ }
    }

    verification def CHK_Us02ReadTheModel {
        doc /* Read the model: the text pane shows the model file, the sketch shows the wiring with its capacity and bottleneck, and PIPE-R1 shows its limit, its verdict and the reason. Nothing is edited. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_02_ReadTheModel.acceptance.textPaneShowsTheModelFile;
            verify US_02_ReadTheModel.acceptance.sketchShowsTheWiring;
            verify US_02_ReadTheModel.acceptance.requirementShowsItsVerdict;
        }
        attribute logicalId : String = "us02-read-the-model";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us02-read-the-model.spec.ts"; }
        action expectTheTextPane { doc /* Open the viewer and expect the text pane to show the model file, with its keywords, its names and its numbers marked. */ }
        action expectTheSketch { doc /* Expect the sketch to show the five servers wired, the caption capacity 1200 and parse marked as the bottleneck. */ }
        action expectTheVerdictOfPipeR1 { doc /* Expect PIPE-R1 to show its limit of 1500, a verdict of FAIL, a reason naming parse at 1200 and the block red. */ }
    }

    verification def CHK_Us03NothingMoves {
        doc /* Raise a server that is not the bottleneck: ingest goes from 2000 to 3000 and nothing moves. The version still grows by one, which is what shows the page redrew before the caption is read. Test sessions only, since the check mutates and teaches nothing new about the shipped state as a monitor. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify US_03_RaiseAServerThatIsNotTheBottleneck.acceptance.nothingMoves; }
        attribute logicalId : String = "us03-nothing-moves";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = false;
        attribute mutating : Boolean = true;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us03-nothing-moves.spec.ts"; }
        action openTheViewer { doc /* Open the viewer on the shipped state and keep its version. */ }
        action raiseIngest { doc /* Raise ingest from 2000 to 3000 and expect the version to grow by one within two seconds. */ }
        action expectNothingMoved { doc /* Expect the capacity, the bottleneck and the verdict of PIPE-R1 unchanged. */ }
        action expectPipeR11Passes { doc /* Expect PIPE-R1.1 to pass with the new value. */ }
        action putTheDemoBack { doc /* Put the demo back as it was found: reset the model and the document through the router, open the viewer and expect the caption capacity 1200, bottleneck parse. */ }
    }

    verification def CHK_Us04BottleneckMoves {
        doc /* Raise the bottleneck: parse goes from 1200 to 1700 and the bottleneck moves to the index pair, then indexA goes from 700 to 900 and PIPE-R1 passes. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_04_RaiseTheBottleneck.acceptance.capacityRises;
            verify US_04_RaiseTheBottleneck.acceptance.requirementPasses;
        }
        attribute logicalId : String = "us04-bottleneck-moves";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 360;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = true;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us04-bottleneck-moves.spec.ts"; }
        action openTheViewer { doc /* Open the viewer on the shipped state, with PIPE-R1 red. */ }
        action raiseParse { doc /* Raise parse from 1200 to 1700 and expect capacity 1400 with the bottleneck at indexA and indexB, PIPE-R1 still FAIL and still red. */ }
        action raiseIndexA { doc /* Raise indexA from 700 to 900 and expect capacity 1600, PIPE-R1 PASS at 1600 against 1500 naming the index pair, and the block no longer red. */ }
        action putTheDemoBack { doc /* Put the demo back as it was found: reset the model and the document through the router, open the viewer and expect the caption capacity 1200, bottleneck parse. */ }
    }

    verification def CHK_Us05TightenLimit {
        doc /* Tighten the limit: the limit of PIPE-R1 goes to 1000 and the verdict passes, then to 2500 and it fails naming parse, with the edited number in the served text where the literal was. Test sessions only, since the check mutates and teaches nothing new about the shipped state as a monitor. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_05_TightenTheLimit.acceptance.loweredLimitPasses;
            verify US_05_TightenTheLimit.acceptance.raisedLimitFails;
            verify US_05_TightenTheLimit.acceptance.editLandsInTheText;
        }
        attribute logicalId : String = "us05-tighten-limit";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = false;
        attribute mutating : Boolean = true;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us05-tighten-limit.spec.ts"; }
        action openTheViewer { doc /* Open the viewer on the shipped state, with PIPE-R1 at FAIL capacity 1200 against 1500, limited by parse. */ }
        action lowerTheLimit { doc /* Lower the limit of PIPE-R1 from 1500 to 1000 and expect PASS, with the block no longer red. */ }
        action raiseTheLimit { doc /* Raise the limit of PIPE-R1 to 2500 and expect FAIL, the block red and the reason naming parse at 1200. */ }
        action expectTheNumberInTheText { doc /* Expect the edited number in the served text where the literal was, and the old literal gone. */ }
        action putTheDemoBack { doc /* Put the demo back as it was found: reset the model and the document through the router, open the viewer and expect the caption capacity 1200, bottleneck parse. */ }
    }

    verification def CHK_Us06ReadDocument {
        doc /* Read the document: the sections carry the document's own numbers, each row carries what the model knows about its requirement, and the requirement whose case nothing runs is INCONCLUSIVE. Nothing is edited. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_06_ReadTheDocument.acceptance.documentNumbersItsOwnSections;
            verify US_06_ReadTheDocument.acceptance.eachRowCarriesWhatTheModelKnows;
            verify US_06_ReadTheDocument.acceptance.unrunCaseIsInconclusive;
        }
        attribute logicalId : String = "us06-read-document";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 360;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/document/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us06-read-document.spec.ts"; }
        action expectTheNumbering { doc /* Open the document and expect the versions line and the sections numbered by the document: PIPE-R1 as 1, its derived requirements 1.1 to 1.5 and PIPE-R2 as 2. */ }
        action expectTheRowOfPipeR14 { doc /* Expect the row of PIPE-R1.4 to carry its short name, its text, its limit, what it is derived from, what satisfies it, its current value, its throughput as a control and its verdict with the reason. */ }
        action expectPipeR2Inconclusive { doc /* Expect PIPE-R2 INCONCLUSIVE, verified by PIPE-VC1, with the reason that the case is declared and no service runs it. */ }
    }

    verification def CHK_Us07ReorderAndNest {
        doc /* Reorder and nest: PIPE-R1.5 moves above PIPE-R1.1 and the numbers follow, PIPE-R2 moves under PIPE-R1 and keeps its relations, and the model is untouched. A pointer drag does not reach the vendored drag library in a headless run, so each move is the mutation the drop would send, through the router, and the document page is loaded afresh to read the result. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_07_ReorderAndNest.acceptance.reorderRenumbers;
            verify US_07_ReorderAndNest.acceptance.nestingKeepsTheRelations;
            verify US_07_ReorderAndNest.acceptance.theModelIsUntouched;
        }
        attribute logicalId : String = "us07-reorder-and-nest";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = true;
        attribute target : String = "/document/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us07-reorder-and-nest.spec.ts"; }
        action openTheViewer { doc /* Open the viewer and keep its version mark. */ }
        action movePipeR15AbovePipeR11 { doc /* Move PIPE-R1.5 above PIPE-R1.1 through the router, open the document and expect it numbered 1.1 with the others shifted to 1.2 to 1.5 in their former order, and nothing else renumbered. */ }
        action movePipeR2UnderPipeR1 { doc /* Move PIPE-R2 under PIPE-R1 through the router, reload and expect it numbered 1.6 inside the list of PIPE-R1, still satisfied by the pipeline, still verified by PIPE-VC1 and still INCONCLUSIVE. */ }
        action expectTheModelUntouched { doc /* Open the viewer and expect the model version, the text and the sketch unchanged. */ }
        action putTheDemoBack { doc /* Put the demo back as it was found: reset the model and the document through the router, open the viewer and expect the caption capacity 1200, bottleneck parse. */ }
    }

    verification def CHK_Us08ShapeTheDocument {
        doc /* Shape the document: a heading takes a number and pushes PIPE-R1 under it, a paragraph of prose appears with no number, an excluded requirement leaves the document for the tray and comes back on restore, and the model is untouched throughout. The heading and prose buttons open a prompt, which the page's dialog handler answers before the click. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_08_ShapeTheDocument.acceptance.headingTakesANumber;
            verify US_08_ShapeTheDocument.acceptance.proseCarriesNoNumber;
            verify US_08_ShapeTheDocument.acceptance.exclusionHidesWithoutRemoving;
            verify US_08_ShapeTheDocument.acceptance.restoringReturnsIt;
        }
        attribute logicalId : String = "us08-shape-the-document";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = true;
        attribute target : String = "/document/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us08-shape-the-document.spec.ts"; }
        action openTheViewer { doc /* Open the viewer and keep its version mark. */ }
        action insertAHeading { doc /* Open the document, insert a heading above PIPE-R1 and expect it numbered 1, PIPE-R1 at 1.1 and its derived requirements 1.1.1 to 1.1.5. */ }
        action addProse { doc /* Add a paragraph of prose under the heading and expect it in place, as the heading's last child, with no number. */ }
        action excludePipeR14 { doc /* Exclude PIPE-R1.4 and expect it in the tray, gone from the document, and PIPE-R1.5 renumbered to 1.1.4. The tray is drawn from the model's own requirement list, so PIPE-R1.4 listed there is the model still listing it. */ }
        action restorePipeR14 { doc /* Restore PIPE-R1.4 and expect it back as the last child of PIPE-R1, numbered 1.1.5, and the tray empty. */ }
        action expectTheModelUntouched { doc /* Open the viewer and expect the model version unchanged and PIPE-R1.4 still listed. */ }
        action putTheDemoBack { doc /* Put the demo back as it was found: reset the model and the document through the router, open the viewer and expect the caption capacity 1200, bottleneck parse. */ }
    }

    verification def CHK_Us09EditFromDocument {
        doc /* Change a value from the document: with the viewer open on a second page, the limit of PIPE-R1 and then the throughput of parse are corrected in the document's rows, and the viewer follows each within two seconds with no reload. The viewer follows on a live update, so the check is skipped where the session carries no stream. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_09_ChangeAValueFromTheDocument.acceptance.editReachesTheViewer;
            verify US_09_ChangeAValueFromTheDocument.acceptance.limitEditReachesTheViewer;
        }
        attribute logicalId : String = "us09-edit-from-document";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = true;
        attribute target : String = "/document/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us09-edit-from-document.spec.ts"; }
        action openBothApps { doc /* Open the document and, on a second page, the viewer, both on the shipped state. */ }
        action changeTheLimit { doc /* Change the limit of PIPE-R1 to 1600 in its row and expect the viewer to show the new limit within two seconds, in the verdict line, the constraint and the served text. */ }
        action changeParse { doc /* Change the throughput of parse to 1700 in the row of PIPE-R1.2 and expect PIPE-R1.2 to pass, PIPE-R1 to stay FAIL naming indexA and indexB at 1400, and the viewer to follow within two seconds: 1700 in the text, capacity 1400 in the caption and the index pair marked. */ }
        action putTheDemoBack { doc /* Put the demo back as it was found: reset the model and the document through the router, open the viewer and expect the caption capacity 1200, bottleneck parse. */ }
    }

    verification def CHK_Us10ViewerToDocument {
        doc /* Change from the viewer, watch the document: with parse already at 1700 and both apps open, indexA is raised from 700 to 900 in the viewer and the document's rows follow within two seconds with no reload. The document follows on a live update, so the check is skipped where the session carries no stream. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify US_10_ChangeFromTheViewerWatchTheDocument.acceptance.documentFollowsTheViewer; }
        attribute logicalId : String = "us10-viewer-to-document";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = true;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us10-viewer-to-document.spec.ts"; }
        action setParseAndOpenBothApps { doc /* Set parse to 1700 through the router, then open the viewer and, on a second page, the document, both showing capacity 1400 with the bottleneck at indexA and indexB. */ }
        action raiseIndexAInTheViewer { doc /* Raise indexA from 700 to 900 in the viewer and expect the document to follow within two seconds: PIPE-R1 PASS at 1600 against 1500 naming the index pair, PIPE-R1.3 PASS, and PIPE-R1.4 still FAIL at 700 against 750. */ }
        action putTheDemoBack { doc /* Put the demo back as it was found: reset the model and the document through the router, open the viewer and expect the caption capacity 1200, bottleneck parse. */ }
    }

    verification def CHK_Us11QueryTheGraph {
        doc /* Query the graph: one query for PIPE-R1's text, verdict and document number is answered in one response from all three services, the served schema carries a type from each, and the playground is there for a visitor to send the same query by hand. Nothing is edited. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify US_11_QueryTheGraph.acceptance.oneResponseCarriesAllThree; }
        attribute logicalId : String = "us11-query-the-graph";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/playground";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us11-query-the-graph.spec.ts"; }
        action sendTheJoinQuery { doc /* Send the join query for PIPE-R1 and expect one answer carrying the shipped text, the verdict FAIL with its reason and the document number 1. */ }
        action readTheSchema { doc /* Read the served schema and expect a type from each of the three subgraphs: Model, Verdict, Document and Node. */ }
        action openThePlayground { doc /* Open the playground and expect it to load. */ }
    }

    verification def CHK_Us12Reset {
        doc /* Reset: a value is edited in the viewer and the document's structure in the document, then the viewer's Reset control brings both apps back to the shipped state within two seconds. Each app follows on a live update, so the check is skipped where the session carries no stream. The reset is the story, and the demo is put back through the router as well, so that a failure part way through leaves it as it was found. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify US_12_Reset.acceptance.bothAppsReturnToTheShippedState; }
        attribute logicalId : String = "us12-reset";
        attribute kind : CheckKind = CheckKind::BrowserCheck;
        attribute frequencyMinutes : Integer = 1440;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = true;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/us12-reset.spec.ts"; }
        action openBothApps { doc /* Open the viewer and, on a second page, the document, both on the shipped state. */ }
        action editInEachApp { doc /* Raise parse to 1700 in the viewer and exclude PIPE-R1.4 in the document, and expect each edit shown. */ }
        action pressReset { doc /* Press Reset in the viewer and expect both apps back at the shipped state within two seconds: the caption, the verdict of PIPE-R1, the tray empty and PIPE-R1.4 numbered 1.4 again. */ }
        action putTheDemoBack { doc /* Put the demo back as it was found: reset the model and the document through the router, open the viewer and expect the caption capacity 1200, bottleneck parse. */ }
    }
}
