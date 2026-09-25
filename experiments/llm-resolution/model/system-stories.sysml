package Experiment_SystemStories {
    doc /* Sixteen system stories, EXP-SR-01 to EXP-SR-16, each with its
           statement in EARS form, and their derivation from the stakeholder
           stories. A story is done as of the merge of the pull request that
           carries its evidence. */
    private import Experiment_StakeholderStories::*;
    private import Experiment_Stakeholders::*;
    private import Experiment_Concerns::*;
    private import Experiment_Context::ExperimentElement;
    private import Federation_Library::*;
    private import VSE_Library::*;
    private import ScalarValues::*;
    private import RequirementDerivation::*;

    requirement <'EXP-SR-01'> EXP_SR_01_ReadTheRegistersAndTheTests : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall read every system story and design constraint of the demo's systems model with its statement, and every top-level Go test function of the repository's main module, from the checkout it runs in.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "read the requirements and the tests from the checkout being run";
        attribute :>> benefit    = "the scores belong to the commit the file names";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement everyRequirement { doc /* Every SR and SC story is read, each with a statement. */ }
            requirement everyTest { doc /* Every top-level test is read once, and test data, hidden folders and other modules are skipped. */ }
            requirement everyKeyDeclared { doc /* Every key a test's name carries is declared in the systems model. */ }
        }
    }

    requirement <'EXP-SR-02'> EXP_SR_02_HideTheKeys : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall remove every requirement key from a test's name and doc comment before any resolver reads the test.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "hide the key a test's name carries";
        attribute :>> benefit    = "the resolvers face the case where the key was never carried";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement keysRemoved { doc /* A name such as TestSR22_X is shown as X, and keys in doc comments are masked. */ }
            requirement noKeyInAQuestion { doc /* No question put to the language model contains a requirement key in its test part. */ }
        }
    }

    requirement <'EXP-SR-03'> EXP_SR_03_AWordOverlapBaseline : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall rank the requirements for each test by the words they share, weighted by how few requirements use each word, propose the best only when its score reaches a threshold fixed in the code, and give the shared words as its reason.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "score a resolver that needs no language model on the same tests";
        attribute :>> benefit    = "the language model's score has something to be compared with";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement rareWordsCount { doc /* A word few requirements use counts for more than a common one. */ }
            requirement thresholdHolds { doc /* Below the threshold the baseline proposes no link. */ }
            requirement reasonIsTheWords { doc /* The reason given is the shared words, heaviest first. */ }
        }
    }

    requirement <'EXP-SR-04'> EXP_SR_04_OneQuestionPerTest : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "When the experiment asks the language model about a test, it shall send the instructions and every requirement as the system message and the test as the user message, with a temperature of 0, a fixed seed, a context of 8192 tokens and a reply schema that admits only a requirement key or none, up to five pieces of evidence and one reason.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "ask one question per test, the same way every time";
        attribute :>> benefit    = "every answer was produced under settings the file records";
        frame concern : ExplanationsNameTheReasons;
        frame concern : SameAnswerTwice;
        requirement :>> acceptance {
            requirement fixedSettings { doc /* Every question carries temperature 0, the seed and a context of 8192. */ }
            requirement schemaAdmitsOnlyKeys { doc /* The schema's requirement field admits the keys and none, and nothing else. */ }
            requirement sharedPrefix { doc /* The system message is the same for every unaltered question, so the server can cache it. */ }
            requirement malformedReplyIsAnError { doc /* A reply that is not the JSON asked for is recorded as an error. */ }
        }
    }

    requirement <'EXP-SR-05'> EXP_SR_05_DeleteTheCitedWords : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "When the language model has linked a test to a requirement, the experiment shall ask again with every word of the cited evidence deleted from the test and from that requirement, and again with as many other words deleted at random from the same places.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "see whether deleting the cited words changes the answer more than deleting others";
        attribute :>> benefit    = "an explanation that names the real reasons can be told from one that does not";
        frame concern : ExplanationsNameTheReasons;
        requirement :>> acceptance {
            requirement citedWordsDeleted { doc /* Every occurrence of the cited words is deleted from the test and the picked requirement, and nothing else. */ }
            requirement controlMatchesTheCount { doc /* The control deletes the same number of other words, chosen with a fixed seed. */ }
            requirement nothingToDelete { doc /* When none of the cited words occurs, no question is asked and the reason is recorded. */ }
        }
    }

    requirement <'EXP-SR-06'> EXP_SR_06_RebuildFromTheCitedWords : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "When the language model has linked a test to a requirement, the experiment shall ask again with the test replaced by the cited evidence alone.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "see whether the cited evidence alone leads to the same answer";
        attribute :>> benefit    = "an explanation that holds the reasons is shown to be enough for the answer";
        frame concern : ExplanationsNameTheReasons;
        requirement :>> acceptance {
            requirement onlyTheEvidence { doc /* The test part of the question holds the evidence and nothing of the test. */ }
        }
    }

    requirement <'EXP-SR-07'> EXP_SR_07_ShuffleTheRequirements : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "When the language model has linked a test to a requirement, the experiment shall ask again with the requirements listed in an order shuffled with a fixed seed, the same order for every test.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "see whether the answer depends on where a requirement sits in the list";
        attribute :>> benefit    = "a link that moves with the list order is known to be fragile";
        frame concern : ExplanationsNameTheReasons;
        requirement :>> acceptance {
            requirement oneShuffle { doc /* Every test is asked with the same shuffled order, a permutation of the original. */ }
        }
    }

    requirement <'EXP-SR-08'> EXP_SR_08_AskTwice : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : SystemsEngineer;
        attribute statement : String = "The experiment shall ask the question about every fourth test a second time, unchanged, and record whether the two replies are identical.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "measure how often the same question gets the same reply";
        attribute :>> benefit    = "the claim that a fixed seed repeats is checked rather than assumed";
        frame concern : SameAnswerTwice;
        requirement :>> acceptance {
            requirement everyFourth { doc /* The first test and every fourth after it are asked twice, whatever the first answer was. */ }
        }
    }

    requirement <'EXP-SR-09'> EXP_SR_09_OneResultsFile : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Operator;
        attribute statement : String = "The experiment shall write one results file in JSON Lines, opening with the commit, the settings, the server's version and the model's digest, and shall add each question and reply to it as the reply arrives.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "hand back one file that holds the whole run";
        attribute :>> benefit    = "nothing is lost if the run stops, and the file says how to repeat it";
        frame concern : RunItOnMyServer;
        frame concern : SameAnswerTwice;
        requirement :>> acceptance {
            requirement writtenAsItGoes { doc /* After n replies the file holds the header and n call lines. */ }
            requirement provenance { doc /* The header holds the commit, the settings, the server's version, the model's digest and a hash of what was read. */ }
        }
    }

    requirement <'EXP-SR-10'> EXP_SR_10_ResumeAStoppedRun : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Operator;
        attribute statement : String = "When the operator gives the results file of a stopped run, the experiment shall reuse its completed replies and ask only the questions it lacks, and if that file was made from different inputs or settings, it shall refuse to resume.";
        attribute pattern : EarsPattern = EarsPattern::Complex;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "pick up a stopped run where it stopped";
        attribute :>> benefit    = "an interruption costs minutes rather than the hour";
        frame concern : RunItOnMyServer;
        requirement :>> acceptance {
            requirement onlyWhatIsMissing { doc /* A resumed run asks no question the earlier file answered. */ }
            requirement refusesAMismatch { doc /* A file made from another corpus or other settings is refused. */ }
        }
    }

    requirement <'EXP-SR-11'> EXP_SR_11_ReportWithIntervals : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall report, for each resolver, precision and recall over the tests that carry a key, and for each probe, the share of changed answers, each with its 95% Wilson interval, and shall list every link proposed for a test that carries no key, unscored.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "read the scores with their uncertainty, and the unjudged links apart";
        attribute :>> benefit    = "no number is read as more certain than it is";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement wilsonIntervals { doc /* The interval for 8 of 10 is 0.49 to 0.94, as in the Wilson formula. */ }
            requirement onlyKeyedTestsScored { doc /* Tests with no key never enter precision or recall. */ }
            requirement probesAndProposalsReported { doc /* Each probe's rate and every proposal for an unkeyed test appear in the report. */ }
        }
    }

    requirement <'EXP-SR-12'> EXP_SR_12_OnlyTheGivenServer : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : ModelOwner;
        attribute statement : String = "The experiment shall send requests only to the server at the address the operator gives.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "keep every request on the named server";
        attribute :>> benefit    = "the systems model is read nowhere else";
        frame concern : StaysInTheNetwork;
        requirement :>> acceptance {
            requirement oneHost { doc /* Every request of a whole run goes to the given host and port. */ }
        }
    }

    requirement <'EXP-SR-13'> EXP_SR_13_StopBeforeAsking : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Operator;
        attribute statement : String = "If the server cannot be reached or does not hold the model, then the experiment shall stop before asking any question, with exit status 2 and a message naming what is missing.";
        attribute pattern : EarsPattern = EarsPattern::UnwantedBehaviour;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "learn at once that the server isn't ready";
        attribute :>> benefit    = "no half-run file is left behind to be mistaken for a result";
        frame concern : RunItOnMyServer;
        requirement :>> acceptance {
            requirement missingModel { doc /* A server without the model stops the run with status 2 before any question. */ }
            requirement unreachableServer { doc /* The script stops with status 2 when the address answers nothing. */ }
        }
    }

    requirement <'EXP-SR-14'> EXP_SR_14_AQuickRun : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Operator;
        attribute statement : String = "When the operator asks for a quick run, the experiment shall run every step on a fixed sample of twelve tests, eight that carry a key and four that do not.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "try the whole chain in minutes before committing an hour";
        attribute :>> benefit    = "a mistake in the setup shows up before the long run";
        frame concern : RunItOnMyServer;
        requirement :>> acceptance {
            requirement twelveTests { doc /* A quick run asks about eight keyed and four unkeyed tests, the same twelve every time. */ }
        }
    }

    requirement <'EXP-SR-15'> EXP_SR_15_ReportFromAFile : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "When given a results file to report on, the experiment shall rebuild the report from that file alone, without contacting a server.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "rebuild the report from a file someone sent";
        attribute :>> benefit    = "the numbers can be checked without the server that made them";
        frame concern : SameAnswerTwice;
        requirement :>> acceptance {
            requirement sameReport { doc /* The report rebuilt from a file equals the report the run wrote. */ }
        }
    }

    requirement <'EXP-SR-16'> EXP_SR_16_TheModelAndTheTestsAgree : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : SystemsEngineer;
        attribute statement : String = "A unit test of the experiment shall fail when its verification register names a Go test that does not exist, or when a Go test named for one of its requirements is missing from that requirement's case.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "keep the experiment's model and its tests in step";
        attribute :>> benefit    = "the evidence each case names is evidence that exists";
        frame concern : SameAnswerTwice;
        requirement :>> acceptance {
            requirement bothWays { doc /* Every test the register names exists in the file it gives, and every test named for a requirement is in that requirement's case. */ }
        }
    }

    // ------------------------------------------------------------ derivation --

    #derivation connection us01Derives {
        doc /* The requirements EXP-US-01 gives rise to. */
        end #original ::> EXP_US_01_RunOnMyOwnServer;
        end #derive   ::> EXP_SR_09_OneResultsFile;
        end #derive   ::> EXP_SR_10_ResumeAStoppedRun;
        end #derive   ::> EXP_SR_13_StopBeforeAsking;
        end #derive   ::> EXP_SR_14_AQuickRun;
    }

    #derivation connection us02Derives {
        doc /* The requirements EXP-US-02 gives rise to. */
        end #original ::> EXP_US_02_ScoresAgainstTheRecordedLinks;
        end #derive   ::> EXP_SR_01_ReadTheRegistersAndTheTests;
        end #derive   ::> EXP_SR_02_HideTheKeys;
        end #derive   ::> EXP_SR_03_AWordOverlapBaseline;
        end #derive   ::> EXP_SR_11_ReportWithIntervals;
    }

    #derivation connection us03Derives {
        doc /* The requirements EXP-US-03 gives rise to. */
        end #original ::> EXP_US_03_ExplanationsPutToTheTest;
        end #derive   ::> EXP_SR_04_OneQuestionPerTest;
        end #derive   ::> EXP_SR_05_DeleteTheCitedWords;
        end #derive   ::> EXP_SR_06_RebuildFromTheCitedWords;
        end #derive   ::> EXP_SR_07_ShuffleTheRequirements;
    }

    #derivation connection us04Derives {
        doc /* The requirements EXP-US-04 gives rise to. */
        end #original ::> EXP_US_04_NothingLeavesTheNetwork;
        end #derive   ::> EXP_SR_12_OnlyTheGivenServer;
    }

    #derivation connection us05Derives {
        doc /* The requirements EXP-US-05 gives rise to. */
        end #original ::> EXP_US_05_RepeatableRuns;
        end #derive   ::> EXP_SR_04_OneQuestionPerTest;
        end #derive   ::> EXP_SR_08_AskTwice;
        end #derive   ::> EXP_SR_09_OneResultsFile;
        end #derive   ::> EXP_SR_15_ReportFromAFile;
        end #derive   ::> EXP_SR_16_TheModelAndTheTestsAgree;
    }
}
