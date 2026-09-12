package Federation_Library {
    doc /* Definitions particular to the model of the demo: the requirement
           scheme's enumerations, the design-constraint form, and the metadata
           that ties elements to decision records and verification actions to
           their evidence. */
    private import ScalarValues::*;

    enum def VerificationMethod {
        doc /* The four methods of the requirements scheme. */
        Test; Analysis; Inspection; Demonstration;
    }

    enum def EarsPattern {
        doc /* The EARS pattern a requirement statement follows. */
        Ubiquitous; EventDriven; StateDriven; UnwantedBehaviour; Complex;
    }

    requirement def DesignConstraint {
        doc /* A plain-statement constraint of the light scheme (SC-nn). */
        subject system;
        attribute statement : String;
        requirement acceptance[0..*] {
            doc /* What shows the constraint is honoured. */
        }
    }

    metadata def DecisionRecord {
        doc /* Applied to an element a decision record shaped, naming the record. */
        attribute id : String;
        attribute title : String;
    }

    metadata def Evidence {
        doc /* Applied to a verification action: what kind of evidence it is and
               where it lives. Kinds: go-test, record, validator, make-target,
               workflow, analysis, inspection, checkly. */
        attribute kind : String;
        attribute location : String;
    }

    metadata def InformationalOnly {
        doc /* Applied to a concern no story frames, whose doc says what carries it. */
    }

    enum def CheckKind {
        doc /* The kinds of check the monitoring service runs. */
        ApiCheck; BrowserCheck; MultiStepCheck; PlaywrightSuite;
        UrlMonitor; SslMonitor; DnsMonitor; TcpMonitor; HeartbeatMonitor;
    }
}
