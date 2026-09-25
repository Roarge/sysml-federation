package Experiment_Core {
    doc /* The language model experiment: whether a language model running on
           the operator's own server can recover the trace links between the
           demo's systems model and its Go tests once the keys are hidden, and
           whether the model's explanations name what its answers depended on.
           It is modelled apart from the demo because it is no part of the
           demo, and it reads the demo's registers as data (AD-0032). The
           definitions of the demo's library are reused, so a story, a
           constraint, a decision record and a piece of evidence are written
           here as they are there. See the README beside this folder. */
    private import VSE_Library::*;
    private import Federation_Library::*;
    private import Experiment_Stakeholders::*;
    private import Experiment_Concerns::*;
    private import Experiment_Context::*;
    private import Experiment_StakeholderStories::*;
    private import Experiment_SystemStories::*;
    private import Experiment_DesignConstraints::*;
    private import Experiment_UseCases::*;
    private import Experiment_LogicalArchitecture::*;
    private import Experiment_VerificationCases::*;
    private import Experiment_ValidationCases::*;
}
