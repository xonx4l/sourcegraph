import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'jsdom',
        setupFiles: [
            require.resolve('./src/testSetup.test.ts'),
            require.resolve('../testing/src/reactCleanup.ts'),
            require.resolve('../testing/src/mockMatchMedia.ts'),
            require.resolve('../testing/src/mockUniqueId.ts'),
        ],
    },
})
