package Experiment_LogicalArchitecture {
    doc /* What the experiment is made of, one part per Go file, and every
           allocation as a satisfy. */
    private import Experiment_Context::*;
    private import Experiment_SystemStories::*;
    private import Experiment_DesignConstraints::*;
    private import Federation_Library::DecisionRecord;

    part def CorpusReader :> ExperimentElement { doc /* corpus.go and words.go: reads the registers and the tests, and hides the keys. */ }
    part def ModelWiki :> ExperimentElement { doc /* wiki.go: reads the demo's systems model into elements and their relationships, both ways. */ }
    part def Tools :> ExperimentElement { doc /* tools.go: find, links, doc, path, code, grep and read, with what each task hides. */ }
    part def BaselineResolver :> ExperimentElement { doc /* baseline.go: the word-overlap ranking and its threshold, and the ranking the search tool uses. */ }
    part def Browser :> ExperimentElement { doc /* browse.go: one question per tool call, then the final question. */ }
    part def ViewKeeper :> ExperimentElement { doc /* view.go: the view the language model builds, and its SysML v2 rendering. */ }
    part def IncidentCase :> ExperimentElement { doc /* incident.go and incident-key.json: the incident as reported, its key, its scoring and the citation check. */ }
    part def QuestionWriter :> ExperimentElement { doc /* prompt.go: the instructions for each task and the state of a browse as a question. */ }
    part def ModelClient :> ExperimentElement { doc /* ollama.go: the chat call, the step and final schemas and the server's facts. */ }
    part def ProbeBuilder :> ExperimentElement { doc /* probes.go: the deletions and the rebuild for a test, and the removals for the incident. */ }
    part def Runner :> ExperimentElement { doc /* run.go: the order of the questions, the repeats, and resuming. */ }
    part def ResultsStore :> ExperimentElement { doc /* results.go: the JSON Lines file, written as it goes and read back. */ }
    part def Reporter :> ExperimentElement { doc /* report.go: the scores, the rates, their intervals and the proposals. */ }
    part def RunScript :> ExperimentElement { doc /* run.sh and main.go: the one command, the preflight and the flags. */ }
    part def TraceCheck :> ExperimentElement { doc /* trace_test.go: the agreement between this model and the tests. */ }

    part experiment : Experiment {
        doc /* The experiment as built. */
        @DecisionRecord { id = "AD-0032"; title = "A language model experiment kept outside the product"; }
        part corpus : CorpusReader;
        part wiki : ModelWiki;
        part tools : Tools;
        part baseline : BaselineResolver;
        part browser : Browser;
        part viewKeeper : ViewKeeper;
        part incident : IncidentCase;
        part questions : QuestionWriter;
        part client : ModelClient;
        part probes : ProbeBuilder;
        part runner : Runner;
        part store : ResultsStore;
        part reporter : Reporter;
        part script : RunScript;
        part trace : TraceCheck;
    }

    satisfy EXP_SR_01_ReadTheRegistersAndTheTests by experiment.corpus;
    satisfy EXP_SR_01_ReadTheRegistersAndTheTests by experiment.wiki;
    satisfy EXP_SR_02_HideTheKeys by experiment.corpus;
    satisfy EXP_SR_02_HideTheKeys by experiment.tools;
    satisfy EXP_SR_03_AWordOverlapBaseline by experiment.baseline;
    satisfy EXP_SR_04_BrowseTheSystemsModel by experiment.browser;
    satisfy EXP_SR_04_BrowseTheSystemsModel by experiment.questions;
    satisfy EXP_SR_04_BrowseTheSystemsModel by experiment.client;
    satisfy EXP_SR_05_DeleteTheCitedWords by experiment.probes;
    satisfy EXP_SR_06_RebuildFromTheCitedWords by experiment.probes;
    satisfy EXP_SR_08_AskTwice by experiment.runner;
    satisfy EXP_SR_09_OneResultsFile by experiment.store;
    satisfy EXP_SR_10_ResumeAStoppedRun by experiment.runner;
    satisfy EXP_SR_11_ReportWithIntervals by experiment.reporter;
    satisfy EXP_SR_12_OnlyTheGivenServer by experiment.client;
    satisfy EXP_SR_13_StopBeforeAsking by experiment.script;
    satisfy EXP_SR_14_AQuickRun by experiment.runner;
    satisfy EXP_SR_15_ReportFromAFile by experiment.reporter;
    satisfy EXP_SR_16_TheModelAndTheTestsAgree by experiment.trace;
    satisfy EXP_SR_17_ToolsOverTheSystemsModel by experiment.tools;
    satisfy EXP_SR_18_ToolsOverTheCode by experiment.tools;
    satisfy EXP_SR_19_TheViewAsWorkingState by experiment.viewKeeper;
    satisfy EXP_SR_20_ExplainAnIncident by experiment.incident;
    satisfy EXP_SR_20_ExplainAnIncident by experiment.runner;
    satisfy EXP_SR_21_SuspectedMismatches by experiment.client;
    satisfy EXP_SR_21_SuspectedMismatches by experiment.reporter;
    satisfy EXP_SR_22_IncidentProbes by experiment.probes;
    satisfy EXP_SR_22_IncidentProbes by experiment.runner;
    satisfy EXP_SR_23_CitationsChecked by experiment.incident;
    satisfy EXP_SR_23_CitationsChecked by experiment.reporter;
    satisfy EXP_SC_01_StandardLibraryInItsOwnModule by experiment;
    satisfy EXP_SC_02_NoEmptyInterface by experiment;
    satisfy EXP_SC_03_AboutTwoHours by experiment.runner;
    satisfy EXP_SC_04_ResultsStayOutOfTheRepository by experiment.store;
}
