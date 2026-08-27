package <'WH'> Warehouse {
    doc /* A packing line: two entry stations feed one exit. */

    private import ScalarValues::Real;
    private import ISQ::*;
    private import SI::*;
    private import RequirementDerivation::*;

    abstract part def Station {
        attribute capacity : Real;
    }

    item def Parcel;

    port def ParcelIn { in item parcels : Parcel; }
    port def ParcelOut { out item parcels : Parcel; }

    part def Robot :> Station {
        attribute rate : Real;
        port take : ParcelIn;
        port give : ParcelOut;
    }

    part def Line :> Station {
        attribute cycleTime : DurationValue;
        attribute shifts : Real;
    }

    requirement def RateRequirement {
        doc /* The subject shall handle the required parcel rate. */
        subject station : Station;
        attribute required : Real;
        require constraint { required <= station.capacity }
    }

    requirement def CycleRequirement {
        subject station : Line;
        require constraint { station.cycleTime <= 30[s] }
    }

    part <'WH-L1'> line : Line {
        doc /* pick and label are entries, ship is the exit. */
        attribute :>> shifts = 2;
        part <'WH-A'> pick : Robot { attribute :>> rate = 40; }
        part <'WH-B'> pack : Robot { attribute :>> rate = 30; }
        part <'WH-C'> label : Robot { attribute :>> rate = 25; }
        part <'WH-D'> ship : Robot { attribute :>> rate = 50; }
        connect pick.give to pack.take;
        connect pack.give to ship.take;
        connect label.give to ship.take;
        satisfy lineRate by line;
        satisfy shipRate by ship;
        satisfy packRate by pack;
        satisfy cycleLimit by line;
    }

    requirement <'WH-R1'> lineRate : RateRequirement {
        doc /* The line shall handle the required parcel rate */
        subject station :> line;
        attribute :>> required = 60;
    }

    requirement <'WH-R1.1'> shipRate : RateRequirement {
        subject station :> line.ship;
        attribute :>> required = (lineRate.required + 20) / 2;
    }

    requirement packRate : RateRequirement {
        subject station :> line.pack;
        attribute :>> required = lineRate.required - 10 * 3;
    }

    requirement <'WH-R2'> cycleLimit : CycleRequirement {
        doc /* A cycle shall finish within the limit */
        subject station :> line;
    }

    #derivation connection {
        end #original ::> lineRate;
        end #derive ::> shipRate;
        end #derive ::> packRate;
    }

    verification def CycleTest {
        subject station : Line;
        objective { verify cycleLimit; }
    }

    verification <'WH-VC1'> cycleTest : CycleTest {
        subject station :> line;
    }
}
