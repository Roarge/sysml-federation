package Fixture_VerificationCases {
    doc /* One case per system story, with the Go tests as evidence. */
    private import Fixture_Library::*;
    private import Fixture_SystemStories::*;

    verification def VC_SR_02 {
        doc /* Serving results. */
        objective { verify SR_02_ServeResults; }
        action <'TestSR02_ReturnsARankedPage'> rankedPage {
            @Evidence { kind = "go-test"; location = "adapter/serve/serve_test.go"; }
            doc /* TestSR02_ReturnsARankedPage checks the page is ranked. */
        }
        action <'record: 2026-09-01 SR-02'> servedByHand {
            @Evidence { kind = "record"; location = "docs/notes.md#a-record"; }
            doc /* A query served by hand, the page ranked. */
        }
        action <'record: 2026-09-02 SR-02'> servedAgain {
            @Evidence { kind = "record"; location = "docs/gone.md#a-record"; }
            doc /* The same query, recorded in a file the checkout no longer has. */
        }
        action <'inspection: SR-02'> reviewed {
            @Evidence { kind = "inspection"; location = "review checklist"; }
            doc /* The page read against the story in a review. */
        }
    }
}
