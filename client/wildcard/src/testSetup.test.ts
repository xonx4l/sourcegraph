// nsc: copied from /home/noah/Sourcegraph/sourcegraph/client/branded/src/testSetup.test.ts
import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers'
import * as matchers from '@testing-library/jest-dom/matchers'
import { expect } from 'vitest'

declare module 'vitest' {
    interface Assertion<T = any> extends jest.Matchers<void, T>, TestingLibraryMatchers<T, void> {}
}
expect.extend(matchers)
