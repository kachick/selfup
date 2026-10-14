#!/usr/bin/env bash

set -euo pipefail

# Fetch open PRs with checks and mergeability
prs_json=$(gh pr list \
  --state open \
  --json number,title,author,isDraft,mergeable,statusCheckRollup)

# Allowed bot authors across different GitHub API representations
# (e.g. "app/dependabot" in GraphQL vs "dependabot[bot]" in REST/webhooks)
allowed_bots='["app/dependabot", "dependabot[bot]", "app/renovate", "renovate[bot]", "app/gkeeper", "gkeeper[bot]"]'

# Find all PRs that match criteria:
# 1. Author is in allowed bots list
# 2. Not draft, and mergeable
# 3. Has at least 1 check run
# 4. All check runs are COMPLETED
# 5. All check runs concluded with SUCCESS (strict check: no skipped/neutral allowed)
target_prs=$(echo "$prs_json" | jq -r --argjson bots "$allowed_bots" '
  [
    .[] |
    select(
      (.author.login as $a | $bots | index($a)) and
      (.isDraft | not) and
      (.mergeable == "MERGEABLE") and
      ((.statusCheckRollup | length) > 0) and
      (.statusCheckRollup | all(.status == "COMPLETED")) and
      (.statusCheckRollup | all(.conclusion == "SUCCESS"))
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
