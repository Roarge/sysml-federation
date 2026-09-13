package Federation_VerificationCases {
    doc /* The verification case register. One case per system story, VC_SR_nn,
           and one per design constraint, VC_SC_nn. Each names the element the
           story is about as its subject, reads that story's acceptance in its
           objective, and owns one action per piece of evidence that verifies
           it. The action's metadata says what kind of evidence it is and where
           it lives: a Go test by its function name and file, a recorded run by
           its row in the example's verification record, a reference tool, a
           make target, a workflow step, an analysis or an inspection. A case
           whose evidence has not been produced yet carries its objective and no
           action. */
    private import Federation_SystemStories::*;
    private import Federation_DesignConstraints::*;
    private import Federation_Context::*;
    private import Federation_LogicalArchitecture::*;
    private import Federation_Library::Evidence;

    // ------------------------------------------------- launch and packaging --

    verification def VC_SR_01 {
        doc /* One command: the published image is the measured one, and the one
               docker run line serves the demo from a host with no credential. */
        subject demo : Demo;
        objective { verify SR_01_OneCommand.acceptance; }
        action <'publish.yml: Move latest to the measured image'> latestIsTheMeasuredImage {
            @Evidence { kind = "workflow"; location = ".github/workflows/publish.yml"; }
            doc /* The tag moves latest to the image whose layers were read back, so the one command pulls what was measured. */
        }
        action <'record: 2026-08-29 SR-01'> oneCommandUnauthenticated {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The one docker run line pulled the published image unauthenticated and served the four paths on port 8080. */
        }
    }

    verification def VC_SR_02 {
        doc /* Ready within ten seconds: a startup test in the supervisor and a
               measured start of the built image. */
        subject demo : Demo;
        objective {
            verify SR_02_ReadyWithinTenSeconds.acceptance;
            verify SR_02_ReadyWithinTenSeconds.readinessSla;
        }
        action <'TestSR02_ReadyWithinTenSeconds'> readyWithinTenSeconds {
            @Evidence { kind = "go-test"; location = "cmd/sysml-federation/main_test.go"; }
            doc /* Time to readiness of the three subgraphs, the router and the user interface server stays inside ten seconds. */
        }
        action <'record: 2026-08-28 SR-02'> containerStartMeasured {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A container polled every 200 ms answered the viewer after 0.75 s and a query through the router after 0.77 s. */
        }
    }

    verification def VC_SR_03 {
        doc /* No outbound network: the router child's environment, the
               loopback-only routing, the analysis of the outbound paths and the
               run with the route withheld. */
        subject demo : Demo;
        objective { verify SR_03_NoOutboundNetwork.acceptance; }
        action <'TestSR03_RouterEnvironmentDisablesTelemetry'> routerEnvironment {
            @Evidence { kind = "go-test"; location = "cmd/sysml-federation/main_test.go"; }
            doc /* The router process's environment carries the four variables that switch every outbound path off. */
        }
        action <'TestSR03_RouterConfigurationFileIsAnExplicitOptIn'> configurationFileOptIn {
            @Evidence { kind = "go-test"; location = "cmd/sysml-federation/main_test.go"; }
            doc /* With the variable unset the child's environment is the nine variables and carries no CONFIG_PATH. With it set, CONFIG_PATH carries the path and the nine are still set. */
        }
        action <'TestSR03_RoutingURLsAreLoopback'> routingIsLoopback {
            @Evidence { kind = "go-test"; location = "examples/pipeline/compose_test.go"; }
            doc /* The execution configuration names only loopback addresses. */
        }
        action <'the router outbound paths'> outboundPathsAnalysed {
            @Evidence { kind = "analysis"; location = "docs/decisions/AD-0013-telemetry-off.md"; }
            doc /* The router's code paths that could reach an address outside the container are read, and each is closed by a variable the image sets. */
        }
        action <'record: 2026-08-29 no outbound connection'> noConnectionWithTheRouteWithheld {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The image run with no network interface at debug log level reached healthy, logged 22 start-up lines and named no address but the loopback over nine minutes. */
        }
        action <'record: 2026-08-29 SR-03'> fourVariablesInTheChild {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The router child's environment read from the running container carries the four variables, and the router logged that usage tracking is disabled. */
        }
    }

    verification def VC_SR_04 {
        doc /* Four paths on one port: the paths and the redirect, on the test
               server, on the running stack and on the container. */
        subject demo : Demo;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        action <'TestSR04_FourPathsOnOnePort'> fourPathsAnswer {
            @Evidence { kind = "go-test"; location = "cmd/sysml-federation/main_test.go"; }
            doc /* The four paths answer on the one published port and the root redirects to the viewer. */
        }
        action <'TestSR04_ProxyStreamsServerSentEvents'> proxyStreamsEvents {
            @Evidence { kind = "go-test"; location = "cmd/sysml-federation/main_test.go"; }
            doc /* The proxy on that port streams server-sent events rather than buffering them. */
        }
        action <'record: 2026-08-27 SR-04'> pathsOnTheRunningStack {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The running stack answered 302 to the viewer in one hop and then 200 on each of the three app paths. */
        }
        action <'record: 2026-08-28 SR-04'> pathsOnTheContainer {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The container answered the redirect, the three paths, a query and a held subscription all on port 8080. */
        }
        action <'router.check.ts'> pathsAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
            doc /* router-playground reads the playground as HTML, router-root-redirect reads the 302 to the viewer without following it, and router-health-not-proxied expects a 404 on the health path, each hourly against a running instance. */
        }
    }

    verification def VC_SR_05 {
        doc /* Two platforms: the manifest read back after the push, and a cross
               build that resolves the router binary of the other architecture. */
        subject demo : Demo;
        objective { verify SR_05_TwoPlatforms.acceptance; }
        action <'publish.yml: Read the manifest back'> bothPlatformsInTheManifest {
            @Evidence { kind = "workflow"; location = ".github/workflows/publish.yml"; }
            doc /* The step reads the registry manifest and fails unless both platforms are listed. */
        }
        action <'record: 2026-08-28 cross-build router binary'> arm64BinaryResolves {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* An arm64 build gave an ARM aarch64 router and an ARM aarch64 supervisor, so the pinned multi-platform index resolves per architecture. */
        }
    }

    verification def VC_SR_06 {
        doc /* Size budget: the layer sizes read back after the push, and the
               per-platform figures measured from an exported build. */
        subject demo : Demo;
        objective {
            verify SR_06_SizeBudget.acceptance;
            verify SR_06_SizeBudget.sizeBudget;
        }
        action <'publish.yml: Read the manifest back'> layersWithinTheBudget {
            @Evidence { kind = "workflow"; location = ".github/workflows/publish.yml"; }
            doc /* The same step reads the compressed layer sizes of each platform and fails above the budget. */
        }
        action <'record: 2026-08-28 SR-06'> layersMeasured {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Compressed layers came to 44,845,388 bytes on amd64 and 41,470,306 on arm64, both inside 80 MB. */
        }
    }

    verification def VC_SR_07 {
        doc /* Licence carried in the image: the contents inspected, and the
               text read out of a running container. */
        subject demo : Demo;
        objective { verify SR_07_LicenceCarriedInTheImage.acceptance; }
        action <'image contents'> licenceBesideTheBinary {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The image is read for the router's Apache-2.0 text beside the copied binary. */
        }
        action <'record: 2026-08-28 SR-07'> licenceReadFromTheImage {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The licence file is the Apache License version 2.0 text, readable by the user the image runs as, and the binary reports the pinned release. */
        }
    }

    verification def VC_SR_08 {
        doc /* Licences carried in the repository: the notice and the text
               beside the one vendored file, with the checksum tested. */
        subject repository : Repository;
        objective { verify SR_08_LicencesCarriedInTheRepository.acceptance; }
        action <'TestSR08_VendoredSortableChecksum'> vendoredChecksum {
            @Evidence { kind = "go-test"; location = "examples/pipeline/ui/ui_test.go"; }
            doc /* The vendored file is the pinned release, matched by checksum. */
        }
        action <'repository root'> noticeAndMitTextInPlace {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The root is read for a notice naming both third-party works, with the MIT text beside the vendored file. */
        }
    }

    verification def VC_SR_09 {
        doc /* State lives in memory: a fresh store is the shipped state, a
               restart returns to it, and a stop under load loses nothing that
               was meant to survive. */
        subject demo : Demo;
        objective { verify SR_09_StateLivesInMemory.acceptance; }
        action <'TestSR09_AFreshStoreIsTheShippedState'> freshStore {
            @Evidence { kind = "go-test"; location = "adapter/projection/projection_test.go"; }
            doc /* A store built from the shipped source carries the shipped values and version 1. */
        }
        action <'TestSR09_RestartReturnsToShipped'> restartReturnsToShipped {
            @Evidence { kind = "go-test"; location = "examples/pipeline/document/document_test.go"; }
            doc /* Edits followed by a restart of the document service show the shipped tree again. */
        }
        action <'record: 2026-08-27 SR-09'> restartAfterEdits {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A stop and restart of the whole stack after the recorded edits came back on model version 1, document version 1 and the shipped verdict. */
        }
    }

    verification def VC_SR_10 {
        doc /* No resource from another origin: the embedded assets scanned, and
               the request list of a reload read in a browser. */
        subject demo : Demo;
        objective { verify SR_10_NoResourceFromAnotherOrigin.acceptance; }
        action <'TestSR10_NoResourceFromAnotherOrigin'> assetsScanned {
            @Evidence { kind = "go-test"; location = "examples/pipeline/ui/ui_test.go"; }
            doc /* The embedded assets carry no absolute URL to any origin. */
        }
        action <'record: 2026-08-28 SR-10'> requestListOnReload {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A reload of the viewer made eight requests, every one of them to the one published port, with no font and no other host. */
        }
    }

    // -------------------------------------------------------- the model viewer --

    verification def VC_SR_11 {
        doc /* Model text with the language visible: the tokeniser read against
               the reserved words, and the served file drawn in the viewer. */
        subject demo : Demo;
        objective { verify SR_11_ModelTextWithTheLanguageVisible.acceptance; }
        action <'tokeniser against the reserved words'> tokeniserInspected {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The tokeniser's word list is read against the reserved words of the specification. */
        }
        action <'record: 2026-08-28 SR-11'> notationDrawn {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The text pane drew keywords, numbers, quoted names and doc bodies distinctly, and the tokeniser returned a contiguous run of 28 tokens. */
        }
    }

    verification def VC_SR_12 {
        doc /* Sketch drawn from the wiring: the served wiring and a second
               wiring with fan-in from two depths, and the sketch following an
               edit. */
        subject demo : Demo;
        objective { verify SR_12_SketchDrawnFromTheWiring.acceptance; }
        action <'record: 2026-08-28 SR-12'> sketchAndSecondWiring {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Five boxes left to right with the throughput on each, the capacity in the caption and the bottleneck stroked red, and a second wiring with fan-in from two depths drew the same way. */
        }
        action <'record: 2026-08-28 SR-12, SR-14, SR-22'> sketchFollowsAnEdit {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Each edit in the viewer's panel redrew the sketch with the new capacity and the bottleneck where it had moved to. */
        }
    }

    verification def VC_SR_13 {
        doc /* The viewer's editable set: the controls on the page counted
               against the five throughputs and the one limit. */
        subject demo : Demo;
        objective { verify SR_13_TheViewersEditableSet.acceptance; }
        action <'record: 2026-08-28 SR-13, SR-38'> controlsCounted {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Six controls in the viewer, the five server throughputs and the global limit, and none for a derived limit, a latency or a capacity. */
        }
    }

    verification def VC_SR_14 {
        doc /* Verdict and reason in the viewer: the requirements column as
               served, and the same column after each edit. */
        subject demo : Demo;
        objective { verify SR_14_VerdictAndReasonInTheViewer.acceptance; }
        action <'record: 2026-08-28 SR-14, SR-15'> verdictsAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Every requirement block carried its verdict and its reason, including the inconclusive one naming the verification case that nothing runs. */
        }
        action <'record: 2026-08-28 SR-12, SR-14, SR-22'> verdictsAfterEdits {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Each edit moved the verdict and the reason of the requirements it bore on, through the three states of the worked example. */
        }
    }

    verification def VC_SR_15 {
        doc /* Failing requirement in red: the red block present while the
               verdict is FAIL and absent otherwise. */
        subject demo : Demo;
        objective { verify SR_15_FailingRequirementInRed.acceptance; }
        action <'record: 2026-08-28 SR-14, SR-15'> failingBlockInRed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The failing requirement was red and the passing and inconclusive ones were not. */
        }
    }

    // -------------------------------------------------------------- the adapter --

    verification def VC_SR_16 {
        doc /* Generic projection: a second fixture with other names and another
               wiring projects every field the requirement lists. */
        subject demo : Demo;
        objective { verify SR_16_GenericProjection.acceptance; }
        action <'TestSR16_WarehouseTextVersionAndIdentifiers'> textVersionAndIdentifiers {
            @Evidence { kind = "go-test"; location = "adapter/model/fixture_test.go"; }
            doc /* The second fixture projects its text, its version and its identifiers. */
        }
        action <'TestSR16_WarehousePartsAttributesPortsAndWiring'> partsAttributesPortsAndWiring {
            @Evidence { kind = "go-test"; location = "adapter/model/fixture_test.go"; }
            doc /* The second fixture projects every part with its attributes, its ports, its child parts and its connections. */
        }
        action <'TestSR16_WarehouseRequirementsAndRelationships'> requirementsAndRelationships {
            @Evidence { kind = "go-test"; location = "adapter/model/fixture_test.go"; }
            doc /* The second fixture projects every requirement with its constrained quantity, comparison and limit, and every derivation, satisfaction and verification relationship. */
        }
        action <'TestSR16_WarehouseMutations'> mutationsOnTheSecondFixture {
            @Evidence { kind = "go-test"; location = "adapter/model/fixture_test.go"; }
            doc /* Mutations on the second fixture patch its text and its projection as they do on the example. */
        }
    }

    verification def VC_SR_17 {
        doc /* No example identifiers in the adapter: the source searched, and
               the second fixture sharing no name with the example. */
        subject demo : Demo;
        objective { verify SR_17_NoExampleIdentifiersInTheAdapter.acceptance; }
        action <'TestSR17_NoExampleIdentifiersInTheAdapter'> noExampleIdentifiers {
            @Evidence { kind = "go-test"; location = "adapter/adapter_test.go"; }
            doc /* The adapter source carries no identifier from the example model. */
        }
        action <'TestSR17_SecondFixtureSharesNoNamesWithTheExample'> secondFixtureSharesNoNames {
            @Evidence { kind = "go-test"; location = "adapter/adapter_test.go"; }
            doc /* The second fixture shares no name with the example, so a pass on both cannot come from a name in the adapter. */
        }
        action <'adapter source'> sourceSearched {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The adapter source is read for the example's own words as well as its identifiers. */
        }
    }

    verification def VC_SR_18 {
        doc /* Refusal of unsupported syntax: one fixture per rejection the
               parser distinguishes, each naming file, line and column. */
        subject demo : Demo;
        objective { verify SR_18_RefusalOfUnsupportedSyntax.acceptance; }
        action <'TestSR18_StructuralRejections'> structuralRejections {
            @Evidence { kind = "go-test"; location = "adapter/syntax/parser_test.go"; }
            doc /* Each structural construct outside the supported subset is refused at the place it appears. */
        }
        action <'TestSR18_ExpressionAndPortRejections'> expressionAndPortRejections {
            @Evidence { kind = "go-test"; location = "adapter/syntax/parser_test.go"; }
            doc /* Each unsupported expression and port form is refused at the place it appears. */
        }
        action <'TestSR18_RequirementRejections'> requirementRejections {
            @Evidence { kind = "go-test"; location = "adapter/syntax/parser_test.go"; }
            doc /* Each unsupported requirement form is refused at the place it appears. */
        }
    }

    verification def VC_SR_19 {
        doc /* Quantity, comparison and limit from the constraint: the example's
               seven constraints, and the shapes the adapter refuses. */
        subject demo : Demo;
        objective { verify SR_19_QuantityComparisonAndLimitFromTheConstraint.acceptance; }
        action <'TestSR19_ExampleConstraints'> exampleConstraints {
            @Evidence { kind = "go-test"; location = "adapter/model/constraint_test.go"; }
            doc /* The example's six throughput constraints and its latency constraint project their quantity, comparison and limit. */
        }
        action <'TestSR19_ConstraintShapes'> constraintShapes {
            @Evidence { kind = "go-test"; location = "adapter/model/constraint_test.go"; }
            doc /* Each accepted shape of the rule projects the same three fields whichever way round it is written. */
        }
        action <'TestSR19_AWalkedChainProjectsThePartThatOwnsTheAttribute'> walkedChain {
            @Evidence { kind = "go-test"; location = "adapter/model/constraint_test.go"; }
            doc /* A feature chain rooted at the subject projects the part that owns the last segment. */
        }
        action <'TestSR19_ARequirementThatBindsNoSubjectIsRefused'> noSubjectRefused {
            @Evidence { kind = "go-test"; location = "adapter/model/constraint_test.go"; }
            doc /* A requirement whose constraint binds no subject is refused. */
        }
        action <'TestSR19_OtherShapesAreRefused'> otherShapesRefused {
            @Evidence { kind = "go-test"; location = "adapter/model/constraint_test.go"; }
            doc /* A constraint of any other shape is refused rather than guessed at. */
        }
    }

    verification def VC_SR_20 {
        doc /* Connection direction: the example wiring and a fixture with the
               ports the other way round. */
        subject demo : Demo;
        objective { verify SR_20_ConnectionDirection.acceptance; }
        action <'TestSR20_ConnectionsFollowEndOrderAndPortDirections'> endOrderAndPortDirections {
            @Evidence { kind = "go-test"; location = "adapter/model/model_test.go"; }
            doc /* Each connection projects from its first end to its second, and a reversed pair of ports is refused. */
        }
    }

    verification def VC_SR_21 {
        doc /* Short names as identifiers: the example's own identifiers, and an
               element that declares no short name. */
        subject demo : Demo;
        objective { verify SR_21_ShortNamesAsIdentifiers.acceptance; }
        action <'TestSR21_PartIdentifiersAreShortNamesWithQualifiedFallback'> partIdentifiers {
            @Evidence { kind = "go-test"; location = "adapter/model/model_test.go"; }
            doc /* A part is identified by its declared short name, and by its qualified name where it declares none. */
        }
        action <'TestSR21_RequirementAndVerificationIdentifiers'> requirementIdentifiers {
            @Evidence { kind = "go-test"; location = "adapter/model/model_test.go"; }
            doc /* Requirements and verification cases are identified the same way. */
        }
    }

    verification def VC_SR_22 {
        doc /* Edits patch the source: six Go tests across the model and
               projection packages, and the served text read back in the viewer. */
        subject demo : Demo;
        objective { verify SR_22_EditsPatchTheSource.acceptance; }
        action <'TestSR22_SetAttributePatchesTextAndProjectionTogether'> patchesTogether {
            @Evidence { kind = "go-test"; location = "adapter/model/patch_test.go"; }
            doc /* One mutation changes the literal in the text and the value in the projection together. */
        }
        action <'TestSR22_ConcurrentReadsDuringPatches'> concurrentReads {
            @Evidence { kind = "go-test"; location = "adapter/model/patch_test.go"; }
            doc /* Reads during a patch never see the text and the projection disagree. */
        }
        action <'TestSR22_ASharedLiteralIsNotEditable'> sharedLiteralNotEditable {
            @Evidence { kind = "go-test"; location = "adapter/model/patch_test.go"; }
            doc /* A literal two values share is not editable, because patching it would move both. */
        }
        action <'TestSR22_SetAttributePatchesTheTextAndBumpsTheVersion'> patchesAndBumpsTheVersion {
            @Evidence { kind = "go-test"; location = "adapter/projection/projection_test.go"; }
            doc /* A mutation through the store patches the served text and grows the version by one. */
        }
        action <'TestSR22_TextAndProjectionAgreeUnderConcurrentEdits'> agreeUnderConcurrentEdits {
            @Evidence { kind = "go-test"; location = "adapter/projection/projection_test.go"; }
            doc /* Concurrent edits leave the served text and the projection agreeing at every version. */
        }
        action <'TestSR22_AMutationAnswersFromTheModelItProduced'> answersFromTheModelItProduced {
            @Evidence { kind = "go-test"; location = "adapter/projection/projection_test.go"; }
            doc /* A mutation answers from the model it produced rather than from whatever the store holds by the time it replies. */
        }
        action <'record: 2026-08-28 SR-12, SR-14, SR-22'> editedNumbersInTheServedText {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Every value edited in the viewer appeared in the served text exactly where the original literal was. */
        }
    }

    verification def VC_SR_23 {
        doc /* Expression-bound values are evaluated: the operators, the
               parentheses, the units and the chains, and the derived limits
               following the global limit. */
        subject demo : Demo;
        objective { verify SR_23_ExpressionBoundValuesAreEvaluated.acceptance; }
        action <'TestSR23_DerivedLimitsOfTheExample'> derivedLimitsOfTheExample {
            @Evidence { kind = "go-test"; location = "adapter/model/constraint_test.go"; }
            doc /* The example's five derived limits evaluate to the values the worked example gives. */
        }
        action <'TestSR23_OperatorsParenthesesChainsAndUnits'> operatorsParenthesesChainsAndUnits {
            @Evidence { kind = "go-test"; location = "adapter/model/eval_test.go"; }
            doc /* Each operator, a parenthesised expression, a feature chain and a literal with a unit evaluate as written. */
        }
        action <'TestSR23_ExpressionBoundValuesAreNotEditable'> boundValuesAreNotEditable {
            @Evidence { kind = "go-test"; location = "adapter/model/eval_test.go"; }
            doc /* A value the evaluator produced is marked as bound rather than as a literal anyone may patch. */
        }
        action <'TestSR23_RedefinitionWithoutAValue'> redefinitionWithoutAValue {
            @Evidence { kind = "go-test"; location = "adapter/model/eval_test.go"; }
            doc /* A redefinition that binds no value of its own takes the value it redefines, and an attribute no declaration binds reports no value. */
        }
        action <'TestSR23_ADefinitionsExpressionUsesEachUsagesOwnValues'> definitionExpressionPerUsage {
            @Evidence { kind = "go-test"; location = "adapter/model/eval_test.go"; }
            doc /* An expression written on a definition evaluates against each usage's own values. */
        }
        action <'TestSR23_DerivedLimitsFollowTheGlobalLimit'> derivedLimitsFollowTheGlobalLimit {
            @Evidence { kind = "go-test"; location = "adapter/model/patch_test.go"; }
            doc /* A change to the global limit moves every limit derived from it. */
        }
        action <'TestSR23_SetLimitMovesTheDerivedLimits'> setLimitMovesTheDerivedLimits {
            @Evidence { kind = "go-test"; location = "adapter/projection/projection_test.go"; }
            doc /* The same movement is visible through the store a query reads. */
        }
    }

    verification def VC_SR_24 {
        doc /* Expression-bound values are read-only: an edit that targets one
               is refused and the value is unchanged. */
        subject demo : Demo;
        objective { verify SR_24_ExpressionBoundValuesAreReadOnly.acceptance; }
        action <'TestSR24_ExpressionBoundValuesAreRefused'> boundValuesRefusedInModel {
            @Evidence { kind = "go-test"; location = "adapter/model/patch_test.go"; }
            doc /* A mutation on a derived limit is refused and the limit stands. */
        }
        action <'TestSR24_ExpressionBoundAndUnboundValuesAreRefused'> boundAndUnboundRefusedInProjection {
            @Evidence { kind = "go-test"; location = "adapter/projection/projection_test.go"; }
            doc /* The store refuses both a value bound by an expression and a value bound to nothing at all. */
        }
        action <'record: 2026-08-27 SR-24'> derivedLimitRefusedThroughTheRouter {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A derived limit set through the router came back refused, with no change to the model version and no event on the subscription. */
        }
        action <'refusals.multistep.spec.ts'> derivedLimitRefusedAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/refusals.multistep.spec.ts"; }
            doc /* The refusals check sets PIPE-R1.3's limit through the router and expects the sentence that the value is not a literal in the source, with no data and the model version unchanged. */
        }
    }

    verification def VC_SR_25 {
        doc /* Rejection of invalid values: empty, text, negative, infinite and
               not-a-number inputs refused in both packages and in both apps. */
        subject demo : Demo;
        objective { verify SR_25_RejectionOfInvalidValues.acceptance; }
        action <'TestSR25_InvalidValuesAreRefused'> refusedInModel {
            @Evidence { kind = "go-test"; location = "adapter/model/patch_test.go"; }
            doc /* Empty, text, negative, infinite and not-a-number inputs are refused and the previous value stands. */
        }
        action refusedInProjection {
            @Evidence { kind = "go-test"; location = "adapter/projection/projection_test.go"; }
            doc /* TestSR25_InvalidValuesAreRefused in the projection package, where the same inputs reach the store a query reads. */
        }
        action <'record: 2026-08-27 SR-25'> negativeThroughTheRouter {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A negative throughput sent through the router came back with the service's own message and left the version alone. */
        }
        action <'record: 2026-08-28 SR-25'> badInputInBothApps {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Letters, a negative number and an emptied field were each refused in the viewer and in the document, and every refusal put the served value back. */
        }
        action <'record: 2026-08-28 Refusal text'> refusalTextReachesTheClient {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A refusal from the document service reached the client as the whole of the first error message, unwrapped. */
        }
        action <'refusals.multistep.spec.ts'> negativeValueRefusedAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/refusals.multistep.spec.ts"; }
            doc /* The refusals check sets PIPE-S1's throughput to -5 through the router and expects the sentence that the value must be a finite, non-negative number, with no data and the model version unchanged. */
        }
    }

    verification def VC_SR_26 {
        doc /* Model version events: a subscriber receives the new version after
               each accepted mutation. */
        subject demo : Demo;
        objective { verify SR_26_ModelVersionEvents.acceptance; }
        action <'TestSR26_ModelChangedReachesEverySubscriber'> everySubscriber {
            @Evidence { kind = "go-test"; location = "adapter/serve/server_test.go"; }
            doc /* Every subscriber receives the new version after each accepted mutation. */
        }
        action <'TestSR26_SubscriptionsAreComposedOverWebSocket'> composedOverWebSocket {
            @Evidence { kind = "go-test"; location = "examples/pipeline/compose_test.go"; }
            doc /* The composed configuration carries the subscription over the transport the router will use. */
        }
        action <'record: 2026-08-27 SR-26'> versionAfterEachMutation {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A subscription held through the router carried the new version within a second of each mutation. */
        }
    }

    verification def VC_SR_27 {
        doc /* Document version events: a subscriber receives the new document
               version after each accepted change. */
        subject demo : Demo;
        objective { verify SR_27_DocumentVersionEvents.acceptance; }
        action <'TestSR27_DocumentChangedReachesEverySubscriber'> everySubscriber {
            @Evidence { kind = "go-test"; location = "examples/pipeline/document/document_test.go"; }
            doc /* Every subscriber receives the new document version after each accepted change. */
        }
    }

    // ------------------------------------------------------ the capacity service --

    verification def VC_SR_28 {
        doc /* Rollup by maximum flow: table cases for the wiring shapes, and a
               differential test against the recursive minimum and sum. */
        subject demo : Demo;
        objective { verify SR_28_RollupByMaximumFlow.acceptance; }
        action <'TestSR28_Rollup'> rollupTableCases {
            @Evidence { kind = "go-test"; location = "examples/pipeline/capacity/flow/flow_test.go"; }
            doc /* Chain, fan-out, fan-in, nested branches, a cycle, a disconnected child, zero and missing values and a leaf each give the capacity the rule says. */
        }
        action <'TestSR28_FlowAgreesWithMinAndSum'> agreesWithMinAndSum {
            @Evidence { kind = "go-test"; location = "examples/pipeline/capacity/flow/flow_test.go"; }
            doc /* On series-parallel wirings the maximum flow agrees with the recursive minimum and sum. */
        }
    }

    verification def VC_SR_29 {
        doc /* Bottleneck as the minimum cut: the three states of the worked
               example and a tie case. */
        subject demo : Demo;
        objective { verify SR_29_BottleneckAsTheMinimumCut.acceptance; }
        action <'TestSR29_Analyse'> minimumCut {
            @Evidence { kind = "go-test"; location = "examples/pipeline/capacity/flow/verdict_test.go"; }
            doc /* The reported set is the source-side minimum cut, which stays the same where several cuts are minimal. */
        }
        action <'TestSR29_EntityResolversFollowTheWorkedExample'> entityResolvers {
            @Evidence { kind = "go-test"; location = "examples/pipeline/capacity/capacity_test.go"; }
            doc /* The capacity and bottleneck a query reads follow the three states of the worked example. */
        }
    }

    verification def VC_SR_30 {
        doc /* Verdicts and reasons: one case per verdict kind, per precedence
               rule and per reason template. */
        subject demo : Demo;
        objective { verify SR_30_VerdictsAndReasons.acceptance; }
        action <'TestSR30_Verdict'> verdictKindsAndPrecedence {
            @Evidence { kind = "go-test"; location = "examples/pipeline/capacity/flow/verdict_test.go"; }
            doc /* Each verdict kind, each precedence rule between them and each reason template is exercised. */
        }
        action <'TestSR30_LatencyIsInconclusiveThroughTheEntity'> latencyIsInconclusive {
            @Evidence { kind = "go-test"; location = "examples/pipeline/capacity/capacity_test.go"; }
            doc /* A requirement constraining a quantity the service does not compute reads inconclusive through the graph, naming its verification case. */
        }
    }

    verification def VC_SR_31 {
        doc /* Two configured names: the service carries the quantity and the
               attribute and nothing else of the model. */
        subject demo : Demo;
        objective { verify SR_31_TwoConfiguredNames.acceptance; }
        action <'TestSR31_NoExampleWordsInTheService'> noExampleWords {
            @Evidence { kind = "go-test"; location = "examples/pipeline/capacity/capacity_test.go"; }
            doc /* The service source carries no word of the example beyond the two configured names. */
        }
        action <'service schema, configuration and source'> twoNamesInspected {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The schema, the configuration and the source are read together, so that the declared field set is the whole of what the service selects. */
        }
    }

    verification def VC_SR_32 {
        doc /* No copy of the model: the service holds nothing between
               requests. */
        subject demo : Demo;
        objective { verify SR_32_NoCopyOfTheModel.acceptance; }
        action <'TestSR32_NoStore'> noStore {
            @Evidence { kind = "go-test"; location = "examples/pipeline/capacity/capacity_test.go"; }
            doc /* A restart between two identical requests changes nothing, and the service has no store to change. */
        }
    }

    // -------------------------------------------------- the requirements document --

    verification def VC_SR_33 {
        doc /* Numbering from the tree alone: the shipped tree, the tree after
               each kind of edit, and the numbering read in the app. */
        subject demo : Demo;
        objective { verify SR_33_NumberingFromTheTreeAlone.acceptance; }
        action <'TestSR33_ShippedNumbering'> shippedNumbering {
            @Evidence { kind = "go-test"; location = "examples/pipeline/document/tree/tree_test.go"; }
            doc /* The shipped tree numbers in dotted decimal from the tree alone, with prose skipped and unnumbered. */
        }
        action <'TestSR33_ShippedDocument'> shippedDocument {
            @Evidence { kind = "go-test"; location = "examples/pipeline/document/document_test.go"; }
            doc /* The shipped document the service serves carries that numbering. */
        }
        action <'record: 2026-08-28 SR-33, SR-37'> numberingAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A freshly started stack showed the prose unnumbered first, then the sections and their children numbered in tree order. */
        }
        action <'record: 2026-08-28 SR-33, SR-34'> numberingAfterAHeading {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A heading inserted above a section renumbered it and its children, and an exclusion and a restore renumbered them again. */
        }
        action <'record: 2026-08-28 Depth limit'> depthLimit {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* At the sixth level the app says so and withholds the two operations that would nest deeper, offering them at every level above. */
        }
    }

    verification def VC_SR_34 {
        doc /* Editorial operations in the app: one step per operation against
               the shipped tree. */
        subject demo : Demo;
        objective { verify SR_34_EditorialOperationsInTheApp.acceptance; }
        action <'record: 2026-08-28 SR-34'> dragsAndNesting {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A node moved above another, a node moved to the foot of a list and a node nested inside an empty child list each landed where they were aimed. */
        }
        action <'record: 2026-08-28 SR-33, SR-34'> headingProseAndExclusion {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A heading inserted as a parent, a prose node added, a heading's text edited in place, a requirement excluded and the same requirement restored. */
        }
    }

    verification def VC_SR_35 {
        doc /* Editorial operations in the service: one case per operation, with
               the restore rule and the renumbering. */
        subject demo : Demo;
        objective { verify SR_35_EditorialOperationsInTheService.acceptance; }
        action <'TestSR35_Operations'> operationsOnTheTree {
            @Evidence { kind = "go-test"; location = "examples/pipeline/document/tree/tree_test.go"; }
            doc /* Each operation applies to the tree, and a restored requirement returns as the last child of its former parent. */
        }
        action <'TestSR35_MutationsRenumberAndBumpTheVersion'> mutationsRenumber {
            @Evidence { kind = "go-test"; location = "examples/pipeline/document/document_test.go"; }
            doc /* Each accepted operation renumbers the document and grows the document version. */
        }
        action <'TestSR35_IncludeAndEditText'> includeAndEditText {
            @Evidence { kind = "go-test"; location = "examples/pipeline/document/document_test.go"; }
            doc /* Excluding and restoring a requirement and editing the text of a prose node behave as the operation list says. */
        }
    }

    verification def VC_SR_36 {
        doc /* Document operations leave the model unchanged: structurally, that
               the document service reaches no adapter package, and behaviourally
               through the running stack. */
        subject demo : Demo;
        objective { verify SR_36_DocumentOperationsLeaveTheModelUnchanged.acceptance; }
        action <'TestSR36_TheDocumentServiceReachesNoAdapterPackage'> noAdapterPackage {
            @Evidence { kind = "go-test"; location = "examples/pipeline/isolation_test.go"; }
            doc /* The document service reaches no adapter package, which is what makes the model version unreachable rather than merely unchanged. */
        }
        action <'record: 2026-08-27 SR-36'> versionsThroughTheRouter {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Three document operations driven through the router left the model version where it was, and a control edit showed that reading it can change. */
        }
        action <'record: 2026-08-28 SR-36'> countersInTheHeader {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* The document counter rose through every operation driven by hand and the model counter did not move. */
        }
    }

    verification def VC_SR_37 {
        doc /* What a requirement row shows: the row of each requirement of the
               example read against the projection. */
        subject demo : Demo;
        objective { verify SR_37_WhatARequirementRowShows.acceptance; }
        action <'record: 2026-08-28 SR-33, SR-37'> rowsAsServed {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Each row carried its short name, text, limit with its unit, its relationships, the part that satisfies it, the case that verifies it, the current value where the analysis returns one, and the verdict with its reason. */
        }
    }

    verification def VC_SR_38 {
        doc /* The document's editable set: the controls on the page counted
               against the throughputs and the one limit. */
        subject demo : Demo;
        objective { verify SR_38_TheDocumentsEditableSet.acceptance; }
        action <'record: 2026-08-28 SR-13, SR-38'> controlsCounted {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Six controls in the document, the same editable set as the viewer, and none for any other value. */
        }
    }

    // --------------------------------------------------- sharing through the router --

    verification def VC_SR_39 {
        doc /* Live reflection in the other app: a change in one app reaching the
               other within two seconds and without a reload, timed, including
               after a reconnection and after a frozen tab. */
        subject demo : Demo;
        objective {
            verify SR_39_LiveReflectionInTheOtherApp.acceptance;
            verify SR_39_LiveReflectionInTheOtherApp.reflectionSla;
        }
        action <'record: 2026-08-28 SR-39, SR-44'> resetReachesBothApps {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A reset pressed in one app put the shipped values back in both within two seconds. */
        }
        action <'record: 2026-08-28 SR-39 reconnection'> reconnectionAfterAStop {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* With the stack stopped the viewer said so and retried, and after the restart an edit made elsewhere reached it with no reload. */
        }
        action <'record: 2026-08-28 SR-39 background tab'> recoveryFromAFrozenTab {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* A page the browser had frozen two changes behind redrew the served state within a second of becoming visible and went on taking live changes. */
        }
        action <'record: 2026-08-28 SR-39 edits in both apps'> editsInBothApps {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Edits made in the document appeared in the viewer's text and sketch, and edits made in the viewer moved the document's verdicts, both without a reload. */
        }
    }

    verification def VC_SR_40 {
        doc /* Every request through the router: the calls the apps make read
               against the one endpoint. */
        subject demo : Demo;
        objective { verify SR_40_EveryRequestThroughTheRouter.acceptance; }
        action <'network calls of the apps'> networkCallsInspected {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The apps' queries, mutations and subscriptions are read, and every one of them goes to the router. */
        }
    }

    verification def VC_SR_41 {
        doc /* Services isolated: the module dependency graph and each service's
               outbound calls. */
        subject demo : Demo;
        objective { verify SR_41_ServicesIsolated.acceptance; }
        action <'TestSR41_NoServiceImportsAnother'> noServiceImportsAnother {
            @Evidence { kind = "go-test"; location = "examples/pipeline/isolation_test.go"; }
            doc /* Each service imports only its own subpackages. */
        }
        action <'TestSR41_NoServiceIsConfiguredWithAnAddress'> noConfiguredAddress {
            @Evidence { kind = "go-test"; location = "examples/pipeline/isolation_test.go"; }
            doc /* No service is configured with the address of another, so none could call one if it tried. */
        }
        action <'module dependency graph'> dependencyGraphInspected {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The dependency graph and each service's outbound calls are read together. */
        }
        action <'record: 2026-08-27 SR-41'> importSweep {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* An import sweep over the three services found each reaching only its own subpackages. */
        }
    }

    verification def VC_SR_42 {
        doc /* Composition that cannot drift: the committed configuration
               compared with the schema files it was composed from. */
        subject repository : Repository;
        objective { verify SR_42_CompositionThatCannotDrift.acceptance; }
        action <'TestSR42_ComposedSchemasMatchTheSchemaFiles'> composedSchemasMatch {
            @Evidence { kind = "go-test"; location = "examples/pipeline/compose_test.go"; }
            doc /* The schemas embedded in the committed configuration are identical to the schema files in the repository. */
        }
    }

    verification def VC_SR_43 {
        doc /* One query, three services: one response carrying fields from all
               three, and a served schema with types from each. */
        subject demo : Demo;
        objective { verify SR_43_OneQueryThreeServices.acceptance; }
        action <'record: 2026-08-27 SR-43'> oneQueryThreeServices {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* One query asked for a requirement's text, verdict and document number and one response carried all three, with introspection listing types from each subgraph. */
        }
        action <'router.check.ts'> joinAgainstARunningInstance {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
            doc /* router-join asks PIPE-R1 for its text, verdict, verdictReason and documentNumber through the router every hour and expects the shipped text, a verdict and document number 1 in one answer, with router-version reading the version through the same path. */
        }
    }

    verification def VC_SR_44 {
        doc /* Reset from either app: the two reset mutations, and a reset
               pressed by hand in each app. */
        subject demo : Demo;
        objective { verify SR_44_ResetFromEitherApp.acceptance; }
        action <'TestSR44_ResetRestoresTheShippedValuesAndGrowsTheVersion'> resetRestoresTheValues {
            @Evidence { kind = "go-test"; location = "adapter/projection/projection_test.go"; }
            doc /* Reset puts the shipped model values back and grows the version rather than returning it to one. */
        }
        action <'TestSR44_ResetRestoresTheShippedTreeAndGrowsTheVersion'> resetRestoresTheTree {
            @Evidence { kind = "go-test"; location = "examples/pipeline/document/document_test.go"; }
            doc /* Reset puts the shipped document structure back and grows the document version. */
        }
        action <'record: 2026-08-27 SR-44'> resetThroughTheRouter {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Both reset mutations sent through the router brought the shipped verdict back, and a restart of the stack gave version one again. */
        }
        action <'record: 2026-08-28 SR-39, SR-44'> resetFromEitherApp {
            @Evidence { kind = "record"; location = "examples/pipeline/README.md#verification-record"; }
            doc /* Reset pressed in each app in turn showed the shipped values and the shipped structure in both. */
        }
    }

    // ---------------------------------------------- the models and the repository --

    verification def VC_SR_45 {
        doc /* A model the reference tools accept: the example model and the
               adapter's second fixture put to both tools. */
        subject demo : Demo;
        objective { verify SR_45_AModelTheReferenceToolsAccept.acceptance; }
        action <'omg-pilot 2026-07'> pilotAcceptsTheExample {
            @Evidence { kind = "validator"; location = "make example-model-check"; }
            doc /* The pilot reads each example file and reports no error and no warning. */
        }
        action <'opensysml v0.6.0'> openSysmlAcceptsTheExample {
            @Evidence { kind = "validator"; location = "make example-model-check"; }
            doc /* The command line validates each example file strictly and exits zero. */
        }
    }

    verification def VC_SR_46 {
        doc /* The model and the repository agree: one test that fails on the
               first name one side carries and the other lacks. */
        subject demoModel : DemoModel;
        objective { verify SR_46_TheModelAndTheRepositoryAgree.acceptance; }
        action <'TestSR46_ModelAndRepositoryAgree'> modelAndRepositoryAgree {
            @Evidence { kind = "go-test"; location = "internal/trace/trace_test.go"; }
            doc /* Identifiers, decision records, Go test names, check files, published images, story coverage, the check inventory and the session services are each read from both sides and compared. */
        }
    }

    verification def VC_SR_47 {
        doc /* The demo's model accepted by the reference tools: both tools over
               the whole tree, locally and on every change. */
        subject demoModel : DemoModel;
        objective { verify SR_47_TheDemosModelAcceptedByTheReferenceTools.acceptance; }
        action <'omg-pilot 2026-07'> pilotAcceptsTheModel {
            @Evidence { kind = "validator"; location = "make model-check"; }
            doc /* The pilot reads every file under the model tree and reports no error and no warning. */
        }
        action <'opensysml v0.6.0'> openSysmlAcceptsTheModel {
            @Evidence { kind = "validator"; location = "make model-check"; }
            doc /* The command line validates the whole tree strictly, and a warning counts as a refusal. */
        }
        action <'model.yml'> validatedOnEveryChange {
            @Evidence { kind = "workflow"; location = ".github/workflows/model.yml"; }
            doc /* A pull request or a push that changes a file under the model tree puts the model to both tools. */
        }
    }

    verification def VC_SR_48 {
        doc /* An optional check session: the files of the check project, one
               action per file, with the checks, the monitors, the groups and
               the alert channels each declares. Every check has its own case
               in the check register. */
        subject checkSuite : CheckSuite;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        action <'router.check.ts'> routerChecks {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
            doc /* Declares the seven API checks on the router: router-version, router-join, router-capacity-bottleneck, router-introspection, router-playground, router-root-redirect and router-health-not-proxied. */
        }
        action <'refusals.multistep.spec.ts'> refusalsWalk {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/refusals.multistep.spec.ts"; }
            doc /* Drives the refusals check: the model version, two edits the router refuses, and the version again. */
        }
        action <'monitors.check.ts'> monitors {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
            doc /* Declares monitor-viewer, monitor-document, monitor-root and session-heartbeat, and, when a hostname is set, demo-certificate, demo-dns and demo-tcp. */
        }
        action <'groups.check.ts'> groups {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/groups.check.ts"; }
            doc /* Declares the five groups router, viewer, document, crossApp and session, with their locations, concurrency and retry policy. */
        }
        action <'alerts.check.ts'> alertChannels {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/alerts.check.ts"; }
            doc /* Declares the email and webhook alert channels, each only when its address is in the environment at deploy time. */
        }
        action <'stories.check.ts'> storyChecks {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/stories.check.ts"; }
            doc /* Constructs one check per manifest entry that names a spec file: the twelve browser checks us01-launch to us12-reset, the refusals walk, arithmetic-walk and document-operations, each with its group, its schedule and its tags from the entry. */
        }
        action <'suite.check.ts'> storySuite {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/suite.check.ts"; }
            doc /* Declares story-suite, the twelve browser specs run as one Playwright project under the shipped configuration. */
        }
        action <'arithmetic-walk.multistep.spec.ts'> arithmeticWalk {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/arithmetic-walk.multistep.spec.ts"; }
            doc /* Drives the arithmetic-walk check: the four rows of the capacity table through the router, then the reset. */
        }
        action <'document-operations.multistep.spec.ts'> documentOperations {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/document-operations.multistep.spec.ts"; }
            doc /* Drives the document-operations check: the six editorial mutations through the router with the model version read before and after, then the reset. */
        }
        action <'dashboard.check.ts'> dashboard {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/dashboard.check.ts"; }
            doc /* Declares sysml-federation-dashboard, the public dashboard of every check tagged demo, under a slug from the environment or derived from the account id. */
        }
        action <'status-page.check.ts'> statusPage {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/status-page.check.ts"; }
            doc /* Declares sysml-federation-status with one card, The demo, and the four services Viewer, Document, Router and Session, and, only when CHECKLY_INCIDENTS is 1, one automation rule per service tagged with its group's name. */
        }
        action <'maintenance.check.ts'> maintenanceWindow {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/maintenance.check.ts"; }
            doc /* Declares weekly, the maintenance window over every check tagged demo, constructed only when CHECKLY_MAINTENANCE is 1. */
        }
        action <'private-location.check.ts'> privateLocation {
            @Evidence { kind = "checkly"; location = "checkly/__checks__/private-location.check.ts"; }
            doc /* Declares demo-private, the private location every group runs from in place of its public locations, constructed only when CHECKLY_PRIVATE_LOCATION_SLUG names its slug. */
        }
    }

    // -------------------------------------------------------- design constraints --

    verification def VC_SC_01 {
        doc /* Language and licence: the module's own dependency list read
               against the permitted set. */
        subject demo : Demo;
        objective { verify SC_01_LanguageAndLicence.acceptance; }
        action <'go.mod'> permittedDependenciesOnly {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The module file is read for anything beyond the permitted dependencies and what those require. */
        }
    }

    verification def VC_SC_02 {
        doc /* No empty interface in hand-written code: the analyser over the
               platform pairs, with the reviewed exceptions counted. */
        subject demo : Demo;
        objective { verify SC_02_NoEmptyInterfaceInHandWrittenCode.acceptance; }
        action <'make vet'> analyserPasses {
            @Evidence { kind = "analysis"; location = "Makefile"; }
            doc /* The empty-interface analyser runs across the platform pairs and fails on a value position the rule forbids. */
        }
    }

    verification def VC_SC_03 {
        doc /* Test first: no implementation file in a package that has no
               test. */
        subject repository : Repository;
        objective { verify SC_03_TestFirst.acceptance; }
        action <'package layout'> everyPackageHasATest {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The package layout is read for an implementation file in a package that carries no test. */
        }
    }

    verification def VC_SC_04 {
        doc /* Tracked by allowlist: the two gates that catch a tracked path no
               rule names and a source file no rule reaches. */
        subject repository : Repository;
        objective { verify SC_04_TrackedByAllowlist.acceptance; }
        action <'make check-tracked'> noTrackedPathIsExcluded {
            @Evidence { kind = "make-target"; location = "Makefile"; }
            doc /* No tracked path is excluded by the ignore rules. */
        }
        action <'make check-allowlist'> everySourceFileIsReached {
            @Evidence { kind = "make-target"; location = "Makefile"; }
            doc /* Every source file on disk is one the ignore rules would track. */
        }
    }

    verification def VC_SC_05 {
        doc /* No build step for the web apps: the apps are read as they are
               served. */
        subject demo : Demo;
        objective { verify SC_05_NoBuildStepForTheWebApps.acceptance; }
        action <'no build step'> noBuildStep {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The apps are read for a build step or a toolchain between the source and what the binary serves, and there is neither. */
        }
    }

    verification def VC_SC_06 {
        doc /* Readable in an afternoon: the line totals measured and reported,
               a figure above its estimate recorded rather than corrected. */
        subject demo : Demo;
        objective { verify SC_06_ReadableInAnAfternoon.acceptance; }
        action <'TestSC06_LineBudgets'> lineBudgets {
            @Evidence { kind = "go-test"; location = "examples/pipeline/ui/ui_test.go"; }
            doc /* The hand-written line totals are measured and reported beside the figures the constraint gives. */
        }
    }

    verification def VC_SC_07 {
        doc /* Continuous integration runs unit tests and publishes on tags:
               three workflows and no others. */
        subject repository : Repository;
        objective { verify SC_07_ContinuousIntegrationRunsUnitTestsAndPublishesOnTags.acceptance; }
        action <'test.yml'> unitTestsOnEveryChange {
            @Evidence { kind = "workflow"; location = ".github/workflows/test.yml"; }
            doc /* The unit tests run on pull requests and on pushes to the main branch. */
        }
        action <'publish.yml'> imagePublishedOnTags {
            @Evidence { kind = "workflow"; location = ".github/workflows/publish.yml"; }
            doc /* The image is built and published on version tags. */
        }
        action <'model.yml'> modelValidatedOnChange {
            @Evidence { kind = "workflow"; location = ".github/workflows/model.yml"; }
            doc /* The demo's own model is put to both reference tools when a file under the model tree changes. */
        }
    }
}
