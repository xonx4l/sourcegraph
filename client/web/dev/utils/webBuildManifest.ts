type Entry = 'src/enterprise/main.tsx' | 'src/enterprise/embed/embedMain.tsx' | 'src/enterprise/app/main.tsx'

export interface WebBuildManifest {
    /** Base URL for asset paths. */
    url?: string

    /**
     * A map of entrypoint (such as "src/enterprise/main.tsx") to its JavaScript and CSS assets.
     */
    assets: Partial<Record<Entry, { js: string; css?: string }>>

    /** Additional HTML <script> tags to inject in dev mode. */
    devInjectHTML?: string
}
