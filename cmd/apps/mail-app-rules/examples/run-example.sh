#!/bin/bash

# Example script to run the IMAP DSL processor

# Check if the required environment variables are set
if [ -z "$IMAP_SERVER" ] || [ -z "$IMAP_USERNAME" ]; then
  echo "Please set the following environment variables:"
  echo "  IMAP_SERVER   - Your IMAP server address"
  echo "  IMAP_USERNAME - Your IMAP username"
  echo "  IMAP_PASSWORD - Your IMAP password (or you'll be prompted)"
  exit 1
fi

# If password is not set, prompt for it
if [ -z "$IMAP_PASSWORD" ]; then
  echo -n "Enter your IMAP password: "
  read -s IMAP_PASSWORD
  echo
  export IMAP_PASSWORD
fi

# Build the application if needed
if [ ! -f "../../../imap-dsl" ]; then
  echo "Building the application..."
  (cd ../../.. && go build -o imap-dsl ./cmd/apps/mail-app-rules)
fi

# Run the application with the recent-emails.yaml example
echo "Running the IMAP DSL processor with recent-emails.yaml..."
../../../imap-dsl \
  -rule examples/recent-emails.yaml \
  -server "$IMAP_SERVER" \
  -username "$IMAP_USERNAME" \
  -mailbox "INBOX"

echo
echo "To try other examples, run:"
echo "../../../imap-dsl -rule examples/from-specific-sender.yaml -server \$IMAP_SERVER -username \$IMAP_USERNAME"
echo "../../../imap-dsl -rule examples/important-emails.yaml -server \$IMAP_SERVER -username \$IMAP_USERNAME"
echo "../../../imap-dsl -rule examples/date-range-search.yaml -server \$IMAP_SERVER -username \$IMAP_USERNAME"
echo "../../../imap-dsl -rule examples/full-message-content.yaml -server \$IMAP_SERVER -username \$IMAP_USERNAME" 