package Federation_LogicalArchitecture {
    doc /* What the demo is made of, and what the project around it is made of.

           Three composites hold the whole. The first, demo, carries the V2
           composition board and the V4 deployment board in one place, because
           the parts they draw are the same parts seen twice. The second,
           project, carries the repository, the demo's own model and the check
           suite, which act on the demo from outside it. The third, session,
           carries the optional check session that runs beside the demo.

           The allocations live here too. Every system story is satisfied by the
           element or elements it was allocated to, and a story allocated both to
           something inside the demo and to the repository carries one satisfy in
           each composite. Each design constraint is satisfied by the elements it
           binds. A decision record is applied to the definition or the usage it
           shaped, so that the reason for a part can be read beside the part. */
    private import Federation_InterfaceTypes::*;
    private import Federation_Context::*;
    private import Federation_Domain::*;
    private import Federation_SystemStories::*;
    private import Federation_DesignConstraints::*;
    private import Federation_Library::DecisionRecord;
    private import Federation_Library::CheckKind;
    private import ScalarValues::*;

    // ------------------------------------------------- the three subgraphs --

    part def Adapter :> DemoElement {
        doc /* The SysML adapter, the subject of the V5 board and the allocation
               element ADAPTER of the requirements scheme. Four Go packages under
               adapter/, arranged as a pipeline: syntax, model, projection and
               serve. The model file is read once at startup and never again.

               One path runs backwards through the pipeline. setAttribute and
               setLimit patch the literal at its span in the source text, the
               text is parsed again, and the version counter moves on by one
               (SR-22), so the text the viewer shows and the projection the
               router queries are rebuilt together and cannot disagree.

               The refusal path is as much of the definition as the accepting
               one. A construct outside the subset is refused (SR-18), a
               constraint of a shape the reader does not recognise is refused
               (SR-19), and port directions that disagree with the order of a
               connection's ends are refused (SR-20). Every refusal names the
               file, the line and the column of the first offending token. */
        @DecisionRecord { id = "AD-0003"; title = "The adapter reads files rather than fronting a repository"; }
        @DecisionRecord { id = "AD-0004"; title = "Editing through the projection as scaffolding"; }
        @DecisionRecord { id = "AD-0005"; title = "A curated generic projection rather than generated metamodel types"; }
        @DecisionRecord { id = "AD-0014"; title = "Subscriptions as version events with client refetch"; }
        @DecisionRecord { id = "AD-0015"; title = "A hand-written strict subset parser"; }
        @DecisionRecord { id = "AD-0018"; title = "Short names as entity keys with the qualified name as fallback"; }
        port listen : GraphQLServerPort;
        part syntax {
            doc /* adapter/syntax. The lexer and the parser. Text becomes tokens
                   and tokens become a tree in which every node carries its byte
                   range in the source. It accepts a strict subset of the textual
                   notation and has no opinion about meaning. */
        }
        part model {
            doc /* adapter/model. The resolver: qualified names, short names and
                   feature chains through a requirement's subject and its part
                   usages. It evaluates the bound expressions the subset allows,
                   reads the shape of a constraint into a quantity, a comparison
                   and a limit, checks port directions against connection end
                   order, and keeps the source text with an index from every
                   numeric literal to its span. */
        }
        part projection {
            doc /* adapter/projection. Resolved elements become the seven types
                   the schema publishes: Model, Part, Attribute, Port,
                   Connection, Requirement and VerificationCase. It holds nothing
                   the schema does not show. */
        }
        part serve {
            doc /* adapter/serve. The gqlgen server: the entity resolvers,
                   subscriptions over WebSocket, the store that hands each
                   operation one snapshot of the current model, and the version
                   counter that store moves on at every accepted mutation and at
                   reset. */
        }
        attribute listenAddress : String = "127.0.0.1:3011";
        attribute subgraphName : String = "model";
    }

    part def CapacityService :> DemoElement {
        doc /* The capacity subgraph, the allocation element CAPACITY. It
               computes capacity and bottleneck on a part, and verdict and
               verdictReason on a requirement, on every read, from the fields its
               requires clause names. It holds no copy of the model, so there is
               nothing in it that can fall behind the source. */
        @DecisionRecord { id = "AD-0006"; title = "An idealised capacity model"; }
        @DecisionRecord { id = "AD-0007"; title = "Rollup as maximum flow with the source-side minimum cut"; }
        @DecisionRecord { id = "AD-0008"; title = "Quantity, comparison and limit read from the constraint"; }
        @DecisionRecord { id = "AD-0009"; title = "Connection direction from the order of the ends"; }
        @DecisionRecord { id = "AD-0024"; title = "Verdict reasons built from the capacity service templates"; }
        port listen : GraphQLServerPort;
        attribute listenAddress : String = "127.0.0.1:3012";
        attribute subgraphName : String = "capacity";
        attribute quantity : String = "capacity";
        attribute attributeRead : String = "throughput";
    }

    part def DocumentService :> DemoElement {
        doc /* The document subgraph, the allocation element DOCUMENT. It owns an
               ordered tree and the numbering that follows from it, ships one
               structure, and never reads the model. */
        @DecisionRecord { id = "AD-0025"; title = "The document owns its structure and nothing else"; }
        port listen : GraphQLServerPort;
        attribute listenAddress : String = "127.0.0.1:3013";
        attribute subgraphName : String = "document";
    }

    // --------------------------------------- the router and the UI server --

    part def Router :> DemoElement {
        doc /* The federation router, the allocation element ROUTER. A vendor
               binary copied into the image and run as a child process with
               static configuration. Its environment is built in full by the
               supervisor, all nine variables, so that nothing is inherited from
               the host.

               A configuration file is handed to it only when the operator names
               one, and in that case the file governs its telemetry. The three
               tracing attributes below describe that file and apply only
               then. */
        @DecisionRecord { id = "AD-0010"; title = "The router as a child process from the copied binary"; }
        @DecisionRecord { id = "AD-0013"; title = "Telemetry disabled by environment baked into the image"; }
        @DecisionRecord { id = "AD-0014"; title = "Subscriptions as version events with client refetch"; }
        port listen : GraphQLServerPort;
        port subgraphs : GraphQLClientPort;
        port traces : OtlpClientPort;
        attribute listenAddress : String = "127.0.0.1:3002";
        attribute routerVersion : String = "0.343.1";
        attribute tracingConfigFile : String[0..1];
        attribute parentBasedSampler : Boolean = true;
        attribute samplingRate : Real = 1.0;
    }

    part def UiServer {
        doc /* The user interface server. It serves both apps from an embedded
               filesystem, proxies /graphql and /playground to the router, and
               redirects / to /viewer/. It is the only listener the outside can
               reach. */
        port web : HttpServerPort;
        port upstream : GraphQLClientPort;
        attribute listenAddress : String = "0.0.0.0:8080";
    }

    part def Supervisor {
        doc /* The process that is PID 1. It starts the three subgraphs as
               goroutines, waits until each answers its health endpoint, starts
               the router as a child process, waits for the router to report
               itself ready, and only then opens port 8080. */
        attribute subcommands : String = "serve, adapter, capacity, document, ui, healthcheck";
    }

    // ------------------------------------------------------- the two apps --

    part def ModelViewer :> DemoElement {
        doc /* The viewer, the allocation element UI-MODEL. It shows the model's
               text beside a sketch of its wiring, an edit panel of six controls,
               and one block per requirement with the verdict on it. Served at
               /viewer. */
        @DecisionRecord { id = "AD-0017"; title = "Vanilla web apps with one vendored file"; }
        @DecisionRecord { id = "AD-0026"; title = "The viewer shows the model's text beside a sketch of its wiring"; }
    }

    part def RequirementsDocument :> DemoElement {
        doc /* The document app, the allocation element UI-DOC. It shows the same
               requirements as a numbered document, with the editorial operations
               that shape it. Served at /document. */
        @DecisionRecord { id = "AD-0017"; title = "Vanilla web apps with one vendored file"; }
    }

    // ---------------------------------------------------- what is read in --

    part def ExampleModel :> DemoElement {
        doc /* The file examples/pipeline/model.sysml, the allocation element
               MODEL. It is content rather than code: the adapter reads it once
               at startup, and it is validated locally against the two reference
               tools rather than in continuous integration. The pipeline it
               describes is not a running system anywhere. */
        @DecisionRecord { id = "AD-0019"; title = "SysML 2.0 formal as the target, validated with the pilot and OpenSysML"; }
    }

    // --------------------------------------------- the project around it --

    part def Project {
        doc /* The repository, the model of the demo and the check suite as one
               thing. */
    }

    // ------------------------------------------- the check suite's insides --

    part def Check {
        doc /* One check or monitor of the monitoring project. The logical
               identifier is the name both the register and the manifest know it
               by. A check with no frequency runs on no schedule of its own. */
        attribute logicalId : String;
        attribute kind : CheckKind;
        attribute frequencyMinutes : Integer[0..1];
        attribute deployed : Boolean;
        attribute mutating : Boolean;
        attribute target : String;
    }

    part def CheckGroup {
        doc /* A group of checks sharing a schedule, a concurrency limit and a
               set of run locations. A group is marked mutating when a check in
               it changes the demo's state. */
        attribute concurrency : Integer;
        attribute mutating : Boolean;
        attribute locations : String[1..3];
    }

    part def AlertChannel {
        doc /* Where an alert is delivered: a mailbox or a webhook endpoint. */
        attribute channelKind : String;
    }

    part def Dashboard {
        doc /* The public board that shows how the checks have been running. */
    }

    part def StatusPage {
        doc /* The public status page, with one component per part of the demo a
               visitor would ask about. */
        part components : StatusComponent[4];
    }

    part def StatusComponent {
        doc /* One line of the status page, up or down. */
    }

    part def GatedFeature {
        doc /* A construct declared only when an environment flag is set, because
               the feature it needs is paid for rather than included. */
        attribute enabledBy : String;
    }

    part def CheckSuiteProject :> CheckSuite {
        doc /* The check suite as the monitoring project it declares: five
               groups, the checks themselves, the alert channels, the dashboard,
               the status page, the heartbeat and the three constructs that
               appear only on a paid plan.

               The inventory of checks is not here. It lives in the check case
               register and in the manifest the check project reads, and those
               two are compared with each other rather than with this. */
        part groups : CheckGroup[5];
        part checks : Check[*];
        part alertChannels : AlertChannel[0..2];
        part dashboard : Dashboard;
        part statusPage : StatusPage;
        part heartbeat : Check;
        part automationRules : GatedFeature;
        part maintenanceWindow : GatedFeature;
        part privateLocation : GatedFeature;
    }

    // --------------------------------------------------- the V2 and V4 box --

    part demo : Demo {
        doc /* One container, one process tree, one port.

               The composition side. The three subgraphs resolve disjoint fields
               and the router merges them, so one requirement carries its text
               and limit from the adapter, its verdict and reason from the
               capacity service, and its number and inclusion from the document
               service. One query for a requirement is answered with three
               fetches: the adapter for the requirement and every field the
               capacity service requires, the capacity service's entity resolver
               with those fields as the representation, and the document
               service's entity resolver with the key alone.

               The deployment side. PID 1 is the supervisor. The three subgraphs
               run as goroutines on loopback ports, the router runs as a child
               process on a loopback port, and the user interface server is the
               only thing bound to the published port. The router's health path
               is not proxied, so /viewer, /document, /shared, /graphql and
               /playground are all that port answers, with / redirecting to the
               viewer. The image is built for two platforms by
               cross-compilation, and the attributes below carry the budget, the
               environment the supervisor sets and the layers the final stage is
               made of. */
        @DecisionRecord { id = "AD-0001"; title = "Federation over a single GraphQL service"; }
        @DecisionRecord { id = "AD-0002"; title = "Cosmo as the federation platform"; }
        @DecisionRecord { id = "AD-0011"; title = "One binary, one process tree, one port, one UI server"; }
        attribute compressedSizeCeilingMegabytes : Integer = 80;
        attribute environment : String = "DO_NOT_TRACK=1, COSMO_TELEMETRY_DISABLED=true, TRACING_ENABLED=false, METRICS_OTLP_ENABLED=false, PROMETHEUS_ENABLED=false";
        attribute layers : String = "distroless static-debian13 nonroot, /router and /router.LICENSE, /sysml-federation, /app/config.json, /app/model.sysml";
        part supervisor : Supervisor;
        part uiServer : UiServer;
        part viewer : ModelViewer;
        part documentApp : RequirementsDocument;
        part router : Router;
        part adapter : Adapter;
        part capacity : CapacityService;
        part document : DocumentService;
        part exampleModel : ExampleModel;
        interface toRouter : UiProxy connect uiServer.upstream to router.listen;
        interface modelFetch : RouterToSubgraph connect router.subgraphs to adapter.listen;
        interface capacityFetch : RouterToSubgraph connect router.subgraphs to capacity.listen;
        interface documentFetch : RouterToSubgraph connect router.subgraphs to document.listen;
        flow versionEvents of VersionEvent from adapter.listen.events to router.subgraphs.events;

        // IMAGE, the demo as a whole.
        satisfy SR_01_OneCommand by demo;
        satisfy SR_02_ReadyWithinTenSeconds by demo;
        satisfy SR_03_NoOutboundNetwork by demo;
        satisfy SR_04_FourPathsOnOnePort by demo;
        satisfy SR_05_TwoPlatforms by demo;
        satisfy SR_06_SizeBudget by demo;
        satisfy SR_07_LicenceCarriedInTheImage by demo;

        // ROUTER.
        satisfy SR_03_NoOutboundNetwork by router;
        satisfy SR_04_FourPathsOnOnePort by router;
        satisfy SR_39_LiveReflectionInTheOtherApp by router;
        satisfy SR_42_CompositionThatCannotDrift by router;
        satisfy SR_43_OneQueryThreeServices by router;

        // ADAPTER.
        satisfy SR_09_StateLivesInMemory by adapter;
        satisfy SR_16_GenericProjection by adapter;
        satisfy SR_17_NoExampleIdentifiersInTheAdapter by adapter;
        satisfy SR_18_RefusalOfUnsupportedSyntax by adapter;
        satisfy SR_19_QuantityComparisonAndLimitFromTheConstraint by adapter;
        satisfy SR_20_ConnectionDirection by adapter;
        satisfy SR_21_ShortNamesAsIdentifiers by adapter;
        satisfy SR_22_EditsPatchTheSource by adapter;
        satisfy SR_23_ExpressionBoundValuesAreEvaluated by adapter;
        satisfy SR_24_ExpressionBoundValuesAreReadOnly by adapter;
        satisfy SR_25_RejectionOfInvalidValues by adapter;
        satisfy SR_26_ModelVersionEvents by adapter;
        satisfy SR_31_TwoConfiguredNames by adapter;
        satisfy SR_36_DocumentOperationsLeaveTheModelUnchanged by adapter;
        satisfy SR_39_LiveReflectionInTheOtherApp by adapter;
        satisfy SR_41_ServicesIsolated by adapter;
        satisfy SR_43_OneQueryThreeServices by adapter;
        satisfy SR_44_ResetFromEitherApp by adapter;

        // CAPACITY.
        satisfy SR_28_RollupByMaximumFlow by capacity;
        satisfy SR_29_BottleneckAsTheMinimumCut by capacity;
        satisfy SR_30_VerdictsAndReasons by capacity;
        satisfy SR_31_TwoConfiguredNames by capacity;
        satisfy SR_32_NoCopyOfTheModel by capacity;
        satisfy SR_41_ServicesIsolated by capacity;
        satisfy SR_43_OneQueryThreeServices by capacity;

        // DOCUMENT.
        satisfy SR_09_StateLivesInMemory by document;
        satisfy SR_27_DocumentVersionEvents by document;
        satisfy SR_33_NumberingFromTheTreeAlone by document;
        satisfy SR_35_EditorialOperationsInTheService by document;
        satisfy SR_36_DocumentOperationsLeaveTheModelUnchanged by document;
        satisfy SR_39_LiveReflectionInTheOtherApp by document;
        satisfy SR_41_ServicesIsolated by document;
        satisfy SR_43_OneQueryThreeServices by document;
        satisfy SR_44_ResetFromEitherApp by document;

        // UI-MODEL.
        satisfy SR_10_NoResourceFromAnotherOrigin by viewer;
        satisfy SR_11_ModelTextWithTheLanguageVisible by viewer;
        satisfy SR_12_SketchDrawnFromTheWiring by viewer;
        satisfy SR_13_TheViewersEditableSet by viewer;
        satisfy SR_14_VerdictAndReasonInTheViewer by viewer;
        satisfy SR_15_FailingRequirementInRed by viewer;
        satisfy SR_25_RejectionOfInvalidValues by viewer;
        satisfy SR_39_LiveReflectionInTheOtherApp by viewer;
        satisfy SR_40_EveryRequestThroughTheRouter by viewer;
        satisfy SR_44_ResetFromEitherApp by viewer;

        // UI-DOC.
        satisfy SR_10_NoResourceFromAnotherOrigin by documentApp;
        satisfy SR_25_RejectionOfInvalidValues by documentApp;
        satisfy SR_34_EditorialOperationsInTheApp by documentApp;
        satisfy SR_37_WhatARequirementRowShows by documentApp;
        satisfy SR_38_TheDocumentsEditableSet by documentApp;
        satisfy SR_39_LiveReflectionInTheOtherApp by documentApp;
        satisfy SR_40_EveryRequestThroughTheRouter by documentApp;
        satisfy SR_44_ResetFromEitherApp by documentApp;

        // MODEL.
        satisfy SR_45_AModelTheReferenceToolsAccept by exampleModel;

        // The design constraints, where the constraint register binds them.
        satisfy SC_01_LanguageAndLicence by adapter;
        satisfy SC_01_LanguageAndLicence by capacity;
        satisfy SC_01_LanguageAndLicence by document;
        satisfy SC_01_LanguageAndLicence by demo;
        satisfy SC_02_NoEmptyInterfaceInHandWrittenCode by adapter;
        satisfy SC_02_NoEmptyInterfaceInHandWrittenCode by capacity;
        satisfy SC_02_NoEmptyInterfaceInHandWrittenCode by document;
        satisfy SC_02_NoEmptyInterfaceInHandWrittenCode by demo;
        satisfy SC_05_NoBuildStepForTheWebApps by viewer;
        satisfy SC_05_NoBuildStepForTheWebApps by documentApp;
        satisfy SC_06_ReadableInAnAfternoon by adapter;
        satisfy SC_06_ReadableInAnAfternoon by capacity;
        satisfy SC_06_ReadableInAnAfternoon by document;
        satisfy SC_06_ReadableInAnAfternoon by demo;
        satisfy SC_06_ReadableInAnAfternoon by viewer;
        satisfy SC_06_ReadableInAnAfternoon by documentApp;
    }

    // ---------------------------------------------- the project composite --

    part project : Project {
        doc /* The repository, the model and the check suite. */
        part repository : Repository {
            doc /* The repository as a thing that is built, tested and
                   published. The publish chain runs from end to end without a
                   control plane in it. The maintainer composes the router
                   configuration on their own machine and commits the output,
                   then pushes a version tag. The tag starts the workflow, the
                   workflow builds the image for both platforms by
                   cross-compilation, pushes it to the registry under the semver
                   version alone, and reads the manifest back, failing the job if
                   either platform is missing or over budget. The moving tag
                   follows only once both platforms have passed and only for a
                   version with no pre-release suffix. */
            @DecisionRecord { id = "AD-0012"; title = "Composition as a maintainer step with committed output and a drift test"; }
            @DecisionRecord { id = "AD-0016"; title = "Generated code exempt from the empty-interface rule"; }
            @DecisionRecord { id = "AD-0020"; title = "The image published by a tag-triggered workflow to GHCR"; }
            @DecisionRecord { id = "AD-0027"; title = "The adapter's expected scale raised to 2750 lines"; }
            @DecisionRecord { id = "AD-0028"; title = "The line figures are estimates and never a limit"; }
            @DecisionRecord { id = "AD-0030"; title = "Model validation in continuous integration"; }
        }
        part demoModel : DemoModel {
            doc /* The model under model/, this register among the rest of it. It
                   is read by the two reference tools and by the agreement test,
                   and never by the demo. */
            @DecisionRecord { id = "AD-0021"; title = "The Markdown architecture description as the record, A3 sheets as the overview"; }
            @DecisionRecord { id = "AD-0022"; title = "Track internal/assert and internal/tabletest"; }
            @DecisionRecord { id = "AD-0023"; title = "The light requirements scheme"; }
            @DecisionRecord { id = "AD-0029"; title = "The demo's own model in SysML v2"; }
        }
        part checkSuite : CheckSuiteProject {
            doc /* The optional check project, which drives the demo through a
                   tunnel and is no part of running it. */
        }

        // REPO.
        satisfy SR_08_LicencesCarriedInTheRepository by repository;
        satisfy SR_42_CompositionThatCannotDrift by repository;
        satisfy SR_46_TheModelAndTheRepositoryAgree by repository;
        satisfy SR_47_TheDemosModelAcceptedByTheReferenceTools by repository;
        satisfy SR_48_AnOptionalCheckSession by repository;

        // DEMOMODEL.
        satisfy SR_46_TheModelAndTheRepositoryAgree by demoModel;
        satisfy SR_47_TheDemosModelAcceptedByTheReferenceTools by demoModel;

        // CHECKS.
        satisfy SR_48_AnOptionalCheckSession by checkSuite;

        // The design constraints the repository carries.
        satisfy SC_01_LanguageAndLicence by repository;
        satisfy SC_03_TestFirst by repository;
        satisfy SC_04_TrackedByAllowlist by repository;
        satisfy SC_07_ContinuousIntegrationRunsUnitTestsAndPublishesOnTags by repository;
    }

    // ------------------------------------------------- the check session --

    package CheckSession {
        doc /* The optional check session beside the demo, written as the compose
               file that starts it. Every part that maps to a service carries the
               service name, so that the two can be compared. */

        part def CheckSession {
            doc /* The session as a whole: the runner, the tunnel, the collector,
                   the viewer, the demo under test and the external services they
                   talk to. */
        }

        part def SessionRunner :> DemoElement {
            doc /* The session script in a Node container. It creates the checks,
                   runs them, tears them down again, and reports that the
                   scheduled job ran. It is an element the requirements are
                   allocated to because the optional session story is satisfied
                   by it. */
            port control : HttpClientPort;
            port pings : HttpClientPort;
            attribute script : String = "checkly/scripts/session.sh";
        }

        part def TunnelConnector {
            doc /* The tunnel connector beside the demo. One usage stands for the
                   named tunnel, which needs a hostname and a token, and the
                   other for the quick tunnel, which needs neither. Exactly one
                   of the two runs. */
            port origin : HttpClientPort;
            port edge : TunnelPort;
            attribute mode : String;
        }

        part def TraceCollector {
            doc /* The collector that receives the router's spans over OTLP. It
                   forwards every span to the viewer and only the spans a check
                   marked as its own to the monitoring service's ingest. */
            port receive : OtlpPort;
            port toViewer : OtlpClientPort;
            port toIngest : OtlpClientPort;
            port publicReceive : OtlpPort;
            attribute config : String;
            attribute keepsOnly : String = "trace_state checkly=true, on the ingest pipeline";
        }

        part def TraceViewer {
            doc /* The trace viewer, where a session's spans are read. */
            port receive : OtlpPort;
            attribute ui : String = "localhost:16686";
        }

        part def PrivateLocationAgent {
            doc /* The private-location route, which runs the checks from beside
                   the demo instead of from a public location. It is available on
                   a paid plan only. */
        }

        part session : CheckSession {
            doc /* The optional check session beside the demo: the runner, the
                   tunnel, the collector, the viewer, and the external services
                   they talk to. Exactly one of the two tunnels runs. */
            part runner : SessionRunner {
                attribute composeService : String = "checkly";
            }
            part namedTunnel : TunnelConnector[0..1] {
                attribute composeService : String = "tunnel-named";
                :>> mode = "named";
            }
            part quickTunnel : TunnelConnector[0..1] {
                attribute composeService : String = "tunnel-quick";
                :>> mode = "quick";
            }
            part collector : TraceCollector {
                attribute composeService : String = "otel-collector";
            }
            part viewer : TraceViewer {
                attribute composeService : String = "jaeger";
            }
            part agent : PrivateLocationAgent[0..1] {
                attribute composeService : String = "checkly-agent";
            }
            ref part demo :> Federation_LogicalArchitecture::demo {
                attribute composeService : String = "demo";
            }
            ref part checkly : Federation_Context::CheckSession::ChecklyCloud;
            ref part edge : Federation_Context::CheckSession::TunnelEdge;
            ref part receiver : Federation_Context::CheckSession::AlertReceiver;
            interface traffic : CheckTraffic connect checkly.runners to edge.publicSide;
            interface tunnel : TunnelLink connect edge.origin to namedTunnel.edge;
            interface toDemo : HttpOnOnePort connect namedTunnel.origin to demo.web;
            interface export : TraceExport connect demo.router.traces to collector.receive;
            interface toViewer : TraceForward connect collector.toViewer to viewer.receive;
            interface toIngest : TraceForward connect collector.toIngest to checkly.ingest;
            interface control : ProjectControl connect runner.control to checkly.api;
            interface pings : Heartbeating connect runner.pings to checkly.api;
            interface alerting : Alerting connect checkly.alerts to receiver.webhook;
            flow routerSpans of Span from demo.router.traces to collector.receive;
            satisfy SR_48_AnOptionalCheckSession by runner;
        }
    }
}
