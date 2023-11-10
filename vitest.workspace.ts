import { readFileSync, existsSync } from 'fs'
import path from 'path'

import glob from 'glob'
import { load } from 'js-yaml'
import { defineWorkspace } from 'vitest/config'

interface PnpmWorkspaceFile {
    packages: string[]
}
const workspaceFile = load(readFileSync(path.join(__dirname, 'pnpm-workspace.yaml'), 'utf8')) as PnpmWorkspaceFile
const projectRoots = workspaceFile.packages
    .flatMap(p => glob.sync(`${p}/`, { cwd: __dirname }))
    .map(p => p.replace(/\/$/, ''))
    .filter(dir => existsSync(path.join(dir, 'vitest.config.ts')))

export default defineWorkspace(projectRoots)
