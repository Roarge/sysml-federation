package Experiment_SystemStories {
    doc /* Twenty-two system stories, EXP-SR-01 to EXP-SR-23, each with its
           statement in EARS form, and their derivation from the stakeholder
           stories. A story is done as of the merge of the pull request that
           carries its evidence. EXP-SR-07, which shuffled the requirement
           list of a single question, was withdrawn when the single question
           gave way to browsing, and its number is not reused. */
    private import Experiment_StakeholderStories::*;
    private import Experiment_Stakeholders::*;
    private import Experiment_Concerns::*;
    private import Experiment_Context::ExperimentElement;
    private import Federation_Library::*;
    private import VSE_Library::*;
    private import ScalarValues::*;
    private import RequirementDerivation::*;

    requirement <'EXP-SR-01'> EXP_SR_01_ReadTheRegistersAndTheTests : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall read every element of the demo's systems model with its kind, its description, its attributes and its relationships, every system story and design constraint with its statement, and every top-level Go test function of the repository's main module, from the checkout it runs in.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "read the requirements and the tests from the checkout being run";
        attribute :>> benefit    = "the scores belong to the commit the file names";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement everyRequirement { doc /* Every SR and SC story is read, each with a statement. */ }
            requirement everyTest { doc /* Every top-level test is read once, and test data, hidden folders and other modules are skipped. */ }
            requirement everyKeyDeclared { doc /* Every key a test's name carries is declared in the systems model. */ }
            requirement everyReferenceResolved { doc /* Every reference a relationship of the demo's systems model makes resolves to an element read. */ }
            requirement linksBothWays { doc /* A relationship from one element to another is shown on both, under its name on the first and its reverse on the second. */ }
            requirement shortUniqueIDs { doc /* Every element has one identifier, its short name where it has one and otherwise the shortest part of its qualified name that no other element shares. */ }
        }
    }

    requirement <'EXP-SR-02'> EXP_SR_02_HideTheKeys : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall remove every requirement key from a test's name and doc comment before any resolver reads the test, and while the language model links tests it shall hide the verification register's names of Go tests from every tool and mask every requirement key in the code the tools show.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "hide the key a test's name carries";
        attribute :>> benefit    = "the resolvers face the case where the key was never carried";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement keysRemoved { doc /* A name such as TestSR22_X is shown as X, and keys in doc comments are masked. */ }
            requirement noKeyInAQuestion { doc /* No question put to the language model contains a requirement key in its task. */ }
            requirement registerHidden { doc /* No tool answer in the test task names a Go test the verification register names. */ }
            requirement codeMasked { doc /* A requirement key in code the tools show in the test task is masked. */ }
        }
    }

    requirement <'EXP-SR-03'> EXP_SR_03_AWordOverlapBaseline : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall rank the requirements for each test by the words they share, weighted by how few requirements use each word, propose the best only when its score reaches a threshold fixed in the code, and give the shared words as its reason, and the language model's search of the systems model shall rank elements by the same weighting.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "score a resolver that needs no language model on the same tests";
        attribute :>> benefit    = "the language model's score has something to be compared with";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement rareWordsCount { doc /* A word few requirements use counts for more than a common one. */ }
            requirement thresholdHolds { doc /* Below the threshold the baseline proposes no link. */ }
            requirement reasonIsTheWords { doc /* The reason given is the shared words, heaviest first. */ }
            requirement searchMatchesTheBaseline { doc /* A search for a test's text among the requirements ranks them in the baseline's order. */ }
        }
    }

    requirement <'EXP-SR-04'> EXP_SR_04_BrowseTheSystemsModel : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "When the experiment asks the language model about a test or an incident, it shall ask one question per tool call, each carrying the fixed instructions for the task as the system message and the task, the viewpoint, the current view, the calls left and the last tool's answer as the user message, with a temperature of 0, a fixed seed, a context of 8192 tokens and a reply schema that admits one tool call and edits to the view, and it shall end with a final question when the language model is done or its calls are spent.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "let the language model read the systems model the way a person reads a wiki, a page at a time, the same way every time";
        attribute :>> benefit    = "a systems model too large for one question can still be read, and every answer was produced under settings the file records";
        frame concern : ExplanationsNameTheReasons;
        frame concern : SameAnswerTwice;
        requirement :>> acceptance {
            requirement fixedSettings { doc /* Every question carries temperature 0, the seed and a context of 8192. */ }
            requirement stepSchema { doc /* The reply schema admits one of the tools or done, its arguments, and up to three elements to expose and three to prune. */ }
            requirement sharedPrefix { doc /* The system message is the same for every question of one kind of task, so the server can cache it. */ }
            requirement stateNotHistory { doc /* The user message carries the view and the last tool's answer, and nothing older. */ }
            requirement budgetHolds { doc /* A test is given at most eight tool calls and an incident at most twenty, then the final question. */ }
            requirement malformedReplyIsAnError { doc /* A reply that is not the JSON asked for is recorded as an error. */ }
            requirement answerFirst { doc /* The final schema for a test lists the links before the evidence and the reason, so the evidence is written after the answer. */ }
        }
    }

    requirement <'EXP-SR-05'> EXP_SR_05_DeleteTheCitedWords : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "When the language model has linked a test in the probe sample to a requirement, the experiment shall browse again with every word of the cited evidence deleted from the test and from every tool's answer, again with as many other words deleted at random, and again with the uncited words the test and that requirement share deleted, the rarest first, until at least as many words are gone.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "see whether deleting the cited words changes the answer more than deleting others, including others just as distinctive";
        attribute :>> benefit    = "an explanation that names the real reasons can be told from one that does not";
        frame concern : ExplanationsNameTheReasons;
        requirement :>> acceptance {
            requirement citedWordsDeleted { doc /* Every occurrence of the cited words is deleted from the test and from every tool's answer in the probe's browse, and nothing else. */ }
            requirement controlMatchesTheCount { doc /* The control deletes the same number of other words, chosen with a fixed seed. */ }
            requirement rareSharedControl { doc /* The second control deletes uncited words the test and the picked requirement share, the rarest first, until at least as many words are gone as the deletion took. */ }
            requirement nothingToDelete { doc /* When none of the cited words occurs, no question is asked and the reason is recorded. */ }
            requirement sampleFixed { doc /* The probe sample is the first test and every fourth after it among the tests the language model linked to a requirement, in the order the tests are read. */ }
        }
    }

    requirement <'EXP-SR-06'> EXP_SR_06_RebuildFromTheCitedWords : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "When the language model has linked a test in the probe sample to a requirement, the experiment shall browse again with the test replaced by the cited evidence alone and the test's own code hidden from the tools.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "see whether the cited evidence alone leads to the same answer";
        attribute :>> benefit    = "an explanation that holds the reasons is shown to be enough for the answer";
        frame concern : ExplanationsNameTheReasons;
        requirement :>> acceptance {
            requirement onlyTheEvidence { doc /* The task holds the evidence and nothing else of the test. */ }
            requirement ownCodeHidden { doc /* No tool answer in the rebuilt browse shows the test's own function. */ }
        }
    }

    requirement <'EXP-SR-08'> EXP_SR_08_AskTwice : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : SystemsEngineer;
        attribute statement : String = "The experiment shall browse every fourth test a second time, unchanged, and record whether the two final replies are identical.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "measure how often the same question gets the same reply";
        attribute :>> benefit    = "the claim that a fixed seed repeats is checked rather than assumed";
        frame concern : SameAnswerTwice;
        requirement :>> acceptance {
            requirement everyFourth { doc /* The first test and every fourth after it are browsed twice, whatever the first answer was. */ }
        }
    }

    requirement <'EXP-SR-09'> EXP_SR_09_OneResultsFile : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Operator;
        attribute statement : String = "The experiment shall write one results file in JSON Lines, opening with the commit, the settings, the server's version, the model's digest and a hash of the systems model and the tests it read, and shall add each question and reply to it as the reply arrives.";
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
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall report, for each resolver, precision and recall of the requirement links over the tests that carry a key, and for each probe, the share of changed answers on all links and on correct links alone, split into changes to none and to another requirement, each with its 95% Wilson interval. It shall compare the two hidden-key resolvers test by test with McNemar's exact test, count as found only evidence that occurs as written in the test, count the links to elements of other kinds by kind with the share that lies within three links of the recorded requirement beside the same share for every element, and list every link proposed for a test that carries no key, unscored, once with its reasons and once without them for blind judgement.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "read the scores with their uncertainty, and the unjudged links apart";
        attribute :>> benefit    = "no number is read as more certain than it is";
        frame concern : MeasuredAgainstTheRecord;
        requirement :>> acceptance {
            requirement wilsonIntervals { doc /* The interval for 8 of 10 is 0.49 to 0.94, as in the Wilson formula. */ }
            requirement onlyKeyedTestsScored { doc /* Tests with no key never enter precision or recall. */ }
            requirement probesAndProposalsReported { doc /* Each probe's rate and every proposal for an unkeyed test appear in the report. */ }
            requirement pairedComparison { doc /* The keyed tests are counted by which hidden-key resolver got each right, McNemar's exact p is given, and the word overlap's best-ranked answer is scored as well as its thresholded one. */ }
            requirement changesSplit { doc /* A changed answer is counted as a change to none or to another requirement. */ }
            requirement correctLinksOnly { doc /* Every probe is also reported on the links that were right. */ }
            requirement blindList { doc /* The proposals for unkeyed tests are listed a second time, shuffled, without the resolver or its reason. */ }
            requirement groundedAsWritten { doc /* A piece of evidence counts as found only if it occurs, apart from case, in the test's fields. */ }
            requirement otherKindsAgainstABaseRate { doc /* Links to elements that are not requirements are counted by kind, and the share within three links of the recorded requirement is given beside the share of every element that is. */ }
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
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Operator;
        attribute statement : String = "When the operator asks for a quick run, the experiment shall run every step on a fixed sample of twelve tests, eight that carry a key and four that do not, and browse the incident once.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "try the whole chain in minutes before committing an hour";
        attribute :>> benefit    = "a mistake in the setup shows up before the long run";
        frame concern : RunItOnMyServer;
        requirement :>> acceptance {
            requirement twelveTests { doc /* A quick run browses eight keyed and four unkeyed tests, the same twelve every time, and the incident once. */ }
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

    requirement <'EXP-SR-17'> EXP_SR_17_ToolsOverTheSystemsModel : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : OnCallEngineer;
        attribute statement : String = "When the language model calls find, links, doc, path or code, the experiment shall answer from the systems model it read, with the smallest text that answers the call and within the call's limit.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "search the systems model, follow its links both ways, find the paths between two elements and the code an element names";
        attribute :>> benefit    = "a program does the reading a program can do, and the language model's tokens go on judgement";
        frame concern : ExplainTheOutage;
        requirement :>> acceptance {
            requirement linksGrouped { doc /* An element's links are listed by relationship, the reverse ones included, as identifiers. */ }
            requirement pathsAtMostThreeHops { doc /* A path between two elements has at most three links, and at most three paths are given, shortest first. */ }
            requirement capsHeld { doc /* No answer is longer than its limit, and a cut list says how many more there were. */ }
            requirement unknownAnswered { doc /* An identifier the systems model lacks is answered as unknown, and the browse goes on. */ }
            requirement codeFromTheModel { doc /* The code an element names is its evidence locations, the paths its description names and where its distinctive values occur in the code. */ }
        }
    }

    requirement <'EXP-SR-18'> EXP_SR_18_ToolsOverTheCode : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : OnCallEngineer;
        attribute statement : String = "When the language model calls grep or read, the experiment shall answer from the built system's files in the checkout, its Go, JavaScript, TypeScript, YAML and JSON files, Dockerfiles and shell scripts outside model, docs, experiments and test data, within the call's limit.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "match what the systems model says against what the code does";
        attribute :>> benefit    = "a place where the system was not built as modelled can be found";
        frame concern : ExplainTheOutage;
        requirement :>> acceptance {
            requirement builtSystemOnly { doc /* A search finds nothing under model, docs, experiments or test data. */ }
            requirement linesAround { doc /* A read shows the lines around the one asked for, numbered, within its limit. */ }
        }
    }

    requirement <'EXP-SR-19'> EXP_SR_19_TheViewAsWorkingState : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : OnCallEngineer;
        attribute statement : String = "The experiment shall keep the view the language model builds, adding and removing elements as it asks, show it in every question of the browse, and write it out at the end as a SysML v2 viewpoint definition that frames the task's concern and a view that exposes each element.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Inspection);
        attribute :>> capability = "see the part of the systems model the language model judged relevant, in the systems model's own language";
        attribute :>> benefit    = "the view can be read, checked and kept like any other view of the systems model";
        frame concern : ExplainTheOutage;
        requirement :>> acceptance {
            requirement exposeAndPrune { doc /* An element exposed with a note is in the view until it is pruned. */ }
            requirement writtenAsSysML { doc /* The view is written as a viewpoint definition with the task as its framed concern and a view exposing each element by its qualified name. */ }
        }
    }

    requirement <'EXP-SR-20'> EXP_SR_20_ExplainAnIncident : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : OnCallEngineer;
        attribute statement : String = "When the run reaches the incident, the experiment shall give the language model the incident as the engineer on call reports it, let it browse the whole systems model and the code, ask it why the incident happened, its cause, its mechanism, the code to look at, its consequences and the path it followed, and score the answer item by item against a key committed before the run.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "ask for a timely account of an outage from the systems model and the code";
        attribute :>> benefit    = "the account can be judged against answers written down before anyone saw it";
        frame concern : ExplainTheOutage;
        requirement :>> acceptance {
            requirement keyCommitted { doc /* The key is a file in the module, and every element it names exists in the systems model. */ }
            requirement wholeModel { doc /* The incident's tools hide nothing of the systems model and mask nothing in the code. */ }
            requirement scoredItemByItem { doc /* Each item of the key is reported as found or missed. */ }
        }
    }

    requirement <'EXP-SR-21'> EXP_SR_21_SuspectedMismatches : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : SystemsEngineer;
        attribute statement : String = "The experiment shall ask in every final question whether the language model suspects the system was not built as the systems model says, and shall list every suspicion with the element, what the systems model says, what the system shows and why.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "hear where the systems model and the code disagree";
        attribute :>> benefit    = "a link that has decayed, or never held, is found while it can still be fixed";
        frame concern : ExplainTheOutage;
        requirement :>> acceptance {
            requirement askedEveryTime { doc /* Every final schema has a field for suspected mismatches. */ }
            requirement listed { doc /* The report lists every suspicion with its element, both accounts and the reason. */ }
        }
    }

    requirement <'EXP-SR-22'> EXP_SR_22_IncidentProbes : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall browse the incident again unchanged, again with the alert alone, again with the supervisor's transition on the router's exit removed from the systems model, and again with an unrelated link removed instead.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "see whether the account follows the evidence it rests on";
        attribute :>> benefit    = "an account that does not change when its support is taken away is known not to rest on it";
        frame concern : ExplanationsNameTheReasons;
        requirement :>> acceptance {
            requirement fourVariants { doc /* The incident is browsed as reported, twice, with the alert alone, and with each of the two removals. */ }
            requirement onlyTheTargetRemoved { doc /* A removal takes away the named element or link, and everything else the tools show is unchanged. */ }
        }
    }

    requirement <'EXP-SR-23'> EXP_SR_23_CitationsChecked : UserStory {
        @StoryMeta { status = StoryStatus::inProgress; }
        subject :>> system : ExperimentElement;
        stakeholder :>> role : Evaluator;
        attribute statement : String = "The experiment shall check every element, link and code path a final answer cites against what the run's tools could show, and mark each citation found or missing.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "tell an account that cites what is there from one that cites what is not";
        attribute :>> benefit    = "the one part of an explanation a program can check is checked";
        frame concern : ExplanationsNameTheReasons;
        frame concern : ExplainTheOutage;
        requirement :>> acceptance {
            requirement missingElementFlagged { doc /* An identifier the tools could not show is marked missing. */ }
            requirement unlinkedStepFlagged { doc /* Two consecutive steps of a path that no link joins, and no code the first names, are marked missing. */ }
            requirement removedElementFlagged { doc /* In a browse with an element removed, a citation of that element is marked missing. */ }
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
        end #derive   ::> EXP_SR_17_ToolsOverTheSystemsModel;
        end #derive   ::> EXP_SR_18_ToolsOverTheCode;
    }

    #derivation connection us03Derives {
        doc /* The requirements EXP-US-03 gives rise to. */
        end #original ::> EXP_US_03_ExplanationsPutToTheTest;
        end #derive   ::> EXP_SR_04_BrowseTheSystemsModel;
        end #derive   ::> EXP_SR_05_DeleteTheCitedWords;
        end #derive   ::> EXP_SR_06_RebuildFromTheCitedWords;
        end #derive   ::> EXP_SR_22_IncidentProbes;
        end #derive   ::> EXP_SR_23_CitationsChecked;
    }

    #derivation connection us04Derives {
        doc /* The requirements EXP-US-04 gives rise to. */
        end #original ::> EXP_US_04_NothingLeavesTheNetwork;
        end #derive   ::> EXP_SR_12_OnlyTheGivenServer;
    }

    #derivation connection us05Derives {
        doc /* The requirements EXP-US-05 gives rise to. */
        end #original ::> EXP_US_05_RepeatableRuns;
        end #derive   ::> EXP_SR_04_BrowseTheSystemsModel;
        end #derive   ::> EXP_SR_08_AskTwice;
        end #derive   ::> EXP_SR_09_OneResultsFile;
        end #derive   ::> EXP_SR_15_ReportFromAFile;
        end #derive   ::> EXP_SR_16_TheModelAndTheTestsAgree;
    }

    #derivation connection us06Derives {
        doc /* The requirements EXP-US-06 gives rise to. */
        end #original ::> EXP_US_06_ExplainTheOutage;
        end #derive   ::> EXP_SR_04_BrowseTheSystemsModel;
        end #derive   ::> EXP_SR_17_ToolsOverTheSystemsModel;
        end #derive   ::> EXP_SR_18_ToolsOverTheCode;
        end #derive   ::> EXP_SR_19_TheViewAsWorkingState;
        end #derive   ::> EXP_SR_20_ExplainAnIncident;
        end #derive   ::> EXP_SR_21_SuspectedMismatches;
        end #derive   ::> EXP_SR_23_CitationsChecked;
    }
}
