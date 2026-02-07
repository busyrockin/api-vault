import { mkdir, writeFile } from 'node:fs/promises';
import { join, basename, extname } from 'node:path';
import { execFileSync } from 'node:child_process';

const GITHUB_API = 'https://api.github.com';
const GITHUB_RAW = 'https://raw.githubusercontent.com';
const SKILLS_DIR = '.claude/commands';

// Directories to search for skill files in the source repo, in priority order
const SEARCH_DIRS = ['skills', 'commands', '.claude/commands', ''];

function httpGet(url, { json = false } = {}) {
  const args = ['-sL', '--retry', '2'];
  const token = process.env.GITHUB_TOKEN;
  if (token) {
    args.push('-H', `Authorization: Bearer ${token}`);
  }
  if (json) {
    args.push('-H', 'Accept: application/vnd.github.v3+json');
  }
  args.push('-H', 'User-Agent: api-vault-skills-cli');
  // Write HTTP status code to stderr, body to stdout
  args.push('-w', '\n%{http_code}', url);

  let output;
  try {
    output = execFileSync('curl', args, {
      encoding: 'utf-8',
      maxBuffer: 10 * 1024 * 1024,
      timeout: 30000,
    });
  } catch {
    throw new Error(`Network request failed for ${url}`);
  }

  // Split body and status code
  const lines = output.trimEnd().split('\n');
  const statusCode = parseInt(lines.pop(), 10);
  const body = lines.join('\n');

  if (statusCode === 404) {
    return null;
  }
  if (statusCode === 403) {
    throw new Error(
      'GitHub API rate limit exceeded. Set GITHUB_TOKEN env var to increase your limit.'
    );
  }
  if (statusCode < 200 || statusCode >= 300) {
    throw new Error(`GitHub API returned HTTP ${statusCode}`);
  }

  return json ? JSON.parse(body) : body;
}

function parseRepo(input) {
  // Support owner/repo or full GitHub URLs
  const urlMatch = input.match(/github\.com\/([^/]+)\/([^/]+?)(?:\.git)?$/);
  if (urlMatch) {
    return { owner: urlMatch[1], repo: urlMatch[2] };
  }

  const parts = input.split('/');
  if (parts.length !== 2 || !parts[0] || !parts[1]) {
    throw new Error(
      `Invalid repository format: "${input}". Expected owner/repo (e.g., acme/claude-skills)`
    );
  }
  return { owner: parts[0], repo: parts[1] };
}

function isSkillFile(filename) {
  return extname(filename).toLowerCase() === '.md';
}

export async function add(repoArg, flags = {}) {
  const { owner, repo } = parseRepo(repoArg);
  const branch = flags.branch || null;
  const sourceDir = flags.dir || null;

  // 1. Get repo info (default branch)
  console.log(`Fetching repository info for ${owner}/${repo}...`);
  const repoInfo = httpGet(`${GITHUB_API}/repos/${owner}/${repo}`, { json: true });
  if (!repoInfo) {
    throw new Error(`Repository ${owner}/${repo} not found.`);
  }
  const targetBranch = branch || repoInfo.default_branch;

  // 2. Get the full file tree
  console.log(`Scanning ${targetBranch} branch for skill files...`);
  const tree = httpGet(
    `${GITHUB_API}/repos/${owner}/${repo}/git/trees/${targetBranch}?recursive=1`,
    { json: true }
  );
  if (!tree) {
    throw new Error(`Could not read repository tree for branch ${targetBranch}.`);
  }

  // 3. Find skill files (.md files in the appropriate directories)
  const allFiles = tree.tree
    .filter((item) => item.type === 'blob')
    .map((item) => item.path);

  let skillFiles = [];

  if (sourceDir) {
    // User specified a directory explicitly
    const prefix = sourceDir.replace(/\/$/, '') + '/';
    skillFiles = allFiles.filter(
      (f) => f.startsWith(prefix) && isSkillFile(f) && !f.includes('/', prefix.length)
    );
    if (skillFiles.length === 0) {
      throw new Error(`No .md skill files found in "${sourceDir}" directory.`);
    }
  } else {
    // Auto-detect: search directories in priority order
    for (const dir of SEARCH_DIRS) {
      const prefix = dir ? dir + '/' : '';
      const found = allFiles.filter((f) => {
        if (!isSkillFile(f)) return false;
        if (dir === '') {
          return !f.includes('/');
        }
        return f.startsWith(prefix) && !f.includes('/', prefix.length);
      });
      if (found.length > 0) {
        skillFiles = found;
        if (dir) {
          console.log(`Found skills in ${dir}/ directory.`);
        } else {
          console.log(`Found skills in repository root.`);
        }
        break;
      }
    }
  }

  // Filter out common non-skill markdown files (all-uppercase names like README, CHANGELOG)
  skillFiles = skillFiles.filter((f) => {
    const name = basename(f, '.md');
    if (!f.includes('/') && name === name.toUpperCase() && name.length > 1) {
      return false;
    }
    return true;
  });

  if (skillFiles.length === 0) {
    throw new Error(
      `No skill files found in ${owner}/${repo}. ` +
      `Expected .md files in one of: ${SEARCH_DIRS.filter(Boolean).join(', ')} directories, or in the repo root.`
    );
  }

  console.log(`Found ${skillFiles.length} skill(s) to install.\n`);

  // 4. Create the local skills directory
  const installDir = join(process.cwd(), SKILLS_DIR);
  await mkdir(installDir, { recursive: true });

  // 5. Download and install each skill file
  const installed = [];
  const failed = [];

  for (const filePath of skillFiles) {
    const filename = basename(filePath);
    const destPath = join(installDir, filename);
    const skillName = basename(filename, '.md');

    try {
      const url = `${GITHUB_RAW}/${owner}/${repo}/${targetBranch}/${filePath}`;
      const content = httpGet(url);
      if (content === null) {
        throw new Error('File not found');
      }
      await writeFile(destPath, content, 'utf-8');
      installed.push(skillName);
      console.log(`  + ${skillName}`);
    } catch (err) {
      failed.push({ name: skillName, error: err.message });
      console.error(`  x ${skillName} (${err.message})`);
    }
  }

  // 6. Summary
  console.log('');
  if (installed.length > 0) {
    console.log(`Installed ${installed.length} skill(s) to ${SKILLS_DIR}/`);
    console.log(`Source: ${owner}/${repo} (${targetBranch})`);
  }
  if (failed.length > 0) {
    console.log(`Failed to install ${failed.length} skill(s).`);
  }
  if (installed.length > 0) {
    console.log(`\nUse these skills as slash commands in Claude Code:`);
    for (const name of installed) {
      console.log(`  /${name}`);
    }
  }
}
