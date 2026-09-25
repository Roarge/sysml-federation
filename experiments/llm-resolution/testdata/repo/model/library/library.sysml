package Fixture_Library {
    doc /* The definitions the fixture model uses. */
    requirement def UserStory;
    requirement def DesignConstraint;
    metadata def Evidence {
        attribute kind : ScalarValues::String;
        attribute location : ScalarValues::String;
    }
    metadata def DecisionRecord {
        attribute id : ScalarValues::String;
        attribute title : ScalarValues::String;
    }
}
