Feature: Pipeline contract
  Scenario: Bootstrap and re-apply produces no churn
    Given a fresh repository initialized by Moltark
    When the repository is applied
    And a follow-up plan is executed
    Then the plan reports no pending changes

  Scenario: Drift in owned fields is surfaced by doctor
    Given a repository with drifted owned fields
    When doctor is executed
    Then doctor reports drift detected
