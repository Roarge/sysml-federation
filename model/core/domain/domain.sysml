package Federation_Domain {
    doc /* The things the demo moves about: what a browser asks for and is
           given, what the services publish, the image the demo ships in, and
           the items the check session carries. Nothing here is a component or
           a port. These are the payloads the ports and interfaces name. */
    private import ScalarValues::*;

    item def Page {
        doc /* A page served on the one published port: the viewer, the
               document app, the shared client module or the playground. */
    }

    item def GraphQLQuery {
        doc /* A read sent to the router, which plans it across the three
               subgraphs and merges the answer. */
    }

    item def GraphQLMutation {
        doc /* A write sent to the router: one of the adapter's three model
               mutations or one of the document service's seven. */
    }

    item def VersionEvent {
        doc /* What a subscription delivers when the model or the document
               tree changes, carrying the counter the apps redraw against. */
        attribute version : Integer;
    }

    item def ContainerImage {
        doc /* The published image, one manifest for linux/amd64 and one for
               linux/arm64 behind a single name. */
    }

    item def RouterConfiguration {
        doc /* config.json, the composed router configuration. It is produced
               outside the build by wgc and committed, so the image carries a
               configuration that was composed on a machine rather than in the
               container. */
    }

    enum def Verdict {
        doc /* The verdict the capacity service returns for a requirement, the
               four words of VerdictKind in the SysML v2 Systems Library. */
        PASS; FAIL; INCONCLUSIVE; ERROR;
    }

    enum def NodeKind {
        doc /* What a node in the requirements document tree is: a heading, a
               paragraph of prose, or a requirement drawn from the model. */
        HEADING; PROSE; REQUIREMENT;
    }

    attribute def EntityKey :> String {
        doc /* The value by which every service agrees it is talking about the
               same object. It is the SysML short name, with the qualified name
               as the fallback where an element declares no short name, per the
               decision on short names as entity keys (AD-0018). */
    }

    item def TraceContext {
        doc /* The W3C trace headers the monitoring service injects into a
               request. The tracestate header carries checkly=true, which is
               what marks a span as belonging to a check run rather than to a
               visitor. */
        attribute traceparent : String;
        attribute tracestate : String;
    }

    item def Span {
        doc /* One timed operation as it is exported, carrying the trace it
               belongs to, its name, and whether the check marking reached
               it. */
        attribute traceId : String;
        attribute spanName : String;
        attribute checklyMarked : Boolean;
    }

    item def Trace {
        doc /* The router span and the subgraph fetch spans beneath it, all
               under one trace id, which is how a federated query is read back
               as one thing rather than as four unrelated operations. */
    }

    item def CheckResult {
        doc /* What one check run reports: whether it passed, alongside
               whatever the run recorded about how it got there. */
        attribute passed : Boolean;
    }

    item def Alert {
        doc /* What the monitoring service delivers when a check fails or
               recovers. */
    }

    item def HeartbeatPing {
        doc /* The signal a scheduled job sends to say it ran. Silence is the
               failure, so the absence of this item is what the monitor
               reports on. */
    }

    item def ProjectCommand {
        doc /* An instruction sent to the monitoring service about the check
               project as a whole: test, deploy, trigger or destroy. */
    }
}
