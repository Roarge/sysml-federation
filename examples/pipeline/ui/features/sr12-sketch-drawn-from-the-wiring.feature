@SR-12
Feature: Sketch drawn from the wiring
  As a visitor, I want the viewer to draw the wiring from the model's
  connections, with each throughput, the rolled-up capacity and the bottleneck
  marked on it, so that the bottleneck moving is seen rather than read out of
  numbers.

  @sketchShowsTheWiring @record
  Scenario: Sketch shows the wiring
    Given the shipped model
    When the viewer opens
    Then the sketch shows the servers left to right as wired, each with its throughput, the pipeline's capacity, and the bottleneck servers marked in red

  @secondWiringDrawsTheSameWay @record
  Scenario: Second wiring draws the same way
    Given a second wiring with fan-in from different points
    When the viewer opens
    Then the sketch is drawn from that model's connections in the same way
