#!/usr/bin/env node
import React from 'react';
import {render, Text, Box} from 'ink';
import meow from 'meow';

// Handle CLI with meow
const cli = meow(
  `
  Usage
    $ basic-example

  Options
    --help    Show this help message
    --version Show version
`,
  {
    importMeta: import.meta,
    flags: {
      help: {
        type: 'boolean',
        alias: 'h',
      },
      version: {
        type: 'boolean',
        alias: 'v',
      },
    },
  }
);

const BasicExample = () => (
  <Box flexDirection="column" padding={1}>
    <Box marginBottom={1}>
      <Text bold color="#9D8CFF">Hello from Ink!</Text>
    </Box>
    <Box>
      <Text>This is a basic example of using Ink with React.</Text>
    </Box>
    <Box marginTop={1}>
      <Text color="#FF6B6B">Press Ctrl+C to exit</Text>
    </Box>
  </Box>
);

render(<BasicExample />); 