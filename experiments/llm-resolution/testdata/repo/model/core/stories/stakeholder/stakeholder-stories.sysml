package Fixture_StakeholderStories {
    private import Fixture_Library::*;
    private import Fixture_SystemStories::*;

    requirement <'US-01'> US_01_Search : UserStory {
        doc /* A reader searches and gets a ranked page back. */
    }

    #derivation connection us01Derives {
        end #original ::> US_01_Search;
        end #derive   ::> SR_01_ParseQueries;
        end #derive   ::> SR_02_ServeResults;
    }
}
