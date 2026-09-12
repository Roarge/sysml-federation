package VSE_Library {
    doc /* Definitions for writing user stories as requirements: a story is a
           requirement usage typed by UserStory, with a stakeholder role, a
           capability, a benefit and acceptance criteria, and a status carried
           as metadata. */
    private import ScalarValues::*;

    enum def StoryStatus {
        doc /* The lifecycle of a story: backlog, ready, in progress, done. */
        backlog; ready; inProgress; done;
    }

    abstract requirement def StakeholderNeed {
        doc /* The abstract base of every stakeholder-facing requirement. */
    }

    requirement def UserStory :> StakeholderNeed {
        doc /* A user story: as the role, I want the capability, so that the
               benefit. The subject is the system the story is about and is
               declared first. A story redefines subject and role with the
               model's own types, gives capability and benefit as strings, and
               nests one acceptance criterion per separately verifiable
               outcome so that a verification case can name one criterion or
               all of them. */
        subject system;
        stakeholder role;
        attribute capability : ScalarValues::String[0..1];
        attribute benefit    : ScalarValues::String[0..1];
        requirement acceptance[0..*] {
            doc /* One acceptance criterion in given, when, then form. */
        }
    }

    metadata def StoryMeta {
        doc /* The status of a story, applied as @StoryMeta { status = StoryStatus::done; }. */
        attribute status : StoryStatus[1];
    }
}
