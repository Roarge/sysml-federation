package Experiment_DesignConstraints {
    doc /* Four design constraints, EXP-SC-01 to EXP-SC-04, each with a plain
           statement. */
    private import Federation_Library::*;
    private import Experiment_Context::ExperimentElement;

    requirement <'EXP-SC-01'> EXP_SC_01_StandardLibraryInItsOwnModule : DesignConstraint {
        subject :>> system : ExperimentElement;
        attribute :>> statement = "The experiment is a Go module of its own under experiments/llm-resolution, outside the product and its image, and imports nothing but the Go standard library.";
        requirement :>> acceptance {
            requirement noRequirements { doc /* The module file lists no dependency. */ }
        }
        doc /* Decisions: AD-0032. */
    }

    requirement <'EXP-SC-02'> EXP_SC_02_NoEmptyInterface : DesignConstraint {
        subject :>> system : ExperimentElement;
        attribute :>> statement = "No hand-written code of the experiment uses the empty interface or any in a value position, as SC-02 asks of the demo.";
        requirement :>> acceptance {
            requirement noneFound { doc /* A scan of the module's source finds neither. */ }
        }
    }

    requirement <'EXP-SC-03'> EXP_SC_03_AboutAnHour : DesignConstraint {
        subject :>> system : ExperimentElement;
        attribute :>> statement = "A full run finishes in about an hour on the operator's server, a 14B model at 4-bit quantisation on a 12 GB graphics card that generates about 22 tokens a second. It is an estimate, not a gate.";
        requirement :>> acceptance {
            requirement measured { doc /* The operator's run records its own duration. */ }
        }
    }

    requirement <'EXP-SC-04'> EXP_SC_04_ResultsStayOutOfTheRepository : DesignConstraint {
        subject :>> system : ExperimentElement;
        attribute :>> statement = "Results files are written under experiments/llm-resolution/results, which git ignores, and are never committed.";
        requirement :>> acceptance {
            requirement ignored { doc /* git reports a file under results/ as ignored. */ }
        }
    }
}
