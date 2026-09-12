package Federation_Stakeholders {
    doc /* Who the demo is for and who carries it. Each definition is a
           stakeholder of the architecture description. The three personas the
           boards draw, the visitor, the model owner and the document owner,
           are also the human actors in the context, and they and the
           maintainer are the actors of the use cases. One definition serves
           in every place rather than being written twice. */

    part def Visitor {
        doc /* Human actor. A developer or architect at a small engineering
               firm, evaluating whether federation could connect the tools
               they already have. "They have Docker, a browser and fifteen
               minutes, and they will not read a manual." */
    }

    part def ModelOwner {
        doc /* Human actor. A systems engineer who writes the SysML v2 model,
               wants it to remain the source of truth, and wants to see that
               an edit made anywhere lands in the served model rather than in
               a copy. */
    }

    part def DocumentOwner {
        doc /* Human actor. A requirements engineer who owns the ordering,
               numbering and inclusion decisions of a requirements document,
               and who "has never seen SysML, and must never need to". */
    }

    part def Maintainer {
        doc /* Human actor. The person who reads the whole repository in an
               afternoon, builds it with Go alone and changes one service
               without the others. */
    }

    part def Contributor {
        doc /* Human actor. A person sending a change, with the same concerns
               as the maintainer. */
    }

    part def AdoptingOrganisation {
        doc /* Human actor. An organisation deciding whether the approach
               fits, asking what the services have to agree on and what it
               would have to replace. */
    }
}
