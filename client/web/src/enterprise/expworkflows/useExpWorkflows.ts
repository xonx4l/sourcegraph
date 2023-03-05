import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'

export const useExpWorkflows = (): boolean =>
    useExperimentalFeatures(features => (features.workflows === undefined ? true : features.workflows))
