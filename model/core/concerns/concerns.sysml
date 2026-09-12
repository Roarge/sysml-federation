package Federation_Concerns {
    doc /* One concern per stakeholder question, in the words the architecture
           description asks them in. Every concern takes a DemoElement as its
           subject, so that a story framing one conforms whatever part of the
           demo the story is about. */
    private import Federation_Stakeholders::*;
    private import Federation_Context::DemoElement;
    private import Federation_Library::InformationalOnly;

    concern def LaunchAndSurprise {
        doc /* Can I launch it, and does it show me something I would not have
               believed? */
        subject system : DemoElement;
        stakeholder s : Visitor;
    }

    concern def ModelStaysTheSource {
        doc /* Does the model stay the source, and does an edit anywhere land
               in it? */
        subject system : DemoElement;
        stakeholder s : ModelOwner;
    }

    concern def ShapeWithoutTouching {
        doc /* Can I shape the document without touching the model, and still
               see what the model knows? */
        subject system : DemoElement;
        stakeholder s : DocumentOwner;
    }

    concern def ReadBuildChange {
        @InformationalOnly;
        doc /* Can I read it in an afternoon, build it with Go alone, and
               change one service without the others? No story frames this
               concern. It is carried by the design constraints SC-01, SC-05
               and SC-06, which frame it. */
        subject system : DemoElement;
        stakeholder maintainer : Maintainer;
        stakeholder contributor : Contributor;
    }

    concern def AgreeAndReplace {
        doc /* What do the services have to agree on, and what would I have to
               replace? */
        subject system : DemoElement;
        stakeholder s : AdoptingOrganisation;
    }

    concern def PublishedClaimsStayTrue {
        doc /* Does what the repository says about the demo stay true of what
               ships? Its source is the account of what shipped and what did
               not, together with the decision that the line figures are
               estimates and never a limit (AD-0028). */
        subject system : DemoElement;
        stakeholder s : Maintainer;
    }
}
