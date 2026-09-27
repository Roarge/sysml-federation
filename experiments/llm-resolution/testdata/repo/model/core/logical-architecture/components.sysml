package Fixture_LogicalArchitecture {
    doc /* What the fixture system is made of. */
    private import Fixture_Library::*;
    private import Fixture_SystemStories::*;
    private import Fixture_DesignConstraints::*;
    private import Fixture_Behaviour::*;

    part def Parser {
        doc /* The parser, in adapter/parse. The quokka route is not its
               business. */
    }

    part def Server {
        doc /* The query server, in adapter/serve/serve.go. It listens on its
               own address and serves one ranked page per query. */
        @DecisionRecord { id = "AD-0001"; title = "One server for every query"; }
        attribute listenAddress : ScalarValues::String = "127.0.0.1:4011";
        ref part parserChild : Parser;
        exhibit state lifecycle : ServerStates;
    }

    part system {
        doc /* The system as built. */
        part server : Server;
        part parser : Parser;
        bind server.parserChild = parser;
        perform action answer : AnswerAQuery;
        allocate answer.parse to parser;
        allocate answer.serve to server;
        satisfy SR_01_ParseQueries by parser;
        satisfy SR_02_ServeResults by server;
        satisfy SR_03_ReportLatency by server;
        satisfy SC_01_StandardLibraryOnly by system;
    }
}
