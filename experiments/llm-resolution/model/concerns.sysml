package Experiment_Concerns {
    doc /* One concern per stakeholder question, in the words they would ask it. */
    private import Experiment_Stakeholders::*;
    private import Experiment_Context::ExperimentElement;

    concern def RunItOnMyServer {
        doc /* Can I run it on my own server in one sitting, and hand back one
               file? */
        subject system : ExperimentElement;
        stakeholder s : Operator;
    }

    concern def MeasuredAgainstTheRecord {
        doc /* Are the scores measured against links someone recorded, beside a
               resolver that needs no language model, with their uncertainty
               shown? */
        subject system : ExperimentElement;
        stakeholder s : Evaluator;
    }

    concern def ExplanationsNameTheReasons {
        doc /* When the model explains a link, does the explanation name what
               the answer depended on? */
        subject system : ExperimentElement;
        stakeholder s : Evaluator;
    }

    concern def StaysInTheNetwork {
        doc /* Does anything of the systems model leave the network it lives
               on? */
        subject system : ExperimentElement;
        stakeholder s : ModelOwner;
    }

    concern def ExplainTheOutage {
        doc /* Why is it down, where is the fault, what else is affected, and
               can I check each step of the answer? */
        subject system : ExperimentElement;
        stakeholder s : OnCallEngineer;
    }

    concern def SameAnswerTwice {
        doc /* Would the same inputs give the same links, and could someone else
               check them? */
        subject system : ExperimentElement;
        stakeholder s : SystemsEngineer;
    }
}
