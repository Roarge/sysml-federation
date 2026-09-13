package Federation_DesignConstraints {
    doc /* The design constraint register. A constraint of the light scheme
           (SC-nn) binds how the demo is built rather than what it does, so it
           carries a plain statement instead of an EARS one and traces to no
           story. Each names the whole demo as its subject, the concern it
           frames, one acceptance criterion, and the decisions and allocation
           elements behind it. */
    private import Federation_Library::DesignConstraint;
    private import Federation_Context::DemoElement;
    private import Federation_Concerns::*;
    private import ScalarValues::String;

    requirement <'SC-01'> SC_01_LanguageAndLicence : DesignConstraint {
        subject :>> system : DemoElement;
        attribute :>> statement = "Product code is Go 1.27 under the Apache-2.0 licence. The only permitted third-party Go dependencies beyond the standard library are gqlgen with its own dependencies, go-cmp for tests, and what those require. No dependency that drives a browser is added to the product code, and go-yaml is permitted for the test that reads the compose file. The optional check project under checkly/ is outside the product and is not built, run or read to run the demo.";
        frame concern : ReadBuildChange;
        requirement :>> acceptance {
            requirement permittedDependenciesOnly {
                doc /* The module requires only gqlgen, go-cmp and go-yaml for
                       tests, and what those require. */
            }
        }
        doc /* Decisions: AD-0031. Allocated to: ADAPTER, CAPACITY, DOCUMENT,
               IMAGE, REPO. */
    }

    requirement <'SC-02'> SC_02_NoEmptyInterfaceInHandWrittenCode : DesignConstraint {
        subject :>> system : DemoElement;
        attribute :>> statement = "No interface{} or any in a value position in hand-written code. Generated files marked // Code generated ... DO NOT EDIT. are exempt, and the analyser runs with -generated=false. A forced exception in a hand-written signature carries //nointerface:allow and a baseline bump with a justification in the pull request.";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement analyserPasses {
                doc /* The analyser passes across the five platform pairs. */
            }
        }
        doc /* Decisions: AD-0016. Allocated to: ADAPTER, CAPACITY, DOCUMENT,
               IMAGE. */
    }

    requirement <'SC-03'> SC_03_TestFirst : DesignConstraint {
        subject :>> system : DemoElement;
        attribute :>> statement = "A new implementation file may not appear in a package that has no test. The test is the specification.";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement noImplementationWithoutATest {
                doc /* No implementation file lands in a package without a
                       test. */
            }
        }
        doc /* Decisions: AD-0022. Allocated to: REPO. */
    }

    requirement <'SC-04'> SC_04_TrackedByAllowlist : DesignConstraint {
        subject :>> system : DemoElement;
        attribute :>> statement = "Every new tracked path is named in .gitignore. The implementation phase adds rules for *.html, *.css and *.js under examples/, for a root NOTICE and for docs/a3/*.pdf, places the committed router configuration and its compose input at examples/pipeline/ where the existing one-level rules reach them, keeps the vendored SortableJS out of any directory named vendor/, and extends check-allowlist to the new extensions. docs/decisions/ and .github/workflows/publish.yml are already tracked by existing rules. internal/assert and internal/tabletest are tracked so that tracked tests may import them (AD-0022).";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement gatesPass {
                doc /* Check-tracked and check-allowlist pass. */
            }
        }
        doc /* Decisions: AD-0022. Allocated to: REPO. */
    }

    requirement <'SC-05'> SC_05_NoBuildStepForTheWebApps : DesignConstraint {
        subject :>> system : DemoElement;
        attribute :>> statement = "The web apps are HTML, CSS and native ES modules served from the Go binary, with SortableJS as the only vendored file. No Node toolchain is needed to build, run or read them.";
        frame concern : ReadBuildChange;
        requirement :>> acceptance {
            requirement noBuildStep {
                doc /* No build step exists for the apps. */
            }
        }
        doc /* Decisions: AD-0017. Allocated to: UI-MODEL, UI-DOC. */
    }

    requirement <'SC-06'> SC_06_ReadableInAnAfternoon : DesignConstraint {
        subject :>> system : DemoElement;
        attribute :>> statement = "The hand-written code stays small enough to read in an afternoon. These figures are the expected scale, excluding tests and generated files: all hand-written Go under adapter/ about 2750 lines, each of the capacity and document services about 600, the supervisor with its UI server and command dispatch about 700, each web app about 900 lines of JavaScript and 300 of CSS, with a module shared by both apps counted once. They are estimates, not a gate. Code that needs more lines to be correct takes them, and the figure is then revised to what the work turned out to be. No implementation is trimmed, no fix is refused and no refusal is left unwritten to stay under a number (AD-0028).";
        frame concern : ReadBuildChange;
        requirement :>> acceptance {
            requirement totalsMeasured {
                doc /* The line totals are measured and reported. */
            }
        }
        doc /* Decisions: AD-0017, AD-0027, AD-0028. Allocated to: ADAPTER,
               CAPACITY, DOCUMENT, IMAGE, UI-MODEL, UI-DOC. */
    }

    requirement <'SC-07'> SC_07_ContinuousIntegrationRunsUnitTestsAndPublishesOnTags : DesignConstraint {
        subject :>> system : DemoElement;
        attribute :>> statement = "CI runs the unit tests on pull requests and pushes to main, validates the demo's own model with the two reference tools when a file under model/ changes, and publishes the image on v* tags. The example model is validated locally and its record is in the example README. The full local gate is make check, make model-check runs the validators over the demo's model and make example-model-check over the example. The documentation site is outside this constraint: GitHub Pages builds docs/ with Jekyll on its own infrastructure from the published branch, and no workflow in this repository runs for it (noted 2026-08-27, documentation phase).";
        frame concern : PublishedClaimsStayTrue;
        requirement :>> acceptance {
            requirement threeWorkflowsOnly {
                doc /* The two workflows and the model workflow are the only
                       ones. */
            }
        }
        doc /* Decisions: AD-0012, AD-0020, AD-0030. Allocated to: REPO. */
    }
}
