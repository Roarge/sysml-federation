package Federation_CheckCases {
    doc /* The check register. One case per check and monitor of the check
           project, the check itself as a model element: its attributes equal
           the manifest's entry for it, and its actions are the check's own
           steps, a request and what is asserted of the answer. Each case reads
           the acceptance of the story the check exercises, and its evidence is
           the file that declares or drives it. */
    private import Federation_StakeholderStories::*;
    private import Federation_SystemStories::*;
    private import Federation_Context::*;
    private import Federation_LogicalArchitecture::*;
    private import Federation_Library::*;
    private import ScalarValues::*;

    // ------------------------------------------------ the router, hourly --

    verification def CHK_RouterVersion {
        doc /* The model version through the router: the smallest query the composed graph answers. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_43_OneQueryThreeServices.acceptance; }
        attribute logicalId : String = "router-version";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action postTheVersionQuery { doc /* POST the query for the model's version. */ }
        action expectAVersionAboveZero { doc /* Status 200 and a version greater than 0. */ }
    }

    verification def CHK_RouterJoin {
        doc /* The join query through the router: one requirement with fields from all three subgraphs. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_43_OneQueryThreeServices.acceptance; }
        attribute logicalId : String = "router-join";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action postTheJoinQuery { doc /* POST the requirement query for PIPE-R1 asking for text, verdict, verdictReason and documentNumber. */ }
        action expectTheThreeSubgraphsAnswer { doc /* Status 200, the text equals the shipped sentence, the verdict is not empty and documentNumber is 1. */ }
    }

    verification def CHK_RouterCapacityBottleneck {
        doc /* The capacity and the bottleneck of the pipeline through the router: the minimum cut, computed on the way out. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_29_BottleneckAsTheMinimumCut.acceptance; }
        attribute logicalId : String = "router-capacity-bottleneck";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action postTheCapacityQuery { doc /* POST the part query for PIPE-P1 asking for capacity and the bottleneck's ids. */ }
        action expectACapacityAndABottleneck { doc /* Status 200, a capacity that is not null and a bottleneck that is not empty. */ }
    }

    verification def CHK_RouterIntrospection {
        doc /* Introspection through the router: one schema that carries a type from each subgraph. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_42_CompositionThatCannotDrift.acceptance; }
        attribute logicalId : String = "router-introspection";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action postTheIntrospectionQuery { doc /* POST the schema query asking for the name of every type. */ }
        action expectTypesFromEachSubgraph { doc /* Status 200 and a body that contains Verdict, Document, Node and Model. */ }
    }

    verification def CHK_RouterPlayground {
        doc /* The playground on the one port: an HTML page for a browser that asks for one. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "router-playground";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/playground";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action getThePlayground { doc /* GET the playground path with Accept set to text/html. */ }
        action expectAnHtmlPage { doc /* Status 200 and a body that contains an html tag. */ }
    }

    verification def CHK_RouterRootRedirect {
        doc /* The root on the one port: a redirect to the viewer, read without following it. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "router-root-redirect";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action getTheRootWithoutFollowing { doc /* GET the root with redirects not followed. */ }
        action expectARedirectToTheViewer { doc /* Status 302 and a location header equal to /viewer/. */ }
    }

    verification def CHK_RouterHealthNotProxied {
        doc /* The router's health path is not one of the four: a 404 on the one port is the pass. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "router-health-not-proxied";
        attribute kind : CheckKind = CheckKind::ApiCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/health/ready";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/router.check.ts"; }
        action getTheHealthPath { doc /* GET the router's readiness path, with the check set to pass on an error status. */ }
        action expectNotFound { doc /* Status 404. */ }
    }

    // ------------------------------------------------ the refusals walk --

    verification def CHK_Refusals {
        doc /* Two edits the router refuses, with the model version read before and after to show that nothing moved. Four steps, one request each, and nothing to put back. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective {
            verify US_15_AServedModelThatIsNeverSilentlyWrong.acceptance.boundValuesStayBound;
            verify SR_24_ExpressionBoundValuesAreReadOnly.acceptance.boundValueIsNotEdited;
            verify SR_25_RejectionOfInvalidValues.acceptance.invalidValuesAreRejected;
        }
        attribute logicalId : String = "refusals";
        attribute kind : CheckKind = CheckKind::MultiStepCheck;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1", "eu-west-2", "us-east-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/graphql";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/refusals.multistep.spec.ts"; }
        action readTheVersion { doc /* Read the model version and keep it. */ }
        action refuseADerivedLimit { doc /* Send setLimit for PIPE-R1.3 at 900 and expect the first error to say the value is not a literal in the source, with no data. */ }
        action refuseANegativeThroughput { doc /* Send setAttribute for PIPE-S1's throughput at -5 and expect the first error to say the value must be a finite, non-negative number, with no data. */ }
        action readTheVersionAgain { doc /* Read the model version again and expect the one kept. */ }
    }

    // ------------------------------------------ the session's monitors --

    verification def CHK_MonitorViewer {
        doc /* The viewer answers, every ten minutes. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "monitor-viewer";
        attribute kind : CheckKind = CheckKind::UrlMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/viewer/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action getTheViewer { doc /* GET the viewer's path and expect status 200. */ }
    }

    verification def CHK_MonitorDocument {
        doc /* The document answers, every ten minutes. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "monitor-document";
        attribute kind : CheckKind = CheckKind::UrlMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/document/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action getTheDocument { doc /* GET the document's path and expect status 200. */ }
    }

    verification def CHK_MonitorRoot {
        doc /* The root answers through its redirect, every ten minutes. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_04_FourPathsOnOnePort.acceptance; }
        attribute logicalId : String = "monitor-root";
        attribute kind : CheckKind = CheckKind::UrlMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "/";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action getTheRootFollowingTheRedirect { doc /* GET the root with redirects followed and expect status 200. */ }
    }

    verification def CHK_SessionHeartbeat {
        doc /* The session script is still running: a ping every ten minutes, five minutes of grace, and then a missed ping is a failed session. A heartbeat runs on no schedule and from no location, so the case carries neither. */
        subject runner : Federation_LogicalArchitecture::CheckSession::SessionRunner;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        attribute logicalId : String = "session-heartbeat";
        attribute kind : CheckKind = CheckKind::HeartbeatMonitor;
        attribute deployed : Boolean = true;
        attribute mutating : Boolean = false;
        attribute target : String = "the ping URL";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action expectAPingEveryTenMinutes { doc /* Expect a ping on the monitor's URL within each ten-minute period plus five minutes of grace. */ }
    }

    // ------------------------------- the public hostname, when one is set --

    verification def CHK_DemoCertificate {
        doc /* The certificate the edge presents for the public hostname is valid, with an alert two weeks before it expires. Deployed only when a hostname is set. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        attribute logicalId : String = "demo-certificate";
        attribute kind : CheckKind = CheckKind::SslMonitor;
        attribute frequencyMinutes : Integer = 60;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : String = "whenHostnameSet";
        attribute mutating : Boolean = false;
        attribute target : String = "DEMO_HOSTNAME";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action readTheCertificate { doc /* Complete a TLS handshake with the hostname on port 443 and expect a valid certificate more than 14 days from expiry. */ }
    }

    verification def CHK_DemoDns {
        doc /* The public hostname resolves. The tunnel's CNAME is flattened at the edge, so an A query is what answers. Deployed only when a hostname is set. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        attribute logicalId : String = "demo-dns";
        attribute kind : CheckKind = CheckKind::DnsMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : String = "whenHostnameSet";
        attribute mutating : Boolean = false;
        attribute target : String = "DEMO_HOSTNAME";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action queryTheARecord { doc /* Query the A record of the hostname and expect an answer that is not empty. */ }
    }

    verification def CHK_DemoTcp {
        doc /* Port 443 of the public hostname accepts a connection over IPv4. Deployed only when a hostname is set. */
        subject demo : Demo;
        actor checkly : Federation_Context::CheckSession::ChecklyCloud;
        objective { verify SR_48_AnOptionalCheckSession.acceptance; }
        attribute logicalId : String = "demo-tcp";
        attribute kind : CheckKind = CheckKind::TcpMonitor;
        attribute frequencyMinutes : Integer = 10;
        attribute locations : String[1..3] = ("eu-central-1");
        attribute deployed : String = "whenHostnameSet";
        attribute mutating : Boolean = false;
        attribute target : String = "DEMO_HOSTNAME:443";
        @Evidence { kind = "checkly"; location = "checkly/__checks__/monitors.check.ts"; }
        action connectToPort443 { doc /* Open a TCP connection to the hostname on port 443 over IPv4 and expect it to be accepted. */ }
    }
}
