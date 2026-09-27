@SR-31
Feature: Two configured names
  As a visitor, I want the capacity service to carry exactly two model-
  specific names and select only declared fields of the generic projection, so
  that what the services have to agree on is a few names and a key rather than
  a shared model.

  @twoNamesAndDeclaredFieldsOnly @inspection
  Scenario: Two names and declared fields only
    Given the capacity service's schema, configuration and source
    When they are read
    Then the only model-specific names are the quantity it computes and the attribute it reads, and every field it selects is declared in its requires clause
