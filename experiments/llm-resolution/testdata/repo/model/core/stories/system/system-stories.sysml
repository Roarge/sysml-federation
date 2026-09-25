package Fixture_SystemStories {
    requirement <'SR-01'> SR_01_ParseQueries : UserStory {
        attribute statement : String = "The parser shall reject an empty query and parse a well-formed one into tokens.";
    }
    requirement <'SR-02'> SR_02_ServeResults : UserStory {
        attribute statement : String = "When a query is parsed, the server shall return the \"top\" results as a ranked page.";
    }
    requirement <'SR-03'> SR_03_ReportLatency : UserStory {
        attribute statement : String = "The server shall report the latency of every page it returns.";
    }
}
