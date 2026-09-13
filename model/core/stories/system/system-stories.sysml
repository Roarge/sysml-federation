package Federation_SystemStories {
    doc /* The system story register. Each story restates one requirement of the
           light scheme (SR-nn) as a requirement usage typed UserStory: the EARS
           statement it was written as, the pattern that statement follows, the
           methods that verify it, the capability and benefit in the system's own
           words, the concerns its upstream stories frame, and one acceptance
           criterion per separately verifiable outcome. The derivation
           connections at the end of the file record which stakeholder story each
           system story comes from, one connection per stakeholder story. */
    private import Federation_StakeholderStories::*;
    private import Federation_Stakeholders::*;
    private import Federation_Concerns::*;
    private import Federation_Context::DemoElement;
    private import Federation_Library::*;
    private import VSE_Library::*;
    private import ScalarValues::*;
    private import RequirementDerivation::*;

    // ------------------------------------------------ launch and packaging --

    #derive
    requirement <'SR-01'> SR_01_OneCommand : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The demo shall start from the single command docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation with no other file, argument, volume or network configuration.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "start the whole demo from one command, with no other file, argument, volume or network configuration";
        attribute :>> benefit    = "a curious visitor reaches the running demo before anything has a chance to stop them";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement startsFromOneCommand {
                doc /* Given a host with Docker and no authentication to the
                       registry, when the one command from the README is run,
                       then the demo starts with no other file, argument, volume
                       or network configuration. */
            }
            requirement recordedOnThreePlatforms {
                doc /* Given the same run on Linux, macOS and Windows, when each
                       is recorded in the example README, then the record names
                       the hosts used. */
            }
        }
        doc /* Rationale: the audience has Docker and fifteen minutes, and a
               clone, a compose file or a second command is where a curious
               visitor stops. The package is made public by hand before the
               command works for anyone (AD-0011, AD-0020). The criterion
               recordedOnThreePlatforms has a record for Linux alone, as the
               release notes in article 11 say. Allocated to: IMAGE.
               Upstream: US-01. */
    }

    #derive
    requirement <'SR-02'> SR_02_ReadyWithinTenSeconds : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "When the container starts with the image already pulled, the demo shall serve both apps and the router endpoint within ten seconds.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "serve both apps and the router endpoint within ten seconds of the container starting";
        attribute :>> benefit    = "the start never outlasts the visitor's attention";
        frame concern : LaunchAndSurprise;
        attribute readySeconds : Rational = 10.0;
        requirement readinessSla {
            doc /* Formalises the benefit: ready within ten seconds. */
            require constraint { readySeconds <= 10.0 }
        }
        requirement :>> acceptance {
            requirement readyWithinTenSeconds {
                doc /* Given the image is already pulled, when the container
                       starts, then the three subgraphs, the router and the UI
                       server are all ready within ten seconds, measured by the
                       supervisor's own startup test. */
            }
        }
        doc /* Rationale: a start that outlasts the visitor's attention reads as
               failure. The reference host is Linux with Docker Engine, and the
               slower desktop hosts are covered by the demonstration of SR-01
               (AD-0011). Allocated to: IMAGE. Upstream: US-01. */
    }

    #derive
    requirement <'SR-03'> SR_03_NoOutboundNetwork : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The demo shall make no network connection to any address outside the container, and the image shall set DO_NOT_TRACK=1, COSMO_TELEMETRY_DISABLED=true, TRACING_ENABLED=false and METRICS_OTLP_ENABLED=false in the router process's environment, and shall hand the router a configuration file only when the operator names one in SYSML_FEDERATION_ROUTER_CONFIG_PATH, in which case that file governs the router's telemetry.";
        attribute pattern : EarsPattern = EarsPattern::Complex;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Analysis, VerificationMethod::Demonstration);
        attribute :>> capability = "run with no connection to any address outside the container, the router's telemetry switched off, and a configuration file handed over only when the operator names one";
        attribute :>> benefit    = "the claim that the demo suits an air-gapped machine holds, and telemetry is something the operator asks for rather than something they inherit";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement environmentCarriesTheFourVariables {
                doc /* Given the running container, when the router process's
                       environment is read, then it carries the four variables
                       and the execution configuration names only loopback
                       addresses. */
            }
            requirement outboundPathsAnalysed {
                doc /* Given the router's code paths that could reach an address
                       outside the container, when they are analysed, then none
                       is reached with those variables set. */
            }
            requirement noConnectionAttemptWithTheRouteWithheld {
                doc /* Given the container run with the route to the internet
                       withheld and the log level set to debug, when it starts,
                       then it reaches readiness and logs no connection
                       attempt. */
            }
            requirement configurationFileIsAnOptIn {
                doc /* Given the variable is unset, when the router starts, then
                       its environment is the nine variables the supervisor sets
                       and carries no CONFIG_PATH. Given it is set, then
                       CONFIG_PATH carries that path and the nine variables are
                       still set. */
            }
        }
        doc /* Rationale: the README claims the demo suits an air-gapped machine.
               The router's usage tracker is on by default and either of the
               first two variables switches it off, and the last two are belt and
               braces beside an empty exporter token. A failure to reach the
               tracker is not fatal, so readiness alone proves nothing about
               attempts (AD-0002, AD-0010, AD-0013). Allocated to: IMAGE, ROUTER.
               Upstream: US-01. */
    }

    #derive
    requirement <'SR-04'> SR_04_FourPathsOnOnePort : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The demo shall serve the model viewer at /viewer, the requirements document at /document, the router's GraphQL endpoint at /graphql and the router's playground at /playground on the one published port, and shall redirect / to /viewer.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "serve both apps, the GraphQL endpoint and the playground on one published port, with the root redirecting to the viewer";
        attribute :>> benefit    = "one origin keeps the launch line to a single flag and the browser clear of cross-origin rules";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement fourPathsAnswer {
                doc /* Given the demo is running, when the four paths are
                       requested on the published port, then each is served by
                       the app or the endpoint it names. */
            }
            requirement rootRedirectsToTheViewer {
                doc /* Given the demo is running, when the root path is
                       requested, then it redirects to /viewer. */
            }
        }
        doc /* Rationale: two apps make the sharing of data visible, and one
               origin avoids the browser's cross-origin rules and keeps the
               launch line to one flag. The router's playground path is
               configured so the UI server proxies it without rewriting
               (AD-0011). Allocated to: IMAGE, ROUTER. Upstream: US-01, US-11. */
    }

    #derive
    requirement <'SR-05'> SR_05_TwoPlatforms : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The image shall be published for linux/amd64 and linux/arm64.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "publish the image for linux/amd64 and linux/arm64";
        attribute :>> benefit    = "Apple silicon and ARM servers run the demo as they are";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement bothPlatformsInTheManifest {
                doc /* Given a published tag, when the publish workflow reads the
                       registry manifest, then both platforms are listed and the
                       step fails if either is absent. */
            }
        }
        doc /* Rationale: Apple silicon and ARM servers are common in the
               audience, and the copy of the router binary has to pick the right
               platform (AD-0010, AD-0020). Allocated to: IMAGE. Upstream:
               US-01. */
    }

    #derive
    requirement <'SR-06'> SR_06_SizeBudget : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The compressed size of each platform's image layers shall not exceed 80 MB.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "keep each platform's compressed layers within 80 MB";
        attribute :>> benefit    = "the pull finishes while the visitor is still watching";
        frame concern : LaunchAndSurprise;
        attribute compressedMegabytes : Rational = 80.0;
        requirement sizeBudget {
            doc /* Formalises the benefit: each platform within 80 MB
                   compressed. */
            require constraint { compressedMegabytes <= 80.0 }
        }
        requirement :>> acceptance {
            requirement layersWithinTheBudget {
                doc /* Given a published tag, when the same workflow step reads
                       each platform's layer sizes, then the compressed total is
                       at most 80 MB. */
            }
        }
        doc /* Rationale: the router alone is about 40 MB compressed, which
               leaves about 40 MB for the demo (AD-0010, AD-0020). Allocated to:
               IMAGE. Upstream: US-01. */
    }

    #derive
    requirement <'SR-07'> SR_07_LicenceCarriedInTheImage : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Maintainer;
        attribute statement : String = "The image shall carry the Apache-2.0 licence text of the Cosmo router beside the copied /router binary.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Inspection);
        attribute :>> capability = "carry the router's Apache-2.0 text beside the binary copied into the image";
        attribute :>> benefit    = "the image is published and copied without a licence obligation being missed";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement licenceBesideTheBinary {
                doc /* Given the published image, when its contents are
                       inspected, then the router's Apache-2.0 text sits beside
                       the copied binary. */
            }
        }
        doc /* Rationale: the binary is copied out of the vendor's image, and the
               licence requires its text to travel with it (AD-0002, AD-0010).
               Allocated to: IMAGE. Upstream: US-13. */
    }

    #derive
    requirement <'SR-08'> SR_08_LicencesCarriedInTheRepository : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Maintainer;
        attribute statement : String = "The repository shall carry a root NOTICE naming the Cosmo router (Apache-2.0, version pinned) and SortableJS 1.15.7 (MIT), with the MIT text beside the vendored file.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Inspection);
        attribute :>> capability = "carry a root NOTICE for the router and SortableJS, with the MIT text beside the one vendored file";
        attribute :>> benefit    = "the repository is copied without a licence obligation being missed";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement noticeAndMitTextInPlace {
                doc /* Given the repository, when its root is read, then a NOTICE
                       names the router and SortableJS and the MIT text sits
                       beside the vendored file. */
            }
        }
        doc /* Rationale: SortableJS is the one vendored file (AD-0020).
               Allocated to: REPO. Upstream: US-13. */
    }

    #derive
    requirement <'SR-09'> SR_09_StateLivesInMemory : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The demo shall keep every edited value and every document change in memory only, so that a container restart returns the demo to its shipped state.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "hold every edited value and every document change in memory alone";
        attribute :>> benefit    = "a restart is the whole of the reset";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement restartReturnsTheShippedState {
                doc /* Given values and the document structure have been edited,
                       when the services are restarted, then both show the
                       shipped state. */
            }
        }
        doc /* Rationale: a demo has no users to preserve state for, and a
               restart is the cheapest reset (AD-0004). Allocated to: ADAPTER,
               DOCUMENT. Upstream: US-12. */
    }

    #derive
    requirement <'SR-10'> SR_10_NoResourceFromAnotherOrigin : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The apps shall reference no resource at any origin other than the published port.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Demonstration);
        attribute :>> capability = "reference no font, script or image at any origin other than the published port";
        attribute :>> benefit    = "the apps render fully on a machine with no route to the internet";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement noAbsoluteUrlInTheAssets {
                doc /* Given the embedded assets, when they are scanned, then
                       none carries an absolute URL to another origin. */
            }
            requirement appsRenderWithTheHostOffline {
                doc /* Given the host is offline, when either app is opened, then
                       it renders fully. */
            }
        }
        doc /* Rationale: a font, script or image fetched from elsewhere would
               break the second criterion of US-01 and the air-gap claim with no
               service at fault (AD-0011, AD-0017). The criterion
               appsRenderWithTheHostOffline has no run behind it, as the release
               notes in article 11 say. Allocated to: UI-MODEL, UI-DOC.
               Upstream: US-01. */
    }

    // -------------------------------------------------------- the viewer --

    #derive
    requirement <'SR-11'> SR_11_ModelTextWithTheLanguageVisible : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The model viewer shall render the model's SysML v2 source with keywords, names, numeric literals, strings and comments distinguished.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Inspection, VerificationMethod::Demonstration);
        attribute :>> capability = "render the model source with keywords, names, numeric literals, strings and comments distinguished";
        attribute :>> benefit    = "the visitor sees the notation rather than a form over it";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement notationIsVisible {
                doc /* Given the shipped model, when the viewer opens, then the
                       text pane distinguishes keywords, names, numeric literals,
                       strings and comments. */
            }
            requirement tokeniserMatchesTheReservedWords {
                doc /* Given the tokeniser, when it is read against the
                       language's reserved-word list, then the words it
                       distinguishes are those of the list. */
            }
        }
        doc /* Rationale: the demo is about SysML v2, and the visitor should see
               the notation. No permissively licensed renderer exists, so a small
               tokeniser is written here with the specification's reserved words
               as its source (AD-0017, AD-0026). Allocated to: UI-MODEL.
               Upstream: US-02. */
    }

    #derive
    requirement <'SR-12'> SR_12_SketchDrawnFromTheWiring : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder documentOwner : DocumentOwner;
        attribute statement : String = "The model viewer shall draw the pipeline from the model's connections as a left-to-right graph, and shall show on it each server's throughput, the pipeline's capacity and the bottleneck servers marked in red, all as read through the router.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "draw the wiring from the model's connections, with each throughput, the rolled-up capacity and the bottleneck marked on it";
        attribute :>> benefit    = "the bottleneck moving is seen rather than read out of numbers";
        frame concern : LaunchAndSurprise;
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement sketchShowsTheWiring {
                doc /* Given the shipped model, when the viewer opens, then the
                       sketch shows the servers left to right as wired, each with
                       its throughput, the pipeline's capacity, and the
                       bottleneck servers marked in red. */
            }
            requirement secondWiringDrawsTheSameWay {
                doc /* Given a second wiring with fan-in from different points,
                       when the viewer opens, then the sketch is drawn from that
                       model's connections in the same way. */
            }
        }
        doc /* Rationale: the bottleneck moving is the memorable moment, and it
               has to be seen without reading numbers (AD-0017, AD-0026).
               Allocated to: UI-MODEL. Upstream: US-02, US-03, US-04, US-09. */
    }

    #derive
    requirement <'SR-13'> SR_13_TheViewersEditableSet : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder modelOwner : ModelOwner;
        attribute statement : String = "The model viewer shall present as editable exactly the throughput of each server and the limit of the global throughput requirement, and shall present every other value as read-only.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "offer an editable control for each server's throughput and for the limit of the global throughput requirement, and for nothing else";
        attribute :>> benefit    = "the demo stays explainable and an allocated limit cannot be edited out of its allocation";
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement editableSetIsExact {
                doc /* Given the viewer is open, when its controls are read, then
                       every editable control maps to a server's throughput or to
                       the limit of the global throughput requirement, and no
                       other value has one. */
            }
        }
        doc /* Rationale: the editable set is fixed so that the demo stays
               explainable, and the derived limits are bound by expressions
               (AD-0004, AD-0026). Allocated to: UI-MODEL. Upstream: US-03,
               US-04, US-05. */
    }

    #derive
    requirement <'SR-14'> SR_14_VerdictAndReasonInTheViewer : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder modelOwner : ModelOwner;
        attribute statement : String = "The model viewer shall show for each requirement its verdict and its reason as read through the router.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "show each requirement's verdict and reason as read through the router";
        attribute :>> benefit    = "the visitor sees what the analysis concluded and the reason it gives";
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement verdictAndReasonShown {
                doc /* Given each of the three states of the worked example, when
                       the viewer is open, then every requirement shows its
                       verdict and its reason as read through the router. */
            }
        }
        doc /* Rationale: US-02, US-03, US-04 and US-05 observe the verdict and
               the reason in the viewer, beside the red block SR-15 governs
               (AD-0026). Allocated to: UI-MODEL. Upstream: US-02, US-03, US-04,
               US-05. */
    }

    #derive
    requirement <'SR-15'> SR_15_FailingRequirementInRed : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder modelOwner : ModelOwner;
        attribute statement : String = "While a requirement's verdict is FAIL, the model viewer shall show that requirement's block in red, and while it is anything else the viewer shall show the block without red.";
        attribute pattern : EarsPattern = EarsPattern::StateDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "show a failing requirement's block in red and every other block without it";
        attribute :>> benefit    = "the one colour accent in the viewer carries the whole message";
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement failingBlockIsRed {
                doc /* Given a requirement whose verdict is FAIL, when the viewer
                       shows it, then its block is red. */
            }
            requirement otherVerdictsCarryNoRed {
                doc /* Given a requirement whose verdict is anything else, when
                       the viewer shows it, then its block carries no red. */
            }
        }
        doc /* Rationale: red is the only colour accent in the viewer and it
               carries the whole message, and the two-second bound belongs to
               SR-39 (AD-0026). Allocated to: UI-MODEL. Upstream: US-04,
               US-05. */
    }

    // ------------------------------------------------------- the adapter --

    #derive
    requirement <'SR-16'> SR_16_GenericProjection : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder documentOwner : DocumentOwner;
        attribute statement : String = "The adapter shall project from a model written in the supported subset every part with its attributes, ports, child parts and connections, every requirement with its text, subject, constrained quantity, comparison, limit and the limit's unit as written in the source, every derivation, satisfaction and verification relationship, every verification case, and the model's text and version.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "project every part, requirement, relationship and verification case of a model in the supported subset, together with the model's text and version";
        attribute :>> benefit    = "every other service reads what it needs from one projection and never from the source";
        frame concern : LaunchAndSurprise;
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement everyFieldProjected {
                doc /* Given a second fixture model with different names and
                       different wiring, when it is projected, then every field
                       the statement lists is present. */
            }
        }
        doc /* Rationale: the projection is the whole of what the other services
               see, the document reads relationships from it and the capacity
               service reads quantities from it (AD-0005, AD-0015). Allocated to:
               ADAPTER. Upstream: US-02, US-06. */
    }

    #derive
    requirement <'SR-17'> SR_17_NoExampleIdentifiersInTheAdapter : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : AdoptingOrganisation;
        attribute statement : String = "The adapter shall contain no identifier from the example model.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Inspection);
        attribute :>> capability = "contain no identifier of the example model";
        attribute :>> benefit    = "an adopting reader judges what they would keep by reading the adapter alone";
        frame concern : AgreeAndReplace;
        requirement :>> acceptance {
            requirement noExampleNameInTheSource {
                doc /* Given the adapter source, when it is searched for the
                       example's names, then none is found. */
            }
            requirement secondFixtureProjectsWithoutThem {
                doc /* Given the second fixture of SR-16, when it is projected,
                       then every field is projected and no name of the example
                       is involved. */
            }
        }
        doc /* Rationale: the README's claim is that nothing in the adapter is
               specific to one example, so the example's vocabulary stays in the
               example and the two apps (AD-0005). Allocated to: ADAPTER.
               Upstream: US-14. */
    }

    #derive
    requirement <'SR-18'> SR_18_RefusalOfUnsupportedSyntax : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : ModelOwner;
        attribute statement : String = "If the model source contains a construct outside the supported subset, then the adapter shall refuse to start and shall report the file, the line and the column of the first such construct.";
        attribute pattern : EarsPattern = EarsPattern::UnwantedBehaviour;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "refuse to start on a construct outside the supported subset, naming the file, line and column of the first one";
        attribute :>> benefit    = "a model that cannot be served faithfully is never served silently wrong";
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement refusalNamesThePlace {
                doc /* Given a model with a construct outside the supported
                       subset, when the adapter starts, then it refuses and
                       reports the file, the line and the column of the first
                       such construct. */
            }
            requirement oneFixturePerRejection {
                doc /* Given one fixture per rejection the parser distinguishes,
                       when each is read, then each is refused with its own
                       place. */
            }
        }
        doc /* Rationale: a subset parser that skipped what it did not understand
               would serve a silently wrong model, and failing loudly is the only
               honest behaviour (AD-0015). Allocated to: ADAPTER. Upstream:
               US-15. */
    }

    #derive
    requirement <'SR-19'> SR_19_QuantityComparisonAndLimitFromTheConstraint : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        attribute statement : String = "When a requirement's require constraint compares a feature chain rooted at the requirement's subject with an attribute of the requirement or a literal, the adapter shall project the chain's last segment as the constrained quantity, the operator as the comparison and the other operand's value as the limit, and if a requirement's constraint has any other shape, then the adapter shall refuse to start as in SR-18.";
        attribute pattern : EarsPattern = EarsPattern::Complex;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "read a requirement's constrained quantity, comparison and limit out of its constraint, and refuse a constraint of any other shape";
        attribute :>> benefit    = "one rule tells a throughput requirement from a latency one without either being declared twice";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement quantityComparisonAndLimitRead {
                doc /* Given the example's six throughput requirements and its
                       latency requirement, when they are projected, then each
                       carries the chain's last segment as its quantity, the
                       operator as its comparison and the other operand's value
                       as its limit. */
            }
            requirement anyOtherShapeIsRefused {
                doc /* Given a fixture whose constraint has any other shape, when
                       the adapter starts, then it refuses to start as in
                       SR-18. */
            }
        }
        doc /* Rationale: a generic adapter has no other way to know which
               quantity a requirement constrains or which of its attributes is
               the limit, and this one rule is what lets the capacity service
               tell a throughput requirement from a latency requirement
               (AD-0008). Allocated to: ADAPTER. Upstream: US-06. */
    }

    #derive
    requirement <'SR-20'> SR_20_ConnectionDirection : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The adapter shall project each connect statement as a connection from its first end to its second, and if the first end is not an out port or the second is not an in port, then the adapter shall refuse to start as in SR-18.";
        attribute pattern : EarsPattern = EarsPattern::Complex;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "take a connection's direction from the order of its ends, and refuse a wiring whose ports face the wrong way";
        attribute :>> benefit    = "a wiring error is a model error rather than a capacity that is quietly wrong";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement directionFollowsTheEnds {
                doc /* Given the example wiring, when it is projected, then each
                       connection runs from its first end to its second. */
            }
            requirement reversedPortsAreRefused {
                doc /* Given a fixture whose first end is not an out port or
                       whose second is not an in port, when the adapter starts,
                       then it refuses to start as in SR-18. */
            }
        }
        doc /* Rationale: a connection in the language has ordered ends and no
               direction of its own, so the rollup's direction has to come from
               something stated, and the check makes a wiring error a model error
               (AD-0009). Allocated to: ADAPTER. Upstream: US-02. */
    }

    #derive
    requirement <'SR-21'> SR_21_ShortNamesAsIdentifiers : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The adapter shall identify every projected element by its declared short name, and by its qualified name where none is declared.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "identify every projected element by its declared short name, or by its qualified name where none is declared";
        attribute :>> benefit    = "the three services join on a stable identifier the model's author controls";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement shortNameIsTheIdentifier {
                doc /* Given the example's elements, when they are projected,
                       then each is identified by its declared short name. */
            }
            requirement qualifiedNameWhereNoneIsDeclared {
                doc /* Given a fixture element with no short name, when it is
                       projected, then it is identified by its qualified
                       name. */
            }
        }
        doc /* Rationale: the short name is the only author-controlled stable
               identifier in the language, and it is the entity key every other
               service joins on (AD-0005, AD-0018). Allocated to: ADAPTER.
               Upstream: US-02, US-11. */
    }

    #derive
    requirement <'SR-22'> SR_22_EditsPatchTheSource : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder modelOwner : ModelOwner;
        stakeholder documentOwner : DocumentOwner;
        attribute statement : String = "When a mutation sets a value that is a literal in the model source, the adapter shall replace that literal at its recorded source span in the served text, update the projection and increment the model version, and no query shall observe the served text and the projection disagreeing.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "replace an edited literal at its source span, rebuild the projection and increment the version in one step";
        attribute :>> benefit    = "the text the viewer shows and the projection the document reads never disagree";
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement literalAndProjectionAgree {
                doc /* After each mutation the literal in the served text equals
                       the projected value and the version has grown by one. */
            }
            requirement agreeUnderConcurrentReads {
                doc /* Concurrent reads during a patch observe the text and the
                       projection agreeing. */
            }
        }
        doc /* Rationale: the viewer shows the text and the document shows the
               projection, and the version stands in for a model repository
               (AD-0003, AD-0004, AD-0015). Allocated to: ADAPTER. Upstream:
               US-03, US-04, US-05, US-09. */
    }

    #derive
    requirement <'SR-23'> SR_23_ExpressionBoundValuesAreEvaluated : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        stakeholder visitor : Visitor;
        attribute statement : String = "The adapter shall evaluate an attribute value bound by an expression made of numeric literals with an optional unit, feature reference chains, the operators plus, minus, times and divide, and parentheses.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "evaluate a value bound by an expression over literals, feature references, the four arithmetic operators and parentheses";
        attribute :>> benefit    = "an allocated limit follows the limit it is derived from, as the model says it does";
        frame concern : ShapeWithoutTouching;
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement derivedLimitsFollowTheGlobalLimit {
                doc /* Given the shipped model, when the global limit changes,
                       then every limit bound by an expression over it changes
                       with it. */
            }
            requirement everyOperatorEvaluates {
                doc /* Given a fixture covering each operator and a parenthesised
                       expression, when each value is read, then it is the value
                       the expression gives. */
            }
        }
        doc /* Rationale: the derived limits follow the global limit by
               expression in the model, which is how allocation is written in the
               language (AD-0015). Allocated to: ADAPTER. Upstream: US-06, US-09,
               US-10. */
    }

    #derive
    requirement <'SR-24'> SR_24_ExpressionBoundValuesAreReadOnly : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : ModelOwner;
        attribute statement : String = "If a mutation targets a value bound by an expression, then the adapter shall reject the mutation and leave the value unchanged.";
        attribute pattern : EarsPattern = EarsPattern::UnwantedBehaviour;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "reject a mutation that targets a value bound by an expression";
        attribute :>> benefit    = "an allocation cannot be broken by an edit";
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement boundValueIsNotEdited {
                doc /* Given a derived limit bound by an expression, when it is
                       set through the playground, then the mutation is refused
                       and the value is unchanged. */
            }
        }
        doc /* Rationale: editing an allocated limit would break the allocation,
               and the editable set excludes it (AD-0004). Allocated to: ADAPTER.
               Upstream: US-15. */
    }

    #derive
    requirement <'SR-25'> SR_25_RejectionOfInvalidValues : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "If a submitted value is not a finite, non-negative number, then the adapter shall reject the mutation and leave the previous value in place.";
        attribute pattern : EarsPattern = EarsPattern::UnwantedBehaviour;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "reject a value that is not a finite, non-negative number and leave the previous one in place";
        attribute :>> benefit    = "the first nonsense a visitor types does not stop the demo";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement invalidValuesAreRejected {
                doc /* Given an empty, textual, negative, infinite or
                       not-a-number input, when it is submitted, then the
                       mutation is rejected and the previous value stands. */
            }
        }
        doc /* Rationale: the first thing a visitor types after the numbers is
               nonsense, and the demo has to survive it, with the apps restoring
               the field from the served value (AD-0004). Allocated to: ADAPTER,
               UI-MODEL, UI-DOC. Upstream: US-03. */
    }

    #derive
    requirement <'SR-26'> SR_26_ModelVersionEvents : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        stakeholder visitor : Visitor;
        attribute statement : String = "When a mutation is accepted, the adapter shall emit the new model version on its subscription.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "emit the new model version on a subscription whenever a mutation is accepted";
        attribute :>> benefit    = "each app learns of a change without polling and without a reload";
        frame concern : ShapeWithoutTouching;
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement subscriberReceivesTheVersion {
                doc /* Given a subscriber on the adapter's subscription, when a
                       mutation is accepted, then the subscriber receives the new
                       model version. */
            }
        }
        doc /* Rationale: live push through the router is built from version
               events and client refetches (AD-0003, AD-0014). Allocated to:
               ADAPTER. Upstream: US-09, US-10. */
    }

    #derive
    requirement <'SR-27'> SR_27_DocumentVersionEvents : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        stakeholder visitor : Visitor;
        attribute statement : String = "When a document change is accepted, the document service shall emit the new document version on its subscription.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "emit the new document version on a subscription whenever a document change is accepted";
        attribute :>> benefit    = "each app learns of a document change without polling and without a reload";
        frame concern : ShapeWithoutTouching;
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement subscriberReceivesTheDocumentVersion {
                doc /* Given a subscriber on the document service's
                       subscription, when a document change is accepted, then the
                       subscriber receives the new document version. */
            }
        }
        doc /* Rationale: as SR-26, for the document (AD-0014, AD-0025).
               Allocated to: DOCUMENT. Upstream: US-09, US-10. */
    }

    // ----------------------------------------------- the capacity service --

    #derive
    requirement <'SR-28'> SR_28_RollupByMaximumFlow : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder modelOwner : ModelOwner;
        stakeholder documentOwner : DocumentOwner;
        attribute statement : String = "The capacity service shall compute the capacity of a part with child parts as the maximum flow through the children's wiring with each child's configured attribute as that child's capacity, and shall compute the capacity of a part without children as that part's own configured attribute.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "compute a part's capacity as the maximum flow through its children's wiring, and a childless part's as its own configured attribute";
        attribute :>> benefit    = "a chain and its parallel branches are answered by one rule that stays defined where the simple rules are not";
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement tableCasesForEachWiring {
                doc /* Given table cases for a chain, fan-out and fan-in, nested
                       branches, a cycle, a disconnected child, and zero and
                       missing values, when the capacity is computed, then each
                       is the maximum flow through that wiring. */
            }
            requirement aLeafReadsItsOwnAttribute {
                doc /* Given a part with no children, when its capacity is
                       computed, then it is that part's own configured
                       attribute. */
            }
            requirement differentialAgainstMinimumAndSum {
                doc /* Given series-parallel wirings, when the rollup is compared
                       with the recursive minimum and sum, then the two
                       agree. */
            }
        }
        doc /* Rationale: the README states that a chain's capacity is the
               smallest member and that parallel branches sum, and maximum flow
               computes both for series-parallel wiring and stays defined where
               those rules are not. The leaf rule is what lets the derived
               requirements on single servers use the same verdict rule
               (AD-0006, AD-0007, AD-0009). Allocated to: CAPACITY. Upstream:
               US-02, US-03, US-04, US-05, US-09, US-10. */
    }

    #derive
    requirement <'SR-29'> SR_29_BottleneckAsTheMinimumCut : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder modelOwner : ModelOwner;
        stakeholder documentOwner : DocumentOwner;
        attribute statement : String = "The capacity service shall report as bottleneck the child parts of the source-side minimum cut.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "report the child parts of the source-side minimum cut as the bottleneck";
        attribute :>> benefit    = "the reported set is the same for every maximum flow, ties included";
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement cutIsTheBottleneck {
                doc /* Given the three states of the worked example, when the
                       bottleneck is reported, then it is the child parts of the
                       source-side minimum cut. */
            }
            requirement aTieHasOneAnswer {
                doc /* Given a wiring in which several minimum cuts exist, when
                       the bottleneck is reported, then the source-side cut is
                       reported and is the same for every maximum flow. */
            }
        }
        doc /* Rationale: the reported set has to be defined even when several
               minimum cuts exist, and the source-side cut is the same for every
               maximum flow. Raising a child outside the cut leaves the capacity
               unchanged, and raising one inside it raises the capacity until
               another cut becomes minimal (AD-0007). Allocated to: CAPACITY.
               Upstream: US-02, US-03, US-04, US-05, US-09, US-10. */
    }

    #derive
    requirement <'SR-30'> SR_30_VerdictsAndReasons : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        stakeholder modelOwner : ModelOwner;
        stakeholder documentOwner : DocumentOwner;
        attribute statement : String = "The capacity service shall return for each requirement whose constrained quantity is the quantity it is configured to compute a verdict of PASS when the subject's capacity satisfies the projected comparison against the limit and FAIL otherwise, shall return INCONCLUSIVE for a requirement whose constrained quantity is any other, for a subject with neither children nor the configured attribute, and for a wiring with no entry or no exit part, shall return ERROR when a child's attribute is missing or negative, with the other-quantity INCONCLUSIVE taking precedence over ERROR, ERROR over the remaining INCONCLUSIVE cases, and those over PASS and FAIL, and shall attach a reason built from the templates on the capacity model page.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "return PASS, FAIL, INCONCLUSIVE or ERROR for each requirement by a stated precedence, with a reason built from a fixed template";
        attribute :>> benefit    = "a reader sees what the verdict is and why, in words the service is allowed to know";
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement oneCasePerVerdictKind {
                doc /* Given a case for PASS, FAIL, INCONCLUSIVE and ERROR, when
                       each verdict is returned, then it is the kind the
                       statement gives for that case. */
            }
            requirement precedenceHolds {
                doc /* Given a case for each precedence rule, when two conditions
                       hold at once, then the verdict is the one the ordering
                       names. */
            }
            requirement reasonsFollowTheTemplates {
                doc /* Given a case per reason template, when the reason is
                       attached, then it is built from that template, and the
                       INCONCLUSIVE reason names the requirement's verification
                       case where the projection carries one. */
            }
        }
        doc /* Rationale: keying the verdict on the constrained quantity is what
               makes the latency requirement INCONCLUSIVE and the throughput ones
               evaluable, the enumeration mirrors the language's own verdict
               kinds, and the reason is what both apps show (AD-0006, AD-0008,
               AD-0024). Allocated to: CAPACITY. Upstream: US-02, US-03, US-04,
               US-05, US-06, US-09, US-10. */
    }

    #derive
    requirement <'SR-31'> SR_31_TwoConfiguredNames : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The capacity service shall carry exactly two model-specific names, the quantity it computes and the attribute it reads, and shall select from the adapter only fields of the generic projection, declared in its @requires.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Inspection);
        attribute :>> capability = "carry exactly two model-specific names and select only declared fields of the generic projection";
        attribute :>> benefit    = "what the services have to agree on is a few names and a key rather than a shared model";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement twoNamesAndDeclaredFieldsOnly {
                doc /* Given the capacity service's schema, configuration and
                       source, when they are read, then the only model-specific
                       names are the quantity it computes and the attribute it
                       reads, and every field it selects is declared in its
                       requires clause. */
            }
        }
        doc /* Rationale: the argument is that services need agree only on keys
               and a few names. The field set is a structural dependency on the
               projection and is stated as such, so nobody mistakes the contract
               for smaller than it is (AD-0008). Allocated to: CAPACITY, ADAPTER.
               Upstream: US-11. */
    }

    #derive
    requirement <'SR-32'> SR_32_NoCopyOfTheModel : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        stakeholder visitor : Visitor;
        attribute statement : String = "The capacity service shall hold no model data between requests.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "hold no model data between requests";
        attribute :>> benefit    = "the view is live rather than exported";
        frame concern : ShapeWithoutTouching;
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement restartChangesNothing {
                doc /* Given two identical requests, when the service is
                       restarted between them, then the two answers are the
                       same. */
            }
            requirement theServiceHasNoStore {
                doc /* Given the capacity service, when its source is read, then
                       it carries no store of model data. */
            }
        }
        doc /* Rationale: recomputing on read is what makes the view live rather
               than exported (AD-0014). Allocated to: CAPACITY. Upstream: US-09,
               US-10. */
    }

    // -------------------------------------------- the requirements document --

    #derive
    requirement <'SR-33'> SR_33_NumberingFromTheTreeAlone : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        attribute statement : String = "The document service shall hold an ordered tree of heading, prose and requirement nodes and shall number heading and requirement nodes in dotted decimal from the tree alone, siblings counted from one in tree order and prose nodes skipped and unnumbered.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "number heading and requirement nodes in dotted decimal from the ordered tree alone, skipping prose";
        attribute :>> benefit    = "the numbering is the document's own editorial decision and owes nothing to the model";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement shippedTreeNumbersCorrectly {
                doc /* Given the shipped tree, when it is numbered, then heading
                       and requirement nodes carry dotted decimal numbers with
                       siblings counted from one in tree order. */
            }
            requirement proseIsSkippedAndUnnumbered {
                doc /* Given a tree carrying prose nodes, when it is numbered,
                       then the prose nodes are skipped and carry no number. */
            }
            requirement numberingFollowsEachEdit {
                doc /* Given a tree after each kind of edit, when it is numbered,
                       then the numbers follow the tree alone. */
            }
        }
        doc /* Rationale: numbering is an editorial decision the model does not
               contain (AD-0025). Allocated to: DOCUMENT. Upstream: US-06, US-07,
               US-08. */
    }

    #derive
    requirement <'SR-34'> SR_34_EditorialOperationsInTheApp : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        attribute statement : String = "The requirements document app shall offer the document owner these operations: move a node, nest a node under another, insert a heading as the parent of a node, add a prose node, edit the text of a heading or prose node, exclude a requirement, and restore an excluded requirement.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "offer the document owner move, nest, insert a heading, add prose, edit text, exclude and restore";
        attribute :>> benefit    = "the document reads the way its readers expect, from the app the owner already has open";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement everyOperationIsOffered {
                doc /* Given the shipped tree, when each operation is taken in
                       turn, then the app offers all seven of them. */
            }
        }
        doc /* Rationale: the editorial operations the document owner needs, each
               of which the stories exercise (AD-0017, AD-0025). Allocated to:
               UI-DOC. Upstream: US-07, US-08. */
    }

    #derive
    requirement <'SR-35'> SR_35_EditorialOperationsInTheService : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        attribute statement : String = "When an operation of SR-34 is accepted, the document service shall apply it to the tree, restore an excluded requirement as the last child of its former parent, and renumber.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "apply an accepted operation to the tree, restore an excluded requirement as the last child of its former parent, and renumber";
        attribute :>> benefit    = "the service owns the tree and its numbering, so the app never has to";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement eachOperationIsApplied {
                doc /* Given one case per operation, when the operation is
                       accepted, then the service applies it to the tree and
                       renumbers. */
            }
            requirement restoreReturnsToTheFormerParent {
                doc /* Given an excluded requirement, when it is restored, then
                       it returns as the last child of its former parent. */
            }
        }
        doc /* Rationale: the service owns the tree and the numbering (AD-0025).
               Allocated to: DOCUMENT. Upstream: US-07, US-08. */
    }

    #derive
    requirement <'SR-36'> SR_36_DocumentOperationsLeaveTheModelUnchanged : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        attribute statement : String = "When an operation of SR-34 is accepted, the model version shall be unchanged.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "leave the model version untouched by every editorial operation";
        attribute :>> benefit    = "shaping the document cannot reach the model";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement documentReachesNoAdapterPackage {
                doc /* Given the module dependency graph, when the document
                       service is read, then it reaches no adapter package, which
                       is what makes the model version unreachable rather than
                       merely unchanged. */
            }
            requirement versionUnchangedThroughTheStack {
                doc /* Given the running stack where one store serves the
                       adapter, when each document operation is driven through
                       the composed graph, then the model version read from the
                       adapter's own endpoint is the same before and after. */
            }
        }
        doc /* Rationale: US-07 and US-08 observe that the model text, the sketch
               and the model's relationships stay as they were, which follows
               from the document service holding only requirement keys and
               structure. A test written where the document service holds no
               model state asserts only that an untouched object stays untouched,
               and cannot fail, so the property is verified in two halves
               (AD-0025). Allocated to: DOCUMENT, ADAPTER. Upstream: US-07,
               US-08. */
    }

    #derive
    requirement <'SR-37'> SR_37_WhatARequirementRowShows : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        stakeholder visitor : Visitor;
        attribute statement : String = "The requirements document app shall show for each included requirement its short name, text, limit with its unit, the requirement it derives from or the requirements derived from it, the part that satisfies it, the verification case that verifies it, the current value of its subject where the analysis returns one, and the verdict with its reason.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "show for each included requirement what the model knows about it and what the analysis concluded";
        attribute :>> benefit    = "a requirements reader reviews both without opening a modelling tool";
        frame concern : ShapeWithoutTouching;
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement rowCarriesWhatTheModelKnows {
                doc /* Given the document is open, when the row of each
                       requirement of the example is read against the projection,
                       then it shows every field the statement lists. */
            }
            requirement absentValueShowsNone {
                doc /* Given a requirement for which the analysis returns no
                       current value, when its row is read, then the row shows
                       none. */
            }
        }
        doc /* Rationale: the document is a view over the model plus the
               analysis, and a reader has to see both without a modelling tool
               (AD-0024, AD-0025). Allocated to: UI-DOC. Upstream: US-06, US-07,
               US-09, US-10. */
    }

    #derive
    requirement <'SR-38'> SR_38_TheDocumentsEditableSet : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        attribute statement : String = "The requirements document app shall present as editable exactly the throughput of each derived requirement's subject server and the limit of the global throughput requirement, and shall present every other value as read-only.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "offer the same editable set as the viewer, reached from the document";
        attribute :>> benefit    = "a number is corrected from where the document owner already works";
        frame concern : ShapeWithoutTouching;
        requirement :>> acceptance {
            requirement editableSetIsExact {
                doc /* Given the document is open, when its controls are read,
                       then every editable control maps to a derived
                       requirement's subject server throughput or to the limit of
                       the global throughput requirement, and no other value has
                       one. */
            }
        }
        doc /* Rationale: the same editable set as the viewer, reached from where
               the document owner works (AD-0004). Allocated to: UI-DOC.
               Upstream: US-09. */
    }

    // ----------------------------------------------- sharing through the router --

    #derive
    requirement <'SR-39'> SR_39_LiveReflectionInTheOtherApp : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        stakeholder visitor : Visitor;
        attribute statement : String = "When the model version or the document version changes, each app shall reflect the change within two seconds and without a page reload.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "reflect a model or document version change in each app within two seconds and without a reload";
        attribute :>> benefit    = "both apps are seen to read one graph rather than each other";
        frame concern : ShapeWithoutTouching;
        frame concern : LaunchAndSurprise;
        attribute reflectionSeconds : Rational = 2.0;
        requirement reflectionSla {
            doc /* Formalises the benefit: reflected within two seconds. */
            require constraint { reflectionSeconds <= 2.0 }
        }
        requirement :>> acceptance {
            requirement changeReachesTheOtherApp {
                doc /* Given both apps open in a local browser session, when the
                       model version or the document version changes, then the
                       other app reflects the change within two seconds, timed,
                       and without a page reload. */
            }
        }
        doc /* Rationale: the demo's claim is that both apps read one graph, and
               a reload would leave it a claim. The two-second bound is the
               demo's own success criterion (AD-0014). Allocated to: UI-MODEL,
               UI-DOC, ADAPTER, DOCUMENT, ROUTER. Upstream: US-09, US-10,
               US-12. */
    }

    #derive
    requirement <'SR-40'> SR_40_EveryRequestThroughTheRouter : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The apps shall send every query, mutation and subscription to the router.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Inspection);
        attribute :>> capability = "send every query, mutation and subscription to the router";
        attribute :>> benefit    = "the router stays the only place where the three services meet";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement everyCallGoesToTheRouter {
                doc /* Given the apps, when their network calls are inspected,
                       then every query, mutation and subscription is addressed
                       to the router. */
            }
            requirement appsFailWithoutTheRouter {
                doc /* Given the router is absent, when either app is opened,
                       then it fails rather than reaching a service some other
                       way. */
            }
        }
        doc /* Rationale: the router is the only place where the three services
               meet (AD-0001, AD-0011). Allocated to: UI-MODEL, UI-DOC. Upstream:
               US-10, US-11. */
    }

    #derive
    requirement <'SR-41'> SR_41_ServicesIsolated : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : DocumentOwner;
        stakeholder visitor : Visitor;
        attribute statement : String = "No service shall import, call or read the data of another service.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Inspection);
        attribute :>> capability = "keep every service from importing, calling or reading the data of another";
        attribute :>> benefit    = "no side channel undoes the argument the demo is making";
        frame concern : ShapeWithoutTouching;
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement noServiceReachesAnother {
                doc /* Given the module dependency graph and each service's
                       outbound calls, when they are inspected, then no service
                       imports, calls or reads the data of another. */
            }
        }
        doc /* Rationale: any side channel would undo the argument (AD-0001).
               Allocated to: ADAPTER, CAPACITY, DOCUMENT. Upstream: US-07, US-08,
               US-10. */
    }

    #derive
    requirement <'SR-42'> SR_42_CompositionThatCannotDrift : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : AdoptingOrganisation;
        attribute statement : String = "The committed router configuration shall embed subgraph schemas identical to the schema files in the repository.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Demonstration);
        attribute :>> capability = "embed in the committed router configuration subgraph schemas identical to the repository's schema files";
        attribute :>> benefit    = "the contract between producer and consumer is checked mechanically before deployment rather than discovered in production";
        frame concern : AgreeAndReplace;
        requirement :>> acceptance {
            requirement driftFailsTheTest {
                doc /* Given the committed configuration and the schema files,
                       when the test runs, then it fails on any difference
                       between them. */
            }
            requirement compositionProducesTheCommittedFile {
                doc /* Given the maintainer's compose step, when it is run, then
                       its output is the committed configuration. */
            }
        }
        doc /* Rationale: the README's third condition, the contract checked
               mechanically before deployment, is shown by the repository rather
               than by the demo. Composition needs a toolchain continuous
               integration does not carry, so the maintainer composes and commits
               and a test guards the result (AD-0002, AD-0012). Allocated to:
               ROUTER, REPO. Upstream: US-16. */
    }

    #derive
    requirement <'SR-43'> SR_43_OneQueryThreeServices : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "The demo shall answer, in one response, a query for a requirement's text, verdict and document number, and its served schema shall contain types contributed by all three subgraphs.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Demonstration);
        attribute :>> capability = "answer a requirement's text, verdict and document number in one response, from a schema all three subgraphs contribute to";
        attribute :>> benefit    = "the visitor sees the federation rather than taking it on trust";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement oneResponseCarriesAllThree {
                doc /* Given the running stack, when one query asks for a
                       requirement's text, verdict and document number, then one
                       response carries all three. */
            }
            requirement schemaCarriesAllThreeSubgraphs {
                doc /* Given the playground, when the served schema is read, then
                       it contains types contributed by all three subgraphs. */
            }
        }
        doc /* Rationale: the visitor sees the federation in the playground
               rather than taking it on trust (AD-0001, AD-0002). Allocated to:
               ROUTER, ADAPTER, CAPACITY, DOCUMENT. Upstream: US-11. */
    }

    #derive
    requirement <'SR-44'> SR_44_ResetFromEitherApp : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Visitor;
        attribute statement : String = "When reset is requested from either app, the demo shall show the shipped model values and the shipped document structure in both apps.";
        attribute pattern : EarsPattern = EarsPattern::EventDriven;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Demonstration);
        attribute :>> capability = "return both apps to the shipped model values and the shipped document structure on a request from either";
        attribute :>> benefit    = "the demo is shown again from the beginning";
        frame concern : LaunchAndSurprise;
        requirement :>> acceptance {
            requirement resetMutationsReturnTheShippedState {
                doc /* Given edited values and an edited document structure, when
                       the adapter's and the document service's reset mutations
                       are called, then each returns its shipped state. */
            }
            requirement resetFromEitherApp {
                doc /* Given both apps, when the reset control is used in either,
                       then both show the shipped model values and the shipped
                       document structure. */
            }
        }
        doc /* Rationale: a demo is shown more than once, and the two-second
               bound is SR-39's (AD-0004, AD-0025). Allocated to: ADAPTER,
               DOCUMENT, UI-MODEL, UI-DOC. Upstream: US-12. */
    }

    // ------------------------------------- the example model and the record --

    #derive
    requirement <'SR-45'> SR_45_AModelTheReferenceToolsAccept : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : ModelOwner;
        attribute statement : String = "The example model shall be accepted without error by the OMG pilot implementation release 2026-07 and by the OpenSysML command line.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Demonstration);
        attribute :>> capability = "be accepted without error by the OMG pilot implementation release 2026-07 and by the OpenSysML command line";
        attribute :>> benefit    = "a reader who knows the language loads the file in a real tool and trusts what the demo says about it";
        frame concern : ModelStaysTheSource;
        requirement :>> acceptance {
            requirement bothToolsAcceptTheExample {
                doc /* Given the example model, when the OMG pilot
                       implementation and the OpenSysML command line read it,
                       then neither reports an error. */
            }
            requirement recordedInTheExampleReadme {
                doc /* Given the run is made locally, when it is recorded in the
                       example README, then the record names the releases
                       used. */
            }
        }
        doc /* Rationale: a subset parser cannot prove conformance, and a reader
               who knows the language will load the file in a real tool.
               Acceptance shows the file is valid as the reference tools read it
               rather than conformance in the formal sense (AD-0003, AD-0015,
               AD-0019). Allocated to: MODEL. Upstream: US-17. */
    }

    #derive
    requirement <'SR-46'> SR_46_TheModelAndTheRepositoryAgree : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Maintainer;
        attribute statement : String = "The repository shall carry a unit test that fails when an identifier of a story, a requirement, a constraint or a decision record, a Go test name, a check file or a published image is named in the model without existing in the repository, or exists in the repository without being named in the model.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "keep the model and the repository in agreement by a test that fails on the first identifier, test name, check file or image that one names and the other lacks";
        attribute :>> benefit    = "the record cannot drift from what ships without a check failing";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement identifiers {
                doc /* Every identifier the model writes is declared exactly once
                       in it, every decision record the model names exists under
                       the decisions folder, and every record there is named in
                       the model. */
            }
            requirement requirementsAffected {
                doc /* Every requirement a decision record names as affected is a
                       short name the model declares. */
            }
            requirement goTests {
                doc /* Every Go test the model names exists in the file it names,
                       and every Go test of the requirement scheme is named by
                       the model. */
            }
            requirement checkFiles {
                doc /* Every check file a case names exists in the check project,
                       and every check file of that project is named by a
                       case. */
            }
            requirement images {
                doc /* Every published image is named by exactly one view, and
                       every image a view names is published. */
            }
            requirement coverage {
                doc /* Every story that is done is verified by a case, satisfied
                       by something in the logical architecture, and a derived
                       end of a derivation connection. */
            }
            requirement checkInventory {
                doc /* The check register and the manifest the check project
                       reads carry the same checks with the same fields. */
            }
            requirement sessionParts {
                doc /* The services the compose file of the check session starts
                       are the parts the session composite carries, and nothing
                       else. */
            }
        }
        doc /* Rationale: a general claim about the model is only as good as the
               test that reads it (AD-0029). Allocated to: DEMOMODEL, REPO.
               Upstream: US-18. */
    }

    #derive
    requirement <'SR-47'> SR_47_TheDemosModelAcceptedByTheReferenceTools : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Maintainer;
        attribute statement : String = "The model under model/ shall be accepted without error or warning by the OMG pilot implementation release 2026-07 and by OpenSysML v0.6.0, and continuous integration shall run both on every pull request and push to main that changes a file under model/.";
        attribute pattern : EarsPattern = EarsPattern::Ubiquitous;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test);
        attribute :>> capability = "validate the demo's own model with both reference tools on every change";
        attribute :>> benefit    = "a form one tool refuses never lands on main";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement bothToolsAccept {
                doc /* Given the whole model tree, when the OMG pilot
                       implementation release 2026-07 and OpenSysML v0.6.0 read
                       it, then neither reports an error or a warning. */
            }
            requirement runsInContinuousIntegration {
                doc /* Given a pull request or a push to main that changes a file
                       under model/, when the workflow runs, then it puts the
                       model to both tools. */
            }
        }
        doc /* Rationale: the two tools disagree on forms, and nothing else reads
               this model (AD-0030). Allocated to: DEMOMODEL, REPO. Upstream:
               US-18. */
    }

    #derive
    requirement <'SR-48'> SR_48_AnOptionalCheckSession : UserStory {
        @StoryMeta { status = StoryStatus::done; }
        subject :>> system : DemoElement;
        stakeholder :>> role : Maintainer;
        attribute statement : String = "When Checkly credentials are supplied to the compose profile, the demo shall be reachable through a tunnel, the check suite shall run against it and record a session, the project shall be deployed for the duration and destroyed when the profile stops, and the router's traces shall reach a collector beside the demo. When no credentials are supplied, the demo shall run as docker run runs it.";
        attribute pattern : EarsPattern = EarsPattern::Complex;
        attribute verificationMethods : VerificationMethod[1..*] = (VerificationMethod::Test, VerificationMethod::Demonstration);
        attribute :>> capability = "run the whole check suite against a running instance through a tunnel, deploy it for the session, trace every request, and destroy it on stop, only when asked";
        attribute :>> benefit    = "the demo's claims are checked whenever it runs with credentials, and the demo is untouched otherwise";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement reachableThroughATunnel {
                doc /* Given credentials supplied to the compose profile, when
                       the session starts, then the demo answers on a public
                       hostname through the tunnel. */
            }
            requirement suiteRunsAndRecords {
                doc /* Given the demo is reachable, when the check suite runs
                       against it, then every check runs and the run is recorded
                       as a session. */
            }
            requirement deployedForTheDuration {
                doc /* Given the session is running, when the project is read at
                       the monitoring service, then its checks are deployed for
                       as long as the session lasts. */
            }
            requirement destroyedOnStop {
                doc /* Given a running session, when the compose profile stops,
                       then the deployed project is destroyed. */
            }
            requirement tracesReachTheCollector {
                doc /* Given the session is running, when a request passes
                       through the router, then its trace reaches the collector
                       beside the demo. */
            }
            requirement unchangedWithoutCredentials {
                doc /* Given no credentials, when the demo is started, then it
                       runs exactly as docker run runs it. */
            }
        }
        doc /* Rationale: a monitoring service runs elsewhere and a demo on one
               machine answers only there. Allocated to: CHECKS, REPO. Upstream:
               US-19. */
    }

    // ------------------------------------------------ the derivations, US-01 to US-19 --

    #derivation connection us01Derives {
        doc /* The requirements US-01 gives rise to. */
        end #original ::> US_01_LaunchWithOneCommand;
        end #derive   ::> SR_01_OneCommand;
        end #derive   ::> SR_02_ReadyWithinTenSeconds;
        end #derive   ::> SR_03_NoOutboundNetwork;
        end #derive   ::> SR_04_FourPathsOnOnePort;
        end #derive   ::> SR_05_TwoPlatforms;
        end #derive   ::> SR_06_SizeBudget;
        end #derive   ::> SR_10_NoResourceFromAnotherOrigin;
    }

    #derivation connection us02Derives {
        doc /* The requirements US-02 gives rise to. */
        end #original ::> US_02_ReadTheModel;
        end #derive   ::> SR_11_ModelTextWithTheLanguageVisible;
        end #derive   ::> SR_12_SketchDrawnFromTheWiring;
        end #derive   ::> SR_14_VerdictAndReasonInTheViewer;
        end #derive   ::> SR_16_GenericProjection;
        end #derive   ::> SR_20_ConnectionDirection;
        end #derive   ::> SR_21_ShortNamesAsIdentifiers;
        end #derive   ::> SR_28_RollupByMaximumFlow;
        end #derive   ::> SR_29_BottleneckAsTheMinimumCut;
        end #derive   ::> SR_30_VerdictsAndReasons;
    }

    #derivation connection us03Derives {
        doc /* The requirements US-03 gives rise to. */
        end #original ::> US_03_RaiseAServerThatIsNotTheBottleneck;
        end #derive   ::> SR_12_SketchDrawnFromTheWiring;
        end #derive   ::> SR_13_TheViewersEditableSet;
        end #derive   ::> SR_14_VerdictAndReasonInTheViewer;
        end #derive   ::> SR_22_EditsPatchTheSource;
        end #derive   ::> SR_25_RejectionOfInvalidValues;
        end #derive   ::> SR_28_RollupByMaximumFlow;
        end #derive   ::> SR_29_BottleneckAsTheMinimumCut;
        end #derive   ::> SR_30_VerdictsAndReasons;
    }

    #derivation connection us04Derives {
        doc /* The requirements US-04 gives rise to. */
        end #original ::> US_04_RaiseTheBottleneck;
        end #derive   ::> SR_12_SketchDrawnFromTheWiring;
        end #derive   ::> SR_13_TheViewersEditableSet;
        end #derive   ::> SR_14_VerdictAndReasonInTheViewer;
        end #derive   ::> SR_15_FailingRequirementInRed;
        end #derive   ::> SR_22_EditsPatchTheSource;
        end #derive   ::> SR_28_RollupByMaximumFlow;
        end #derive   ::> SR_29_BottleneckAsTheMinimumCut;
        end #derive   ::> SR_30_VerdictsAndReasons;
    }

    #derivation connection us05Derives {
        doc /* The requirements US-05 gives rise to. */
        end #original ::> US_05_TightenTheLimit;
        end #derive   ::> SR_13_TheViewersEditableSet;
        end #derive   ::> SR_14_VerdictAndReasonInTheViewer;
        end #derive   ::> SR_15_FailingRequirementInRed;
        end #derive   ::> SR_22_EditsPatchTheSource;
        end #derive   ::> SR_28_RollupByMaximumFlow;
        end #derive   ::> SR_29_BottleneckAsTheMinimumCut;
        end #derive   ::> SR_30_VerdictsAndReasons;
    }

    #derivation connection us06Derives {
        doc /* The requirements US-06 gives rise to. */
        end #original ::> US_06_ReadTheDocument;
        end #derive   ::> SR_16_GenericProjection;
        end #derive   ::> SR_19_QuantityComparisonAndLimitFromTheConstraint;
        end #derive   ::> SR_23_ExpressionBoundValuesAreEvaluated;
        end #derive   ::> SR_30_VerdictsAndReasons;
        end #derive   ::> SR_33_NumberingFromTheTreeAlone;
        end #derive   ::> SR_37_WhatARequirementRowShows;
    }

    #derivation connection us07Derives {
        doc /* The requirements US-07 gives rise to. */
        end #original ::> US_07_ReorderAndNest;
        end #derive   ::> SR_33_NumberingFromTheTreeAlone;
        end #derive   ::> SR_34_EditorialOperationsInTheApp;
        end #derive   ::> SR_35_EditorialOperationsInTheService;
        end #derive   ::> SR_36_DocumentOperationsLeaveTheModelUnchanged;
        end #derive   ::> SR_37_WhatARequirementRowShows;
        end #derive   ::> SR_41_ServicesIsolated;
    }

    #derivation connection us08Derives {
        doc /* The requirements US-08 gives rise to. */
        end #original ::> US_08_ShapeTheDocument;
        end #derive   ::> SR_33_NumberingFromTheTreeAlone;
        end #derive   ::> SR_34_EditorialOperationsInTheApp;
        end #derive   ::> SR_35_EditorialOperationsInTheService;
        end #derive   ::> SR_36_DocumentOperationsLeaveTheModelUnchanged;
        end #derive   ::> SR_41_ServicesIsolated;
    }

    #derivation connection us09Derives {
        doc /* The requirements US-09 gives rise to. */
        end #original ::> US_09_ChangeAValueFromTheDocument;
        end #derive   ::> SR_12_SketchDrawnFromTheWiring;
        end #derive   ::> SR_22_EditsPatchTheSource;
        end #derive   ::> SR_23_ExpressionBoundValuesAreEvaluated;
        end #derive   ::> SR_26_ModelVersionEvents;
        end #derive   ::> SR_27_DocumentVersionEvents;
        end #derive   ::> SR_28_RollupByMaximumFlow;
        end #derive   ::> SR_29_BottleneckAsTheMinimumCut;
        end #derive   ::> SR_30_VerdictsAndReasons;
        end #derive   ::> SR_32_NoCopyOfTheModel;
        end #derive   ::> SR_37_WhatARequirementRowShows;
        end #derive   ::> SR_38_TheDocumentsEditableSet;
        end #derive   ::> SR_39_LiveReflectionInTheOtherApp;
    }

    #derivation connection us10Derives {
        doc /* The requirements US-10 gives rise to. */
        end #original ::> US_10_ChangeFromTheViewerWatchTheDocument;
        end #derive   ::> SR_23_ExpressionBoundValuesAreEvaluated;
        end #derive   ::> SR_26_ModelVersionEvents;
        end #derive   ::> SR_27_DocumentVersionEvents;
        end #derive   ::> SR_28_RollupByMaximumFlow;
        end #derive   ::> SR_29_BottleneckAsTheMinimumCut;
        end #derive   ::> SR_30_VerdictsAndReasons;
        end #derive   ::> SR_32_NoCopyOfTheModel;
        end #derive   ::> SR_37_WhatARequirementRowShows;
        end #derive   ::> SR_39_LiveReflectionInTheOtherApp;
        end #derive   ::> SR_40_EveryRequestThroughTheRouter;
        end #derive   ::> SR_41_ServicesIsolated;
    }

    #derivation connection us11Derives {
        doc /* The requirements US-11 gives rise to. */
        end #original ::> US_11_QueryTheGraph;
        end #derive   ::> SR_04_FourPathsOnOnePort;
        end #derive   ::> SR_21_ShortNamesAsIdentifiers;
        end #derive   ::> SR_31_TwoConfiguredNames;
        end #derive   ::> SR_40_EveryRequestThroughTheRouter;
        end #derive   ::> SR_43_OneQueryThreeServices;
    }

    #derivation connection us12Derives {
        doc /* The requirements US-12 gives rise to. */
        end #original ::> US_12_Reset;
        end #derive   ::> SR_09_StateLivesInMemory;
        end #derive   ::> SR_39_LiveReflectionInTheOtherApp;
        end #derive   ::> SR_44_ResetFromEitherApp;
    }

    #derivation connection us13Derives {
        doc /* The requirements US-13 gives rise to. */
        end #original ::> US_13_LicencesTravelWithWhatIsCopied;
        end #derive   ::> SR_07_LicenceCarriedInTheImage;
        end #derive   ::> SR_08_LicencesCarriedInTheRepository;
    }

    #derivation connection us14Derives {
        doc /* The requirements US-14 gives rise to. */
        end #original ::> US_14_NothingOfTheExampleInTheAdapter;
        end #derive   ::> SR_17_NoExampleIdentifiersInTheAdapter;
    }

    #derivation connection us15Derives {
        doc /* The requirements US-15 gives rise to. */
        end #original ::> US_15_AServedModelThatIsNeverSilentlyWrong;
        end #derive   ::> SR_18_RefusalOfUnsupportedSyntax;
        end #derive   ::> SR_24_ExpressionBoundValuesAreReadOnly;
    }

    #derivation connection us16Derives {
        doc /* The requirements US-16 gives rise to. */
        end #original ::> US_16_TheContractCheckedBeforeDeployment;
        end #derive   ::> SR_42_CompositionThatCannotDrift;
    }

    #derivation connection us17Derives {
        doc /* The requirements US-17 gives rise to. */
        end #original ::> US_17_TheExampleAcceptedByTheReferenceTools;
        end #derive   ::> SR_45_AModelTheReferenceToolsAccept;
    }

    #derivation connection us18Derives {
        doc /* The requirements US-18 gives rise to. */
        end #original ::> US_18_TheDemosOwnModelAsTheRecord;
        end #derive   ::> SR_46_TheModelAndTheRepositoryAgree;
        end #derive   ::> SR_47_TheDemosModelAcceptedByTheReferenceTools;
    }

    #derivation connection us19Derives {
        doc /* The requirements US-19 gives rise to. */
        end #original ::> US_19_TheStoriesExercisedAgainstARunningInstance;
        end #derive   ::> SR_48_AnOptionalCheckSession;
    }
}
