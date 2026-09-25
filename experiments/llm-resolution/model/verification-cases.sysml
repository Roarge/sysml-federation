package Experiment_VerificationCases {
    doc /* The verification register. One case per system story, VC_EXP_SR_nn,
           and one per design constraint, VC_EXP_SC_nn. Each names the story
           it verifies in its objective and owns one action per Go test, by
           the test's name and file. The tests were named here before they
           were written. A case whose evidence has not been produced yet
           carries its objective and no action. */
    private import Experiment_SystemStories::*;
    private import Experiment_DesignConstraints::*;
    private import Experiment_Context::Experiment;
    private import Federation_Library::Evidence;

    verification def VC_EXP_SR_01 {
        doc /* Reading the registers and the tests: the real checkout, and a fixture tree for what must be skipped. */
        subject experiment : Experiment;
        objective { verify EXP_SR_01_ReadTheRegistersAndTheTests.acceptance; }
        action <'TestEXPSR01_ReadsEveryRequirementWithItsStatement'> readsEveryRequirement {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/corpus_test.go"; }
            doc /* Every SR and SC story of the checkout is read, each with a statement. */
        }
        action <'TestEXPSR01_ReadsEveryTopLevelTestOnce'> readsEveryTest {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/corpus_test.go"; }
            doc /* A fixture tree's tests are read once each, and test data, hidden folders and other modules are skipped. */
        }
        action <'TestEXPSR01_EveryKeyATestCarriesIsDeclared'> everyKeyDeclared {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/corpus_test.go"; }
            doc /* Every key in a test name of the checkout is a declared requirement. */
        }
        action <'TestEXPSR01_EveryReferenceInTheSystemsModelResolves'> everyReferenceResolved {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/wiki_test.go"; }
            doc /* Every relationship of the checkout's systems model points at an element the wiki read. */
        }
        action <'TestEXPSR01_EveryLinkShowsFromBothEnds'> linksBothWays {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/wiki_test.go"; }
            doc /* A satisfy in a fixture model shows as satisfies on the part and satisfied by on the requirement. */
        }
        action <'TestEXPSR01_ElementsGetShortUniqueIDs'> shortUniqueIDs {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/wiki_test.go"; }
            doc /* A short name is the identifier, and otherwise the shortest suffix of the qualified name no other element shares. */
        }
    }

    verification def VC_EXP_SR_02 {
        doc /* Hiding the keys: in names and doc comments, and in every question the checkout gives rise to. */
        subject experiment : Experiment;
        objective { verify EXP_SR_02_HideTheKeys.acceptance; }
        action <'TestEXPSR02_KeysAreRemovedFromNamesAndDocs'> keysRemoved {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/corpus_test.go"; }
            doc /* Names lose Test and the key, and keys in doc comments are masked. */
        }
        action <'TestEXPSR02_NoKeyReachesAQuestion'> noKeyInAQuestion {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/prompt_test.go"; }
            doc /* No task of a question made from the checkout contains a key. */
        }
        action <'TestEXPSR02_TheRegisterIsHiddenInTheTestTask'> registerHidden {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* No tool answer in the test task names a Go test the verification register names. */
        }
        action <'TestEXPSR02_KeysAreMaskedInCodeAnswers'> codeMasked {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* A key in code a tool shows in the test task comes back masked. */
        }
    }

    verification def VC_EXP_SR_03 {
        doc /* The word-overlap baseline on a small corpus written for the test. */
        subject experiment : Experiment;
        objective { verify EXP_SR_03_AWordOverlapBaseline.acceptance; }
        action <'TestEXPSR03_RareSharedWordsRankFirst'> rareWordsRankFirst {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/baseline_test.go"; }
            doc /* A requirement sharing a rare word outranks one sharing a common word. */
        }
        action <'TestEXPSR03_NoProposalBelowTheThreshold'> noProposalBelowThreshold {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/baseline_test.go"; }
            doc /* A test sharing too little proposes no link. */
        }
        action <'TestEXPSR03_TheReasonIsTheSharedWords'> reasonIsTheSharedWords {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/baseline_test.go"; }
            doc /* The reason is the shared words, heaviest first. */
        }
        action <'TestEXPSR03_AnAcronymsPluralStaysOneWord'> acronymPlural {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/baseline_test.go"; }
            doc /* URLs, IDs and APIs are one word each when identifiers are split, not UR and Ls. */
        }
        action <'TestEXPSR03_SearchForRequirementsRanksAsTheBaseline'> searchMatchesTheBaseline {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* A search among the requirements for a test's text gives the baseline's order. */
        }
    }

    verification def VC_EXP_SR_04 {
        doc /* Browsing, against a stand-in server that records what it is sent. */
        subject experiment : Experiment;
        objective { verify EXP_SR_04_BrowseTheSystemsModel.acceptance; }
        action <'TestEXPSR04_EveryQuestionCarriesTheFixedSettings'> fixedSettings {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/ollama_test.go"; }
            doc /* Temperature 0, the seed and a context of 8192 arrive with every question. */
        }
        action <'TestEXPSR04_TheStepSchemaAdmitsOneToolAndViewEdits'> stepSchema {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/ollama_test.go"; }
            doc /* The step schema admits the seven tools or done, their arguments, and up to three exposures and three prunings. */
        }
        action <'TestEXPSR04_QuestionsOfOneTaskShareOneSystemMessage'> sharedPrefix {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/prompt_test.go"; }
            doc /* Every question of a test browse has the same system message, and every question of an incident browse has another. */
        }
        action <'TestEXPSR04_EachQuestionCarriesTheViewAndOnlyTheLastAnswer'> stateNotHistory {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/browse_test.go"; }
            doc /* The third question of a browse holds the view and the second tool's answer, and nothing of the first answer. */
        }
        action <'TestEXPSR04_ABrowseStopsAtItsBudget'> budgetHolds {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/browse_test.go"; }
            doc /* A language model that never says done gets eight tool calls for a test and twenty for an incident, then the final question. */
        }
        action <'TestEXPSR04_AReplyThatIsNotTheAskedJSONIsAnError'> malformedReply {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/ollama_test.go"; }
            doc /* A reply that is not the JSON asked for comes back as an error with its text kept. */
        }
        action <'TestEXPSR04_TheFinalSchemaAsksForTheLinksBeforeTheEvidence'> answerFirst {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/ollama_test.go"; }
            doc /* The final schema for a test lists links, evidence and reason in that order. */
        }
    }

    verification def VC_EXP_SR_05 {
        doc /* Deleting the cited words, and the two controls beside it. */
        subject experiment : Experiment;
        objective { verify EXP_SR_05_DeleteTheCitedWords.acceptance; }
        action <'TestEXPSR05_DeletionRedactsTheCitedWordsFromEveryToolAnswer'> citedWordsDeleted {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/probes_test.go"; }
            doc /* The cited words go from the task and from every tool answer of the browse, and nothing else changes. */
        }
        action <'TestEXPSR05_ControlRemovesAsManyOtherWords'> controlMatchesTheCount {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/probes_test.go"; }
            doc /* The control deletes as many words, none of them cited, the same ones on every run. */
        }
        action <'TestEXPSR05_TheRareSharedControlDeletesSharedWordsRarestFirst'> rareSharedControl {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/probes_test.go"; }
            doc /* Uncited words the test and the requirement share go, the rarest first, until at least as many are gone as the deletion took. */
        }
        action <'TestEXPSR05_NothingToDeleteIsRecordedAndNotAsked'> nothingToDelete {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* Evidence found nowhere gives a recorded note and no question. */
        }
        action <'TestEXPSR05_TheProbeSampleIsEveryFourthLinkedTest'> sampleFixed {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* The first test and every fourth after it among those linked to a requirement, in the order read. */
        }
    }

    verification def VC_EXP_SR_06 {
        doc /* Rebuilding the test from the cited words. */
        subject experiment : Experiment;
        objective { verify EXP_SR_06_RebuildFromTheCitedWords.acceptance; }
        action <'TestEXPSR06_ReconstructionShowsOnlyTheCitedWords'> onlyTheEvidence {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/probes_test.go"; }
            doc /* The task holds the evidence and nothing else of the test. */
        }
        action <'TestEXPSR06_TheRebuiltBrowseHidesTheTestsOwnCode'> ownCodeHidden {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/probes_test.go"; }
            doc /* No grep or read in the rebuilt browse shows the test's own function. */
        }
    }

    verification def VC_EXP_SR_08 {
        doc /* Browsing again, unchanged. */
        subject experiment : Experiment;
        objective { verify EXP_SR_08_AskTwice.acceptance; }
        action <'TestEXPSR08_EveryFourthTestIsAskedTwice'> everyFourth {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* The first test and every fourth after it are browsed twice. */
        }
    }

    verification def VC_EXP_SR_09 {
        doc /* The results file, against a stand-in server. */
        subject experiment : Experiment;
        objective { verify EXP_SR_09_OneResultsFile.acceptance; }
        action <'TestEXPSR09_EachReplyIsWrittenAsItArrives'> writtenAsItGoes {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* The file holds a call line for each reply the server has given. */
        }
        action <'TestEXPSR09_TheHeaderRecordsWhatARepeatNeeds'> provenance {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* The header holds the commit, the settings, the version, the digest and a hash of the systems model and the tests. */
        }
    }

    verification def VC_EXP_SR_10 {
        doc /* Resuming a stopped run. */
        subject experiment : Experiment;
        objective { verify EXP_SR_10_ResumeAStoppedRun.acceptance; }
        action <'TestEXPSR10_ResumingAsksOnlyWhatIsMissing'> onlyWhatIsMissing {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* A resumed run asks the server only the questions the earlier file lacks. */
        }
        action <'TestEXPSR10_ResumingRefusesADifferentCorpusOrSettings'> refusesAMismatch {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* A file from another corpus or other settings is refused. */
        }
    }

    verification def VC_EXP_SR_11 {
        doc /* The report, from calls written for the test. */
        subject experiment : Experiment;
        objective { verify EXP_SR_11_ReportWithIntervals.acceptance; }
        action <'TestEXPSR11_WilsonIntervalsMatchKnownValues'> wilsonIntervals {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* 8 of 10 gives 0.49 to 0.94, 80 of 100 gives 0.71 to 0.87. */
        }
        action <'TestEXPSR11_OnlyKeyedTestsAreScored'> onlyKeyedScored {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* A test with no key never enters precision or recall. */
        }
        action <'TestEXPSR11_ProbeRatesAndProposalsAreReported'> probesAndProposals {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* Each probe's changed-answer rate and every unkeyed proposal appear. */
        }
        action <'TestEXPSR11_TheResolversAreComparedTestByTest'> pairedComparison {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* Both right, only one right, neither: counted per keyed test, with McNemar's exact p and the best-ranked baseline row. */
        }
        action <'TestEXPSR11_ChangedAnswersAreSplitIntoNoneAndAnother'> changesSplit {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* A change to none and a change to another requirement are counted apart. */
        }
        action <'TestEXPSR11_EveryProbeIsAlsoReportedOnCorrectLinks'> correctLinksOnly {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* Each probe has a row for the links that were right. */
        }
        action <'TestEXPSR11_UnkeyedProposalsAreListedForBlindJudgement'> blindList {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* The proposals appear again, shuffled, without resolver or reason. */
        }
        action <'TestEXPSR11_EvidenceCountsAsFoundOnlyAsWritten'> groundedAsWritten {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* Evidence found only word by word, or found only outside the test, does not count. */
        }
        action <'TestEXPSR11_OtherKindsAreMeasuredAgainstABaseRate'> otherKindsAgainstABaseRate {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* Links to parts and actions are counted by kind, with the share near the recorded requirement beside the share of every element. */
        }
    }

    verification def VC_EXP_SR_12 {
        doc /* Only the given server, over a whole run. */
        subject experiment : Experiment;
        objective { verify EXP_SR_12_OnlyTheGivenServer.acceptance; }
        action <'TestEXPSR12_OnlyTheGivenServerIsContacted'> oneHost {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* Every request of a full run against a stand-in server goes to its address. */
        }
    }

    verification def VC_EXP_SR_13 {
        doc /* Stopping before asking. */
        subject experiment : Experiment;
        objective { verify EXP_SR_13_StopBeforeAsking.acceptance; }
        action <'TestEXPSR13_AMissingModelStopsTheRunBeforeAnyQuestion'> missingModel {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/main_test.go"; }
            doc /* A server without the model gives status 2 and no question. */
        }
        action <'TestEXPSR13_AnUnreachableServerStopsTheScript'> unreachableServer {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/script_test.go"; }
            doc /* run.sh against an address that answers nothing exits with status 2. */
        }
    }

    verification def VC_EXP_SR_14 {
        doc /* The quick run. */
        subject experiment : Experiment;
        objective { verify EXP_SR_14_AQuickRun.acceptance; }
        action <'TestEXPSR14_AQuickRunTakesTwelveTestsAndOneIncident'> twelveTests {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* Eight keyed and four unkeyed tests, the same twelve each time, and one incident browse. */
        }
    }

    verification def VC_EXP_SR_15 {
        doc /* The report rebuilt from a file. */
        subject experiment : Experiment;
        objective { verify EXP_SR_15_ReportFromAFile.acceptance; }
        action <'TestEXPSR15_TheReportIsRebuiltFromAFile'> sameReport {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/main_test.go"; }
            doc /* The report rebuilt from the file equals the one the run wrote, with no server running. */
        }
    }

    verification def VC_EXP_SR_16 {
        doc /* The agreement between this register and the tests. */
        subject experiment : Experiment;
        objective { verify EXP_SR_16_TheModelAndTheTestsAgree.acceptance; }
        action <'TestEXPSR16_TheModelAndTheTestsAgree'> bothWays {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/trace_test.go"; }
            doc /* Every test named here exists in its file, and every TestEXP function is named in its requirement's case. */
        }
    }

    verification def VC_EXP_SR_17 {
        doc /* The tools over the systems model, on a fixture model. */
        subject experiment : Experiment;
        objective { verify EXP_SR_17_ToolsOverTheSystemsModel.acceptance; }
        action <'TestEXPSR17_LinksAreGroupedByRelationship'> linksGrouped {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* A part's links come back grouped, satisfies beside typed by, as identifiers. */
        }
        action <'TestEXPSR17_PathsHaveAtMostThreeHops'> pathsAtMostThreeHops {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* Paths are shortest first, at most three, of at most three links each, and none beyond. */
        }
        action <'TestEXPSR17_EveryAnswerStaysWithinItsLimit'> capsHeld {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* Long answers are cut at their limit and say how many more there were. */
        }
        action <'TestEXPSR17_AnUnknownIdentifierIsAnsweredAsUnknown'> unknownAnswered {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* An identifier the systems model lacks gets an answer saying so. */
        }
        action <'TestEXPSR17_CodeGivesTheSystemsModelsOwnLocations'> codeFromTheModel {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* Evidence locations, paths in the description and where distinctive values occur. */
        }
    }

    verification def VC_EXP_SR_18 {
        doc /* The tools over the code, on the fixture tree. */
        subject experiment : Experiment;
        objective { verify EXP_SR_18_ToolsOverTheCode.acceptance; }
        action <'TestEXPSR18_GrepSearchesTheBuiltSystemOnly'> builtSystemOnly {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* A word in a fixture model file, a doc and test data is found only in the built system's file. */
        }
        action <'TestEXPSR18_ReadShowsTheLinesAroundTheOneAsked'> linesAround {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/tools_test.go"; }
            doc /* Numbered lines around the one asked, within the limit. */
        }
    }

    verification def VC_EXP_SR_19 {
        doc /* The view as working state. The rendered view is also put to both reference tools by hand, on the view of the quick run. */
        subject experiment : Experiment;
        objective { verify EXP_SR_19_TheViewAsWorkingState.acceptance; }
        action <'TestEXPSR19_ExposeAndPruneChangeTheView'> exposeAndPrune {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/view_test.go"; }
            doc /* An exposed element stays with its note until pruned, and an unknown one is refused. */
        }
        action <'TestEXPSR19_TheViewIsWrittenOutAsSysML'> writtenAsSysML {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/view_test.go"; }
            doc /* A viewpoint definition framing the task, and a view exposing each element by its qualified name. */
        }
    }

    verification def VC_EXP_SR_20 {
        doc /* The incident and its key. */
        subject experiment : Experiment;
        objective { verify EXP_SR_20_ExplainAnIncident.acceptance; }
        action <'TestEXPSR20_TheKeyNamesOnlyElementsThatExist'> keyCommitted {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/incident_test.go"; }
            doc /* Every element the committed key names is an element of the checkout's systems model. */
        }
        action <'TestEXPSR20_TheIncidentSeesTheWholeSystemsModel'> wholeModel {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/incident_test.go"; }
            doc /* The incident's tools show the register and mask no key. */
        }
        action <'TestEXPSR20_TheIncidentIsScoredItemByItem'> scoredItemByItem {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/incident_test.go"; }
            doc /* Each item of the key is found or missed in an answer written for the test. */
        }
    }

    verification def VC_EXP_SR_21 {
        doc /* Suspected mismatches. */
        subject experiment : Experiment;
        objective { verify EXP_SR_21_SuspectedMismatches.acceptance; }
        action <'TestEXPSR21_EveryFinalSchemaAsksForSuspectedMismatches'> askedEveryTime {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/ollama_test.go"; }
            doc /* Both final schemas carry the mismatches field. */
        }
        action <'TestEXPSR21_SuspectedMismatchesAreListed'> listed {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/report_test.go"; }
            doc /* Each suspicion appears with its element, both accounts and the reason. */
        }
    }

    verification def VC_EXP_SR_22 {
        doc /* The incident's probes. */
        subject experiment : Experiment;
        objective { verify EXP_SR_22_IncidentProbes.acceptance; }
        action <'TestEXPSR22_TheIncidentIsBrowsedInFiveWays'> fourVariants {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/run_test.go"; }
            doc /* As reported, again unchanged, with the alert alone, and with each of the two removals. */
        }
        action <'TestEXPSR22_ARemovalTakesAwayOnlyItsTarget'> onlyTheTargetRemoved {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/probes_test.go"; }
            doc /* Removing a transition changes its own pages and links and nothing else. */
        }
    }

    verification def VC_EXP_SR_23 {
        doc /* The citation check. */
        subject experiment : Experiment;
        objective { verify EXP_SR_23_CitationsChecked.acceptance; }
        action <'TestEXPSR23_AMissingElementIsFlagged'> missingElementFlagged {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/incident_test.go"; }
            doc /* An identifier no tool could show is marked missing. */
        }
        action <'TestEXPSR23_ConsecutiveStepsMustBeLinked'> unlinkedStepFlagged {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/incident_test.go"; }
            doc /* Two steps no link joins, and no code the first names, are marked missing. */
        }
        action <'TestEXPSR23_ACitationOfARemovedElementIsFlagged'> removedElementFlagged {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/incident_test.go"; }
            doc /* In a browse with a transition removed, citing that transition is marked missing. */
        }
    }

    verification def VC_EXP_SC_01 {
        doc /* Standard library only, in a module of its own. */
        subject experiment : Experiment;
        objective { verify EXP_SC_01_StandardLibraryInItsOwnModule.acceptance; }
        action <'TestEXPSC01_TheModuleHasNoDependencies'> noRequirements {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/constraints_test.go"; }
            doc /* go.mod names this module and requires nothing. */
        }
    }

    verification def VC_EXP_SC_02 {
        doc /* No empty interface. */
        subject experiment : Experiment;
        objective { verify EXP_SC_02_NoEmptyInterface.acceptance; }
        action <'TestEXPSC02_NoEmptyInterfaceInValuePositions'> noneFound {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/constraints_test.go"; }
            doc /* The module's Go files use neither interface{} nor any. */
        }
    }

    verification def VC_EXP_SC_03 {
        doc /* About two hours on the operator's server. The evidence is the operator's run, which has not happened yet. */
        subject experiment : Experiment;
        objective { verify EXP_SC_03_AboutTwoHours.acceptance; }
    }

    verification def VC_EXP_SC_04 {
        doc /* Results stay out of the repository. */
        subject experiment : Experiment;
        objective { verify EXP_SC_04_ResultsStayOutOfTheRepository.acceptance; }
        action <'TestEXPSC04_ResultsAreIgnoredByGit'> ignored {
            @Evidence { kind = "go-test"; location = "experiments/llm-resolution/constraints_test.go"; }
            doc /* git check-ignore reports a file under results/ as ignored. */
        }
    }
}
