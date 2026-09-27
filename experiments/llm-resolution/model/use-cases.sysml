package Experiment_UseCases {
    doc /* Four use cases, one per stakeholder story with an actor who acts:
           running the experiment, reading its scores, testing its
           explanations, and explaining an outage. The successions carry the
           order and nothing else. */
    private import Experiment_StakeholderStories::*;
    private import Experiment_Stakeholders::*;
    private import Experiment_Context::Experiment;

    use case def UC_EXP_01_RunTheExperiment {
        doc /* The operator runs the experiment against the server in their
               network and hands back the results file. */
        subject experiment : Experiment;
        actor operator : Operator;
        objective realises :> EXP_US_01_RunOnMyOwnServer { subject :>> system = experiment; }
        action giveTheAddress { doc /* the server's address in OLLAMA_URL, on a checkout of the repository */ }
        action checkTheServer { doc /* the script finds the server and the model, or stops with status 2 */ }
        action runQuickly { doc /* a quick run of twelve tests shows the chain works in minutes */ }
        action runInFull { doc /* the full run asks the base questions, then the probes, writing each reply as it arrives */ }
        action handBackTheFile { doc /* the operator uploads the one results file */ }
        first giveTheAddress then checkTheServer;
        first checkTheServer then runQuickly;
        first runQuickly then runInFull;
        first runInFull then handBackTheFile;
    }

    use case def UC_EXP_02_ReadTheScores {
        doc /* The evaluator reads each resolver's scores against the recorded
               links and judges the proposals nobody recorded. */
        subject experiment : Experiment;
        actor evaluator : Evaluator;
        objective realises :> EXP_US_02_ScoresAgainstTheRecordedLinks { subject :>> system = experiment; }
        action rebuildTheReport { doc /* the report is rebuilt from the file, with no server */ }
        action readTheScores { doc /* precision and recall for the key rule, the baseline and the language model, each with its interval */ }
        action judgeTheProposals { doc /* the links proposed for tests with no key, listed for a person to judge */ }
        first rebuildTheReport then readTheScores;
        first readTheScores then judgeTheProposals;
    }

    use case def UC_EXP_03_TestTheExplanations {
        doc /* The evaluator reads what happened when the model's cited words
               were deleted or kept alone, when a question was asked again, and
               when a link the incident's account relied on was removed. */
        subject experiment : Experiment;
        actor evaluator : Evaluator;
        objective realises :> EXP_US_03_ExplanationsPutToTheTest { subject :>> system = experiment; }
        action compareDeletionWithControl { doc /* the share of answers that changed when the cited words went, beside the share when other words went */ }
        action readTheReconstruction { doc /* the share of answers the cited words alone reproduced */ }
        action readTheStability { doc /* the share of identical repeats */ }
        action readTheRemoval { doc /* whether the incident's account changed, or kept citing what was removed */ }
        first compareDeletionWithControl then readTheReconstruction;
        first readTheReconstruction then readTheStability;
        first readTheStability then readTheRemoval;
    }

    use case def UC_EXP_04_ExplainAnIncident {
        doc /* The engineer on call says what is down, and reads the language
               model's account of why, where and what else, with every step
               cited. */
        subject experiment : Experiment;
        actor onCall : OnCallEngineer;
        objective realises :> EXP_US_06_ExplainTheOutage { subject :>> system = experiment; }
        action sayWhatIsDown { doc /* the alert, and the container's last line when it is known */ }
        action frameTheViewpoint { doc /* the language model frames the concern the view has to answer */ }
        action buildTheView { doc /* it searches, follows links and reads code, exposing and pruning elements */ }
        action readTheAccount { doc /* why, the cause, the mechanism, the code, the consequences and any suspected mismatch */ }
        action checkTheCitations { doc /* every element, link and file the account cites is looked up */ }
        first sayWhatIsDown then frameTheViewpoint;
        first frameTheViewpoint then buildTheView;
        first buildTheView then readTheAccount;
        first readTheAccount then checkTheCitations;
    }
}
