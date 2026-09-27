package Fixture_Behaviour {
    doc /* What the fixture system does. */

    action def AnswerAQuery {
        doc /* One query, parsed and then served. */
        action parse { doc /* tokens from the query text */ }
        action serve { doc /* the ranked page */ }
        first parse then serve;
    }

    item def ParserExited { doc /* the parser child ended */ }
    item def StopRequested { doc /* the process was asked to stop */ }

    state def ServerStates {
        doc /* How the server stops, from adapter/serve/serve.go. */
        entry; then serving;
        state serving;
        state stopped;
        transition onParserExit first serving accept ParserExited then stopped;
        transition onStop first serving accept StopRequested then stopped;
    }
}
