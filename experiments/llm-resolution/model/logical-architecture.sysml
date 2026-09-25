package Experiment_LogicalArchitecture {
    doc /* What the experiment is made of, one part per Go file, and every
           allocation as a satisfy. */
    private import Experiment_Context::*;
    private import Experiment_SystemStories::*;
    private import Experiment_DesignConstraints::*;
    private import Federation_Library::DecisionRecord;

    part def CorpusReader :> ExperimentElement { doc /* corpus.go and words.go: reads the registers and the tests, and hides the keys. */ }
    part def BaselineResolver :> ExperimentElement { doc /* baseline.go: the word-overlap ranking and its threshold. */ }
    part def QuestionWriter :> ExperimentElement { doc /* prompt.go: the instructions, the requirement list and the test as a question. */ }
    part def ModelClient :> ExperimentElement { doc /* ollama.go: the chat call, the reply schema and the server's facts. */ }
    part def ProbeBuilder :> ExperimentElement { doc /* probes.go: the altered questions of the four probes. */ }
    part def Runner :> ExperimentElement { doc /* run.go: the order of the questions, the repeats, and resuming. */ }
    part def ResultsStore :> ExperimentElement { doc /* results.go: the JSON Lines file, written as it goes and read back. */ }
    part def Reporter :> ExperimentElement { doc /* report.go: the scores, the rates, their intervals and the proposals. */ }
    part def RunScript :> ExperimentElement { doc /* run.sh and main.go: the one command, the preflight and the flags. */ }
    part def TraceCheck :> ExperimentElement { doc /* trace_test.go: the agreement between this model and the tests. */ }

    part experiment : Experiment {
        doc /* The experiment as built. */
        @DecisionRecord { id = "AD-0032"; title = "A language model experiment kept outside the product"; }
        part corpus : CorpusReader;
        part baseline : BaselineResolver;
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
    satisfy EXP_SR_02_HideTheKeys by experiment.corpus;
    satisfy EXP_SR_03_AWordOverlapBaseline by experiment.baseline;
    satisfy EXP_SR_04_OneQuestionPerTest by experiment.questions;
    satisfy EXP_SR_04_OneQuestionPerTest by experiment.client;
    satisfy EXP_SR_05_DeleteTheCitedWords by experiment.probes;
    satisfy EXP_SR_06_RebuildFromTheCitedWords by experiment.probes;
    satisfy EXP_SR_07_ShuffleTheRequirements by experiment.probes;
    satisfy EXP_SR_08_AskTwice by experiment.runner;
    satisfy EXP_SR_09_OneResultsFile by experiment.store;
    satisfy EXP_SR_10_ResumeAStoppedRun by experiment.runner;
    satisfy EXP_SR_11_ReportWithIntervals by experiment.reporter;
    satisfy EXP_SR_12_OnlyTheGivenServer by experiment.client;
    satisfy EXP_SR_13_StopBeforeAsking by experiment.script;
    satisfy EXP_SR_14_AQuickRun by experiment.runner;
    satisfy EXP_SR_15_ReportFromAFile by experiment.reporter;
    satisfy EXP_SR_16_TheModelAndTheTestsAgree by experiment.trace;
    satisfy EXP_SC_01_StandardLibraryInItsOwnModule by experiment;
    satisfy EXP_SC_02_NoEmptyInterface by experiment;
    satisfy EXP_SC_03_AboutAnHour by experiment.runner;
    satisfy EXP_SC_04_ResultsStayOutOfTheRepository by experiment.store;
}
