#!/bin/bash
export GITHUB_TOKEN=$(vault kv get -mount=personal -format json goreleaser | jq -r .data.data.github_token)
export MASTODON_CLIENT_ID=$(vault kv get -mount=personal -format json goreleaser | jq -r .data.data.mastodon_client_id)
export MASTODON_CLIENT_SECRET=$(vault kv get -mount=personal -format json goreleaser | jq -r .data.data.mastodon_client_secret)
export MASTODON_ACCESS_TOKEN=$(vault kv get -mount=personal -format json goreleaser | jq -r .data.data.mastodon_access_token)

goreleaser release --clean