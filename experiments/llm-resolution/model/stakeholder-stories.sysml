package Experiment_StakeholderStories {
    doc /* Five stakeholder stories, EXP-US-01 to EXP-US-05. The first three are
           realised by use cases, and the last two are qualities every run must
           have. None is done until the operator's own run has been judged
           against it. */
    private import VSE_Library::*;
    private import Experiment_Stakeholders::*;
    private import Experiment_Concerns::*;
    private import Experiment_Context::Experiment;

    requirement <'EXP-US-01'> EXP_US_01_RunOnMyOwnServer : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : Experiment;
        stakeholder :>> role : Operator;
        attribute :>> capability = "run the whole experiment against the language model server in my network with one command";
        attribute :>> benefit    = "the systems model never leaves my network and I hand back a single file";
        frame concern : RunItOnMyServer;
        requirement :>> acceptance {
            requirement oneCommand {
                doc /* Given a checkout and the server's address, when the
                       operator runs the script, then every step runs without
                       further input. */
            }
            requirement oneFile {
                doc /* When the run ends or is stopped, then one results file
                       holds everything it produced. */
            }
            requirement withinAnHour {
                doc /* On the operator's server the full run finishes in about
                       an hour. */
            }
        }
    }

    requirement <'EXP-US-02'> EXP_US_02_ScoresAgainstTheRecordedLinks : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : Experiment;
        stakeholder :>> role : Evaluator;
        attribute :>> capability = "see each resolver scored against the links the demo's systems model records, with the keys hidden and the uncertainty shown";
        attribute :>> benefit    = "I can judge the language model against a resolver that needs none, on answers someone wrote down";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement keysHidden {
                doc /* No resolver ever reads a requirement key in a test. */
            }
            requirement baselineBeside {
                doc /* A resolver that uses no language model is scored on the
                       same tests. */
            }
            requirement intervalsShown {
                doc /* Every rate is given with its 95% interval. */
            }
            requirement unjudgedListed {
                doc /* Links proposed for tests that carry no key are listed for
                       a person to judge, never scored. */
            }
        }
    }

    requirement <'EXP-US-03'> EXP_US_03_ExplanationsPutToTheTest : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : Experiment;
        stakeholder :>> role : Evaluator;
        attribute :>> capability = "see whether the model's explanations name what its answers depended on";
        attribute :>> benefit    = "I know whether an explanation can stand as the record of why a link was made";
        frame concern : ExplanationsNameTheReasons;
        requirement :>> acceptance {
            requirement citedWordsTested {
                doc /* The words the model cited are removed and the question is
                       asked again, beside a control that removes other words. */
            }
            requirement stabilityTested {
                doc /* The same question in a different order, and unchanged, is
                       asked again. */
            }
        }
    }

    requirement <'EXP-US-04'> EXP_US_04_NothingLeavesTheNetwork : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : Experiment;
        stakeholder :>> role : ModelOwner;
        attribute :>> capability = "know that the systems model is read only by the server I named";
        attribute :>> benefit    = "the experiment can run where the model is allowed to be";
        frame concern : StaysInTheNetwork;
        requirement :>> acceptance {
            requirement oneServer {
                doc /* The experiment contacts no address but the server's. */
            }
        }
    }

    requirement <'EXP-US-05'> EXP_US_05_RepeatableRuns : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : Experiment;
        stakeholder :>> role : SystemsEngineer;
        attribute :>> capability = "repeat a run from what its results file records, and see how often the same question gets the same answer";
        attribute :>> benefit    = "a link can be checked by someone other than the person who made it";
        frame concern : SameAnswerTwice;
        requirement :>> acceptance {
            requirement provenanceRecorded {
                doc /* The file records the commit, the settings, the server's
                       version and the model's digest. */
            }
            requirement repeatsMeasured {
                doc /* A sample of questions is asked twice and the agreement
                       reported. */
            }
        }
    }
}
