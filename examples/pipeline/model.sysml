package <'PIPE'> QueryPipeline {
    doc /* A query processing pipeline of five servers. Throughput and capacity
         * are query rates in queries per second. The rollup is not in the
         * model: capacity is declared without a value and computed elsewhere. */

    private import ScalarValues::Real;
    private import ISQ::*;
    private import SI::*;
    private import SIPrefixes::*;
    private import MeasurementReferences::*;
    private import RequirementDerivation::*;

    attribute <ms> millisecond : DurationUnit {
        :>> unitConversion : ConversionByPrefix { :>> prefix = milli; :>> referenceUnit = s; }
    }

    abstract part def Component {
        doc /* Anything with a sustainable query rate. */
        attribute capacity : Real;
    }

    item def Query;

    port def QueryInput {
        in item queries : Query;
    }

    port def QueryOutput {
        out item queries : Query;
    }

    part def Server :> Component {
        doc /* A server with a measured throughput in queries per second. */
        attribute throughput : Real;
        port input : QueryInput;
        port output : QueryOutput;
    }

    part def Pipeline :> Component {
        doc /* The whole pipeline. Its latency is declared without a value. */
        attribute latency : DurationValue;
    }

    requirement def ThroughputRequirement {
        doc /* The subject shall sustain at least the required query rate. */
        subject target : Component;
        attribute requiredRate : Real;
        require constraint { target.capacity >= requiredRate }
    }

    requirement def LatencyRequirement {
        doc /* The subject's end-to-end latency shall not exceed the limit. */
        subject target : Pipeline;
        attribute maxLatency : DurationValue;
        require constraint { target.latency <= maxLatency }
    }

    part <'PIPE-P1'> pipeline : Pipeline {
        doc /* Queries enter at ingest, are parsed, indexed on two servers in
             * parallel and served. */
        part <'PIPE-S1'> ingest : Server { attribute :>> throughput = 2000; }
        part <'PIPE-S2'> parse : Server { attribute :>> throughput = 1200; }
        part <'PIPE-S3'> indexA : Server { attribute :>> throughput = 700; }
        part <'PIPE-S4'> indexB : Server { attribute :>> throughput = 700; }
        part <'PIPE-S5'> serve : Server { attribute :>> throughput = 1800; }

        connect ingest.output to parse.input;
        connect parse.output to indexA.input;
        connect parse.output to indexB.input;
        connect indexA.output to serve.input;
        connect indexB.output to serve.input;

        satisfy globalThroughput by pipeline;
        satisfy ingestThroughput by ingest;
        satisfy parseThroughput by parse;
        satisfy indexAThroughput by indexA;
        satisfy indexBThroughput by indexB;
        satisfy serveThroughput by serve;
        satisfy latencyLimit by pipeline;
    }

    requirement <'PIPE-R1'> globalThroughput : ThroughputRequirement {
        doc /* The pipeline shall sustain the required query rate */
        subject target :> pipeline;
        attribute :>> requiredRate = 1500;
    }

    requirement <'PIPE-R1.1'> ingestThroughput : ThroughputRequirement {
        doc /* ingest shall sustain its allocated rate */
        subject target :> pipeline.ingest;
        attribute :>> requiredRate = globalThroughput.requiredRate;
    }

    requirement <'PIPE-R1.2'> parseThroughput : ThroughputRequirement {
        doc /* parse shall sustain its allocated rate */
        subject target :> pipeline.parse;
        attribute :>> requiredRate = globalThroughput.requiredRate;
    }

    requirement <'PIPE-R1.3'> indexAThroughput : ThroughputRequirement {
        doc /* indexA shall sustain its allocated rate */
        subject target :> pipeline.indexA;
        attribute :>> requiredRate = globalThroughput.requiredRate / 2;
    }

    requirement <'PIPE-R1.4'> indexBThroughput : ThroughputRequirement {
        doc /* indexB shall sustain its allocated rate */
        subject target :> pipeline.indexB;
        attribute :>> requiredRate = globalThroughput.requiredRate / 2;
    }

    requirement <'PIPE-R1.5'> serveThroughput : ThroughputRequirement {
        doc /* serve shall sustain its allocated rate */
        subject target :> pipeline.serve;
        attribute :>> requiredRate = globalThroughput.requiredRate;
    }

    requirement <'PIPE-R2'> latencyLimit : LatencyRequirement {
        doc /* End-to-end latency shall not exceed the limit */
        subject target :> pipeline;
        attribute :>> maxLatency = 200[ms];
    }

    #derivation connection {
        end #original ::> globalThroughput;
        end #derive ::> ingestThroughput;
        end #derive ::> parseThroughput;
        end #derive ::> indexAThroughput;
        end #derive ::> indexBThroughput;
        end #derive ::> serveThroughput;
    }

    verification def LatencyTest {
        doc /* Measures end-to-end latency while the required query rate is applied. */
        subject target : Pipeline;
        objective { verify latencyLimit; }
    }

    verification <'PIPE-VC1'> latencyTest : LatencyTest {
        subject target :> pipeline;
    }
}
