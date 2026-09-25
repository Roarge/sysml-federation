package Experiment_ValidationCases {
    doc /* One validation case per stakeholder story. The evidence for each is
           the operator's run on their own server and the reading of its
           results file, and none has happened yet, so each case carries its
           objective and no action. */
    private import Experiment_StakeholderStories::*;
    private import Experiment_Context::Experiment;

    verification def VAL_EXP_US_01 {
        doc /* Run on my own server: the operator's quick run and full run. */
        subject experiment : Experiment;
        objective { verify EXP_US_01_RunOnMyOwnServer.acceptance; }
    }

    verification def VAL_EXP_US_02 {
        doc /* Scores against the recorded links: the report of the operator's run. */
        subject experiment : Experiment;
        objective { verify EXP_US_02_ScoresAgainstTheRecordedLinks.acceptance; }
    }

    verification def VAL_EXP_US_03 {
        doc /* Explanations put to the test: the probe rates of the operator's run. */
        subject experiment : Experiment;
        objective { verify EXP_US_03_ExplanationsPutToTheTest.acceptance; }
    }

    verification def VAL_EXP_US_04 {
        doc /* Nothing leaves the network: the operator's run with the server's address the only one given. */
        subject experiment : Experiment;
        objective { verify EXP_US_04_NothingLeavesTheNetwork.acceptance; }
    }

    verification def VAL_EXP_US_05 {
        doc /* Repeatable runs: the repeat rate of the operator's run, and the report rebuilt from its file. */
        subject experiment : Experiment;
        objective { verify EXP_US_05_RepeatableRuns.acceptance; }
    }
}
