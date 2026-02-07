#!/usr/bin/env node

import { add } from '../lib/add.js';

const USAGE = `Usage: skills <command> [options]

Commands:
  add <owner/repo>    Add skills from a GitHub repository
  help                Show this help message

Examples:
  npx skills add acme/claude-skills
  npx skills add acme/claude-skills --dir commands
  npx skills add acme/claude-skills --branch main`;

function parseArgs(argv) {
  const args = argv.slice(2);
  const command = args[0];
  const positional = [];
  const flags = {};

  for (let i = 1; i < args.length; i++) {
    if (args[i].startsWith('--')) {
      const key = args[i].slice(2);
      const value = args[i + 1] && !args[i + 1].startsWith('--') ? args[++i] : true;
      flags[key] = value;
    } else {
      positional.push(args[i]);
    }
  }

  return { command, positional, flags };
}

async function main() {
  const { command, positional, flags } = parseArgs(process.argv);

  switch (command) {
    case 'add': {
      const repo = positional[0];
      if (!repo) {
        console.error('Error: Missing repository argument.\n');
        console.error('Usage: skills add <owner/repo>');
        process.exit(1);
      }
      await add(repo, flags);
      break;
    }
    case 'help':
    case '--help':
    case '-h':
    case undefined:
      console.log(USAGE);
      break;
    default:
      console.error(`Unknown command: ${command}\n`);
      console.log(USAGE);
      process.exit(1);
  }
}

main().catch((err) => {
  console.error(`Error: ${err.message}`);
  process.exit(1);
});
