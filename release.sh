#!/bin/bash
echo "Getting secrets from Vault..."
secrets=$(vault kv get -mount=personal -format json goreleaser)
echo -ne "\tGithub token..."
GITHUB_TOKEN=$(echo $secrets | jq -r .data.data.github_token)
echo -ne " done\n"

echo -ne "\tMastodon Client ID..."
MASTODON_CLIENT_ID=$(echo $secrets | jq -r .data.data.mastodon_client_id)
echo -ne " done\n"

echo -ne "\tMastodon Client Secret..."
MASTODON_CLIENT_SECRET=$(echo $secrets | jq -r .data.data.mastodon_client_secret)
echo -ne " done\n"

echo -ne "\tMastodon Access Token..."
MASTODON_ACCESS_TOKEN=$(echo $secrets | jq -r .data.data.mastodon_access_token)
echo -ne " done\n"

branch=$(git rev-parse --abbrev-ref HEAD)

if [ "$branch" = "main" ]; then
    goreleaser release --clean
else 
    goreleaser release --skip=announce,publish --clean
fi
