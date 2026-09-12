package Federation_Context {
    doc /* The system of interest is the demo, and everything else here is
           something it touches or something the project does to it. The
           repository, the demo's own model and the check project all act on
           the demo from outside it, which is why they sit beside the browser
           and the registry rather than inside the box.

           The example model file sits at the edge and is easy to mistake for
           the inside. It is content the adapter reads, authored in the
           repository rather than produced by it, and it is ExampleModel in
           Federation_LogicalArchitecture rather than an element declared
           here. The pipeline it describes is not a running system anywhere.

           Environmental impact, none, because a demo on one machine affects
           nothing beyond it. */
    private import Federation_Stakeholders::*;
    private import Federation_Domain::*;
    private import Federation_InterfaceTypes::*;
    private import ScalarValues::*;

    abstract part def DemoElement {
        doc /* An element the requirements are allocated to. */
    }

    part def Demo :> DemoElement {
        doc /* The demo: three subgraphs, one router, one UI server and two
               web apps, all in one container started by one command,
               docker run --rm -p 8080:8080 ghcr.io/roarge/sysml-federation.
               The pipeline the example describes is not running anywhere. The
               allocation element IMAGE of the requirements scheme is this
               whole. */
        port web : HttpServerPort;
    }

    // ------------------------------------------------- external systems --

    part def Browser {
        doc /* External system. The visitor's browser, two tabs or two
               visitors, a desktop browser. */
        port http : HttpClientPort;
    }

    part def DockerHost {
        doc /* External system. The Docker host, Linux, macOS or Windows,
               publishing port 8080. */
        port pull : ImageClientPort;
    }

    part def ContainerRegistry {
        doc /* External system. The GitHub container registry,
               ghcr.io/roarge/sysml-federation, a public package, two
               platforms. */
        port image : ImagePort;
    }

    part def MaintainerMachine {
        doc /* External system. The maintainer's machine: Go, Node for wgc and
               Docker. It pushes a v* tag, and it composes config.json and
               commits it. */
        port tags : TagPort;
    }

    part def GitHubActions {
        doc /* External system. The workflows, which build and publish the
               image on the tag, run the unit tests, and validate the demo's
               own model. */
        port trigger : TagPort;
    }

    part def Repository :> DemoElement {
        doc /* Project element. The repository as a thing that is built, tested
               and published: the build, the tests and the workflows. The
               allocation element REPO of the requirements scheme. */
    }

    part def DemoModel :> DemoElement {
        doc /* Project element. The model under model/, read by the two
               validators rather than by the demo. The allocation element
               DEMOMODEL of the requirements scheme. */
    }

    part def CheckSuite :> DemoElement {
        doc /* Project element. The optional check project, which drives the
               demo through a tunnel. The allocation element CHECKS of the
               requirements scheme. It lands in a later pull request. */
    }

    // --------------------------------------------- environmental effect --

    part def Network {
        doc /* Environmental effect. The host's route to the internet, present
               or withheld by --network none. The demo is built to make no
               connection outside the container while it runs, and a run with
               the route withheld is what shows it. */
        attribute reachable : Boolean;
    }

    // -------------------------------------------------- the check session --

    package CheckSession {
        doc /* The session externals. These appear in no published board and
               are exposed only by the session view, so that the boards the
               articles carry stay as they were drawn. */

        part def ChecklyCloud {
            doc /* External system. The monitoring service: its API, its check
                   runners in the public locations, its trace ingest and its
                   alert delivery. */
            port api : HttpServerPort;
            port runners : HttpClientPort;
            port ingest : OtlpPort;
            port alerts : AlertClientPort;
        }

        part def TunnelEdge {
            doc /* External system. The edge that terminates the public
                   hostname and hands each request to the connector beside the
                   demo. */
            port publicSide : HttpServerPort;
            port origin : HttpClientPort;
        }

        part def AlertReceiver {
            doc /* External system. Where an alert lands: the owner's mailbox
                   and the owner's webhook endpoint. */
            port email : AlertPort;
            port webhook : AlertPort;
        }
    }

    // ------------------------------------------------------ the composite --

    part def DemoContext {
        doc /* V1: the demo among the things it touches. Two crossings of the
               boundary are not port connections and are recorded here instead.
               The Docker host runs the container once, with docker run, and it
               is the maintainer's machine rather than the container that
               composes config.json with wgc and commits it. */
        part demo : Demo;
        part visitor : Visitor;
        part modelOwner : ModelOwner;
        part documentOwner : DocumentOwner;
        part browser : Browser;
        part dockerHost : DockerHost;
        part registry : ContainerRegistry;
        part maintainerMachine : MaintainerMachine;
        part actions : GitHubActions;
        part repository : Repository;
        part demoModel : DemoModel;
        part checkSuite : CheckSuite;
        part network : Network;
        interface pages : HttpOnOnePort connect browser.http to demo.web;
        interface pull : ImagePull connect registry.image to dockerHost.pull;
        interface publish : TagPublish connect maintainerMachine.tags to actions.trigger;
    }
}
