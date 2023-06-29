import { ApolloClient } from '@apollo/client'
import { Observable, from } from 'rxjs'
import { map } from 'rxjs/operators'

import { ErrorLike } from '@sourcegraph/common'
import { getDocumentNode } from '@sourcegraph/http-client'

import {
    GitCommitAncestorFields,
    RepositoryGitCommitResult,
    RepositoryGitCommitVariables,
    Scalars,
} from '../../../../graphql-operations'
import { REPOSITORY_GIT_COMMIT } from '../../../../repo/RevisionsPopover/RevisionsPopoverCommits'

export const queryGitCommits = (
    repo: Scalars['ID'],
    client: ApolloClient<object>
): Observable<GitCommitAncestorFields[] | ErrorLike | null | undefined> =>
    from(
        client.query<RepositoryGitCommitResult, RepositoryGitCommitVariables>({
            query: getDocumentNode(REPOSITORY_GIT_COMMIT),
            variables: { repo, revision: 'main', first: null, query: null },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ node }) => {
            if (!node) {
                throw new Error(`Repository ${repo} not found`)
            }

            if (node.__typename !== 'Repository') {
                throw new Error(`Node is a ${node.__typename}, not a Repository`)
            }

            return node.commit?.ancestors.nodes
        })
    )
