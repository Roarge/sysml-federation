package Federation_Views {
    doc /* The views register: one view for every board and sheet the
           repository publishes, and the viewpoint each of them answers. A
           viewpoint frames the concerns its readers arrive with, a view
           definition says what the drawing looks like, and a view names the
           elements it draws and the image files it corresponds to. Rendering
           is not declared here. The images are drawn by hand and the model
           says which elements each one is about. See model/README.md. */
    private import Federation_Concerns::*;
    private import Federation_Context;
    private import Federation_LogicalArchitecture;
    private import Federation_FunctionalArchitecture;
    private import Federation_StakeholderStories;
    private import Federation_Stakeholders;
    private import Federation_InterfaceTypes;
    private import Federation_VerificationCases;

    // ------------------------------------------------------- viewpoints --
    // One viewpoint per goal question, in the words the board or the sheet
    // asks it in.

    viewpoint def ContextViewpoint {
        doc /* What is inside, what is outside, what crosses the line? */
        frame concern : LaunchAndSurprise;
    }

    viewpoint def CompositionViewpoint {
        doc /* What do the services agree on? */
        frame concern : ModelStaysTheSource;
        frame concern : ShapeWithoutTouching;
        frame concern : ReadBuildChange;
        frame concern : AgreeAndReplace;
    }

    viewpoint def RuntimeViewpoint {
        doc /* How does a change in one tab reach the other? */
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        frame concern : ShapeWithoutTouching;
    }

    viewpoint def DeploymentViewpoint {
        doc /* What runs where, and how does it get built? */
        frame concern : ReadBuildChange;
        frame concern : AgreeAndReplace;
    }

    viewpoint def AdapterViewpoint {
        doc /* How does a model file become a subgraph? */
        frame concern : ModelStaysTheSource;
        frame concern : ReadBuildChange;
    }

    viewpoint def NewcomerViewpoint {
        doc /* What does this demo claim, what is in the box, and what would I
               keep or replace if I adopted it? Why does raising one server
               change nothing and raising another change everything? */
        frame concern : LaunchAndSurprise;
        frame concern : AgreeAndReplace;
    }

    viewpoint def StoryboardViewpoint {
        doc /* What should the visitor experience, in order? */
        frame concern : LaunchAndSurprise;
        frame concern : ModelStaysTheSource;
        frame concern : ShapeWithoutTouching;
    }

    viewpoint def SessionViewpoint {
        doc /* Does what the repository says about the demo stay true of what
               ships, whenever it runs? */
        frame concern : PublishedClaimsStayTrue;
    }

    // -------------------------------------------------- view definitions --
    // What each kind of drawing looks like, and the viewpoint it conforms to.

    view def ContextView {
        doc /* Boxes with name, type and technology, labelled one-directional
               arrows, dashed for what is outside. */
        viewpoint conformsTo : ContextViewpoint;
    }

    view def CompositionView {
        doc /* Nested boxes for what the demo is made of, each service with the
               port it publishes, and the types the three subgraphs have to
               agree on drawn beside them. */
        viewpoint conformsTo : CompositionViewpoint;
    }

    view def RuntimeView {
        doc /* Numbered steps along the arrows, in the order they happen, each
               step carrying the value that travels it, or the screen as it
               stands at one such step. */
        viewpoint conformsTo : RuntimeViewpoint;
    }

    view def DeploymentView {
        doc /* A node for every machine, with the image it runs, and the arrows
               that carry a build from the repository to the host. */
        viewpoint conformsTo : DeploymentViewpoint;
    }

    view def AdapterView {
        doc /* Stages left to right, each labelled with what it reads and what
               it leaves behind. */
        viewpoint conformsTo : AdapterViewpoint;
    }

    view def SheetView {
        doc /* One sheet read at a glance, the model side beside the drawing
               that explains it, sized for print. */
        viewpoint conformsTo : NewcomerViewpoint;
    }

    view def StoryboardView {
        doc /* One frame per story, in the order a visitor meets them, each
               frame carrying the person it belongs to. */
        viewpoint conformsTo : StoryboardViewpoint;
    }

    view def SessionView {
        doc /* Boxes for the services that stand beside the demo, with the
               arrows that carry a request in and a trace back out. */
        viewpoint conformsTo : SessionViewpoint;
    }

    // --------------------------------------------------------- the views --
    // One view per published board or sheet. The image files a view names are
    // the files it corresponds to, and every published image is named once.

    view v1Context : ContextView {
        expose Federation_Context::DemoContext::**;
        expose Federation_LogicalArchitecture::ExampleModel;
        expose Federation_LogicalArchitecture::demo::**;
        expose Federation_Domain::RouterConfiguration;
        doc /* The context board: the demo, the people who use it and the
               machines and services around it.
               docs/img/v1-context.png
               docs/architecture/architecture-views.pdf page 2 */
    }

    view v2Composition : CompositionView {
        expose Federation_LogicalArchitecture::demo::**;
        expose Federation_InterfaceTypes::**;
        expose Federation_Domain::RouterConfiguration;
        doc /* The composition board: the services inside the demo, the ports
               between them, the three subgraph schemas and the one requirement
               the graph merges out of them.
               docs/img/v2-composition.png
               docs/img/v2-subgraph-schemas.png
               docs/img/v2-merged-requirement.png */
    }

    view v3Runtime : RuntimeView {
        expose Federation_FunctionalArchitecture::EditPropagation::**;
        expose Federation_LogicalArchitecture::demo::**;
        doc /* The runtime board: an edit in one tab travelling to the other,
               the same journey for an edit that moves nothing, and the panels
               for a document edit and a reset.
               docs/img/v3-runtime.png
               docs/img/v3-nothing-moves.png */
    }

    view v4Deployment : DeploymentView {
        expose Federation_LogicalArchitecture::demo::**;
        expose Federation_LogicalArchitecture::project::**;
        expose Federation_Context::GitHubActions;
        expose Federation_Context::ContainerRegistry;
        expose Federation_Context::DockerHost;
        doc /* The deployment board: what runs on the host, what the one
               container holds, and the layers the image is built from.
               docs/img/v4-deployment.png
               docs/img/v4-container.png
               docs/img/v4-image-layers.png */
    }

    view v5Adapter : AdapterView {
        expose Federation_LogicalArchitecture::Adapter::**;
        expose Federation_LogicalArchitecture::ExampleModel;
        expose Federation_LogicalArchitecture::Router;
        expose Federation_VerificationCases::VC_SR_45::**;
        doc /* The adapter board: a model file read, projected and served as a
               subgraph.
               docs/img/v5-adapter.png */
    }

    view l0Sheet : SheetView {
        expose Federation_FunctionalArchitecture::TheArgument::**;
        expose Federation_LogicalArchitecture::demo::**;
        expose Federation_Context::ContainerRegistry;
        expose Federation_Context::Browser;
        expose Federation_Context::MaintainerMachine;
        doc /* The L0 sheet: the whole argument on one page, with the legend
               that reads it.
               docs/img/a3-l0-model-side.png
               docs/img/a3-l0-legend.png
               docs/a3/L0-federating-a-systems-model.pdf */
    }

    view l2bSheet : SheetView {
        expose Federation_FunctionalArchitecture::VerdictComputation::**;
        expose Federation_LogicalArchitecture::demo::**;
        doc /* The L2b sheet: the capacity of the wiring, the states it passes
               through, the arithmetic behind each verdict and where it runs,
               with the capacity service holding no copy of the model and
               recomputing on every read.
               docs/img/a3-l2b-model-side.png
               docs/img/a3-l2b-wiring-states.png
               docs/img/a3-l2b-arithmetic.png
               docs/img/a3-l2b-physical.png
               docs/a3/L2b-pipeline-example-capacity-and-verdicts.pdf */
    }

    view storyboard : StoryboardView {
        expose Federation_StakeholderStories::US_01_LaunchWithOneCommand;
        expose Federation_StakeholderStories::US_02_ReadTheModel;
        expose Federation_StakeholderStories::US_03_RaiseAServerThatIsNotTheBottleneck;
        expose Federation_StakeholderStories::US_04_RaiseTheBottleneck;
        expose Federation_StakeholderStories::US_05_TightenTheLimit;
        expose Federation_StakeholderStories::US_06_ReadTheDocument;
        expose Federation_StakeholderStories::US_07_ReorderAndNest;
        expose Federation_StakeholderStories::US_08_ShapeTheDocument;
        expose Federation_StakeholderStories::US_09_ChangeAValueFromTheDocument;
        expose Federation_StakeholderStories::US_10_ChangeFromTheViewerWatchTheDocument;
        expose Federation_StakeholderStories::US_11_QueryTheGraph;
        expose Federation_StakeholderStories::US_12_Reset;
        expose Federation_Stakeholders::Visitor;
        expose Federation_Stakeholders::ModelOwner;
        expose Federation_Stakeholders::DocumentOwner;
        expose Federation_LogicalArchitecture::demo::**;
        doc /* The storyboard: the journey a visitor takes through the twelve
               stories, the three people who take it, and a frame for each
               story that has one.
               docs/img/stories-journey.png
               docs/img/stories-personas.png
               docs/img/us01-launch.png
               docs/img/us03-nothing-moves.png
               docs/img/us04-bottleneck-moves.png
               docs/img/us06-document.png
               docs/img/us09-edit-from-document.png
               docs/img/us10-viewer-to-document.png
               docs/img/us11-query.png
               docs/img/us12-reset.png */
    }

    view fiveViews : SheetView {
        expose Federation_Views::ContextViewpoint;
        expose Federation_Views::CompositionViewpoint;
        expose Federation_Views::RuntimeViewpoint;
        expose Federation_Views::DeploymentViewpoint;
        expose Federation_Views::AdapterViewpoint;
        expose Federation_LogicalArchitecture::demo::**;
        doc /* The five views on one sheet, each with the question it answers,
               and the sketch that introduces the demo before any of them.
               docs/img/architecture-five-views.png
               docs/img/overview-sketch.png */
    }

    view appShots : RuntimeView {
        expose Federation_LogicalArchitecture::ModelViewer;
        expose Federation_LogicalArchitecture::RequirementsDocument;
        expose Federation_FunctionalArchitecture::VerdictComputation::**;
        doc /* The two apps as they appear on screen: the shipped state, the
               bottleneck moved, a passing verdict, an edited limit, a refused
               value, the document tree, the exclusion both apps honour and the
               depth the document is read at.
               docs/img/app-viewer-shipped.png
               docs/img/app-viewer-bottleneck-moved.png
               docs/img/app-viewer-passing.png
               docs/img/app-viewer-limit-edit.png
               docs/img/app-viewer-refusal.png
               docs/img/app-document-tree.png
               docs/img/app-both-exclusion.png
               docs/img/app-document-depth.png */
    }

    view session : SessionView {
        expose Federation_Context::CheckSession::**;
        expose Federation_LogicalArchitecture::CheckSession::**;
        expose Federation_LogicalArchitecture::CheckSession::session::**;
        expose Federation_FunctionalArchitecture::Session::**;
        doc /* The check session beside the demo: the runner, the tunnel, the
               collector, the viewer and the services they talk to. There is no
               published image of this view. */
    }
}
