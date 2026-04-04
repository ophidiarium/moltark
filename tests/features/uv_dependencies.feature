Feature: Preserve uv-managed dependency edits
  Scenario: uv-managed dependencies are preserved during re-apply
    Given a Python repository bootstrapped by Moltark
    When Moltark plan is executed
    Then no dependency drift is reported
    And Moltark apply makes no dependency changes

  Scenario: uv workspace members are derived from parent-relative project paths
    Given a repository declaring a nested uv workspace
    When Moltark apply is executed
    Then the root project writes uv workspace members relative to its path
