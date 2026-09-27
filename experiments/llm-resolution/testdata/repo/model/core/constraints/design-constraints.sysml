package Fixture_DesignConstraints {
    requirement <'SC-01'> SC_01_StandardLibraryOnly : DesignConstraint {
        attribute :>> statement = "The code imports nothing but the standard library.";
    }
}
