#!/usr/bin/env bash

set -euo pipefail

# Fetch open PRs with checks and mergeability
prs_json=$(gh pr list \
  --state open \
  --json number,title,author,isDraft,mergeable,statusCheckRollup)

# Allowed bot authors across different GitHub API representations
# (e.g. "app/dependabot" in GraphQL vs "dependabot[bot]" in REST/webhooks)
allowed_bots='["app/dependabot", "dependabot[bot]", "app/renovate", "renovate[bot]", "app/samuebot", "samuebot[bot]"]'

# NOTE: Security must rely primarily on verified bot logins (author.login) and CI status checks.
# Do not place full trust in PR title parsing alone, as title-based filters can be fragile.
# TODO: Relying solely on semver bump types is imperfect. Consider introducing an LLM-based diff review in the future.
target_prs=$(echo "$prs_json" | jq -r --argjson bots "$allowed_bots" '
  def is_disallowed_dependabot_major($title):
    ($title | capture("^Bump (?<pkg>.+) from (?<old>[^ ]+) to (?<new>[^ ]+)$")) as $m |
    if ($m.pkg | test("cachix")) then
      false
    else
      (($m.old | ltrimstr("v") | split(".")[0]) != ($m.new | ltrimstr("v") | split(".")[0]))
    end;

  [
    .[] |
    select(
      (.author.login as $a | $bots | index($a)) and
      (.isDraft | not) and
      (.mergeable == "MERGEABLE") and
      ((.statusCheckRollup | length) > 0) and
      (.statusCheckRollup | all(.status == "COMPLETED")) and
      (.statusCheckRollup | all(.conclusion == "SUCCESS")) and
      (
        if (.author.login | test("dependabot")) then
          (try is_disallowed_dependabot_major(.title) catch false) | not
        else
          true
        end
      )
    )
  ] | .[].number
')

if [ -z "$target_prs" ]; then
  echo "No eligible bot PR found for merge."
  exit 0
fi

# Do not use --auto: all status checks have already passed, and direct merge
# avoids depending on repository branch protection settings.
for pr in $target_prs; do
  echo "Attempting to merge PR #$pr..."
  if gh pr merge --squash --delete-branch "$pr"; then
    echo "Successfully merged PR #$pr."
  else
    echo "Failed to merge PR #$pr (possibly conflicting with a prior merge). Skipping for next run."
  fi
done
