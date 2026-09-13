package Federation_StakeholderStories {
    doc /* The stakeholder story register. US-01 to US-12 are the storyboard
           stories, one per page of docs/stories/use-cases.pdf, and US-13 to
           US-19 were added with the model to give a story to the requirements
           that traced to a stated obligation rather than to a use case. Each
           story names the whole demo as its subject, one role, one concern it
           frames, and one nested acceptance criterion per separately
           verifiable outcome. */
    private import Federation_Stakeholders::*;
    private import Federation_Concerns::*;
    private import Federation_Context::Demo;
    private import VSE_Library::*;
    private import ScalarValues::*;

    // ------------------------------------------- the storyboard, US-01 to US-12 --

    requirement <'US-01'> US_01_LaunchWithOneCommand : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Visitor;
        attribute :>> capability = "start the whole demo with one command copied from the README";
        attribute :>> benefit    = "I can see it working before deciding whether to read on";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement bothAppsRender {
                doc /* Given the image is already pulled, when the one docker run
                       line from the README is executed, then within ten seconds
                       http://localhost:8080/viewer renders the model viewer and
                       http://localhost:8080/document the requirements document. */
            }
            requirement noRouteToTheInternet {
                doc /* Given the container is running with no route to the
                       internet, when either app is opened, then it renders fully
                       and every request it makes goes to the one published
                       port. */
            }
        }
        doc /* Storyboard page 2 of docs/stories/use-cases.pdf,
               docs/img/us01-launch.png. */
    }

    requirement <'US-02'> US_02_ReadTheModel : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Visitor;
        attribute :>> capability = "see the pipeline model as SysML v2 text beside a sketch of the wiring";
        attribute :>> benefit    = "I understand what is being modelled without knowing the language";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement textPaneShowsTheModelFile {
                doc /* Given the shipped model, when the viewer opens, then the
                       text pane shows the model file with the servers, their
                       throughput values, the connections, the requirements and
                       their short names. */
            }
            requirement sketchShowsTheWiring {
                doc /* Given the shipped model, when the viewer opens, then the
                       sketch shows the five servers left to right as wired, each
                       with its throughput, the rolled-up capacity of 1200, and
                       parse marked as the bottleneck. */
            }
            requirement requirementShowsItsVerdict {
                doc /* Given the shipped model, when the viewer opens, then
                       PIPE-R1 shows its limit of 1500, a verdict of FAIL and a
                       reason naming parse at 1200. */
            }
        }
        doc /* Storyboard page 3 of docs/stories/use-cases.pdf. */
    }

    requirement <'US-03'> US_03_RaiseAServerThatIsNotTheBottleneck : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Visitor;
        attribute :>> capability = "when I raise the throughput of a server that is not the bottleneck, see that the capacity, the verdict and the bottleneck stay as they were";
        attribute :>> benefit    = "I learn that a chain is governed by its worst link";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement nothingMoves {
                doc /* Given the shipped model, when ingest is raised from 2000 to
                       3000, then the capacity stays 1200, PIPE-R1 stays FAIL, the
                       bottleneck stays parse, and the only other visible change is
                       PIPE-R1.1 continuing to pass with its new value. */
            }
            requirement badValueRejected {
                doc /* Given the shipped model, when a throughput is set to a value
                       that is not a number or is negative, then the value is
                       rejected and the previous value stands. */
            }
            requirement zeroStopsTheChain {
                doc /* Given the shipped model, when parse is set to 0, then the
                       capacity becomes 0, PIPE-R1 stays FAIL and the reason names
                       parse at 0. */
            }
        }
        doc /* Storyboard page 4 of docs/stories/use-cases.pdf,
               docs/img/us03-nothing-moves.png. What does not happen: no verdict
               flips, no colour changes, the sketch does not move its bottleneck
               mark. The story uses ingest because raising indexA or indexB would
               leave the capacity alone but could flip that server's own derived
               requirement, which is US-10's lesson rather than this one. */
    }

    requirement <'US-04'> US_04_RaiseTheBottleneck : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Visitor;
        attribute :>> capability = "when I raise the throughput of the bottleneck, see the capacity rise and the bottleneck move";
        attribute :>> benefit    = "I see the model respond as a whole";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement capacityRises {
                doc /* Given the shipped model, when parse is raised from 1200 to 1700,
                       then the capacity becomes 1400, PIPE-R1 stays FAIL, and the
                       bottleneck moves to the pair indexA and indexB. */
            }
            requirement requirementPasses {
                doc /* Given parse at 1700, when indexA is raised from 700 to 900, then
                       the capacity becomes 1600, PIPE-R1 becomes PASS with a reason
                       naming the index pair at 1600 against a limit of 1500, and the
                       requirement block is no longer red. */
            }
        }
        doc /* Storyboard page 5 of docs/stories/use-cases.pdf, docs/img/us04-bottleneck-moves.png. */
    }

    requirement <'US-05'> US_05_TightenTheLimit : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : ModelOwner;
        attribute :>> capability = "change the limit in the throughput requirement and see which servers are responsible when it fails";
        attribute :>> benefit    = "the requirement is checked against the model rather than against a copy";
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement loweredLimitPasses {
                doc /* Given the shipped model, when the limit of PIPE-R1 is changed
                       from 1500 to 1000, then the verdict becomes PASS. */
            }
            requirement raisedLimitFails {
                doc /* Given the shipped model, when the limit of PIPE-R1 is changed
                       from 1500 to 2500, then the verdict becomes FAIL, the
                       requirement block turns red, and the reason names parse as the
                       bottleneck at 1200. */
            }
            requirement editLandsInTheText {
                doc /* Given any value edited in the viewer, when the model text is
                       read, then the edited number appears in the text exactly where
                       the original literal was. */
            }
        }
        doc /* Storyboard page 6 of docs/stories/use-cases.pdf. */
    }

    requirement <'US-06'> US_06_ReadTheDocument : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : DocumentOwner;
        attribute :>> capability = "read the requirements as a numbered document that also shows what the model knows about each one";
        attribute :>> benefit    = "I can review them without opening a modelling tool";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement documentNumbersItsOwnSections {
                doc /* Given the shipped model and the shipped document structure,
                       when the document opens, then it shows PIPE-R1 as section 1
                       with its derived requirements as 1.1 to 1.5 and PIPE-R2 as
                       section 2, numbers that are the document's own and not the
                       model's short names. */
            }
            requirement eachRowCarriesWhatTheModelKnows {
                doc /* Given the document is open, when a requirement is read, then
                       it shows its short name, text, limit, the requirement it is
                       derived from or the requirements derived from it, the part
                       that satisfies it, the verification case that verifies it,
                       the current value of the subject it is checked against, and
                       its verdict with a reason. */
            }
            requirement unrunCaseIsInconclusive {
                doc /* Given the document is open, when PIPE-R2 is read, then its
                       verdict is INCONCLUSIVE with the reason "PIPE-VC1 is declared
                       and no service runs it". */
            }
        }
        doc /* Storyboard page 7 of docs/stories/use-cases.pdf,
               docs/img/us06-document.png. */
    }

    requirement <'US-07'> US_07_ReorderAndNest : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : DocumentOwner;
        attribute :>> capability = "move requirements into a different order and nest one under another";
        attribute :>> benefit    = "the document reads the way my readers expect while the model stays untouched";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement reorderRenumbers {
                doc /* Given the shipped document, when PIPE-R1.5 is moved above
                       PIPE-R1.1, then PIPE-R1.5 becomes 1.1, the others shift to
                       1.2 to 1.5 in their former order, and nothing else changes
                       number. */
            }
            requirement nestingKeepsTheRelations {
                doc /* Given the shipped document, when PIPE-R2 is moved under
                       PIPE-R1, then it is numbered 1.6, and its relationships in
                       the model are unchanged and still shown. */
            }
            requirement theModelIsUntouched {
                doc /* Given a reordered document, when the model viewer is read,
                       then the model text and the sketch are unchanged. */
            }
        }
        doc /* Storyboard page 8 of docs/stories/use-cases.pdf. */
    }

    requirement <'US-08'> US_08_ShapeTheDocument : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : DocumentOwner;
        attribute :>> capability = "add headings and prose and hide a requirement";
        attribute :>> benefit    = "the document carries the editorial decisions the model does not contain";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement headingTakesANumber {
                doc /* Given the shipped document, when a heading "Performance" is
                       inserted as the parent of PIPE-R1, then the heading takes
                       number 1, PIPE-R1 becomes 1.1 and its derived requirements
                       1.1.1 to 1.1.5. */
            }
            requirement proseCarriesNoNumber {
                doc /* Given the document is open, when a prose paragraph is added
                       under the heading, then it appears in place and carries no
                       number. */
            }
            requirement exclusionHidesWithoutRemoving {
                doc /* Given the shipped document, when PIPE-R1.4 is excluded, then
                       it disappears from the document, PIPE-R1.5 becomes 1.4, and
                       the model still lists PIPE-R1.4. */
            }
            requirement restoringReturnsIt {
                doc /* Given PIPE-R1.4 is excluded, when it is restored, then it
                       returns as the last child of PIPE-R1 and is numbered 1.5. */
            }
        }
        doc /* Storyboard page 9 of docs/stories/use-cases.pdf. */
    }

    requirement <'US-09'> US_09_ChangeAValueFromTheDocument : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : DocumentOwner;
        attribute :>> capability = "correct a number inline in the document";
        attribute :>> benefit    = "the model is updated from where I work";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement editReachesTheViewer {
                doc /* Given the viewer is open in another tab, when the throughput
                       of parse is changed from 1200 to 1700 in the row of PIPE-R1.2,
                       then PIPE-R1.2 becomes PASS, PIPE-R1 stays FAIL with a reason
                       now naming indexA and indexB at 1400, and within two seconds
                       the viewer's text shows 1700, its sketch shows capacity 1400
                       with the index pair marked, all without a reload. */
            }
            requirement limitEditReachesTheViewer {
                doc /* Given the document is open, when the limit of PIPE-R1 is
                       changed in its row, then the viewer shows the new limit in
                       the requirement text. */
            }
        }
        doc /* Storyboard page 10 of docs/stories/use-cases.pdf,
               docs/img/us09-edit-from-document.png. */
    }

    requirement <'US-10'> US_10_ChangeFromTheViewerWatchTheDocument : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Visitor;
        attribute :>> capability = "when I edit a value in the viewer with the document open beside it, the document to update by itself";
        attribute :>> benefit    = "I see that both apps read the same data rather than each other";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement documentFollowsTheViewer {
                doc /* Given both apps are open and parse is at 1700, when indexA is
                       raised from 700 to 900 in the viewer, then within two seconds
                       and without a reload the document shows PIPE-R1 as PASS with a
                       reason naming the index pair at 1600 against 1500, PIPE-R1.3
                       as PASS, and PIPE-R1.4 still FAIL with a reason giving 700
                       against its limit of 750. */
            }
        }
        doc /* Storyboard page 11 of docs/stories/use-cases.pdf,
               docs/img/us10-viewer-to-document.png. */
    }

    requirement <'US-11'> US_11_QueryTheGraph : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Visitor;
        attribute :>> capability = "run one query that returns fields from all three services for one requirement";
        attribute :>> benefit    = "I see the federation rather than take it on trust";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement oneResponseCarriesAllThree {
                doc /* Given the demo is running, when one query in the playground
                       asks for PIPE-R1's text, verdict and document number, then one
                       response carries all three, and the served schema contains
                       types contributed by all three subgraphs. */
            }
        }
        doc /* Storyboard page 12 of docs/stories/use-cases.pdf,
               docs/img/us11-query.png. */
    }

    requirement <'US-12'> US_12_Reset : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Visitor;
        attribute :>> capability = "return the demo to its shipped state";
        attribute :>> benefit    = "I can show it again from the beginning";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement bothAppsReturnToTheShippedState {
                doc /* Given values and the document structure have been edited, when
                       the reset control is used in either app, then both apps show
                       the shipped values and the shipped document structure within
                       two seconds. */
            }
        }
        doc /* Storyboard page 13 of docs/stories/use-cases.pdf,
               docs/img/us12-reset.png. */
    }

    // ----------------------------- added with the model, US-13 to US-19 --

    requirement <'US-13'> US_13_LicencesTravelWithWhatIsCopied : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Maintainer;
        attribute :>> capability = "every third-party binary and file the demo ships or vendors to carry its licence text beside it";
        attribute :>> benefit    = "the demo can be published and copied without an obligation being missed";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement licenceInTheImage {
                doc /* Given the published image, when its contents are inspected,
                       then the router's Apache-2.0 text sits beside /router. */
            }
            requirement licencesInTheRepository {
                doc /* Given the repository, when its root is read, then a NOTICE
                       names the router and SortableJS and the MIT text sits beside
                       the vendored file. */
            }
        }
        doc /* Added with the model, to give a story to a requirement that traced
               to a stated obligation rather than to a use case. */
    }

    requirement <'US-14'> US_14_NothingOfTheExampleInTheAdapter : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : AdoptingOrganisation;
        attribute :>> capability = "the adapter to contain nothing specific to the example";
        attribute :>> benefit    = "I can judge what I would keep by reading the adapter alone";
        frame concern : AgreeAndReplace;
        requirement :>> acceptance {
            requirement noExampleNames {
                doc /* Given the adapter source, when it is searched for the
                       example's names, then none is found, and a second fixture with
                       other names projects every field. */
            }
        }
        doc /* Added with the model, to give a story to a requirement that traced
               to a stated obligation rather than to a use case. */
    }

    requirement <'US-15'> US_15_AServedModelThatIsNeverSilentlyWrong : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : ModelOwner;
        attribute :>> capability = "the adapter to refuse a model it cannot serve faithfully, naming the place, and to keep a value bound by an expression out of reach of an edit";
        attribute :>> benefit    = "what is served is my model and not a copy that drifted from it";
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement refusalNamesThePlace {
                doc /* Given a model with a construct outside the supported subset,
                       when the adapter starts, then it refuses with the file, line
                       and column of the first such construct. */
            }
            requirement boundValuesStayBound {
                doc /* Given a value bound by an expression, when an edit targets it,
                       then the edit is refused and the value is unchanged. */
            }
        }
        doc /* Added with the model, to give a story to a requirement that traced
               to a stated obligation rather than to a use case. */
    }

    requirement <'US-16'> US_16_TheContractCheckedBeforeDeployment : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : AdoptingOrganisation;
        attribute :>> capability = "the committed router configuration checked against the subgraph schemas in the repository";
        attribute :>> benefit    = "the contract between producer and consumer is checked mechanically before deployment rather than discovered in production";
        frame concern : AgreeAndReplace;
        requirement :>> acceptance {
            requirement driftFailsTheTest {
                doc /* Given the committed configuration and the schema files, when
                       the test runs, then it fails on any difference between
                       them. */
            }
        }
        doc /* Added with the model, to give a story to a requirement that traced
               to a stated obligation rather than to a use case. */
    }

    requirement <'US-17'> US_17_TheExampleAcceptedByTheReferenceTools : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : ModelOwner;
        attribute :>> capability = "the example model accepted without error by the reference implementations";
        attribute :>> benefit    = "a SysML-literate reader can load the file in a real tool and trust what the demo says about it";
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement bothToolsAccept {
                doc /* Given the example file, when the OMG pilot and OpenSysML read
                       it, then neither reports an error or a warning. */
            }
        }
        doc /* Added with the model, to give a story to a requirement that traced
               to a stated obligation rather than to a use case. */
    }

    requirement <'US-18'> US_18_TheDemosOwnModelAsTheRecord : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Maintainer;
        attribute :>> capability = "the demo described by a SysML v2 model of its own, validated by the reference tools and checked against the documents and the tests";
        attribute :>> benefit    = "the record cannot drift from what ships without a check failing";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement validatedOnEveryChange {
                doc /* Given a pull request that changes the model, when it is
                       opened, then both reference tools accept the model or the
                       check fails. */
            }
            requirement agreementIsTested {
                doc /* Given the model and the repository, when a unit test runs,
                       then it fails if they disagree on an identifier, a test name,
                       a check file or a published image. */
            }
        }
        doc /* Added with the model, to give a story to a requirement that traced
               to a stated obligation rather than to a use case. */
    }

    requirement <'US-19'> US_19_TheStoriesExercisedAgainstARunningInstance : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : Demo;
        stakeholder :>> role : Maintainer;
        attribute :>> capability = "every story's acceptance exercised against a running instance by a monitoring service, with the trace of each request kept";
        attribute :>> benefit    = "the demo's claims are checked whenever it runs with credentials, and not once";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement withCredentials {
                doc /* Given credentials for the monitoring service, when the demo is
                       started with them, then the suite runs against it, records a
                       session and deploys for the duration. */
            }
            requirement withoutCredentials {
                doc /* Given no credentials, when the demo is started, then it runs
                       exactly as docker run runs it. */
            }
        }
        doc /* Added with the model, to give a story to a requirement that traced
               to a stated obligation rather than to a use case. */
    }
}
