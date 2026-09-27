@SR-08
Feature: Licences carried in the repository
  As a maintainer, I want the repository to carry a root NOTICE for the router
  and SortableJS, with the MIT text beside the one vendored file, so that the
  repository is copied without a licence obligation being missed.

  @noticeAndMitTextInPlace
  Scenario: The root's NOTICE names both third-party works, with the MIT text beside the vendored file
    Given the repository
    When its root is read
    Then its NOTICE names the Cosmo router and SortableJS
    And the MIT text of SortableJS sits beside the vendored file, with the vendor's copyright line
