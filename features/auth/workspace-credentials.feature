Feature: Workspace-scoped Bitbucket credentials

  Bitbucket pull request operations must use credentials stored for the selected workspace without reusing credentials for an Atlassian cloud site.

  @authn @security @application
  Scenario: An operator stores site and workspace credentials independently
    Given Atlas has one configured cloud site and one default Bitbucket workspace
    When the operator logs in to the site
    And the operator logs in to the Bitbucket workspace
    Then authentication status reports the site as usable
    And authentication status reports the workspace as usable
    And authentication status contains no email or token values

  @authn @security @regression
  Scenario: A licensed site credential cannot authorize a Bitbucket request
    Given the operator has a usable credential for a licensed cloud site
    And the selected Bitbucket workspace has no stored credential
    When the operator reads a pull request from that workspace
    Then the command fails with an authentication error
    And the error tells the operator to log in to that workspace
    And no Bitbucket request is sent

  @authn @security @application
  Scenario: A Bitbucket request uses the credential for its selected workspace
    Given the operator has different credentials for two Bitbucket workspaces
    When the operator reads a pull request from the second workspace
    Then the request is authenticated with the second workspace credential
    And no cloud-site credential is used

  @authn @security @application
  Scenario: Targeted logout preserves the other credential namespace
    Given the operator has a usable cloud-site credential
    And the operator has a usable Bitbucket workspace credential
    When the operator logs out of the Bitbucket workspace
    Then the cloud site remains usable
    And the Bitbucket workspace is not usable

  @authn @security @regression
  Scenario Outline: An empty targeted logout selector cannot clear the session
    Given the operator has a usable cloud-site credential
    And the operator has a usable Bitbucket workspace credential
    When the operator requests logout with an empty <selector> selector
    Then the command fails with a usage error
    And the cloud site remains usable
    And the Bitbucket workspace remains usable

    Examples:
      | selector  |
      | site      |
      | workspace |

  @authn @security @regression
  Scenario: An unexpected logout argument cannot clear the session
    Given the operator has a usable cloud-site credential
    And the operator has a usable Bitbucket workspace credential
    When the operator requests logout with an unexpected positional argument
    Then the command fails with a usage error
    And the cloud site remains usable
    And the Bitbucket workspace remains usable

  @authn @security @application
  Scenario: A workspace-only session is usable through MCP status
    Given the operator has a usable Bitbucket workspace credential and no cloud-site credential
    When an agent reads Atlas status
    Then the session is reported as signed in and usable
    And the workspace is reported as usable
    And the status contains no email or token values
