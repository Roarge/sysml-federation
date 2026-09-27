@SR-37
Feature: What a requirement row shows
  As a document owner, I want the document app to show for each included
  requirement what the model knows about it and what the analysis concluded,
  so that a requirements reader reviews both without opening a modelling tool.

  @rowCarriesWhatTheModelKnows @record
  Scenario: Row carries what the model knows
    Given the document is open
    When the row of each requirement of the example is read against the projection
    Then it shows every field the statement lists

  @absentValueShowsNone @record
  Scenario: Absent value shows none
    Given a requirement for which the analysis returns no current value
    When its row is read
    Then the row shows none
