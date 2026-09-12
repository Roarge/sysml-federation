package Federation_InterfaceTypes {
    doc /* The port and interface types every later register connects with.
           A port definition says what crosses it and in which direction, and
           the conjugate direction is written out as its own definition rather
           than left to be inferred. An interface definition names its two ends
           and the payload they carry. */
    private import Federation_Domain::*;

    // ------------------------------------------------------------- ports --

    port def HttpServerPort {
        doc /* The listening side of an HTTP endpoint. It accepts requests and
               answers with pages. */
        in item requests : GraphQLQuery;
        out item pages : Page;
    }

    port def HttpClientPort {
        doc /* The calling side of an HTTP endpoint, the conjugate of
               HttpServerPort written out in full. */
        out item requests : GraphQLQuery;
        in item pages : Page;
    }

    port def GraphQLServerPort {
        doc /* The listening side of a GraphQL endpoint. It accepts queries and
               publishes version events on a subscription. */
        in item queries : GraphQLQuery;
        out item events : VersionEvent;
    }

    port def GraphQLClientPort {
        doc /* The calling side of a GraphQL endpoint, the conjugate of
               GraphQLServerPort written out in full. */
        out item queries : GraphQLQuery;
        in item events : VersionEvent;
    }

    port def OtlpPort {
        doc /* Where spans are received, OTLP over HTTP. */
        in item spans : Span;
    }

    port def OtlpClientPort {
        doc /* Where spans are sent from, the conjugate of OtlpPort. */
        out item spans : Span;
    }

    port def TunnelPort {
        doc /* The connector's end of a tunnel. What crosses it is whatever the
               edge received, so it accepts requests and answers with pages the
               way a listening endpoint does. */
        in item requests : GraphQLQuery;
        out item pages : Page;
    }

    port def AlertPort {
        doc /* Where an alert is delivered, a mailbox or a webhook endpoint. */
        in item alerts : Alert;
    }

    port def AlertClientPort {
        doc /* Where an alert is sent from, the conjugate of AlertPort. */
        out item alerts : Alert;
    }

    port def ImagePort {
        doc /* Where an image is served from, the registry's side of a pull. */
        out item images : ContainerImage;
    }

    port def ImageClientPort {
        doc /* Where an image is pulled to, the host's side of a pull. */
        in item images : ContainerImage;
    }

    port def TagPort {
        doc /* Where a version tag is pushed and where one is received. The tag
               is a git reference rather than an item the model carries, so the
               definition names no payload. */
    }

    // -------------------------------------------------------- interfaces --

    interface def HttpOnOnePort {
        doc /* Browser to user interface server on the one published port,
               carrying pages, GraphQL and the subscription stream. */
        end browser : HttpClientPort;
        end server : HttpServerPort;
    }

    interface def RouterToSubgraph {
        doc /* Router to one subgraph, GraphQL over HTTP with subscriptions
               over WebSocket. */
        end router : GraphQLClientPort;
        end subgraph : GraphQLServerPort;
    }

    interface def UiProxy {
        doc /* User interface server to router, proxying /graphql and
               /playground so that the router is never addressable on its
               own. */
        end ui : GraphQLClientPort;
        end router : GraphQLServerPort;
    }

    interface def ImagePull {
        doc /* Registry to Docker host, carrying the container image once. */
        end registry : ImagePort;
        end host : ImageClientPort;
    }

    interface def TagPublish {
        doc /* The maintainer's machine to the workflow, carrying the version
               tag whose push publishes the image. */
        end machine : TagPort;
        end actions : TagPort;
    }

    interface def CheckTraffic {
        doc /* The monitoring service's runners to the tunnel edge, HTTP
               carrying a TraceContext so that a check run is recognisable in
               the trace. */
        end runners : HttpClientPort;
        end edge : HttpServerPort;
    }

    interface def TunnelLink {
        doc /* The edge to the connector beside the demo, carrying whatever the
               public hostname received. */
        end edge : HttpClientPort;
        end tunnelConnector : TunnelPort;
    }

    interface def TraceExport {
        doc /* Router to collector, OTLP over HTTP, carrying spans. */
        end router : OtlpClientPort;
        end collector : OtlpPort;
    }

    interface def TraceForward {
        doc /* Collector onward, carrying spans to the viewer and to the
               monitoring service's ingest. */
        end collector : OtlpClientPort;
        end sink : OtlpPort;
    }

    interface def ProjectControl {
        doc /* A runner to the monitoring service's API, carrying a
               ProjectCommand. */
        end runner : HttpClientPort;
        end api : HttpServerPort;
    }

    interface def Heartbeating {
        doc /* A runner to the monitoring service's API, carrying the
               HeartbeatPing that says a scheduled job ran. */
        end runner : HttpClientPort;
        end api : HttpServerPort;
    }

    interface def Alerting {
        doc /* The monitoring service to a receiver, carrying an alert to a
               mailbox or a webhook endpoint. */
        end service : AlertClientPort;
        end receiver : AlertPort;
    }
}
