import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'happy-dom',
        environmentMatchGlobs: [
            ['src/enterprise/code-monitoring/ManageCodeMonitorPage.test.tsx', 'jsdom'], // needs window.confirm, Request
            ['src/enterprise/code-monitoring/CreateCodeMonitorPage.test.tsx', 'jsdom'], // 'Error: Should not already be working.'
            ['src/hooks/useScrollManager/useScrollManager.test.tsx', 'jsdom'], // for correct scroll counting
            ['src/components/KeyboardShortcutsHelp/KeyboardShortcutsHelp.test.tsx', 'jsdom'], // event.getModifierState
        ],

        setupFiles: [
            require.resolve('./src/testSetup.test.ts'),
            require.resolve('../testing/src/reactCleanup.ts'),
            require.resolve('../testing/src/mockMatchMedia.ts'),
            require.resolve('../testing/src/mockUniqueId.ts'),
            require.resolve('../testing/src/mockDate.ts'),
            require.resolve('../testing/src/fetch.js'),
        ],
    },
})
