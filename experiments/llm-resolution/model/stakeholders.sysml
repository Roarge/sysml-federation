package Experiment_Stakeholders {
    doc /* Who the experiment is for. The operator runs it, and the other three read
           what it produces, each with a different question. */

    part def Operator {
        doc /* Human actor. The person who runs the experiment on a language
               model server in their own network and hands back the results
               file. For this experiment, the author of the articles. */
    }

    part def Evaluator {
        doc /* Human actor. A reader who judges how well a resolver does: expert
               in linking records and in measuring that, and sceptical of any
               score given without the answers it was measured against. */
    }

    part def SystemsEngineer {
        doc /* Human actor. A reader who would have to live with the links in a
               systems model, and answer to an assessor for each of them. */
    }

    part def ModelOwner {
        doc /* Human actor. The person responsible for the systems model, who
               decides where it may be read and by what. */
    }
}
