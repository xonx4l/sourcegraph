import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environmentMatchGlobs: [
            ['**/*.tsx', 'jsdom'],
            ['src/util/(useInputValidation|dom).test.ts', 'jsdom'],
        ],
        setupFiles: [require.resolve('./src/testSetup.test.ts'), require.resolve('../testing/src/reactCleanup.ts')],
    },
})
