import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'jsdom',
        setupFiles: [require.resolve('./src/testSetup.test.ts')],
        singleThread: true, // got `failed to terminate worker` occasionally in Bazel CI
    },
})
