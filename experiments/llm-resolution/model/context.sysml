package Experiment_Context {
    doc /* The experiment among what it touches. It reads a checkout of the
           repository, asks a language model server in the operator's network,
           and writes one results file that the operator hands on. The server
           and the checkout are outside it, and the language model on the server is the
           thing being examined, and nothing of the experiment runs there. */
    private import Experiment_Stakeholders::*;
    private import ScalarValues::*;

    abstract part def ExperimentElement {
        doc /* Anything a story or a concern of the experiment is about. */
    }

    // ------------------------------------------------------------- items --

    item def RequirementText {
        doc /* A requirement's key, name and statement, as read from the demo's
               registers. */
    }

    item def SystemsModelText {
        doc /* The demo's systems model as the experiment reads it: every
               element with its kind, description and attributes, and every
               relationship between them. */
    }

    item def CodeText {
        doc /* The built system's files in the checkout, as the code tools
               search and read them. */
    }

    item def IncidentReport {
        doc /* What the engineer on call says is down: the alert, and the
               container's last line when it is known. */
    }

    item def TestText {
        doc /* A Go test's name, package, file and doc comment, with any
               requirement key removed. */
    }

    item def Question {
        doc /* One chat request: the instructions for the task as the system
               message, and the task, the viewpoint, the view, the calls left
               and the last tool's answer as the user message, with the
               settings and the reply schema. */
    }

    item def Answer {
        doc /* One reply: a tool call with edits to the view, or a final answer
               citing elements, links and files, with the server's timings. */
    }

    item def ResultsFile {
        doc /* One JSON Lines file: the run's header, every question and reply,
               and the summary. */
    }

    // ------------------------------------------------------------- ports --

    port def ChatClientPort {
        doc /* The asking side of the chat endpoint. */
        out item questions : Question;
        in item answers : Answer;
    }

    port def ChatServerPort {
        doc /* The answering side of the chat endpoint, the conjugate written
               out. */
        in item questions : Question;
        out item answers : Answer;
    }

    interface def ChatOverHttp {
        doc /* The experiment to the model server, Ollama's /api/chat over HTTP
               in the operator's network. */
        end client : ChatClientPort;
        end server : ChatServerPort;
    }

    // ------------------------------------------------------ the elements --

    part def Experiment :> ExperimentElement {
        doc /* The experiment: a Go program in its own module under
               experiments/llm-resolution, and the script that runs it. */
        port chat : ChatClientPort;
        out item results : ResultsFile;
    }

    part def ModelServer {
        doc /* External. An Ollama server in the operator's network, holding the
               language model under test. */
        port chat : ChatServerPort;
        attribute model : String;
        attribute numCtx : Integer;
    }

    part def RepositoryCheckout {
        doc /* External. The operator's checkout of this repository, at the
               commit the run records. */
        out item requirements : RequirementText[1..*];
        out item tests : TestText[1..*];
        out item systemsModel : SystemsModelText;
        out item code : CodeText;
    }

    part def ExperimentContext {
        doc /* The experiment among the things it touches. */
        part experiment : Experiment;
        part server : ModelServer;
        part checkout : RepositoryCheckout;
        part operator : Operator;
        part onCall : OnCallEngineer;
        interface chat : ChatOverHttp connect experiment.chat to server.chat;
    }
}
