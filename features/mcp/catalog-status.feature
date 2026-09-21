Feature: MCP catalog and status
  Scenario: compact catalog
    Given an MCP server
    When the client lists MCP tools
    Then the MCP catalog is compact
  Scenario: status signed out
    Given an MCP server
    When the client calls MCP status
    Then MCP status is signed out
  Scenario: prompts
    Given an MCP server
    When the client lists MCP prompts
    Then the prompts are jira-search, confluence-write, pr-review, jsm-customer
